package misc

import (
	"github.com/uber-go/zap"
	"errors"
)

func NewLogLinkMailer(log zap.Logger) (*logLinkMailer, error) {
	if log == nil {
		return nil, errors.New("nil log error")
	}
	return &logLinkMailer{
		sendActivationLinkFn: func(a, c string){
			log.Info("logLinkMailer.SendActivationLink", zap.String("address", a), zap.String("activationCode", c))
		},
		sendPwdResetLinkFn: func(a, c string){
			log.Info("logLinkMailer.SendPwdResetLink", zap.String("address", a), zap.String("resetCode", c))
		},
		sendNewEmailConfirmationLinkFn: func(a, c string){
			log.Info("logLinkMailer.SendNewEmailConfirmationLink", zap.String("address", a), zap.String("confirmationCode", c))
		},
	}, nil
}

type logLinkMailer struct {
	sendActivationLinkFn           func(a, c string)
	sendPwdResetLinkFn             func(a, c string)
	sendNewEmailConfirmationLinkFn func(a, c string)
}

func (l *logLinkMailer) SendActivationLink(address, activationCode string) error {
	l.sendActivationLinkFn(address, activationCode)
	return nil
}

func (l *logLinkMailer) SendPwdResetLink(address, resetCode string) error {
	l.sendPwdResetLinkFn(address, resetCode)
	return nil
}

func (l *logLinkMailer) SendNewEmailConfirmationLink(address, confirmationCode string) error {
	l.sendNewEmailConfirmationLinkFn(address, confirmationCode)
	return nil
}
