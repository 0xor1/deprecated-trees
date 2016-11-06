package helper

import (
	"errors"
	"github.com/pborman/uuid"
	"github.com/uber-go/zap"
	"strings"
)

var (
	IdGenerationErr = errors.New("Failed to generate id")
)

type Entity struct {
	Id string `json:"id"`
}

//returns version 1 uuid string without hyphens
func NewId() (string, error) {
	id := uuid.NewUUID()
	if id == nil {
		return "", IdGenerationErr
	}
	return strings.Replace(id.String(), "-", "", -1), nil
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}

func NewLogLinkMailer(log zap.Logger) (LinkMailer, error) {
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
