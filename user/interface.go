package user

import (
	"bitbucket.org/robsix/core"
)

type User struct {
	core.Entity
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type UserStore interface {
	Authenticate(email string, pwd string) (string, error)
	Register(email, firstName, lastName, pwd string) error
	ForgotPwd(email string) error
	ChangePwd(id, oldPwd, newPwd string) error
	Get(ids []string) ([]*User, error)
}

type GetUserAuthenticationInfo func(email string) (userId, scryptPwd, salt string, N, r, p, keyLen int, err error)
type SaveUserInfo func(email, firstName, lastName , scryptPwd, salt string, N, r, p, keyLen int) error
type SavePwd func(email, scryptPwd, salt string, N, r, p, keyLen int) error
type SendEmail func(address, content string) error
type Get func(ids []string) ([]*User, error)
