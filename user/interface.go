package user

import (
	"bitbucket.org/robsix/core"
	"errors"
	"fmt"
)

var (
	IncorrectPwdErr          = errors.New("pwd incorrect")
	IncorrectUsersErr        = errors.New("incorrect users returned")
	UserNotActivated         = errors.New("user not activated")
	EmailAlreadyInUseErr     = errors.New("email already in use")
	InvalidEmailErr          = errors.New("invalid email")
	FirstNameErr             = errors.New("firstName must be provided")
	LastNameErr              = errors.New("lastName must be provided")
	UserAlreadyActivatedErr  = errors.New("user already activated")
	EmailConfirmationCodeErr = errors.New("email confirmation code is of zero length")
	NewEmailConfirmationErr  = errors.New("new email and confirmation code do not match recod")
)

type Entity struct {
	core.Entity
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Me struct {
	Entity
	Email    string `json:"email"`
	NewEmail string `json:"newEmail,omitempty"`
}

type Api interface {
	Register(email, firstName, lastName, pwd string) error
	ResendActivationEmail(email string) error
	Activate(activationCode string) (id string, err error)
	Authenticate(email string, pwd string) (id string, err error)
	ChangeEmail(id, newEmail string) error
	ResendNewEmailConfirmationEmail(id string) error
	ConfirmNewEmail(email, confirmationCode string) error
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (string, error)
	ChangePwd(id, oldPwd, newPwd string) error
	Get(id string) (*Entity, error)
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}

type InvalidPwdErr struct {
	MinRuneCount  int
	MaxRuneCount  int
	RegexMatchers []string
}

func newInvalidPwdErr(minRuneCount, maxRuneCount int, regexMatchers []string) *InvalidPwdErr {
	return &InvalidPwdErr{
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]string, 0, len(regexMatchers)), regexMatchers...),
	}
}

func (e *InvalidPwdErr) Error() string {
	return fmt.Sprintf(fmt.Sprintf("pwd must be between %d and %d utf8 characters long and match all regexs %v", e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers))
}
