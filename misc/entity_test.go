package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/pborman/uuid"
)

func Test_NewId(t *testing.T) {
	id, err := NewId()
	assert.IsType(t, uuid.UUID{}, id)
	assert.Nil(t, err)
}
