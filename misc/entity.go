package misc

import (
	"bytes"
	"encoding/hex"
	. "github.com/pborman/uuid"
	"time"
)

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

type Entity struct {
	Id Id `json:"id"`
}

type GenEntity func() (*Entity, error)

func NewEntity() *Entity {
	return &Entity{
		Id: NewId(),
	}
}

type NamedEntity struct {
	Entity
	Name string `json:"name"`
}

type AddMemberInternal struct {
	NamedEntity
	Role AccountRole `json:"role"`
}

type AddMemberExternal struct {
	Entity
	Role AccountRole `json:"role"`
}

type GenNamedEntity func(name string) *NamedEntity

func NewNamedEntity(name string) *NamedEntity {
	return &NamedEntity{
		Entity: Entity{
			Id: NewId(),
		},
		Name: name,
	}
}

type CreatedNamedEntity struct {
	NamedEntity
	CreatedOn time.Time `json:"createdOn"`
}

type GenCreatedNamedEntity func(name string) *CreatedNamedEntity

func NewCreatedNamedEntity(name string) *CreatedNamedEntity {
	return &CreatedNamedEntity{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: NewId(),
			},
			Name: name,
		},
		CreatedOn: Now(),
	}
}

type GenId func() Id

//returns version 1 uuid as a byte slice
func NewId() Id {
	id := NewUUID()
	if id == nil {
		panic(idGenerationErr)
	}
	return Id(id)
}

type CommonTimeProps struct{
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
}

type CommonNodeProps struct {
	CreatedNamedEntity
	CommonTimeProps
	Description        string `json:"description"`
	LinkedFilesCount   uint64 `json:"linkedFilesCount"`
	ChatCount          uint64 `json:"chatCount"`
}

type CommonAbstractNodeProps struct {
	MinimumRemainingTime uint64 `json:"minimumRemainingTime"`
	IsParallel           bool   `json:"isParallel"`
}

type AccountMember struct {
	AddMemberInternal
	IsActive           bool   `json:"isActive"`
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
