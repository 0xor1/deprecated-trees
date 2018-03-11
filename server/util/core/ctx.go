package core

import (
	"bitbucket.org/0xor1/task/server/util/avatar"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/mail"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/time"
	"database/sql"
	"encoding/json"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"
	"net/http"
	"regexp"
	"sync"
)

var (
	unauthorizedErr = &err.Err{Code: "", Message: ""}
)

// per request info fields
type Ctx struct {
	me                     *id.Id
	session                *sessions.Session
	requestStartUnixMillis int64
	resp                   http.ResponseWriter
	req                    *http.Request
	queryInfosMtx          *sync.RWMutex
	queryInfos             []*QueryInfo
	retrievedDlms          map[string]int64
	dlmsToUpdate           map[string]interface{}
	cacheItemsToUpdate     map[string]interface{}
	cacheKeysToDelete      map[string]interface{}
	staticResources        *StaticResources
}

func (c *Ctx) TryMe() *id.Id {
	return c.me
}

func (c *Ctx) Me() id.Id {
	if c.me == nil {
		panic(unauthorizedErr)
	}
	return *c.me
}

func (c *Ctx) Log(err error) {
	c.staticResources.LogError(err)
}

func (c *Ctx) AccountExec(query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.staticResources.AccountDb, query, args...)
}

func (c *Ctx) AccountQuery(query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.staticResources.AccountDb, query, args...)
}

func (c *Ctx) AccountQueryRow(query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.staticResources.AccountDb, query, args...)
}

func (c *Ctx) PwdExec(query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.staticResources.PwdDb, query, args...)
}

func (c *Ctx) PwdQuery(query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.staticResources.PwdDb, query, args...)
}

func (c *Ctx) PwdQueryRow(query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.staticResources.PwdDb, query, args...)
}

func (c *Ctx) TreeShardCount() int {
	return len(c.staticResources.TreeShards)
}

func (c *Ctx) TreeExec(shard int, query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.staticResources.TreeShards[shard], query, args...)
}

func (c *Ctx) TreeQuery(shard int, query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.staticResources.TreeShards[shard], query, args...)
}

func (c *Ctx) TreeQueryRow(shard int, query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.staticResources.TreeShards[shard], query, args...)
}

func (c *Ctx) GetCacheValue(val interface{}, key string, dlmKeys []string, args interface{}) bool {
	if key == "" {
		return false
	}
	dlm, e := getDlm(c, dlmKeys)
	if e != nil {
		c.Log(e)
		return false
	}
	if dlm > c.requestStartUnixMillis {
		return false
	}
	jsonBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.staticResources.MasterCacheKey, Key: key, Args: args})
	if e != nil {
		c.Log(e)
		return false
	}
	cnn := c.staticResources.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	start := time.NowUnixMillis()
	jsonBytes, e = redis.Bytes(cnn.Do(cnst.GET, jsonBytes))
	writeQueryInfo(c, &QueryInfo{Query: cnst.GET, Args: jsonBytes, Duration: time.NowUnixMillis() - start})
	if e != nil {
		c.Log(e)
		return false
	}
	if len(jsonBytes) == 0 {
		return false
	}
	e = json.Unmarshal(jsonBytes, val)
	if e != nil {
		c.Log(e)
		return false
	}
	return true
}

func (c *Ctx) SetCacheValue(val interface{}, key string, dlmKeys []string, args interface{}) {
	if val == nil || key == "" {
		panic(err.InvalidArguments)
	}
	valBytes, e := json.Marshal(val)
	if e != nil {
		c.Log(e)
		return
	}
	cacheKeyBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.staticResources.MasterCacheKey, Key: key, Args: args})
	if e != nil {
		c.Log(e)
		return
	}
	for _, dlmKey := range dlmKeys {
		c.dlmsToUpdate[dlmKey] = nil
	}
	c.cacheItemsToUpdate[string(cacheKeyBytes)] = valBytes
}

func (c *Ctx) DeleteDlmKeys(keys []string) {
	for _, key := range keys {
		c.cacheKeysToDelete[key] = nil
	}
}

func (c *Ctx) NameRegexMatchers() []*regexp.Regexp {
	return c.staticResources.NameRegexMatchers
}

func (c *Ctx) PwdRegexMatchers() []*regexp.Regexp {
	return c.staticResources.PwdRegexMatchers
}

func (c *Ctx) NameMinRuneCount() int {
	return c.staticResources.NameMinRuneCount
}

func (c *Ctx) NameMaxRuneCount() int {
	return c.staticResources.NameMaxRuneCount
}

func (c *Ctx) PwdMinRuneCount() int {
	return c.staticResources.PwdMinRuneCount
}

func (c *Ctx) PwdMaxRuneCount() int {
	return c.staticResources.PwdMaxRuneCount
}

func (c *Ctx) MaxProcessEntityCount() int {
	return c.staticResources.MaxProcessEntityCount
}

func (c *Ctx) CryptCodeLen() int {
	return c.staticResources.CryptCodeLen
}

func (c *Ctx) SaltLen() int {
	return c.staticResources.SaltLen
}

func (c *Ctx) ScryptN() int {
	return c.staticResources.ScryptN
}

func (c *Ctx) ScryptR() int {
	return c.staticResources.ScryptR
}

func (c *Ctx) ScryptP() int {
	return c.staticResources.ScryptP
}

func (c *Ctx) ScryptKeyLen() int {
	return c.staticResources.ScryptKeyLen
}

func (c *Ctx) RegionalV1PrivateClient() private.V1Client {
	return c.staticResources.RegionalV1PrivateClient
}

func (c *Ctx) MailClient() mail.Client {
	return c.staticResources.MailClient
}

func (c *Ctx) AvatarClient() avatar.Client {
	return c.staticResources.AvatarClient
}

// helpers

func sqlExec(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) (sql.Result, error) {
	start := time.NowUnixMillis()
	res, e := rs.Exec(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: time.NowUnixMillis() - start})
	return res, e
}

func sqlQuery(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) (isql.Rows, error) {
	start := time.NowUnixMillis()
	rows, e := rs.Query(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: time.NowUnixMillis() - start})
	return rows, e
}

func sqlQueryRow(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) isql.Row {
	start := time.NowUnixMillis()
	row := rs.QueryRow(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: time.NowUnixMillis() - start})
	return row
}

func writeQueryInfo(ctx *Ctx, qi *QueryInfo) {
	ctx.queryInfosMtx.Lock()
	defer ctx.queryInfosMtx.Unlock()
	ctx.queryInfos = append(ctx.queryInfos, qi)
}

func getQueryInfos(ctx *Ctx) []*QueryInfo {
	ctx.queryInfosMtx.RLock()
	defer ctx.queryInfosMtx.RUnlock()
	cpy := make([]*QueryInfo, 0, len(ctx.queryInfos))
	for _, qi := range ctx.queryInfos {
		cpy = append(cpy, qi)
	}
	return cpy
}

func getDlm(ctx *Ctx, dlmKeys []string) (int64, error) {
	panicIfRetrievedDlmsAreMissingEntries := false
	if len(ctx.retrievedDlms) > 0 {
		panicIfRetrievedDlmsAreMissingEntries = true
	}
	dlmsToFetch := make([]interface{}, 0, len(dlmKeys))
	latestDlm := int64(0)
	for _, dlmKey := range dlmKeys {
		dlm, exists := ctx.retrievedDlms[dlmKey]
		if !exists {
			if panicIfRetrievedDlmsAreMissingEntries {
				err.FmtPanic("missing dlm key %q on path %s", dlmKey, ctx.req.URL.Path)
			}
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && dlm > latestDlm {
			latestDlm = dlm
		}
	}
	if len(dlmsToFetch) > 0 {
		cnn := ctx.staticResources.DlmAndDataRedisPool.Get()
		defer cnn.Close()
		start := time.NowUnixMillis()
		dlms, e := redis.Int64s(cnn.Do("MGET", dlmsToFetch...))
		writeQueryInfo(ctx, &QueryInfo{Query: "MGET", Args: dlmsToFetch, Duration: time.NowUnixMillis() - start})
		if e != nil {
			return 0, e
		}
		for _, dlm := range dlms {
			if dlm > latestDlm {
				latestDlm = dlm
			}
		}
	}
	return latestDlm, nil
}

func setDlmsForUpdate(ctx *Ctx, dlmKeys []string) {
	for _, key := range dlmKeys {
		ctx.dlmsToUpdate[key] = nil
	}
}

func doCacheUpdate(ctx *Ctx) {
	if len(ctx.dlmsToUpdate) == 0 && len(ctx.cacheItemsToUpdate) == 0 && len(ctx.cacheKeysToDelete) == 0 {
		return
	}
	setArgs := make([]interface{}, 0, (len(ctx.dlmsToUpdate)*2)+(len(ctx.cacheItemsToUpdate)*2))
	for k := range ctx.dlmsToUpdate {
		setArgs = append(setArgs, k, ctx.requestStartUnixMillis)
	}
	for k, v := range ctx.cacheItemsToUpdate {
		setArgs = append(setArgs, k, v)
	}
	delArgs := make([]interface{}, 0, len(ctx.cacheKeysToDelete))
	for k := range ctx.cacheKeysToDelete {
		delArgs = append(delArgs, k)
	}
	cnn := ctx.staticResources.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	if len(setArgs) > 0 {
		start := time.NowUnixMillis()
		_, e := cnn.Do("MSET", setArgs...)
		writeQueryInfo(ctx, &QueryInfo{Query: "MSET", Args: setArgs, Duration: time.NowUnixMillis() - start})
		if e != nil {
			ctx.Log(e)
		}
	}
	if len(delArgs) > 0 {
		start := time.NowUnixMillis()
		_, e := cnn.Do("DEL", setArgs...)
		writeQueryInfo(ctx, &QueryInfo{Query: "DEL", Args: setArgs, Duration: time.NowUnixMillis() - start})
		if e != nil {
			ctx.Log(e)
		}
	}

}

type valueCacheKey struct {
	MasterKey string      `json:"masterKey"`
	Key       string      `json:"key"`
	Args      interface{} `json:"args"`
}

type QueryInfo struct {
	Query    string      `json:"query"`
	Args     interface{} `json:"args"`
	Duration int64       `json:"duration"`
}
