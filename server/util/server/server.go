package server

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/crypt"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/queryinfo"
	"bitbucket.org/0xor1/task/server/util/static"
	t "bitbucket.org/0xor1/task/server/util/time"
	"time"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"fmt"
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
	err.PanicIf(e)
	return &Server{
		Routes: routes,
		SR:     sr,
	}
}

type Server struct {
	Routes map[string]*endpoint.Endpoint
	SR     *static.Resources
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := &responseWrapper{code: 0, w: w}
	ctx := &_ctx{
		requestStartUnixMillis: t.NowUnixMillis(),
		resp:               resp,
		req:                req,
		retrievedDlms:      map[string]int64{},
		dlmsToUpdate:       map[string]interface{}{},
		cacheItemsToUpdate: map[string]interface{}{},
		cacheKeysToDelete:  map[string]interface{}{},
		queryInfosMtx:      &sync.RWMutex{},
		queryInfos:         make([]*queryinfo.QueryInfo, 0, 10),
		SR:                 s.SR,
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
		s.SR.LogStats(resp.code, req.Method, lowerPath, ctx.requestStartUnixMillis, getQueryInfos(ctx))
	}()
	//must make sure to close the request body
	if req != nil && req.Body != nil {
		defer req.Body.Close()
	}
	//check for special case of api docs first
	if req.Method == cnst.GET && lowerPath == s.SR.ApiDocsRoute {
		writeRawJson(resp, 200, s.SR.ApiDocs)
		return
	}
	//check for special case of api mget
	if req.Method == cnst.GET && lowerPath == s.SR.ApiMGetRoute {
		reqs := map[string]string{}
		err.PanicIf(json.Unmarshal([]byte(req.URL.Query().Get("args")), &reqs))
		responseChan := make(chan *mgetResponse)
		bodyOnly := strings.ToLower(req.URL.Query().Get("format")) == "bodyonly"
		for key, reqUrl := range reqs {
			go func(key, reqUrl string) {
				r, _ := http.NewRequest(cnst.GET, reqUrl, nil)
				for _, c := range req.Cookies() {
					r.AddCookie(c)
				}
				w := &mgetResponseWriter{header: http.Header{}, body: bytes.NewBuffer(make([]byte, 0, 1000))}
				s.ServeHTTP(w, r)
				responseChan <- &mgetResponse{
					BodyOnly: bodyOnly,
					Key: key,
					Code: w.code,
					Header: w.header,
					Body: w.body.Bytes(),
				}
			}(key, reqUrl)
		}
		timeoutChan := time.After(s.SR.ApiMGetTimeout)
		timedOut := false
		fullMGetResponse := map[string]*mgetResponse{}
		for !(len(reqs) == len(fullMGetResponse) || timedOut) {
			select {
			case resp := <-responseChan:
				fullMGetResponse[resp.Key] = resp
			case <-timeoutChan:
				timedOut = true
			}
		}

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
		ctx.session, _ = s.SR.SessionStore.Get(req, s.SR.SessionCookieName)
		if ctx.session != nil {
			iMe := ctx.session.Values["me"]
			if iMe != nil {
				me := iMe.(id.Id)
				ctx.me = &me
			}
		}
		//check for valid me value if endpoint requires active session, and check for X header in POST requests for CSRF prevention
		if ep.RequiresSession && ctx.me == nil || req.Method == cnst.POST && req.Header.Get("X-Client") == "" {
			writeJson(resp, http.StatusUnauthorized, unauthorizedErr)
			return
		}
		//setup _ctx
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
			err.FmtPanic("private endpoints don't support POST Form data")
		}
		args = ep.ProcessForm(resp, req)
	}
	//process private ts and key args
	if ep.IsPrivate {
		ts, e := strconv.ParseInt(reqQueryValues.Get("ts"), 10, 64)
		err.PanicIf(e)
		//if the timestamp the req was sent is over a minute ago, reject the request
		if t.NowUnixMillis()-ts > 60000 {
			err.FmtPanic("suspicious private request sent over a minute ago")
		}
		key, e := base64.RawURLEncoding.DecodeString(reqQueryValues.Get("_"))
		err.PanicIf(e)
		// check the args/timestamp/key are valid
		if !bytes.Equal(key, ep.PrivateKeyGen(argsBytes, reqQueryValues.Get("ts"))) {
			err.FmtPanic("invalid private request keys don't match")
		}
		//check redis cache to ensure key has not appeared in the last minute, to prevent replay attacks
		cnn := s.SR.PrivateKeyRedisPool.Get()
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
		if me, ok := iId.(id.Id); !ok {
			err.FmtPanic("isAuthentication did not return id.Id type")
		} else {
			ctx.me = &me //set me on _ctx for logging info in defer above
			ctx.session.Values["me"] = me
			ctx.session.Values["AuthedOn"] = t.NowUnixMillis()
			ctx.session.Save(req, resp)
			writeJsonOk(ctx.resp, me)
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

type mgetResponseWriter struct{
	code int
	header http.Header
	body *bytes.Buffer
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


type mgetResponse struct{
	BodyOnly bool       `json:"bodyOnly"`
	Key  string         `json:"key"`
	Code int 			`json:"code"`
	Header http.Header  `json:"header"`
	Body []byte			`json:"body"`
}

func (r *mgetResponse) MarshalJSON() ([]byte, error) {
	if !r.BodyOnly {
		h, _ := json.Marshal(r.Header)
		return []byte(fmt.Sprintf(`{"key":%q,"code":%d,"header":%s,"body":%s}`, r.Key, r.Code, h, r.Body)), nil
	} else {
		return r.Body, nil
	}
}
