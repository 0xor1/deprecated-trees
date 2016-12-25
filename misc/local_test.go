package misc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/uber-go/zap"
)

func Test_NewLogLinkMailer_NilLogErr(t *testing.T) {
	linkMailer, err:= NewLogLinkMailer(nil)
	assert.Nil(t, linkMailer)
	assert.NotNil(t, err)
}

func Test_NewLogLinkMailer_Success(t *testing.T) {
	linkMailer, err:= NewLogLinkMailer(zap.New(zap.NewTextEncoder()))
	assert.NotNil(t, linkMailer)
	assert.Nil(t, err)
}

func Test_LogLinkMailer_SendActivationLink(t *testing.T) {
	var address, code string
	linkMailer := &logLinkMailer{
		sendActivationLinkFn: func(a, c string){
			address = a
			code = c
		},
	}
	err := linkMailer.SendActivationLink("test_address", "test_code")
	assert.Equal(t, "test_address", address)
	assert.Equal(t, "test_code", code)
	assert.Nil(t, err)
}

func Test_LogLinkMailer_SendNewEmailConfirmationLink(t *testing.T) {
	var address, code string
	linkMailer := &logLinkMailer{
		sendPwdResetLinkFn: func(a, c string){
			address = a
			code = c
		},
	}
	err := linkMailer.SendPwdResetLink("test_address", "test_code")
	assert.Equal(t, "test_address", address)
	assert.Equal(t, "test_code", code)
	assert.Nil(t, err)
}

func Test_LogLinkMailer_SendPwdResetLink(t *testing.T) {
	var address, code string
	linkMailer := &logLinkMailer{
		sendNewEmailConfirmationLinkFn: func(a, c string){
			address = a
			code = c
		},
	}
	err := linkMailer.SendNewEmailConfirmationLink("test_address", "test_code")
	assert.Equal(t, "test_address", address)
	assert.Equal(t, "test_code", code)
	assert.Nil(t, err)
}
