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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/0xor1/panic"
	"github.com/garyburd/redigo/redis"
	gorillacontext "github.com/gorilla/context"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
			_, exists := routes[lowerPath]
			panic.If(exists, "duplicate endpoint path %q", lowerPath)
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
	panic.IfNotNil(e)
	fileServerDir, e := filepath.Abs(sr.FileServerDir)
	panic.IfNotNil(e)
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
	timeoutCtx, cancel := context.WithTimeout(req.Context(), 2*time.Second) // no request should be allowed to take more than a second
	defer cancel()
	req = req.WithContext(timeoutCtx)
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
		gorillacontext.Clear(req) //required for gorilla cookie session usage, or resources will leak
		r := recover()
		if r != nil {
			e, ok := r.(*err.Http)
			if ok && e != nil {
				writeJson(resp, e.Code, e.Message)
			} else {
				writeJson(resp, http.StatusInternalServerError, "internal server error")
			}
			ctx.LogIf(r.(error))
		}
		s.SR.LogStats(resp.code, lowerPath, ctx.requestStartUnixMillis, ctx.getQueryInfos())
	}()
	// panic with a timeout error if we timeout and a value hasn't already been written
	defer func() {
		r := recover()
		if r != nil {
			e, ok := r.(error)
			err.HttpPanicf(ok && e == context.DeadlineExceeded, http.StatusServiceUnavailable, "request was taking too long to process, try again later")
			panic.IfNotNil(r)
		}
	}()
	//must make sure to close the request body
	if req != nil && req.Body != nil {
		defer req.Body.Close()
	}
	//set common headers
	resp.Header().Set("X-Frame-Options", "DENY")
	resp.Header().Set("X-XSS-Protection", "1; mode=block")
	//resp.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src %s", s.SR.ClientHost)) TODO figure out what is wrong with the vue pwa setup that makes this break it
	resp.Header().Set("Cache-Control", "no-cache, no-store")
	resp.Header().Set("X-Version", s.SR.Version)
	//check for none api call
	if req.Method == http.MethodGet && !strings.HasPrefix(lowerPath, "/api/") {
		s.FileServer.ServeHTTP(resp, req)
		return
	}
	//check for special case of api docs first
	if lowerPath == s.SR.ApiDocsRoute {
		writeRawJson(resp, 200, s.SR.ApiDocs)
		return
	}
	//check for special case of api logout
	if lowerPath == s.SR.ApiLogoutRoute {
		var e error
		ctx.session, e = s.SR.SessionStore.Get(req, s.SR.SessionCookieName)
		panic.IfNotNil(e)
		ctx.session.Options.MaxAge = -1
		ctx.session.Values = map[interface{}]interface{}{}
		ctx.session.Save(req, resp)
		writeJsonOk(resp, nil)
		return
	}
	reqQueryValues := req.URL.Query()
	panic.If(reqQueryValues.Get("region") == "", "missing region query param")
	// act as a proxy for other regions if necessary, this will only happen in stg and pro environments as lcl and dev are one box environments
	region := cnst.Region(strings.ToLower(reqQueryValues.Get("region")))
	region.Validate()
	if !(s.SR.Env == cnst.LclEnv || s.SR.Env == cnst.DevEnv) && region != s.SR.Region {
		req.URL.Host = fmt.Sprintf("%s-%s-api.%s", s.SR.Env, region, s.SR.NakedHost)
		proxyResp, e := http.DefaultClient.Do(req)
		panic.IfNotNil(e)
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
	//check for special case of api mdo
	if lowerPath == s.SR.ApiMDoRoute {
		mDoReqs := map[string]mDoReq{}
		bodyBytes, e := ioutil.ReadAll(req.Body)
		panic.IfNotNil(e)
		panic.IfNotNil(json.Unmarshal(bodyBytes, &mDoReqs))
		fullMGetResponse := map[string]*mgetResponse{}
		fullMGetResponseMtx := &sync.Mutex{}
		includeHeaders := ctx.queryBoolVal("headers", false)
		does := make([]func(), 0, len(mDoReqs))
		for key := range mDoReqs {
			does = append(does, func(key string, reqData mDoReq) func() {
				return func() {
					argsBytes, e := json.Marshal(reqData.Args)
					panic.IfNotNil(e)
					r, _ := http.NewRequest(http.MethodPost, reqData.Path+"?region="+reqData.Region.String(), bytes.NewReader(argsBytes))
					for _, c := range req.Cookies() {
						r.AddCookie(c)
					}
					for name := range req.Header {
						r.Header.Add(name, req.Header.Get(name))
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
			}(key, mDoReqs[key]))
		}
		panic.IfNotNil(panic.SafeGoGroup(does...))
		writeJsonOk(ctx.resp, fullMGetResponse)
		return
	}
	//get endpoint
	ep := s.Routes[lowerPath]
	// check for 404
	err.HttpPanicf(ep == nil, http.StatusNotFound, "not found")
	// only none private endpoints use sessions
	if !ep.IsPrivate {
		//get a cookie session
		var e error
		ctx.session, e = s.SR.SessionStore.Get(req, s.SR.SessionCookieName)
		panic.IfNotNil(e)
		if ctx.session != nil {
			iMe := ctx.session.Values["me"]
			if iMe != nil {
				me := iMe.(id.Id)
				ctx.me = &me
			}
		}
		//check for valid me value if endpoint requires active session, and check for X header in POST requests for CSRF prevention
		err.HttpPanicf(ep.RequiresSession && ctx.me == nil || req.Method == http.MethodPost && req.Header.Get("X-Client") == "", http.StatusUnauthorized, "unauthorized")
	}
	//process args
	var e error
	var argsBytes []byte
	var args interface{}
	if ep.GetArgsStruct != nil {
		argsBytes, e = ioutil.ReadAll(req.Body)
		panic.IfNotNil(e)
	} else if ep.ProcessForm != nil {
		// private endpoints dont support post requests with form data
		err.HttpPanicf(ep.IsPrivate, http.StatusBadRequest, "private endpoints don't support POST Form data")
		args = ep.ProcessForm(resp, req)
	}
	//process private ts and key args
	if ep.IsPrivate {
		ts, e := strconv.ParseInt(reqQueryValues.Get("ts"), 10, 64)
		panic.IfNotNil(e)
		//if the timestamp the req was sent is over a minute ago, reject the request
		err.HttpPanicf(t.NowUnixMillis()-ts > 60000, http.StatusBadRequest, "suspicious private request sent over a minute ago")
		key, e := base64.RawURLEncoding.DecodeString(reqQueryValues.Get("_"))
		panic.IfNotNil(e)
		// check the args/timestamp/key are valid
		err.HttpPanicf(!bytes.Equal(key, ep.PrivateKeyGen(argsBytes, reqQueryValues.Get("ts"))), http.StatusUnauthorized, "invalid private request keys don't match")
		//check redis cache to ensure key has not appeared in the last minute, to prevent replay attacks
		cnn := s.SR.PrivateKeyRedisPool.Get()
		defer cnn.Close()
		cnn.Send("MULTI")
		cnn.Send("SETNX", reqQueryValues.Get("_"), "")
		cnn.Send("EXPIRE", reqQueryValues.Get("_"), 61)
		vals, e := redis.Ints(cnn.Do("EXEC"))
		panic.IfNotNil(e)
		panic.If(len(vals) != 2, "vals should have exactly two integer values")
		err.HttpPanicf(vals[0] != 1, http.StatusUnauthorized, "private request key duplication, replay attack detection")
		panic.If(vals[1] != 1, "failed to set expiry on private request key")
		//at this point private request is valid
	}
	if len(argsBytes) > 0 {
		args = ep.GetArgsStruct()
		panic.IfNotNil(json.Unmarshal(argsBytes, args))
	}
	//if this endpoint is the authentication endpoint it should return just the users id.Id, add it to the session cookie
	result := ep.CtxHandler(ctx, args)
	if ep.IsAuthentication {
		me, ok := result.(id.Identifiable)
		panic.If(!ok, "isAuthentication did not return id.Identifiable type")
		i := me.Id()
		ctx.me = &i //set me on _ctx for logging info in defer above
		ctx.session.Values["me"] = i
		ctx.session.Values["AuthedOn"] = t.NowUnixMillis()
		ctx.session.Save(req, resp)
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
	panic.IfNotNil(e)
	writeRawJson(w, code, bodyBytes)
}

func writeRawJson(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(code)
	_, e := w.Write(body)
	panic.IfNotNil(e)
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

type mDoReq struct {
	Region cnst.Region            `json:"region"`
	Path   string                 `json:"path"`
	Args   map[string]interface{} `json:"args"`
}
