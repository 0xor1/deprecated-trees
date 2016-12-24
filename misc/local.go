package misc

import (
	"errors"
	"github.com/uber-go/zap"
)

func NewLogLinkMailer(log zap.Logger) (*logLinkMailer, error) {
	if log == nil {
		return nil, errors.New("nil log error")
	}
	return &logLinkMailer{
		log: log,
	}, nil
}

type logLinkMailer struct {
	log zap.Logger
}

func (l *logLinkMailer) SendActivationLink(address, activationCode string) error {
	l.log.Info("logLinkMailer.SendActivationLink", zap.String("address", address), zap.String("activationCode", activationCode))
	return nil
}

func (l *logLinkMailer) SendPwdResetLink(address, resetCode string) error {
	l.log.Info("logLinkMailer.SendPwdResetLink", zap.String("address", address), zap.String("resetCode", resetCode))
	return nil
}

func (l *logLinkMailer) SendNewEmailConfirmationLink(address, confirmationCode string) error {
	l.log.Info("logLinkMailer.SendNewEmailConfirmationLink", zap.String("address", address), zap.String("confirmationCode", confirmationCode))
	return nil
}
