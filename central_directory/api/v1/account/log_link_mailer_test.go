package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newLogLinkMailer_nilLogPanic(t *testing.T) {
	defer func(){
		err := recover().(error)
		assert.Equal(t, err, nilLogErr)
	}()
	newLogLinkMailer(nil)
}

func Test_newLogLinkMailer_success(t *testing.T) {
	linkMailer := newLogLinkMailer(NewLog(nil))

	assert.NotNil(t, linkMailer)
}

func Test_logLinkMailer_all_methods(t *testing.T) {
	linkMailer := newLogLinkMailer(NewLog(nil))

	err := linkMailer.sendMultipleAccountPolicyEmail("1")
	assert.Nil(t, err)
	err = linkMailer.sendActivationLink("1", "2")
	assert.Nil(t, err)
	err = linkMailer.sendPwdResetLink("1", "2")
	assert.Nil(t, err)
	err = linkMailer.sendNewEmailConfirmationLink("1", "2", "3")
	assert.Nil(t, err)
}
