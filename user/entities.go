package user

import (
	"bitbucket.org/robsix/core"
	"time"
)

type User struct {
	core.Entity
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Me struct {
	User
	Email    string `json:"email"`
	NewEmail string `json:"newEmail,omitempty"`
}

type FullUserInfo struct {
	Me
	RegistrationTime         time.Time  `json:"registrationTime"`
	ActivationCode           string     `json:"activationCode,omitempty"`
	ActivationTime           *time.Time `json:"activationTime,omitempty"`
	NewEmailConfirmationCode string     `json:"newEmailConfirmationCode,omitempty"`
	ResetPwdCode             string     `json:"resetPwdCode,omitempty"`
	ScryptSalt               []byte     `json:"scryptSalt"`
	ScryptPwd                []byte     `json:"scryptPwd"`
	ScryptN                  int        `json:"scryptN"`
	ScryptR                  int        `json:"scryptR"`
	ScryptP                  int        `json:"scryptP"`
	ScryptKeyLen             int        `json:"scryptKeyLen"`
}

func (u *FullUserInfo) IsActivated() bool {
	return len(u.ActivationCode) == 0
}

func (u *FullUserInfo) ToMe() *Me {
	return &Me{
		User: User{
			Entity: core.Entity{
				Id: u.Id,
			},
			FirstName: u.FirstName,
			LastName:  u.LastName,
		},
		Email:    u.Email,
		NewEmail: u.NewEmail,
	}
}

func (u *FullUserInfo) ToUser() *User {
	return &User{
		Entity: core.Entity{
			Id: u.Id,
		},
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
