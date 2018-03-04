package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	StaticResources   *StaticResources
}

func (e *Endpoint) ValidateEndpoint() {
	if (e.Method != GET && e.Method != POST) || (e.ProcessForm != nil && e.Method != POST) || (e.ProcessForm != nil && len(e.FormStruct) == 0) || e.CtxHandler == nil  {
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
		if e.IsPrivate {
			InvalidOperationErr.Panic() //private endpoints dont support sending form data
		}
	} else if args != nil {
		argsBytes, err := json.Marshal(args)
		if err != nil {
			return nil, err
		}
		if e.IsPrivate {
			e.addPrivateArgsToReqQuery(req, argsBytes)
		}
		if e.Method == GET {
			req.URL.Query().Set("args", string(argsBytes))
		} else if e.Method == POST {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(argsBytes))
		} else {
			return nil, invalidEndpointErr
		}

	}
	return req, nil
}

func (e *Endpoint) addPrivateArgsToReqQuery(r *http.Request, argsBytes []byte) {
	ts := fmt.Sprintf("%d", NowUnixMillis())
	key := ScryptKey(append(argsBytes, []byte(ts)...), e.StaticResources.RegionalV1PrivateClientSecret, e.StaticResources.ScryptN, e.StaticResources.ScryptR, e.StaticResources.ScryptP, e.StaticResources.ScryptKeyLen)
	r.URL.Query().Set("_", base64.URLEncoding.EncodeToString(key))
	r.URL.Query().Set("ts", ts)
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
