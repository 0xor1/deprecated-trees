package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"testing"
)

func Test_logLinkMailer_all_methods(t *testing.T) {
	linkMailer := NewLogLinkMailer(NewLog(nil))

	linkMailer.sendMultipleAccountPolicyEmail("1")
	linkMailer.sendActivationLink("1", "2")
	linkMailer.sendPwdResetLink("1", "2")
	linkMailer.sendNewEmailConfirmationLink("1", "2", "3")
}
