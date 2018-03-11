package core

import (
	"bitbucket.org/0xor1/task/server/util/avatar"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/crypt"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/mail"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/time"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

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
	RegionalV1PrivateClient private.V1Client
	// mail client for sending emails
	MailClient mail.Client
	// avatar client for storing avatar images
	AvatarClient avatar.Client
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
		requestStartUnixMillis: time.NowUnixMillis(),
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
			e, ok := r.(*err.Err)
			if ok && e != nil {
				writeJson(resp, http.StatusInternalServerError, e)
			} else {
				writeJson(resp, http.StatusInternalServerError, err.External)
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
			iMyId := ctx.session.Values["me"]
			if iMyId != nil {
				id := iMyId.(id.Id)
				ctx.me = &id
			}
		}
		//check for valid me value if endpoint requires active session, and check for X header in POST requests for CSRF prevention
		if ep.RequiresSession && ctx.me == nil || req.Method == cnst.POST && req.Header.Get("X-Client") == "" {
			writeJson(resp, http.StatusUnauthorized, unauthorizedErr)
			return
		}
		//setup ctx
	}
	//process args
	var e error
	var argsBytes []byte
	var args interface{}
	reqQueryValues := req.URL.Query()
	if ep.Method == cnst.GET && ep.GetArgsStruct != nil {
		argsBytes = []byte(reqQueryValues.Get("args"))
	} else if ep.Method == cnst.POST && ep.GetArgsStruct != nil {
		argsBytes, e = ioutil.ReadAll(req.Body)
		err.PanicIf(e)
	} else if ep.Method == cnst.POST && ep.ProcessForm != nil {
		if ep.IsPrivate {
			// private endpoints dont support post requests with form data
			panic(invalidEndpointErr)
		}
		args = ep.ProcessForm(resp, req)
	}
	//process private ts and key args
	if ep.IsPrivate {
		ts, e := strconv.ParseInt(reqQueryValues.Get("ts"), 10, 64)
		err.PanicIf(e)
		//if the timestamp the req was sent is over a minute ago, reject the request
		if time.NowUnixMillis()-ts > 60000 {
			err.FmtPanic("suspicious private request sent over a minute ago")
		}
		key, e := base64.RawURLEncoding.DecodeString(reqQueryValues.Get("_"))
		err.PanicIf(e)
		// check the args/timestamp/key are valid
		if !bytes.Equal(key, crypt.ScryptKey(append(argsBytes, []byte(reqQueryValues.Get("ts"))...), sr.RegionalV1PrivateClientSecret, sr.ScryptN, sr.ScryptR, sr.ScryptP, sr.ScryptKeyLen)) {
			err.FmtPanic("invalid private request keys don't match")
		}
		//check redis cache to ensure key has not appeared in the last minute, to prevent replay attacks
		cnn := sr.PrivateKeyRedisPool.Get()
		defer cnn.Close()
		cnn.Send("MULTI")
		cnn.Send("SETNX", reqQueryValues.Get("_"), "")
		cnn.Send("EXPIRE", reqQueryValues.Get("_"), 60)
		vals, e := redis.Ints(cnn.Do("EXEC"))
		err.PanicIf(e)
		if len(vals) != 2 {
			err.FmtPanic("vals should have exactly two integer values")
		}
		if vals[0] != 1 {
			err.FmtPanic("private request key duplication, replay attack detection")
		}
		if vals[1] != 1 {
			err.FmtPanic("failed to set expiry on private request key")
		}
		//at this point private request is valid
	}
	if len(argsBytes) > 0 {
		args = ep.GetArgsStruct()
		err.PanicIf(json.Unmarshal(argsBytes, args))
	}
	//if this endpoint is the authentication endpoint it should return just the users id.Id, add it to the session cookie
	if ep.IsAuthentication {
		iId := ep.CtxHandler(ctx, args)
		if myId, ok := iId.(id.Id); !ok {
			err.FmtPanic("isAuthentication did not return id.Id type")
		} else {
			ctx.me = &myId //set me on ctx for logging info in defer above
			ctx.session.Values["me"] = myId
			ctx.session.Values["AuthedOn"] = time.NowUnixMillis()
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
	bodyBytes, e := json.Marshal(body)
	err.PanicIf(e)
	writeRawJson(w, code, bodyBytes)
}

func writeRawJson(w http.ResponseWriter, code int, body []byte) {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	_, e := w.Write(body)
	err.PanicIf(e)
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
