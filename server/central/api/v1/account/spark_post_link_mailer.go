package account

import (
	. "bitbucket.org/0xor1/task/server/util"
)

const (
	sparkPostApiKey = "e38dfdcfead6c487718c658e67504102d25077e6"
)

type sparkPostMailer struct {
}

func (l *sparkPostMailer) sendMultipleAccountPolicyEmail(address string) {
	NotImplementedErr.Panic()
}

func (l *sparkPostMailer) sendActivationLink(address, activationCode string) {
	NotImplementedErr.Panic()
}

func (l *sparkPostMailer) sendPwdResetLink(address, resetCode string) {
	NotImplementedErr.Panic()
}

func (l *sparkPostMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) {
	NotImplementedErr.Panic()
}
