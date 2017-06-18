package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewApi_success(t *testing.T) {
	api := NewApi(&mockInternalRegionClient{}, &mockLinkMailer{}, &mockAvatarStore{}, nil, nil, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, &isql.MockDB{}, &isql.MockDB{})
	assert.NotNil(t, api)
}

func Test_NewLogLinkMailer_nilCriticalParamErr(t *testing.T) {
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
	NewSparkPostLinkMailer()
}
