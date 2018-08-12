package server

import (
	"database/sql"
	"encoding/json"
	"github.com/0xor1/isql"
	"github.com/0xor1/trees/server/util/avatar"
	"github.com/0xor1/trees/server/util/cachekey"
	"github.com/0xor1/trees/server/util/err"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/mail"
	"github.com/0xor1/trees/server/util/private"
	"github.com/0xor1/trees/server/util/queryinfo"
	"github.com/0xor1/trees/server/util/static"
	"github.com/0xor1/trees/server/util/time"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
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
	fixedTreeReadSlaveMtx  *sync.RWMutex
	fixedTreeReadSlave     isql.DBCore
	cache                  *bool
	profile                *bool
	retrievedDlms          map[string]int64
	dlmsToUpdate           map[string]interface{}
	cacheItemsToUpdate     map[string]interface{}
	SR                     *static.Resources
}

func (c *_ctx) TryMe() *id.Id {
	return c.me
}

func (c *_ctx) Me() id.Id {
	c.ReturnNowIf(c.me == nil, http.StatusUnauthorized, "please login")
	return *c.me
}

func (c *_ctx) LogIf(err error) bool {
	if err != nil {
		c.SR.LogError(err)
		return true
	}
	return false
}

func (c *_ctx) ReturnNowIf(condition bool, httpStatus int, messageFmt string, messageArgs ...interface{}) {
	err.HttpPanicf(condition, httpStatus, messageFmt, messageArgs...)
}

func (c *_ctx) ReturnBadRequestNowIf(condition bool, messageFmt string, messageArgs ...interface{}) {
	err.HttpPanicf(condition, http.StatusBadRequest, messageFmt, messageArgs...)
}

func (c *_ctx) ReturnUnauthorizedNowIf(condition bool) {
	err.HttpPanicf(condition, http.StatusUnauthorized, "unauthorized")
}

func (c *_ctx) AccountExec(query string, args ...interface{}) (sql.Result, error) {
	return c.sqlExec(c.SR.AccountDb, query, args...)
}

func (c *_ctx) AccountQuery(query string, args ...interface{}) (isql.Rows, error) {
	return c.sqlQuery(c.SR.AccountDb, query, args...)
}

func (c *_ctx) AccountQueryRow(query string, args ...interface{}) isql.Row {
	return c.sqlQueryRow(c.SR.AccountDb, query, args...)
}

func (c *_ctx) PwdExec(query string, args ...interface{}) (sql.Result, error) {
	return c.sqlExec(c.SR.PwdDb, query, args...)
}

func (c *_ctx) PwdQuery(query string, args ...interface{}) (isql.Rows, error) {
	return c.sqlQuery(c.SR.PwdDb, query, args...)
}

func (c *_ctx) PwdQueryRow(query string, args ...interface{}) isql.Row {
	return c.sqlQueryRow(c.SR.PwdDb, query, args...)
}

func (c *_ctx) TreeShardCount() int {
	return len(c.SR.TreeShards)
}

func (c *_ctx) TreeExec(shard int, query string, args ...interface{}) (sql.Result, error) {
	return c.sqlExec(c.SR.TreeShards[shard], query, args...)
}

func (c *_ctx) TreeQuery(shard int, query string, args ...interface{}) (isql.Rows, error) {
	return c.sqlQuery(c.getFixedTreeReadSlave(shard), query, args...)
}

func (c *_ctx) TreeQueryRow(shard int, query string, args ...interface{}) isql.Row {
	return c.sqlQueryRow(c.getFixedTreeReadSlave(shard), query, args...)
}

func (c *_ctx) GetCacheValue(val interface{}, key *cachekey.Key) bool {
	if !c.SR.CachingEnabled || key.Key == "" || !c.useCache() {
		return false
	}
	dlm, e := c.getDlm(key.DlmKeys)
	if c.LogIf(e) {
		// if the error is a timeout error to redis assume it's unavailable for  the rest of the request and skip it
		if to, ok := e.(net.Error); ok && to.Timeout() {
			f := false
			c.cache = &f
		}
		return false
	}
	argsJsonBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.SR.MasterCacheKey, Key: key.Key, DlmKeys: key.SortedDlmKeys(), Dlm: dlm, Args: key.Args})
	if c.LogIf(e) {
		return false
	}
	cnn := c.SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	start := time.NowUnixMillis()
	jsonBytes, e := redis.Bytes(cnn.Do("GET", string(argsJsonBytes)))
	c.writeQueryInfo("GET", string(argsJsonBytes), start)
	if e == redis.ErrNil {
		return false
	}
	if c.LogIf(e) {
		return false
	}
	if len(jsonBytes) == 0 {
		return false
	}
	e = json.Unmarshal(jsonBytes, val)
	if c.LogIf(e) {
		return false
	}
	return true
}

func (c *_ctx) SetCacheValue(val interface{}, key *cachekey.Key) {
	if !c.SR.CachingEnabled || !c.useCache() {
		return
	}
	dlm, e := c.getDlm(key.DlmKeys)
	if c.LogIf(e) {
		// if the error is a timeout error to redis assume it's unavailable for  the rest of the request and skip it
		if to, ok := e.(net.Error); ok && to.Timeout() {
			f := false
			c.cache = &f
		}
		return
	}
	valBytes, e := json.Marshal(val)
	if c.LogIf(e) {
		return
	}
	cacheKeyBytes, e := json.Marshal(&valueCacheKey{MasterKey: c.SR.MasterCacheKey, Key: key.Key, DlmKeys: key.SortedDlmKeys(), Dlm: dlm, Args: key.Args})
	if c.LogIf(e) {
		return
	}
	c.cacheItemsToUpdate[string(cacheKeyBytes)] = valBytes
}

func (c *_ctx) TouchDlms(cacheKeys *cachekey.Key) {
	if !c.SR.CachingEnabled {
		return
	}
	for key := range cacheKeys.DlmKeys {
		c.dlmsToUpdate[key] = nil
	}
}

func (c *_ctx) ClientScheme() string {
	return c.SR.ClientScheme
}

func (c *_ctx) ClientHost() string {
	return c.SR.ClientHost
}

func (c *_ctx) NameRegexMatchers() []*regexp.Regexp {
	return c.SR.NameRegexMatchers
}

func (c *_ctx) DisplayNameRegexMatchers() []*regexp.Regexp {
	return c.SR.DisplayNameRegexMatchers
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

func (c *_ctx) DisplayNameMinRuneCount() int {
	return c.SR.DisplayNameMinRuneCount
}

func (c *_ctx) DisplayNameMaxRuneCount() int {
	return c.SR.DisplayNameMaxRuneCount
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

func (c *_ctx) useCache() bool {
	if c.cache == nil {
		val := c.queryBoolVal("cache", true)
		c.cache = &val
	}
	return *c.cache
}

func (c *_ctx) doProfile() bool {
	if c.profile == nil {
		val := c.queryBoolVal("profile", false)
		c.profile = &val
	}
	return *c.profile
}

func (c *_ctx) sqlExec(rs isql.DBCore, query string, args ...interface{}) (sql.Result, error) {
	start := time.NowUnixMillis()
	res, e := rs.ExecContext(c.req.Context(), query, args...)
	c.writeQueryInfo(query, args, start)
	return res, e
}

func (c *_ctx) sqlQuery(rs isql.DBCore, query string, args ...interface{}) (isql.Rows, error) {
	start := time.NowUnixMillis()
	rows, e := rs.QueryContext(c.req.Context(), query, args...)
	c.writeQueryInfo(query, args, start)
	return rows, e
}

func (c *_ctx) sqlQueryRow(rs isql.DBCore, query string, args ...interface{}) isql.Row {
	start := time.NowUnixMillis()
	row := rs.QueryRowContext(c.req.Context(), query, args...)
	c.writeQueryInfo(query, args, start)
	return row
}

func (c *_ctx) writeQueryInfo(query string, args interface{}, startUnixMillis int64) {
	if !c.doProfile() {
		return
	}
	c.queryInfosMtx.Lock()
	defer c.queryInfosMtx.Unlock()
	c.queryInfos = append(c.queryInfos, queryinfo.New(query, args, startUnixMillis))
}

func (c *_ctx) getQueryInfos() []*queryinfo.QueryInfo {
	c.queryInfosMtx.RLock()
	defer c.queryInfosMtx.RUnlock()
	cpy := make([]*queryinfo.QueryInfo, 0, len(c.queryInfos))
	for _, qi := range c.queryInfos {
		cpy = append(cpy, qi)
	}
	return cpy
}

func (c *_ctx) getDlm(dlmKeys map[string]bool) (int64, error) {
	dlmsToFetch := make([]interface{}, 0, len(dlmKeys))
	latestDlm := int64(0)
	for dlmKey := range dlmKeys {
		dlm, exists := c.retrievedDlms[dlmKey]
		if !exists {
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && dlm > latestDlm {
			latestDlm = dlm
		}
	}
	if len(dlmsToFetch) > 0 {
		cnn := c.SR.DlmAndDataRedisPool.Get()
		defer cnn.Close()
		start := time.NowUnixMillis()
		dlms, e := redis.Int64s(cnn.Do("MGET", dlmsToFetch...))
		c.writeQueryInfo("MGET", dlmsToFetch, start)
		if e != nil {
			return 0, e
		}
		for i, dlm := range dlms {
			c.retrievedDlms[dlmsToFetch[i].(string)] = dlm
			if dlm > latestDlm {
				latestDlm = dlm
			}
		}
	}
	return latestDlm, nil
}

func (c *_ctx) doCacheUpdate() {
	if !c.SR.CachingEnabled || (len(c.dlmsToUpdate) == 0 && len(c.cacheItemsToUpdate) == 0) {
		return
	}
	setArgs := make([]interface{}, 0, (len(c.dlmsToUpdate)*2)+(len(c.cacheItemsToUpdate)*2))
	for k := range c.dlmsToUpdate {
		setArgs = append(setArgs, k, c.requestStartUnixMillis)
	}
	for k, v := range c.cacheItemsToUpdate {
		setArgs = append(setArgs, k, v)
	}
	cnn := c.SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	if len(setArgs) > 0 {
		start := time.NowUnixMillis()
		_, e := cnn.Do("MSET", setArgs...)
		c.writeQueryInfo("MSET", setArgs, start)
		c.LogIf(e)
	}
}

func (c *_ctx) getFixedTreeReadSlave(shard int) isql.DBCore {
	c.fixedTreeReadSlaveMtx.RLock()
	if c.fixedTreeReadSlave == nil {
		c.fixedTreeReadSlaveMtx.RUnlock()
		c.fixedTreeReadSlaveMtx.Lock()
		defer c.fixedTreeReadSlaveMtx.Unlock()
		slaves := c.SR.TreeShards[shard].Slaves()
		if len(slaves) == 0 {
			c.fixedTreeReadSlave = c.SR.TreeShards[shard].Primary()
		} else {
			c.fixedTreeReadSlave = slaves[rand.Intn(len(slaves))]
		}
	} else {
		c.fixedTreeReadSlaveMtx.RUnlock()
	}
	return c.fixedTreeReadSlave
}

type valueCacheKey struct {
	MasterKey string      `json:"masterKey"`
	Key       string      `json:"key"`
	DlmKeys   []string    `json:"dlmKeys"`
	Dlm       int64       `json:"dlm"`
	Args      interface{} `json:"args"`
}

func (c *_ctx) queryBoolVal(name string, def bool) bool {
	switch strings.ToLower(c.req.URL.Query().Get(name)) {
	case "1", "y", "yes", "t", "true":
		return true
	case "0", "n", "no", "f", "false":
		return false
	default:
		return def
	}
}
