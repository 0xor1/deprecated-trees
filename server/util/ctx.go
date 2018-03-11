package util

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// per request info fields
type Ctx struct {
	myId                   *Id
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

func (c *Ctx) TryMyId() *Id {
	return c.myId
}

func (c *Ctx) MyId() Id {
	if c.myId == nil {
		unauthorizedErr.Panic()
	}
	return *c.myId
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
	dlm, err := getDlm(c, dlmKeys)
	if err != nil {
		c.Log(err)
		return false
	}
	if dlm > c.requestStartUnixMillis {
		return false
	}
	jsonBytes, err := json.Marshal(&valueCacheKey{MasterKey: c.staticResources.MasterCacheKey, Key: key, Args: args})
	if err != nil {
		c.Log(err)
		return false
	}
	cnn := c.staticResources.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	start := NowUnixMillis()
	jsonBytes, err = redis.Bytes(cnn.Do(GET, jsonBytes))
	writeQueryInfo(c, &QueryInfo{Query: GET, Args: jsonBytes, Duration: NowUnixMillis() - start})
	if err != nil {
		c.Log(err)
		return false
	}
	if len(jsonBytes) == 0 {
		return false
	}
	err = json.Unmarshal(jsonBytes, val)
	if err != nil {
		c.Log(err)
		return false
	}
	return true
}

func (c *Ctx) SetCacheValue(val interface{}, key string, dlmKeys []string, args interface{}) {
	if val == nil || key == "" {
		InvalidArgumentsErr.Panic()
	}
	valBytes, err := json.Marshal(val)
	if err != nil {
		c.Log(err)
		return
	}
	cacheKeyBytes, err := json.Marshal(&valueCacheKey{MasterKey: c.staticResources.MasterCacheKey, Key: key, Args: args})
	if err != nil {
		c.Log(err)
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

func (c *Ctx) DlmKeyForSystem() string {
	return "sys"
}

func (c *Ctx) DlmKeyForAccountMaster(accountId Id) string {
	return dlmKeyFor(c, "amstr", accountId)
}

func (c *Ctx) DlmKeyForAccount(accountId Id) string {
	return dlmKeyFor(c, "a", accountId)
}

func (c *Ctx) DlmKeyForAccountActivities(accountId Id) string {
	return dlmKeyFor(c, "aa", accountId)
}

func (c *Ctx) DlmKeyForAccountMember(accountId Id) string {
	return dlmKeyFor(c, "am", accountId)
}

func (c *Ctx) DlmKeyForAllAccountMembers(accountId Id) string {
	return dlmKeyFor(c, "ams", accountId)
}

func (c *Ctx) DlmKeyForProjectMaster(projectId Id) string {
	return dlmKeyFor(c, "pmstr", projectId)
}

func (c *Ctx) DlmKeyForProject(projectId Id) string {
	return dlmKeyFor(c, "p", projectId)
}

func (c *Ctx) DlmKeyForProjectActivities(projectId Id) string {
	return dlmKeyFor(c, "pa", projectId)
}

func (c *Ctx) DlmKeyForProjectMember(projectMemberId Id) string {
	return dlmKeyFor(c, "pm", projectMemberId)
}

func (c *Ctx) DlmKeyForAllProjectMembers(projectId Id) string {
	return dlmKeyFor(c, "pms", projectId)
}

func (c *Ctx) DlmKeyForTask(taskId Id) string {
	return dlmKeyFor(c, "t", taskId)
}

func (c *Ctx) DlmKeyForTasks(taskIds []Id) []string {
	strs := make([]string, 0, len(taskIds))
	for _, id := range taskIds {
		strs = append(strs, c.DlmKeyForTask(id))
	}
	return strs
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

func (c *Ctx) RegionalV1PrivateClient() RegionalV1PrivateClient {
	return c.staticResources.RegionalV1PrivateClient
}

func (c *Ctx) MailClient() MailClient {
	return c.staticResources.MailClient
}

func (c *Ctx) AvatarClient() AvatarClient {
	return c.staticResources.AvatarClient
}

// helpers

func sqlExec(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) (sql.Result, error) {
	start := NowUnixMillis()
	res, err := rs.Exec(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: NowUnixMillis() - start})
	return res, err
}

func sqlQuery(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) (isql.Rows, error) {
	start := NowUnixMillis()
	rows, err := rs.Query(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: NowUnixMillis() - start})
	return rows, err
}

func sqlQueryRow(ctx *Ctx, rs isql.ReplicaSet, query string, args ...interface{}) isql.Row {
	start := NowUnixMillis()
	row := rs.QueryRow(query, args...)
	writeQueryInfo(ctx, &QueryInfo{Query: query, Args: args, Duration: NowUnixMillis() - start})
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

func dlmKeyFor(c *Ctx, typeKey string, id Id) string {
	return typeKey + ":" + id.String()
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
				panic(&missingDlmErr{
					dlmKey:  dlmKey,
					reqPath: ctx.req.URL.Path,
				})
			}
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && dlm > latestDlm {
			latestDlm = dlm
		}
	}
	if len(dlmsToFetch) > 0 {
		cnn := ctx.staticResources.DlmAndDataRedisPool.Get()
		defer cnn.Close()
		start := NowUnixMillis()
		dlms, err := redis.Int64s(cnn.Do("MGET", dlmsToFetch...))
		writeQueryInfo(ctx, &QueryInfo{Query: "MGET", Args: dlmsToFetch, Duration: NowUnixMillis() - start})
		if err != nil {
			return 0, err
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
		start := NowUnixMillis()
		_, err := cnn.Do("MSET", setArgs...)
		writeQueryInfo(ctx, &QueryInfo{Query: "MSET", Args: setArgs, Duration: NowUnixMillis() - start})
		if err != nil {
			ctx.Log(err)
		}
	}
	if len(delArgs) > 0 {
		start := NowUnixMillis()
		_, err := cnn.Do("DEL", setArgs...)
		writeQueryInfo(ctx, &QueryInfo{Query: "DEL", Args: setArgs, Duration: NowUnixMillis() - start})
		if err != nil {
			ctx.Log(err)
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

type MailClient interface {
	Send(sendTo []string, content string)
}

type AvatarClient interface {
	MaxAvatarDim() uint
	Save(key string, mimeType string, data io.Reader)
	Delete(key string)
	DeleteAll()
}

type RegionalV1PrivateClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, myId Id, myName string, myDisplayName *string) (int, error)
	DeleteAccount(region string, shard int, account, myId Id) error
	AddMembers(region string, shard int, account, myId Id, members []*AddMemberPrivate) error
	RemoveMembers(region string, shard int, account, myId Id, members []Id) error
	MemberIsOnlyAccountOwner(region string, shard int, account, myId Id) (bool, error)
	SetMemberName(region string, shard int, account, myId Id, newName string) error
	SetMemberDisplayName(region string, shard int, account, myId Id, newDisplayName *string) error
	MemberIsAccountOwner(region string, shard int, account, myId Id) (bool, error)
}

// Collection of application static resources, in lcl and dev "onebox" environments all values must be set
// but in stg and prd environments central and regional endpoints are physically separated and so not all values are valid
// e.g. account and pwd dbs are only initialised on central service, whilst redis pool and tree shards are only initialised
// for regional endpoints.
type StaticResources struct {
	// server address eg "127.0.0.1:8787"
	ServerAddress string
	// must be one of "lcl", "dev", "stg", "prd"
	Env string
	// must be one of "lcl", "dev", "central", "use", "usw", "euw"
	Region string
	// commit sha
	Version string
	// api docs path
	ApiDocsRoute string
	// session cookie name
	SessionCookieName string
	// session cookie store
	SessionStore *sessions.CookieStore
	// lowercase paths to endpoints map
	Routes map[string]*Endpoint
	// indented json api docs
	ApiDocs []byte
	// incremental base64 value
	MasterCacheKey string
	// regexes that account names must match to be valid during account creation or name setting
	NameRegexMatchers []*regexp.Regexp
	// regexes that account pwds must match to be valid during account creation or pwd setting
	PwdRegexMatchers []*regexp.Regexp
	// minimum number of runes required for a valid account name
	NameMinRuneCount int
	// maximum number of runes required for a valid account name
	NameMaxRuneCount int
	// minimum number of runes required for a valid account pwd
	PwdMinRuneCount int
	// maximum number of runes required for a valid account pwd
	PwdMaxRuneCount int
	// max number of entities that can be processed at once, also used for max limit value on queries
	MaxProcessEntityCount int
	// length of cryptographic codes, used in email links for validating email addresses and resetting pwds
	CryptCodeLen int
	// length of salts used for pwd hashing
	SaltLen int
	// scrypt N value
	ScryptN int
	// scrypt R value
	ScryptR int
	// scrypt P value
	ScryptP int
	// scrypt key length
	ScryptKeyLen int
	// regional v1 private client secret
	RegionalV1PrivateClientSecret []byte
	// regional v1 private client used by central endpoints
	RegionalV1PrivateClient RegionalV1PrivateClient
	// mail client for sending emails
	MailClient MailClient
	// avatar client for storing avatar images
	AvatarClient AvatarClient
	// error logging function
	LogError func(error)
	// stats logging function
	LogStats func(status int, method, path string, reqStartUnixMillis int64, queryInfos []*QueryInfo)
	// account sql connection
	AccountDb isql.ReplicaSet
	// pwd sql connection
	PwdDb isql.ReplicaSet
	// tree shard sql connections
	TreeShards map[int]isql.ReplicaSet
	// redis pool for caching layer
	DlmAndDataRedisPool iredis.Pool
	// redis pool for private request keys to check for replay attacks
	PrivateKeyRedisPool iredis.Pool
}

func (sr *StaticResources) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := &responseWrapper{code: 0, w: w}
	ctx := &Ctx{
		requestStartUnixMillis: NowUnixMillis(),
		resp:               resp,
		req:                req,
		retrievedDlms:      map[string]int64{},
		dlmsToUpdate:       map[string]interface{}{},
		cacheItemsToUpdate: map[string]interface{}{},
		cacheKeysToDelete:  map[string]interface{}{},
		queryInfosMtx:      &sync.RWMutex{},
		queryInfos:         make([]*QueryInfo, 0, 10),
		staticResources:    sr,
	}
	//always do case insensitive path routing
	lowerPath := strings.ToLower(req.URL.Path)
	// defer func handles logging panic errors and returning 500s and logging request/response/database/cache stats to datadog in none lcl env
	defer func() {
		context.Clear(req) //required for guerilla cookie session usage, or resources will leak
		r := recover()
		if r != nil {
			pErr, ok := r.(PermissionedError)
			if ok && pErr != nil && pErr.IsPublic() {
				writeJson(resp, http.StatusInternalServerError, pErr)
			} else {
				writeJson(resp, http.StatusInternalServerError, internalServerErr)
			}
			err := r.(error)
			if err != nil {
				sr.LogError(err)
			}
		}
		sr.LogStats(resp.code, req.Method, lowerPath, ctx.requestStartUnixMillis, getQueryInfos(ctx))
	}()
	//must make sure to close the request body
	if req != nil && req.Body != nil {
		defer req.Body.Close()
	}
	//check for special case of api docs first
	if lowerPath == sr.ApiDocsRoute {
		writeRawJson(resp, 200, sr.ApiDocs)
		return
	}
	//get endpoint
	ep := sr.Routes[lowerPath]
	// check for 404
	if ep == nil || ep.Method != strings.ToUpper(req.Method) {
		http.NotFound(resp, req)
		return
	}
	// only none private endpoints use sessions
	if !ep.IsPrivate {
		//get a cookie session
		ctx.session, _ = sr.SessionStore.Get(req, sr.SessionCookieName)
		if ctx.session != nil {
			iMyId := ctx.session.Values["myId"]
			if iMyId != nil {
				id := iMyId.(Id)
				ctx.myId = &id
			}
		}
		//check for valid myId value if endpoint requires active session, and check for X header in POST requests for CSRF prevention
		if ep.RequiresSession && ctx.myId == nil || req.Method == POST && req.Header.Get("X-Client") == "" {
			writeJson(resp, http.StatusUnauthorized, unauthorizedErr)
			return
		}
		//setup ctx
	}
	//process args
	var err error
	var argsBytes []byte
	var args interface{}
	reqQueryValues := req.URL.Query()
	if ep.Method == GET && ep.GetArgsStruct != nil {
		argsBytes = []byte(reqQueryValues.Get("args"))
	} else if ep.Method == POST && ep.GetArgsStruct != nil {
		argsBytes, err = ioutil.ReadAll(req.Body)
		PanicIf(err)
	} else if ep.Method == POST && ep.ProcessForm != nil {
		if ep.IsPrivate {
			// private endpoints dont support post requests with form data
			invalidEndpointErr.Panic()
		}
		args = ep.ProcessForm(resp, req)
	}
	//process private ts and key args
	if ep.IsPrivate {
		ts, err := strconv.ParseInt(reqQueryValues.Get("ts"), 10, 64)
		PanicIf(err)
		//if the timestamp the req was sent is over a minute ago, reject the request
		if NowUnixMillis()-ts > 60000 {
			FmtPanic("suspicious private request sent over a minute ago")
		}
		key, err := base64.RawURLEncoding.DecodeString(reqQueryValues.Get("_"))
		PanicIf(err)
		// check the args/timestamp/key are valid
		if !bytes.Equal(key, ScryptKey(append(argsBytes, []byte(reqQueryValues.Get("ts"))...), sr.RegionalV1PrivateClientSecret, sr.ScryptN, sr.ScryptR, sr.ScryptP, sr.ScryptKeyLen)) {
			FmtPanic("invalid private request keys don't match")
		}
		//check redis cache to ensure key has not appeared in the last minute, to prevent replay attacks
		cnn := sr.PrivateKeyRedisPool.Get()
		defer cnn.Close()
		cnn.Send("MULTI")
		cnn.Send("SETNX", reqQueryValues.Get("_"), "")
		cnn.Send("EXPIRE", reqQueryValues.Get("_"), 60)
		vals, err := redis.Ints(cnn.Do("EXEC"))
		PanicIf(err)
		if len(vals) != 2 {
			FmtPanic("vals should have exactly two integer values")
		}
		if vals[0] != 1 {
			FmtPanic("private request key duplication, replay attack detection")
		}
		if vals[1] != 1 {
			FmtPanic("failed to set expiry on private request key")
		}
		//at this point private request is valid
	}
	if len(argsBytes) > 0 {
		args = ep.GetArgsStruct()
		PanicIf(json.Unmarshal(argsBytes, args))
	}
	//if this endpoint is the authentication endpoint it should return just the users Id, add it to the session cookie
	if ep.IsAuthentication {
		iId := ep.CtxHandler(ctx, args)
		if myId, ok := iId.(Id); !ok {
			FmtPanic("isAuthentication did not return Id type")
		} else {
			ctx.myId = &myId //set myId on ctx for logging info in defer above
			ctx.session.Values["myId"] = myId
			ctx.session.Values["AuthedOn"] = NowUnixMillis()
			ctx.session.Save(req, resp)
			writeJsonOk(ctx.resp, myId)
		}
	} else {
		writeJsonOk(ctx.resp, ep.CtxHandler(ctx, args))
	}
}

func writeJsonOk(w http.ResponseWriter, body interface{}) {
	writeJson(w, http.StatusOK, body)
}

func writeJson(w http.ResponseWriter, code int, body interface{}) {
	bodyBytes, err := json.Marshal(body)
	PanicIf(err)
	writeRawJson(w, code, bodyBytes)
}

func writeRawJson(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	_, err := w.Write(body)
	PanicIf(err)
}

func NewLocalAvatarStore(relDirPath string, maxAvatarDim uint) AvatarClient {
	if relDirPath == "" {
		InvalidArgumentsErr.Panic()
	}
	wd, err := os.Getwd()
	PanicIf(err)
	absDirPath := path.Join(wd, relDirPath)
	os.MkdirAll(absDirPath, os.ModeDir)
	return &localAvatarStore{
		mtx:          &sync.Mutex{},
		maxAvatarDim: maxAvatarDim,
		absDirPath:   absDirPath,
	}
}

type localAvatarStore struct {
	mtx          *sync.Mutex
	maxAvatarDim uint
	absDirPath   string
}

func (s *localAvatarStore) MaxAvatarDim() uint {
	return s.maxAvatarDim
}

func (s *localAvatarStore) Save(key string, mimeType string, data io.Reader) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	avatarBytes, err := ioutil.ReadAll(data)
	PanicIf(err)
	PanicIf(ioutil.WriteFile(path.Join(s.absDirPath, key), avatarBytes, os.ModePerm))
}

func (s *localAvatarStore) Delete(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	PanicIf(os.Remove(path.Join(s.absDirPath, key)))
}

func (s *localAvatarStore) DeleteAll() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	PanicIf(os.RemoveAll(s.absDirPath))
}

func NewLocalMailClient() MailClient {
	return &localMailClient{}
}

type localMailClient struct{}

func (s *localMailClient) Send(sendTo []string, content string) {
	fmt.Println(sendTo, content)
}

type responseWrapper struct {
	code int
	w    http.ResponseWriter
}

func (r *responseWrapper) Header() http.Header {
	return r.w.Header()
}

func (r *responseWrapper) Write(data []byte) (int, error) {
	return r.w.Write(data)
}

func (r *responseWrapper) WriteHeader(code int) {
	r.code = code
	r.w.WriteHeader(code)
}
