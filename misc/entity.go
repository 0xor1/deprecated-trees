package misc

import (
	"bytes"
	"encoding/hex"
	. "github.com/pborman/uuid"
)

type Id UUID

func (id Id) String() string {
	return hex.EncodeToString(id)
}

func (id Id) Equal(other Id) bool {
	return bytes.Equal(id, other)
	//return id.String() == other.String()
}

type Entity struct {
	Id Id `json:"id"`
}

type NamedEntity struct {
	Entity
	Name string `json:"name"`
}

type GenNewId func() (Id, error)

//returns version 1 uuid as a byte slice
func NewId() (Id, error) {
	id := NewUUID()
	if id == nil {
		return nil, idGenerationErr
	}
	return Id(id), nil
}
