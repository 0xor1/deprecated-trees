package endpoint

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/0xor1/panic"
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/ctx"
	"github.com/0xor1/trees/server/util/time"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	queryString = "Query_String"
	body        = "Body"
	form        = "Form"
)

type Endpoint struct {
	Note                     string
	Path                     string
	IsPrivate                bool
	RequiresSession          bool
	ExampleResponseStructure interface{}
	IsAuthentication         bool
	FormStruct               map[string]string
	ProcessForm              func(http.ResponseWriter, *http.Request) interface{}
	GetArgsStruct            func() interface{}
	CtxHandler               func(ctx ctx.Ctx, args interface{}) interface{}
	PrivateKeyGen            func(argsBytes []byte, ts string) []byte
}

func (ep *Endpoint) ValidateEndpoint() {
	panic.If((ep.ProcessForm != nil && ep.IsPrivate) || // if processForm is passed it must not be a private call, private endpoints dont support forms
		(ep.ProcessForm != nil && len(ep.FormStruct) == 0) || // if processForm is passed FormStruct must be given for documentation
		ep.CtxHandler == nil, // every endpoint needs a handler
		"invalid endpoint")
}

func (ep *Endpoint) GetEndpointDocumentation() *endpointDocumentation {
	var argsLocation *string
	var note *string
	if ep.Note != "" {
		note = &ep.Note
	}
	var argsStruct interface{}
	if ep.GetArgsStruct != nil {
		argsStruct = ep.GetArgsStruct()
		argsLocation = &body
	} else if ep.ProcessForm != nil {
		argsStruct = ep.FormStruct
		argsLocation = &form
	}
	var isAuth *bool
	if ep.IsAuthentication {
		isAuth = &ep.IsAuthentication
	}
	return &endpointDocumentation{
		Note:                     note,
		Method:                   http.MethodPost,
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

func (ep *Endpoint) createRequest(baseURl string, region cnst.Region, args interface{}, buildForm func() (io.ReadCloser, string)) (*http.Request, error) {
	reqUrl, e := url.Parse(baseURl + ep.Path)
	if e != nil {
		return nil, e
	}
	urlVals := url.Values{}
	urlVals.Set("region", string(region))
	var body io.ReadCloser
	var contentType string
	if buildForm != nil {
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
		body = ioutil.NopCloser(bytes.NewBuffer(argsBytes))
	}
	reqUrl.RawQuery = urlVals.Encode()
	req, e := http.NewRequest(http.MethodPost, reqUrl.String(), body)
	if e != nil {
		return nil, e
	}
	req.Header.Add("X-Client", "go-client")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req, nil
}

func (ep *Endpoint) DoRequest(css *clientsession.Store, baseUrl string, region cnst.Region, args interface{}, buildForm func() (io.ReadCloser, string), respVal interface{}) (interface{}, error) {
	region.Validate()
	req, e := ep.createRequest(baseUrl, region, args, buildForm)
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
	if resp.StatusCode == http.StatusOK {
		if respVal != nil {
			e = json.NewDecoder(resp.Body).Decode(respVal)
			if e != nil {
				return nil, e
			}
		}
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(string(bodyBytes))
	}
	return respVal, nil
}
