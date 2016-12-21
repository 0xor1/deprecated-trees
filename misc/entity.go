package misc

import (
	"errors"
	"github.com/pborman/uuid"
	"github.com/uber-go/zap"
)

var (
	IdGenerationErr = errors.New("Failed to generate id")
)

type Entity struct {
	Id uuid.UUID `json:"id"`
}

type CentralEntity struct {
	Entity
	Region string `json:"region"`
	Shard  int    `json:"shard"`
}

//returns version 1 uuid as a byte slice
func NewId() (uuid.UUID, error) {
	id := uuid.NewUUID()
	if id == nil {
		return nil, IdGenerationErr
	}
	return id, nil
}

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
