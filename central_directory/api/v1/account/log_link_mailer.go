package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/uber-go/zap"
)

func newLogLinkMailer(log Log) linkMailer {
	if log == nil {
		panic(nilLogErr)
	}
	return &logLinkMailer{
		log: log,
	}
}

type logLinkMailer struct {
	log Log
}

func (l *logLinkMailer) sendMultipleAccountPolicyEmail(address string) error {
	l.log.Info(zap.String("address", address))
	return nil
}

func (l *logLinkMailer) sendActivationLink(address, activationCode string) error {
	l.log.Info(zap.String("address", address), zap.String("activationCode", activationCode))
	return nil
}

func (l *logLinkMailer) sendPwdResetLink(address, resetCode string) error {
	l.log.Info(zap.String("address", address), zap.String("resetCode", resetCode))
	return nil
}

func (l *logLinkMailer) sendNewEmailConfirmationLink(currentAddress, newAddress, confirmationCode string) error {
	l.log.Info(zap.String("currentAddress", currentAddress), zap.String("newAddress", newAddress), zap.String("confirmationCode", confirmationCode))
	return nil
}
