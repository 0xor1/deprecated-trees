package user

import (
	"bitbucket.org/robsix/task_center/misc"
	"errors"
	"fmt"
	. "github.com/pborman/uuid"
	"time"
)

var (
	NilUserStoreErr          = errors.New("nil userStore passed to Api")
	NilPwdStoreErr           = errors.New("nil pwdStore passed to Api")
	NilLinkMailerErr         = errors.New("nil linkMailer passed to Api")
	NilLogErr                = errors.New("nil log passed to Api")
	NoSuchUserErr            = errors.New("no such user")
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

type SearchTermTooShortErr struct {
	MinRuneCount int
}

func (e *SearchTermTooShortErr) Error() string {
	return fmt.Sprintf(fmt.Sprintf("search term must be at least %d characters long", e.MinRuneCount))
}

type User struct {
	misc.CentralEntity
	Username string `json:"username"`
}

type Me struct {
	User
	Email    string  `json:"email"`
	NewEmail *string `json:"newEmail,omitempty"`
}

type Api interface {
	Register(username, region, email, pwd string) error
	ResendActivationEmail(email string) error
	Activate(activationCode string) (id UUID, err error)
	Authenticate(username, pwd string) (id UUID, err error)
	ChangeUsername(id UUID, newUsername string) error
	ChangeEmail(id UUID, newEmail string) error
	ResendNewEmailConfirmationEmail(id UUID) error
	ConfirmNewEmail(email, confirmationCode string) (UUID, error)
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error)
	ChangePwd(id UUID, oldPwd, newPwd string) error
	GetMe(id UUID) (*Me, error)
	Delete(id UUID) error
	Get(ids []UUID) ([]*User, error)
	Search(search string, limit int) ([]*User, error)
}

type FullUserInfo struct {
	Me
	RegistrationTime         time.Time
	ActivationCode           *string
	ActivationTime           *time.Time
	NewEmailConfirmationCode *string
	ResetPwdCode             *string
}

func (u *FullUserInfo) isActivated() bool {
	return u.ActivationCode == nil
}

type PwdInfo struct {
	Salt   []byte
	Pwd    []byte
	N      int
	R      int
	P      int
	KeyLen int
}

type UserStore interface {
	Create(user *FullUserInfo) error
	GetByUsername(username string) (*FullUserInfo, error)
	GetByEmail(email string) (*FullUserInfo, error)
	GetById(id UUID) (*FullUserInfo, error)
	GetByActivationCode(activationCode string) (*FullUserInfo, error)
	GetByNewEmailConfirmationCode(confirmationCode string) (*FullUserInfo, error)
	GetByResetPwdCode(resetPwdCode string) (*FullUserInfo, error)
	GetByIds(ids []UUID) ([]*User, error)
	Search(search string, limit int) ([]*User, error)
	Update(user *FullUserInfo) error
	Delete(id UUID) error
}

type PwdStore interface {
	Create(userId UUID, pwdInfo *PwdInfo) error
	Get(userId UUID) (*PwdInfo, error)
	Update(userId UUID, pwdInfo *PwdInfo) error
	Delete(userId UUID) error
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}
