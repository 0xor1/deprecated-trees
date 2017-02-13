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

func Test_IdEqual(t *testing.T) {
	id1, _ := NewId()
	id2, _ := NewId()
	assert.True(t, id1.Equal(id1))
	assert.False(t, id1.Equal(id2))
}
