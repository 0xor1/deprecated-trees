package misc

import (
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewId(t *testing.T) {
	id, err := NewId()
	assert.IsType(t, uuid.UUID{}, id)
	assert.Nil(t, err)
}
