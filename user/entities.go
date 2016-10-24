package user

import (
	"bitbucket.org/robsix/core"
	"time"
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

type User struct {
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

func (u *User) isActivated() bool {
	return len(u.ActivationCode) == 0
}

func (u *User) toMe() *Me {
	return &Me{
		Entity: Entity{
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

func (u *User) toEntity() *Entity {
	return &Entity{
		Entity: core.Entity{
			Id: u.Id,
		},
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
