package misc

import (
	"errors"
	. "github.com/pborman/uuid"
)

var (
	idGenerationErr = errors.New("Failed to generate id")
)

type Entity struct {
	Id UUID `json:"id"`
}

type GenNewId func() (UUID, error)

//returns version 1 uuid as a byte slice
func NewId() (UUID, error) {
	id := NewUUID()
	if id == nil {
		return nil, idGenerationErr
	}
	return id, nil
}
