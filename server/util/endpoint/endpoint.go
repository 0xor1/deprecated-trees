package endpoint

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/time"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	queryString        = "Query_String"
	body               = "Body"
	form               = "Form"
	invalidEndpointErr = &err.Err{Code: "u_s_ie", Message: "invalid endpoint"}
)

type Endpoint struct {
	Note                     string
	Method                   string
	Path                     string
	IsPrivate                bool
	RequiresSession          bool
	ExampleResponseStructure interface{}
	IsAuthentication         bool
	PermissionDlmKeys        func(ctx ctx.Ctx, args interface{}) []string
	ValueDlmKeys             func(ctx ctx.Ctx, args interface{}) []string
	FormStruct               map[string]string
	ProcessForm              func(http.ResponseWriter, *http.Request) interface{}
	GetArgsStruct            func() interface{}
	PermissionCheck          func(ctx ctx.Ctx, args interface{})
	CtxHandler               func(ctx ctx.Ctx, args interface{}) interface{}
	PrivateKeyGen            func(argsBytes []byte, ts string) []byte
}

func (ep *Endpoint) ValidateEndpoint() {
	if (ep.Method != cnst.GET && ep.Method != cnst.POST) || // only GET and POST methods supported for read and write operations respectively
		(ep.ProcessForm != nil && ep.Method != cnst.POST) || // if processForm is passed it must be a POST call
		(ep.ProcessForm != nil && ep.IsPrivate) || // if processForm is passed it must not be a private call, private endpoints dont support forms
		(ep.ProcessForm != nil && len(ep.FormStruct) == 0) || // if processForm is passed FormStruct must be given for documentation
		ep.CtxHandler == nil { // every endpoint needs a handler
		panic(invalidEndpointErr)
	}
}

func (ep *Endpoint) GetEndpointDocumentation() *endpointDocumentation {
	var argsLocation *string
	if ep.GetArgsStruct != nil {
		argsLocation = &queryString
		if ep.Method == cnst.POST {
			if ep.ProcessForm != nil {
				argsLocation = &form
			} else {
				argsLocation = &body
			}
		}
	}
	var note *string
	if ep.Note != "" {
		note = &ep.Note
	}
	var argsStruct interface{}
	if ep.GetArgsStruct != nil {
		argsStruct = ep.GetArgsStruct()
	} else if ep.ProcessForm != nil {
		argsStruct = ep.FormStruct
	}
	var isAuth *bool
	if ep.IsAuthentication {
		isAuth = &ep.IsAuthentication
	}
	return &endpointDocumentation{
		Note:                     note,
		Method:                   ep.Method,
		Path:                     ep.Path,
		RequiresSession:          ep.RequiresSession,
		ArgsLocation:             argsLocation,
		ArgsStructure:            argsStruct,
		ExampleResponseStructure: ep.ExampleResponseStructure,
		IsAuthentication:         isAuth,
	}
}

type endpointDocumentation struct {
	Note                     *string     `json:"note,omitempty"`
	Method                   string      `json:"method"`
	Path                     string      `json:"path"`
	RequiresSession          bool        `json:"requiresSession"`
	ArgsLocation             *string     `json:"argsLocation,omitempty"`
	ArgsStructure            interface{} `json:"argsStructure,omitempty"`
	ExampleResponseStructure interface{} `json:"exampleResponseStructure,omitempty"`
	IsAuthentication         *bool       `json:"isAuthentication,omitempty"`
}

func (ep *Endpoint) createRequest(host string, args interface{}, buildForm func() (io.ReadCloser, string)) (*http.Request, error) {
	reqUrl, e := url.Parse(host + ep.Path)
	if e != nil {
		return nil, e
	}
	urlVals := url.Values{}
	var body io.ReadCloser
	var contentType string
	if buildForm != nil {
		if ep.Method != cnst.POST {
			return nil, invalidEndpointErr
		}
		if ep.IsPrivate {
			return nil, err.InvalidOperation //private endpoints dont support sending form data
		}
		body, contentType = buildForm()
	} else if args != nil {
		argsBytes, e := json.Marshal(args)
		if e != nil {
			return nil, e
		}
		if ep.IsPrivate {
			ts := fmt.Sprintf("%d", time.NowUnixMillis())
			key := ep.PrivateKeyGen(argsBytes, ts)
			urlVals.Set("_", base64.RawURLEncoding.EncodeToString(key))
			urlVals.Set("ts", ts)
		}
		if ep.Method == cnst.GET {
			urlVals.Set("args", string(argsBytes))
		} else if ep.Method == cnst.POST {
			body = ioutil.NopCloser(bytes.NewBuffer(argsBytes))
		} else {
			return nil, invalidEndpointErr
		}
	}
	reqUrl.RawQuery = urlVals.Encode()
	req, e := http.NewRequest(ep.Method, reqUrl.String(), body)
	if e != nil {
		return nil, e
	}
	req.Header.Add("X-Client", "go-client")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func (ep *Endpoint) DoRequest(css *clientsession.Store, host string, args interface{}, buildForm func() (io.ReadCloser, string), respVal interface{}) (interface{}, error) {
	req, e := ep.createRequest(host, args, buildForm)
	if e != nil {
		return nil, e
	}
	if css != nil {
		for name, value := range css.Cookies {
			req.AddCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}
	resp, e := http.DefaultClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if css != nil {
		for _, cookie := range resp.Cookies() {
			css.Cookies[cookie.Name] = cookie.Value
		}
	}
	if e != nil {
		return nil, e
	}
	if respVal != nil {

		e = json.NewDecoder(resp.Body).Decode(respVal)
		if e != nil {
			return nil, e
		}
	}
	return respVal, nil
}
