package server

import (
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/crypt"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/queryinfo"
	"bitbucket.org/0xor1/trees/server/util/static"
	t "bitbucket.org/0xor1/trees/server/util/time"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/0xor1/panic"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func New(sr *static.Resources, endpointSets ...[]*endpoint.Endpoint) *Server {
	routes := map[string]*endpoint.Endpoint{}
	privateKeyGen := func(argsBytes []byte, ts string) []byte {
		return crypt.ScryptKey(append(argsBytes, []byte(ts)...), sr.RegionalV1PrivateClientSecret, sr.ScryptN, sr.ScryptR, sr.ScryptP, sr.ScryptKeyLen)
	}
	for _, endpointSet := range endpointSets {
		for _, ep := range endpointSet {
			ep.ValidateEndpoint()
			lowerPath := strings.ToLower(ep.Path)
			if _, exists := routes[lowerPath]; exists {
				err.FmtPanic("duplicate endpoint path %q", lowerPath)
			}
			routes[lowerPath] = ep
			ep.PrivateKeyGen = privateKeyGen
		}
	}
	routeDocs := make([]interface{}, 0, len(routes))
	for _, endpointSet := range endpointSets {
		for _, ep := range endpointSet {
			if !ep.IsPrivate {
				routeDocs = append(routeDocs, ep.GetEndpointDocumentation())
			}
		}
	}
	var e error
	sr.ApiDocs, e = json.MarshalIndent(routeDocs, "", "    ")
	panic.If(e)
	fileServerDir, e := filepath.Abs(sr.FileServerDir)
	panic.If(e)
	return &Server{
		Routes:     routes,
		SR:         sr,
		FileServer: http.FileServer(http.Dir(fileServerDir)),
	}
}

type Server struct {
	Routes     map[string]*endpoint.Endpoint
	SR         *static.Resources
	FileServer http.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := &responseWrapper{code: 0, w: w}
	//setup _ctx
	ctx := &_ctx{
		requestStartUnixMillis: t.NowUnixMillis(),
		resp:                  resp,
		req:                   req,
		retrievedDlms:         map[string]int64{},
		dlmsToUpdate:          map[string]interface{}{},
		cacheItemsToUpdate:    map[string]interface{}{},
		queryInfosMtx:         &sync.RWMutex{},
		queryInfos:            make([]*queryinfo.QueryInfo, 0, 10),
		fixedTreeReadSlaveMtx: &sync.RWMutex{},
		SR: s.SR,
	}
	//always do case insensitive path routing
	lowerPath := strings.ToLower(req.URL.Path)
	// defer func handles logging panic errors and returning 500s and logging request/response/database/cache stats to datadog in none lcl env
	defer func() {
		context.Clear(req) //required for guerrilla cookie session usage, or resources will leak
		r := recover()
		if r != nil {
			e, ok := r.(*err.Err)
			if ok && e != nil {
				if e == err.InsufficientPermission {
					http.NotFound(resp, req)
				} else {
					writeJson(resp, http.StatusInternalServerError, e)
				}
			} else {
				writeJson(resp, http.StatusInternalServerError, err.External)
			}
			er := r.(error)
			if er != nil {
				s.SR.LogError(er)
			}
		}
		s.SR.LogStats(resp.code, req.Method, lowerPath, ctx.requestStartUnixMillis, ctx.getQueryInfos())
	}()
	//must make sure to close the request body
	if req != nil && req.Body != nil {
		defer req.Body.Close()
	}
	//set common headers
	resp.Header().Set("X-Frame-Options", "DENY")
	resp.Header().Set("X-XSS-Protection", "1; mode=block")
	resp.Header().Set("Content-Security-Policy", "default-src *.project-trees.com")
	resp.Header().Set("Cache-Control", "private, must-revalidate, max-stale=0, max-age=0")
	resp.Header().Set("X-Version", s.SR.Version)
	//check for none api call
	if req.Method == http.MethodGet && !strings.HasPrefix(lowerPath, "/api/") {
		s.FileServer.ServeHTTP(resp, req)
		return
	}
	//check for special case of api docs first
	if req.Method == http.MethodGet && lowerPath == s.SR.ApiDocsRoute {
		writeRawJson(resp, 200, s.SR.ApiDocs)
		return
	}
	//check for special case of api logout
	if req.Method == http.MethodPost && lowerPath == s.SR.ApiLogoutRoute {
		var e error
		ctx.session, e = s.SR.SessionStore.Get(req, s.SR.SessionCookieName)
		panic.If(e)
		ctx.session.Options.MaxAge = -1
		ctx.session.Values = map[interface{}]interface{}{}
		ctx.session.Save(req, resp)
		writeJsonOk(resp, nil)
		return
	}
	reqQueryValues := req.URL.Query()
	panic.IfTruef(reqQueryValues.Get("region") == "", "missing region query param")
	// act as a proxy for other regions if necessary, this will only happen in stg and pro environments as lcl and dev are one box environments
	lowerRegion := strings.ToLower(reqQueryValues.Get("region"))
	if !(s.SR.Env == cnst.LclEnv || s.SR.Env == cnst.DevEnv) && lowerRegion != s.SR.Region {
		req.URL.Host = fmt.Sprintf("%s-%s-api.project-trees.com")
		proxyResp, e := http.DefaultClient.Do(req)
		panic.If(e)
		if proxyResp != nil && proxyResp.Body != nil {
			defer proxyResp.Body.Close()
		}
		for k, vv := range proxyResp.Header { // copy headers
			for _, v := range vv {
				resp.Header().Add(k, v)
			}
		}
		for _, cookie := range proxyResp.Cookies() { //copy cookies
			http.SetCookie(resp, cookie)
		}
		resp.WriteHeader(proxyResp.StatusCode)
		io.Copy(resp, proxyResp.Body)
		return
	}
	//check for special case of api mget
	if req.Method == http.MethodGet && lowerPath == s.SR.ApiMGetRoute {
		reqs := map[string]string{}
		panic.If(json.Unmarshal([]byte(req.URL.Query().Get("args")), &reqs))
		fullMGetResponse := map[string]*mgetResponse{}
		fullMGetResponseMtx := &sync.Mutex{}
		includeHeaders := ctx.queryBoolVal("headers", false)
		gets := make([]func(), 0, len(reqs))
		for key, reqUrl := range reqs {
			gets = append(gets, func(key, reqUrl string) func() {
				return func() {
					r, _ := http.NewRequest(http.MethodGet, reqUrl, nil)
					for _, c := range req.Cookies() {
						r.AddCookie(c)
					}
					w := &mgetResponseWriter{header: http.Header{}, body: bytes.NewBuffer(make([]byte, 0, 1000))}
					s.ServeHTTP(w, r)
					fullMGetResponseMtx.Lock()
					defer fullMGetResponseMtx.Unlock()
					fullMGetResponse[key] = &mgetResponse{
						includeHeaders: includeHeaders,
						Code:           w.code,
						Header:         w.header,
						Body:           w.body.Bytes(),
					}
				}
			}(key, reqUrl))
		}
		panic.If(panic.SafeGoGroup(s.SR.ApiMGetTimeout, gets...))
		writeJsonOk(ctx.resp, fullMGetResponse)
		return
	}
	//get endpoint
	ep := s.Routes[lowerPath]
	// check for 404
	if ep == nil || ep.Method != strings.ToUpper(req.Method) {
		http.NotFound(resp, req)
		return
	}
	// only none private endpoints use sessions
	if !ep.IsPrivate {
		//get a cookie session
		var e error
		ctx.session, e = s.SR.SessionStore.Get(req, s.SR.SessionCookieName)
		panic.If(e)
		if ctx.session != nil {
			iMe := ctx.session.Values["me"]
			if iMe != nil {
				me := iMe.(id.Id)
				ctx.me = &me
			}
		}
		//check for valid me value if endpoint requires active session, and check for X header in POST requests for CSRF prevention
		if ep.RequiresSession && ctx.me == nil || req.Method == http.MethodPost && req.Header.Get("X-Client") == "" {
			writeJson(resp, http.StatusUnauthorized, unauthorizedErr)
			return
		}
	}
	//process args
	var e error
	var argsBytes []byte
	var args interface{}
	if ep.Method == http.MethodGet && ep.GetArgsStruct != nil {
		argsBytes = []byte(reqQueryValues.Get("args"))
	} else if ep.Method == http.MethodPost && ep.GetArgsStruct != nil {
		argsBytes, e = ioutil.ReadAll(req.Body)
		panic.If(e)
	} else if ep.Method == http.MethodPost && ep.ProcessForm != nil {
		if ep.IsPrivate {
			// private endpoints dont support post requests with form data
			err.FmtPanic("private endpoints don't support POST Form data")
		}
		args = ep.ProcessForm(resp, req)
	}
	//process private ts and key args
	if ep.IsPrivate {
		ts, e := strconv.ParseInt(reqQueryValues.Get("ts"), 10, 64)
		panic.If(e)
		//if the timestamp the req was sent is over a minute ago, reject the request
		if t.NowUnixMillis()-ts > 60000 {
			err.FmtPanic("suspicious private request sent over a minute ago")
		}
		key, e := base64.RawURLEncoding.DecodeString(reqQueryValues.Get("_"))
		panic.If(e)
		// check the args/timestamp/key are valid
		if !bytes.Equal(key, ep.PrivateKeyGen(argsBytes, reqQueryValues.Get("ts"))) {
			err.FmtPanic("invalid private request keys don't match")
		}
		//check redis cache to ensure key has not appeared in the last minute, to prevent replay attacks
		cnn := s.SR.PrivateKeyRedisPool.Get()
		defer cnn.Close()
		cnn.Send("MULTI")
		cnn.Send("SETNX", reqQueryValues.Get("_"), "")
		cnn.Send("EXPIRE", reqQueryValues.Get("_"), 61)
		vals, e := redis.Ints(cnn.Do("EXEC"))
		panic.If(e)
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
		panic.If(json.Unmarshal(argsBytes, args))
	}
	//if this endpoint is the authentication endpoint it should return just the users id.Id, add it to the session cookie
	result := ep.CtxHandler(ctx, args)
	if ep.IsAuthentication {
		if me, ok := result.(id.Identifiable); !ok {
			err.FmtPanic("isAuthentication did not return id.Identifiable type")
		} else {
			i := me.Id()
			ctx.me = &i //set me on _ctx for logging info in defer above
			ctx.session.Values["me"] = i
			ctx.session.Values["AuthedOn"] = t.NowUnixMillis()
			ctx.session.Save(req, resp)
		}
	}
	ctx.doCacheUpdate()
	if ctx.doProfile() {
		writeJsonOk(ctx.resp, &profileResponse{
			Duration:   t.NowUnixMillis() - ctx.requestStartUnixMillis,
			QueryInfos: ctx.getQueryInfos(),
			Result:     result,
		})
	} else {
		writeJsonOk(ctx.resp, result)
	}
}

func writeJsonOk(w http.ResponseWriter, body interface{}) {
	writeJson(w, http.StatusOK, body)
}

func writeJson(w http.ResponseWriter, code int, body interface{}) {
	bodyBytes, e := json.Marshal(body)
	panic.If(e)
	writeRawJson(w, code, bodyBytes)
}

func writeRawJson(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(code)
	_, e := w.Write(body)
	panic.If(e)
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

type mgetResponseWriter struct {
	code   int
	header http.Header
	body   *bytes.Buffer
}

func (r *mgetResponseWriter) Header() http.Header {
	return r.header
}

func (r *mgetResponseWriter) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *mgetResponseWriter) WriteHeader(code int) {
	r.code = code
}

type mgetResponse struct {
	includeHeaders bool
	Code           int         `json:"code"`
	Header         http.Header `json:"header"`
	Body           []byte      `json:"body"`
}

func (r *mgetResponse) MarshalJSON() ([]byte, error) {
	if r.includeHeaders {
		h, _ := json.Marshal(r.Header)
		if r.Code == 200 {
			return []byte(fmt.Sprintf(`{"code":%d,"header":%s,"body":%s}`, r.Code, h, r.Body)), nil
		} else {
			return []byte(fmt.Sprintf(`{"code":%d,"header":%s,"body":%q}`, r.Code, h, r.Body)), nil
		}
	} else {
		if r.Code == 200 {
			return []byte(fmt.Sprintf(`{"code":%d,"body":%s}`, r.Code, r.Body)), nil
		} else {
			return []byte(fmt.Sprintf(`{"code":%d,"body":%q}`, r.Code, r.Body)), nil
		}
	}
}

type profileResponse struct {
	Duration   int64                  `json:"duration"`
	QueryInfos []*queryinfo.QueryInfo `json:"queryInfos"`
	Result     interface{}            `json:"result"`
}
