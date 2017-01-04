package account

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"bitbucket.org/robsix/task_center/misc"
)

func Test_NewLogLinkMailer_nilLogErr(t *testing.T) {
	linkMailer, err := NewLogLinkMailer(nil)

	assert.Nil(t, linkMailer)
	assert.Equal(t, err, nilLogErr)
}

func Test_NewLogLinkMailer_success(t *testing.T) {
	linkMailer, err := NewLogLinkMailer(misc.NewLog(nil))

	assert.NotNil(t, linkMailer)
	assert.Nil(t, err)
}