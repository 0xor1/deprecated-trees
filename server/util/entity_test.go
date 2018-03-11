package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewId(t *testing.T) {
	id := NewId()
	assert.IsType(t, Id{}, id)
}

func Test_IdEqual(t *testing.T) {
	id1 := NewId()
	id2 := NewId()
	assert.True(t, id1.Equal(id1))
	assert.False(t, id1.Equal(id2))
}

func Test_IdCopy(t *testing.T) {
	id1 := NewId()
	id2 := id1.Copy()
	assert.True(t, id1.Equal(id2))
	tmp := id2[0]
	id2[0] = id2[1]
	id2[1] = tmp
	assert.False(t, id1.Equal(id2))
}
