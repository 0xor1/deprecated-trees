package id

import (
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/time"
	"bytes"
	"encoding/base64"
	"github.com/oklog/ulid"
	"math/rand"
	"net/http"
	"sync"
)

var (
	entropyMtx = &sync.Mutex{}
	entropy    = rand.New(rand.NewSource(time.NowUnixMillis()))
)

//returns ulid as a byte slice
func New() Id {
	entropyMtx.Lock() //rand source is not safe for concurrent use :(
	defer entropyMtx.Unlock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return Id(id[:])
}

func Parse(id string) Id {
	b, e := base64.RawURLEncoding.DecodeString(id)
	err.HttpPanicf(e != nil || len(b) != 16, http.StatusBadRequest, "invalid id")
	return Id(b)
}

type Id []byte

func (id Id) MarshalJSON() ([]byte, error) {
	return []byte(`"` + id.String() + `"`), nil
}

func (id *Id) UnmarshalJSON(data []byte) error {
	*id = Parse(string(bytes.Trim(data, `"`)))
	return nil
}

func (id Id) String() string {
	return base64.RawURLEncoding.EncodeToString(id)
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id, other)
}

func (id Id) Copy() Id {
	return Id(append(make([]byte, 0, 16), id...))
}

type Identifiable interface {
	Id() Id
}
