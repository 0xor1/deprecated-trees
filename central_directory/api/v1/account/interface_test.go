package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewSqlApi_success(t *testing.T) {
	api := NewSqlApi(&mockInternalRegionClient{}, &mockLinkMailer{}, nil, nil, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, &isql.MockDB{}, &isql.MockDB{}, NewLog(nil))
	assert.NotNil(t, api)
}

func Test_NewLogLinkMailer_nilLogPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	NewLogLinkMailer(nil)
}

func Test_NewLogLinkMailer_success(t *testing.T) {
	linkMailer := NewLogLinkMailer(NewLog(nil))

	assert.NotNil(t, linkMailer)
}

func Test_NewEmailLinkMailer_notImplementedPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.Equal(t, NotImplementedErr, err)
	}()
	NewEmailLinkMailer()
}
