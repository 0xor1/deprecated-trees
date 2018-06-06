package session

import (
	"net/http"
	"github.com/gorilla/securecookie"
	"bitbucket.org/0xor1/trees/server/util/id"
	"encoding/gob"
	"github.com/0xor1/panic"
)

var HeaderName = "X-Session"

type Store interface{
	Get(r *http.Request) (*Session)
	Save(w http.ResponseWriter, s *Session)
}

type Session struct{
	Me id.Id
	AuthedOn int64
}

func New(keyPairs ...[]byte) *headerStore {
	gob.Register(&Session{})
	return &headerStore{
		codecs: securecookie.CodecsFromPairs(keyPairs...),
	}
}

//wrapper around gorilla secure cookies codecs
type headerStore struct {
	codecs []securecookie.Codec
}

func (hs *headerStore) Get(r *http.Request) (*Session) {
	value := r.Header.Get(HeaderName)
	if value == "" {
		return nil
	}
	s := &Session{}
	panic.If(securecookie.DecodeMulti(HeaderName, value, s, hs.codecs...))
	return s
}

func (hs *headerStore) Save(w http.ResponseWriter, s *Session) {
	value, e := securecookie.EncodeMulti(HeaderName, s, hs.codecs...)
	panic.If(e)
	w.Header().Set(HeaderName, value)
}
