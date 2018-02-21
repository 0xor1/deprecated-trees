package util

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"runtime/debug"
	"time"
)

var (
	unauthorizedErr   = &AppError{Code: "p_ns", Message: "unauthorized"}
	internalServerErr = &AppError{Code: "p_is", Message: "internal server err"}
)

type HttpHandler interface {
	Handle(handler func(s *session, args interface{}) interface{}, argStructGen func() interface{}, requiresSession bool) httprouter.Handle
}

// handles:
// * sessions
// * csrf token validation
// * parsing request args
// * writing responses
// * err logging
func NewHttpHandler(sessionKeyPairs []*AuthEncrKeyPair, sessionMaxAge int, sessionName string) HttpHandler {
	if len(sessionKeyPairs) == 0 || sessionMaxAge < 60 || sessionName == "" {
		InvalidArgumentsErr.Panic()
	}
	keys := make([][]byte, 0, len(sessionKeyPairs)*2)
	for _, skp := range sessionKeyPairs {
		skp.validate()
		keys = append(keys, skp.AuthKey64, skp.EncrKey32)
	}
	sessionStore := sessions.NewCookieStore(keys...)
	sessionStore.Options.Secure = true
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.MaxAge = sessionMaxAge
	//TODO register session types with gob
	return &httpHandler{
		sessionName:  sessionName,
		sessionStore: sessionStore,
	}
}

type httpHandler struct {
	sessionName  string
	sessionStore *sessions.CookieStore
}

func (h *httpHandler) Handle(handler func(s *session, args interface{}) interface{}, argStructGen func() interface{}, requiresSession bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Args) {
		defer func() {
			r := recover()
			if r != nil {
				pErr := r.(PermissionedError)
				if pErr != nil && pErr.IsPublic() {
					writeJson(w, http.StatusInternalServerError, pErr)
				} else {
					writeJson(w, http.StatusInternalServerError, internalServerErr)
				}
				fmt.Println(r)
				fmt.Println(string(debug.Stack()))
			}
		}()
		if r != nil && r.Body != nil {
			defer r.Body.Close()
		}
		var s *session
		cookieSession, _ := h.sessionStore.Get(r, h.sessionName)
		if cookieSession != nil {
			i := cookieSession.Values["a"]
			if i != nil {
				s = i.(*session)
			}
		}
		if requiresSession && (s == nil || r.Header.Get("X-Csrf-Token") != s.CsrfToken) {
			writeJson(w, http.StatusUnauthorized, unauthorizedErr)
			return
		}
		args := argStructGen()
		argsStr := r.URL.Query().Get("args")
		if argsStr != "" {
			PanicIf(json.Unmarshal([]byte(argsStr), args))
		} else { //check body
			PanicIf(json.NewDecoder(r.Body).Decode(args))
		}
		writeJson(w, http.StatusOK, handler(s, args))
	}
}

type AuthEncrKeyPair struct {
	AuthKey64 []byte
	EncrKey32 []byte
}

func (s *AuthEncrKeyPair) validate() {
	if len(s.AuthKey64) != 64 || len(s.EncrKey32) != 32 {
		InvalidArgumentsErr.Panic()
	}
}

func writeJsonOk(w http.ResponseWriter, body interface{}) {
	writeJson(w, 200, body)
}

func writeJson(w http.ResponseWriter, code int, body interface{}) {
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json;charset=utf-8")
	bodyBytes, err := json.Marshal(body)
	PanicIf(err)
	_, err = w.Write(bodyBytes)
	PanicIf(err)
}
