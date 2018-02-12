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

func Test_IdGreaterThanOrEqualTo(t *testing.T) {
	id1 := NewId()
	id2 := NewId()
	assert.True(t, id2.GreaterThanOrEqualTo(id1))
}

func Test_IdCopy(t *testing.T) {
	id1 := NewId()
	id2 := id1.Copy()
	assert.True(t, id1.Equal(id2))
	tmp := []byte(id2)[0]
	([]byte(id2))[0] = []byte(id2)[1]
	([]byte(id2))[1] = tmp
	assert.False(t, id1.Equal(id2))
}
