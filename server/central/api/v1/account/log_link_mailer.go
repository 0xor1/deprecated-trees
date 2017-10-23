package account

import (
	"fmt"
)

type logLinkMailer struct {
}

func (l *logLinkMailer) sendMultipleAccountPolicyEmail(address string) {
	fmt.Printf("sendMultipleAccountPolicyEmail:\n\taddress: %s\n", address)
}

func (l *logLinkMailer) sendActivationLink(address, activationCode string) {
	fmt.Printf("sendActivationLink:\n\taddress: %s\n\tactivationCode: %s\n", address, activationCode)
}

func (l *logLinkMailer) sendPwdResetLink(address, resetCode string) {
	fmt.Printf("sendPwdResetLink:\n\taddress: %s\n\tresetCode: %s\n", address, resetCode)
}

func (l *logLinkMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) {
	fmt.Printf("sendNewEmailConfirmationLink:\n\tcurrentAddress: %s\n\tnewAddress: %s\n\tconfirmationCode: %s\n", currentAddress, newAddress, confirmationCode)
}
