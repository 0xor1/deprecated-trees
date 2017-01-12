package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewId(t *testing.T) {
	id, err := NewId()
	assert.IsType(t, Id{}, id)
	assert.Nil(t, err)
}
