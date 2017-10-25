package misc

import (
	"bytes"
	"encoding/hex"
	. "github.com/pborman/uuid"
	"time"
)

//returns version 1 uuid as a byte slice
func NewId() Id {
	id := NewUUID()
	if id == nil {
		idGenerationErr.Panic()
	}
	return Id(id)
}

type Id UUID

func (id Id) String() string {
	return hex.EncodeToString(id)
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

func Now() time.Time {
	return time.Now().UTC()
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

type Activity struct {
	OccurredOn time.Time `json:"occurredOn"`
	Member     Id        `json:"member"`
	Item       Id        `json:"item"`
	ItemType   string    `json:"itemType"`
	ItemName   string    `json:"itemName"`
	Action     string    `json:"action"`
	NewValue   *string   `json:"newValue,omitempty"`
}
