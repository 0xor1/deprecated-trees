package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

const (
	sparkPostApiKey = "e38dfdcfead6c487718c658e67504102d25077e6"
)

type sparkPostMailer struct {
}

func (l *sparkPostMailer) sendMultipleAccountPolicyEmail(address string) {
	panic(NotImplementedErr)
}

func (l *sparkPostMailer) sendActivationLink(address, activationCode string) {
	panic(NotImplementedErr)
}

func (l *sparkPostMailer) sendPwdResetLink(address, resetCode string) {
	panic(NotImplementedErr)
}

func (l *sparkPostMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) {
	panic(NotImplementedErr)
}
