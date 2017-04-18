package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func Test_IdGreaterThanOrEqualTo(t *testing.T) {
	id1, _ := NewId()
	id2, _ := NewId()
	assert.True(t, id2.GreaterThanOrEqualTo(id1))
}

func Test_IdCopy(t *testing.T) {
	id1, _ := NewId()
	id2 := id1.Copy()
	assert.True(t, id1.Equal(id2))
	tmp := []byte(id2)[0]
	([]byte(id2))[0] = []byte(id2)[1]
	([]byte(id2))[1] = tmp
	assert.False(t, id1.Equal(id2))
}

func Test_NewEntity(t *testing.T) {
	e, err := NewEntity()
	now := time.Now().UTC()
	assert.Nil(t, err)
	assert.NotNil(t, e.Id)
	assert.True(t, now.Add(-1*time.Millisecond).Before(e.CreatedOn) && now.Add(1*time.Millisecond).After(e.CreatedOn))
}

func Test_NewNamedEntity(t *testing.T) {
	e, err := NewNamedEntity("ali")
	now := time.Now().UTC()
	assert.Nil(t, err)
	assert.NotNil(t, e.Id)
	assert.True(t, now.Add(-1*time.Millisecond).Before(e.CreatedOn) && now.Add(1*time.Millisecond).After(e.CreatedOn))
	assert.Equal(t, "ali", e.Name)
}
