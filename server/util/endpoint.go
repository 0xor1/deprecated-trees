package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Endpoint struct {
	Method            string
	Path              string
	RequiresSession   bool
	ResponseStructure interface{}
	IsAuthentication  bool
	PermissionDlmKeys func(ctx Ctx, args interface{}) []string
	ValueDlmKeys      func(ctx Ctx, args interface{}) []string
	ProcessForm       func(http.ResponseWriter, *http.Request) interface{}
	GetArgsStruct     func() interface{}
	PermissionCheck   func(ctx Ctx, args interface{})
	CentralHandler    func(ctx CentralCtx, args interface{}) interface{}
	RegionalHandler   func(ctx RegionalCtx, args interface{}) interface{}
}

func (e *Endpoint) Validate() {
	if (e.Method != GET && e.Method != POST) || (e.ProcessForm != nil && e.Method != POST) || ((e.CentralHandler == nil && e.RegionalHandler == nil) || (e.CentralHandler != nil && e.RegionalHandler != nil)) {
		invalidEndpointErr.Panic()
	}
}

var (
	queryString = "Query_String"
	body        = "Body"
	form        = "Form"
)

func (e *Endpoint) Documentation() *endpointDocumentation {
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
	return &endpointDocumentation{
		Method:            e.Method,
		Path:              e.Path,
		RequiresSession:   e.RequiresSession,
		ArgsLocation:      argsLocation,
		ArgsStructure:     e.GetArgsStruct(),
		ResponseStructure: e.ResponseStructure,
		IsAuthentication:  e.IsAuthentication,
	}
}

func (e *Endpoint) CreateRequest(schema, host string, args interface{}, buildForm func(r *http.Request, args interface{})) (*http.Request, error) {
	req := &http.Request{
		Method: e.Method,
		URL: &url.URL{
			Scheme: schema,
			Path:   e.Path,
		},
		Header: http.Header{},
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

func (e *Endpoint) DoRequest(schema, host string, args interface{}, buildForm func(r *http.Request, args interface{}), respVal interface{}) (interface{}, error) {
	req, err := e.CreateRequest(schema, host, args, buildForm)
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

type endpointDocumentation struct {
	Method            string      `json:"method"`
	Path              string      `json:"path"`
	RequiresSession   bool        `json:"requiresSession"`
	ArgsLocation      *string     `json:"argsLocation,omitempty"`
	ArgsStructure     interface{} `json:"argsStructure,omitempty"`
	ResponseStructure interface{} `json:"responseStructure,omitempty"`
	IsAuthentication  bool        `json:"isAuthentication"`
}
