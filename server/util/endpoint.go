package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Endpoint struct {
	Note              string
	Method            string
	Path              string
	IsPrivate         bool
	RequiresSession   bool
	ResponseStructure interface{}
	IsAuthentication  bool
	PermissionDlmKeys func(ctx *Ctx, args interface{}) []string
	ValueDlmKeys      func(ctx *Ctx, args interface{}) []string
	FormStruct        map[string]string
	ProcessForm       func(http.ResponseWriter, *http.Request) interface{}
	GetArgsStruct     func() interface{}
	PermissionCheck   func(ctx *Ctx, args interface{})
	CtxHandler        func(ctx *Ctx, args interface{}) interface{}
	RawHandler        func(w http.ResponseWriter, r *http.Request)
}

func (e *Endpoint) ValidateEndpoint() {
	if (e.Method != GET && e.Method != POST) || (e.ProcessForm != nil && e.Method != POST) || (e.ProcessForm != nil && len(e.FormStruct) == 0) || (e.CtxHandler != nil && e.RawHandler != nil) || (e.CtxHandler == nil && e.RawHandler == nil) {
		invalidEndpointErr.Panic()
	}
}

var (
	queryString = "Query_String"
	body        = "Body"
	form        = "Form"
)

func (e *Endpoint) GetEndpointDocumentation() *endpointDocumentation {
	var argsLocation *string
	if e.GetArgsStruct != nil {
		argsLocation = &queryString
		if e.Method == POST {
			if e.ProcessForm != nil {
				argsLocation = &form
			} else {
				argsLocation = &body
			}
		}
	}
	var note *string
	if e.Note != "" {
		note = &e.Note
	}
	var argsStruct interface{}
	if e.GetArgsStruct != nil {
		argsStruct = e.GetArgsStruct()
	} else if e.ProcessForm != nil {
		argsStruct = e.FormStruct
	}
	var isAuth *bool
	if e.IsAuthentication {
		isAuth = &e.IsAuthentication
	}
	return &endpointDocumentation{
		Note:              note,
		Method:            e.Method,
		Path:              e.Path,
		RequiresSession:   e.RequiresSession,
		ArgsLocation:      argsLocation,
		ArgsStructure:     argsStruct,
		ResponseStructure: e.ResponseStructure,
		IsAuthentication:  isAuth,
	}
}

func (e *Endpoint) createRequest(host string, args interface{}, buildForm func(r *http.Request, args interface{})) (*http.Request, error) {
	req, err := http.NewRequest(e.Method, host+e.Path, nil)
	if err != nil {
		return nil, err
	}
	if buildForm != nil {
		if e.Method != POST {
			return nil, invalidEndpointErr
		}
		buildForm(req, args)
	} else if args != nil {
		argsBytes, err := json.Marshal(args)
		if err != nil {
			return nil, err
		}
		if e.Method == GET {
			req.URL.RawQuery = url.Values{"args": []string{string(argsBytes)}}.Encode()
		} else if e.Method == POST {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(argsBytes))
		} else {
			return nil, invalidEndpointErr
		}

	}
	return req, nil
}

func (e *Endpoint) DoRequest(host string, args interface{}, buildForm func(r *http.Request, args interface{}), respVal interface{}) (interface{}, error) {
	req, err := e.createRequest(host, args, buildForm)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if respVal != nil {
		err = json.NewDecoder(resp.Body).Decode(respVal)
		if err != nil {
			return nil, err
		}
	}
	return respVal, nil
}

func (e *Endpoint) handleRequest(ctx *Ctx) {
	var args interface{}
	if e.Method == GET && e.GetArgsStruct != nil {
		args = e.GetArgsStruct()
		PanicIf(json.Unmarshal([]byte(ctx.r.URL.Query().Get("args")), args))
	} else if e.Method == POST && e.GetArgsStruct != nil {
		args = e.GetArgsStruct()
		PanicIf(json.NewDecoder(ctx.r.Body).Decode(args))
	} else if e.Method == POST && e.ProcessForm != nil {
		args = e.ProcessForm(ctx.w, ctx.r)
	}
	if e.CtxHandler != nil {
		e.CtxHandler(ctx, args)
	} else {
		e.RawHandler(ctx.w, ctx.r)
	}
}

type endpointDocumentation struct {
	Note              *string     `json:"note,omitempty"`
	Method            string      `json:"method"`
	Path              string      `json:"path"`
	RequiresSession   bool        `json:"requiresSession"`
	ArgsLocation      *string     `json:"argsLocation,omitempty"`
	ArgsStructure     interface{} `json:"argsStructure,omitempty"`
	ResponseStructure interface{} `json:"responseStructure,omitempty"`
	IsAuthentication  *bool       `json:"isAuthentication,omitempty"`
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
