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

type Entity struct {
	Id        Id        `json:"id"`
	CreatedOn time.Time `json:"createdOn"`
}

type GenEntity func() (*Entity, error)

func NewEntity() (*Entity, error) {
	id, err := NewId()
	if err != nil {
		return nil, err
	}
	return &Entity{
		Id:        id,
		CreatedOn: time.Now().UTC(),
	}, nil
}

type NamedEntity struct {
	Entity
	Name string `json:"name"`
}

type GenNamedEntity func(name string) (*NamedEntity, error)

func NewNamedEntity(name string) (*NamedEntity, error) {
	entity, err := NewEntity()
	if err != nil {
		return nil, err
	}
	return &NamedEntity{
		Entity: *entity,
		Name:   name,
	}, nil
}

type GenId func() (Id, error)

//returns version 1 uuid as a byte slice
func NewId() (Id, error) {
	id := NewUUID()
	if id == nil {
		return nil, idGenerationErr
	}
	return Id(id), nil
}
