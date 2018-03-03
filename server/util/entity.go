package util

import (
	"bytes"
	"encoding/base64"
	. "github.com/pborman/uuid"
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

//returns Version 1 uuid as a byte slice
func NewId() Id {
	id := NewUUID()
	if id == nil {
		idGenerationErr.Panic()
	}
	return Id(id)
}

func ParseId(id string) Id {
	bytes, err := base64.StdEncoding.DecodeString(id)
	if err != nil || len(bytes) != 16 {
		idParseErr.Panic()
	}
	return Id(bytes)
}

type Id UUID

func (id Id) String() string {
	return base64.StdEncoding.EncodeToString(id)
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id, other)
}

func (id Id) GreaterThanOrEqualTo(other Id) bool {
	return bytes.Compare(id, other) > -1
}

func (id Id) Copy() Id {
	return Id(append(make([]byte, 0, 16), []byte(id)...))
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
