package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewMemApi_success(t *testing.T) {
	api := NewMemApi(&mockInternalRegionApi{}, nil, nil, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, NewLog(nil))

	assert.NotNil(t, api)
}

func Test_NewSqlApi_notImplementedErrPanic(t *testing.T) {
	defer func(){
		err := recover().(error)
		assert.Equal(t, NotImplementedErr, err)
	}()
	NewSqlApi(&mockInternalRegionApi{}, nil, nil, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, NewLog(nil))
}
