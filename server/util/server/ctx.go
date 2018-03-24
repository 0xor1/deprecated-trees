package server

import (
	"bitbucket.org/0xor1/task/server/util/avatar"
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/mail"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/queryinfo"
	"bitbucket.org/0xor1/task/server/util/static"
	"bitbucket.org/0xor1/task/server/util/time"
	"database/sql"
	"encoding/json"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var (
	unauthorizedErr = &err.Err{Code: "u_s_u", Message: "unauthorized"}
)

// per request info fields
type _ctx struct {
	me                     *id.Id
	session                *sessions.Session
	requestStartUnixMillis int64
	resp                   http.ResponseWriter
	req                    *http.Request
	queryInfosMtx          *sync.RWMutex
	queryInfos             []*queryinfo.QueryInfo
	retrievedDlms          map[string]int64
	dlmsToUpdate           map[string]interface{}
	cacheItemsToUpdate     map[string]interface{}
	SR                     *static.Resources
}

func (c *_ctx) TryMe() *id.Id {
	return c.me
}

func (c *_ctx) Me() id.Id {
	if c.me == nil {
		panic(unauthorizedErr)
	}
	return *c.me
}

func (c *_ctx) Log(err error) {
	c.SR.LogError(err)
}

func (c *_ctx) AccountExec(query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.SR.AccountDb, query, args...)
}

func (c *_ctx) AccountQuery(query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.SR.AccountDb, query, args...)
}

func (c *_ctx) AccountQueryRow(query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.SR.AccountDb, query, args...)
}

func (c *_ctx) PwdExec(query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.SR.PwdDb, query, args...)
}

func (c *_ctx) PwdQuery(query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.SR.PwdDb, query, args...)
}

func (c *_ctx) PwdQueryRow(query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.SR.PwdDb, query, args...)
}

func (c *_ctx) TreeShardCount() int {
	return len(c.SR.TreeShards)
}

func (c *_ctx) TreeExec(shard int, query string, args ...interface{}) (sql.Result, error) {
	return sqlExec(c, c.SR.TreeShards[shard], query, args...)
}

func (c *_ctx) TreeQuery(shard int, query string, args ...interface{}) (isql.Rows, error) {
	return sqlQuery(c, c.SR.TreeShards[shard], query, args...)
}

func (c *_ctx) TreeQueryRow(shard int, query string, args ...interface{}) isql.Row {
	return sqlQueryRow(c, c.SR.TreeShards[shard], query, args...)
}

func (c *_ctx) GetCacheValue(val interface{}, key *cachekey.Key, args ...interface{}) bool {
	if !c.SR.CachingEnabled || key.KeyVal == "" || queryBoolVal(c, "skipCache") {
		return false
	}
	dlm, e := getDlm(c, key.DlmKeys)
	if e != nil {
		c.Log(e)
		return false
	}
	jsonBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.SR.MasterCacheKey, Key: key.KeyVal, DlmKeys: key.DlmKeys, Dlm: dlm, Args: args})
	if e != nil {
		c.Log(e)
		return false
	}
	cnn := c.SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	start := time.NowUnixMillis()
	jsonBytes, e = redis.Bytes(cnn.Do("GET", string(jsonBytes)))
	writeQueryInfo(c, "GET", string(jsonBytes), start)
	if e == redis.ErrNil {
		return false
	}
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

func (c *_ctx) SetCacheValue(val interface{}, key *cachekey.Key, args ...interface{}) {
	if !c.SR.CachingEnabled {
		return
	}
	dlm, e := getDlm(c, key.DlmKeys)
	if e != nil {
		c.Log(e)
		return
	}
	valBytes, e := json.Marshal(val)
	if e != nil {
		c.Log(e)
		return
	}
	cacheKeyBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.SR.MasterCacheKey, Key: key.KeyVal, DlmKeys: key.DlmKeys, Dlm: dlm, Args: args})
	if e != nil {
		c.Log(e)
		return
	}
	c.cacheItemsToUpdate[string(cacheKeyBytes)] = valBytes
}

func (c *_ctx) TouchDlms(cacheKeys *cachekey.Key) {
	if !c.SR.CachingEnabled {
		return
	}
	for _, key := range cacheKeys.DlmKeys {
		c.dlmsToUpdate[key] = nil
	}
}

func (c *_ctx) NameRegexMatchers() []*regexp.Regexp {
	return c.SR.NameRegexMatchers
}

func (c *_ctx) PwdRegexMatchers() []*regexp.Regexp {
	return c.SR.PwdRegexMatchers
}

func (c *_ctx) NameMinRuneCount() int {
	return c.SR.NameMinRuneCount
}

func (c *_ctx) NameMaxRuneCount() int {
	return c.SR.NameMaxRuneCount
}

func (c *_ctx) PwdMinRuneCount() int {
	return c.SR.PwdMinRuneCount
}

func (c *_ctx) PwdMaxRuneCount() int {
	return c.SR.PwdMaxRuneCount
}

func (c *_ctx) MaxProcessEntityCount() int {
	return c.SR.MaxProcessEntityCount
}

func (c *_ctx) CryptCodeLen() int {
	return c.SR.CryptCodeLen
}

func (c *_ctx) SaltLen() int {
	return c.SR.SaltLen
}

func (c *_ctx) ScryptN() int {
	return c.SR.ScryptN
}

func (c *_ctx) ScryptR() int {
	return c.SR.ScryptR
}

func (c *_ctx) ScryptP() int {
	return c.SR.ScryptP
}

func (c *_ctx) ScryptKeyLen() int {
	return c.SR.ScryptKeyLen
}

func (c *_ctx) RegionalV1PrivateClient() private.V1Client {
	return c.SR.RegionalV1PrivateClient
}

func (c *_ctx) MailClient() mail.Client {
	return c.SR.MailClient
}

func (c *_ctx) AvatarClient() avatar.Client {
	return c.SR.AvatarClient
}

// helpers

func sqlExec(ctx *_ctx, rs isql.ReplicaSet, query string, args ...interface{}) (sql.Result, error) {
	start := time.NowUnixMillis()
	res, e := rs.Exec(query, args...)
	writeQueryInfo(ctx, query, args, start)
	return res, e
}

func sqlQuery(ctx *_ctx, rs isql.ReplicaSet, query string, args ...interface{}) (isql.Rows, error) {
	start := time.NowUnixMillis()
	rows, e := rs.Query(query, args...)
	writeQueryInfo(ctx, query, args, start)
	return rows, e
}

func sqlQueryRow(ctx *_ctx, rs isql.ReplicaSet, query string, args ...interface{}) isql.Row {
	start := time.NowUnixMillis()
	row := rs.QueryRow(query, args...)
	writeQueryInfo(ctx, query, args, start)
	return row
}

func writeQueryInfo(ctx *_ctx, query string, args interface{}, startUnixMillis int64) {
	ctx.queryInfosMtx.Lock()
	defer ctx.queryInfosMtx.Unlock()
	ctx.queryInfos = append(ctx.queryInfos, queryinfo.New(query, args, startUnixMillis))
}

func getQueryInfos(ctx *_ctx) []*queryinfo.QueryInfo {
	ctx.queryInfosMtx.RLock()
	defer ctx.queryInfosMtx.RUnlock()
	cpy := make([]*queryinfo.QueryInfo, 0, len(ctx.queryInfos))
	for _, qi := range ctx.queryInfos {
		cpy = append(cpy, qi)
	}
	return cpy
}

func getDlm(ctx *_ctx, dlmKeys []string) (int64, error) {
	dlmsToFetch := make([]interface{}, 0, len(dlmKeys))
	latestDlm := int64(0)
	for _, dlmKey := range dlmKeys {
		dlm, exists := ctx.retrievedDlms[dlmKey]
		if !exists {
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && dlm > latestDlm {
			latestDlm = dlm
		}
	}
	if len(dlmsToFetch) > 0 {
		cnn := ctx.SR.DlmAndDataRedisPool.Get()
		defer cnn.Close()
		start := time.NowUnixMillis()
		dlms, e := redis.Int64s(cnn.Do("MGET", dlmsToFetch...))
		writeQueryInfo(ctx, "MGET", dlmsToFetch, start)
		if e != nil {
			return 0, e
		}
		for i, dlm := range dlms {
			ctx.retrievedDlms[dlmsToFetch[i].(string)] = dlm
			if dlm > latestDlm {
				latestDlm = dlm
			}
		}
	}
	return latestDlm, nil
}

func doCacheUpdate(ctx *_ctx) {
	if !ctx.SR.CachingEnabled || (len(ctx.dlmsToUpdate) == 0 && len(ctx.cacheItemsToUpdate) == 0) {
		return
	}
	setArgs := make([]interface{}, 0, (len(ctx.dlmsToUpdate)*2)+(len(ctx.cacheItemsToUpdate)*2))
	for k := range ctx.dlmsToUpdate {
		setArgs = append(setArgs, k, ctx.requestStartUnixMillis)
	}
	for k, v := range ctx.cacheItemsToUpdate {
		setArgs = append(setArgs, k, v)
	}
	cnn := ctx.SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	if len(setArgs) > 0 {
		start := time.NowUnixMillis()
		_, e := cnn.Do("MSET", setArgs...)
		writeQueryInfo(ctx, "MSET", setArgs, start)
		if e != nil {
			ctx.Log(e)
		}
	}
}

type valueCacheKey struct {
	MasterKey string      `json:"masterKey"`
	Key       string      `json:"key"`
	DlmKeys   []string    `json:"dlmKeys"`
	Dlm       int64       `json:"dlm"`
	Args      interface{} `json:"args"`
}

func queryBoolVal(ctx *_ctx, name string) bool {
	switch strings.ToLower(ctx.req.URL.Query().Get(name)) {
	case "1", "y", "yes", "t", "true":
		return true
	default:
		return false
	}
}
