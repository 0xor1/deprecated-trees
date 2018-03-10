package util

import (
	"bytes"
	"encoding/base64"
	"github.com/oklog/ulid"
	"time"
	"sync"
	"math/rand"
)

func Now() time.Time {
	return time.Now().UTC()
}

func NowUnixMillis() int64 {
	return Now().UnixNano() / 1000000
}

var(
	entropyMtx = &sync.Mutex{}
	entropy = rand.New(rand.NewSource(NowUnixMillis()))
)

//returns ulid as a byte slice
func NewId() Id {
	entropyMtx.Lock() //rand source is not safe for concurrent use :(
	defer entropyMtx.Unlock()
	id := ulid.MustNew(ulid.Timestamp(Now()), entropy)
	return Id(id[:])
}

func ParseId(id string) Id {
	bytes, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil || len(bytes) != 16 {
		idParseErr.Panic()
	}
	return Id(bytes)
}

type Id []byte

func (id Id) MarshalJSON() ([]byte, error) {
	return []byte(`"`+id.String()+`"`), nil
}

func (id *Id) UnmarshalJSON(data []byte) error {
	*id = ParseId(string(bytes.Trim(data, `"`)))
	return nil
}

func (id Id) String() string {
	return base64.RawURLEncoding.EncodeToString(id)
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id, other)
}

func (id Id) GreaterThanOrEqualTo(other Id) bool {
	return bytes.Compare(id, other) > -1
}

func (id Id) Copy() Id {
	return Id(append(make([]byte, 0, 16), id...))
}

type AddMemberPrivate struct {
	Id          Id          `json:"id"`
	Name        string      `json:"name"`
	DisplayName *string     `json:"displayName"`
	Role        AccountRole `json:"role"`
}

type AddMemberPublic struct {
	Id   Id          `json:"id"`
	Role AccountRole `json:"role"`
}

type AddProjectMember struct {
	Id   Id          `json:"id"`
	Role ProjectRole `json:"role"`
}

type Activity struct {
	OccurredOn time.Time `json:"occurredOn"`
	Member     Id        `json:"member"`
	Item       Id        `json:"item"`
	ItemType   string    `json:"itemType"`
	Action     string    `json:"action"`
	ItemName   *string   `json:"itemName,omitempty"`
	ExtraInfo  *string   `json:"extraInfo,omitempty"`
}
