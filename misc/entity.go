package misc

import (
	"errors"
	. "github.com/pborman/uuid"
	"encoding/hex"
)

var (
	idGenerationErr = errors.New("Failed to generate id")
)

type Id UUID

func (id Id) String() string {
	return hex.EncodeToString(id)
}

type Entity struct {
	Id Id `json:"id"`
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
