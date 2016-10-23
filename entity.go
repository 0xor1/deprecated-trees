package core

import (
	"errors"
	"github.com/pborman/uuid"
	"strings"
)

var (
	IdGenerationErr = errors.New("Failed to generate id")
)

type Entity struct {
	Id string `json:"id"`
}

//returns version 1 uuid string without hyphens
func NewId() (string, error) {
	id := uuid.NewUUID()
	if id == nil {
		return "", IdGenerationErr
	}
	return strings.Replace(id.String(), "-", "", -1), nil
}
