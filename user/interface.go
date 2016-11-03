package user

import (
	"bitbucket.org/robsix/core"
	"errors"
	"fmt"
)

var (
	NilLinkMailerErr         = errors.New("nil linkMailer passed to Api")
	NilLogErr                = errors.New("nil log passed to Api")
	NilNeoDbErr              = errors.New("nil Neo Db passed to NeoApi")
	IncorrectPwdErr          = errors.New("password incorrect")
	UserNotActivated         = errors.New("user not activated")
	InvalidEmailErr          = errors.New("invalid email")
	EmailAlreadyInUseErr     = errors.New("email already in use")
	UsernameAlreadyInUseErr  = errors.New("username already in use")
	UserAlreadyActivatedErr  = errors.New("user already activated")
	EmailConfirmationCodeErr = errors.New("email confirmation code is of zero length")
	NewEmailErr              = errors.New("newEmail is of zero length")
	NewEmailConfirmationErr  = errors.New("new email and confirmation code do not match those recorded")
)

type InvalidStringParamErr struct {
	ParamPurpose  string
	MinRuneCount  int
	MaxRuneCount  int
	RegexMatchers []string
}

func (e *InvalidStringParamErr) Error() string {
	return fmt.Sprintf(fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.ParamPurpose, e.MinRuneCount, e.MaxRuneCount, e.RegexMatchers))
}

type User struct {
	core.Entity
	Username string `json:"username"`
}

type Me struct {
	User
	Email    string `json:"email"`
	NewEmail string `json:"newEmail,omitempty"`
}

type Api interface {
	Register(username, email, pwd string) error
	ResendActivationEmail(email string) error
	Activate(activationCode string) (id string, err error)
	Authenticate(username, pwd string) (id string, err error)
	ChangeUsername(id, newUsername string) error
	ChangeEmail(id, newEmail string) error
	ResendNewEmailConfirmationEmail(id string) error
	ConfirmNewEmail(email, confirmationCode string) error
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (string, error)
	ChangePwd(id, oldPwd, newPwd string) error
	GetMe(id string) (*Me, error)
	Delete(id string) error
	Get(ids []string) ([]*User, error)
	Search(search string, limit int) ([]*User, error)
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}
