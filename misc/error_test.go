package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NilCriticalParamPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.Equal(t, "code: 0, msg: nil yo", err.Error())
	}()
	NilCriticalParamPanic("yo")
}

func Test_Error(t *testing.T) {
	e := &Error{Code: 123, Msg: "yo ho ho"}
	assert.Equal(t, "code: 123, msg: yo ho ho", e.Error())
}

func Test_ErrorRef(t *testing.T) {
	id, _ := NewId()
	e := &ErrorRef{Id: id}
	assert.Equal(t, "errorRef: "+id.String(), e.Error())
}
