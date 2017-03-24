package account

import (

)

const(
	sparkPostApiKey = "e38dfdcfead6c487718c658e67504102d25077e6"
)

type sparkPostMailer struct {
}

func (l *sparkPostMailer) sendMultipleAccountPolicyEmail(address string) error {
	return nil
}

func (l *sparkPostMailer) sendActivationLink(address, activationCode string) error {
	return nil
}

func (l *sparkPostMailer) sendPwdResetLink(address, resetCode string) error {
	return nil
}

func (l *sparkPostMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) error {
	return nil
}
