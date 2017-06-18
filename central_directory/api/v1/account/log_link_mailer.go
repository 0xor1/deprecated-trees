package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/uber-go/zap"
)

type logLinkMailer struct {
	log Log
}

func (l *logLinkMailer) sendMultipleAccountPolicyEmail(address string) {
	l.log.Info(zap.String("address", address))
}

func (l *logLinkMailer) sendActivationLink(address, activationCode string) {
	l.log.Info(zap.String("address", address), zap.String("activationCode", activationCode))
}

func (l *logLinkMailer) sendPwdResetLink(address, resetCode string) {
	l.log.Info(zap.String("address", address), zap.String("resetCode", resetCode))
}

func (l *logLinkMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) {
	l.log.Info(zap.String("currentAddress", currentAddress), zap.String("newAddress", newAddress), zap.String("confirmationCode", confirmationCode))
}
