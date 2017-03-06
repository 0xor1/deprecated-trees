package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_logLinkMailer_all_methods(t *testing.T) {
	linkMailer := NewLogLinkMailer(NewLog(nil))

	err := linkMailer.sendMultipleAccountPolicyEmail("1")
	assert.Nil(t, err)
	err = linkMailer.sendActivationLink("1", "2")
	assert.Nil(t, err)
	err = linkMailer.sendPwdResetLink("1", "2")
	assert.Nil(t, err)
	err = linkMailer.sendNewEmailConfirmationLink("1", "2", "3")
	assert.Nil(t, err)
}
