package misc

import (
	"errors"
	. "github.com/pborman/uuid"
)

var (
	IdGenerationErr = errors.New("Failed to generate id")
)

type Entity struct {
	Id UUID `json:"id"`
}

type CentralEntity struct {
	Entity
	Region string `json:"region"`
	Shard  int    `json:"shard"`
}

//returns version 1 uuid as a byte slice
func NewId() (UUID, error) {
	id := NewUUID()
	if id == nil {
		return nil, IdGenerationErr
	}
	return id, nil
}
