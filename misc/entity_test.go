package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func Test_NewEntity(t *testing.T) {
	e := NewEntity()
	assert.NotNil(t, e.Id)
}

func Test_NewNamedEntity(t *testing.T) {
	e := NewNamedEntity("ali")
	assert.NotNil(t, e.Id)
	assert.Equal(t, "ali", e.Name)
}

func Test_NewCreatedNamedEntity(t *testing.T) {
	e := NewCreatedNamedEntity("ali")
	now := time.Now().UTC()
	assert.NotNil(t, e.Id)
	assert.True(t, now.Add(-1*time.Millisecond).Before(e.CreatedOn) && now.Add(1*time.Millisecond).After(e.CreatedOn))
	assert.Equal(t, "ali", e.Name)
}
