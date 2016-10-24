package user

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

type Store interface {
	GetByEmail(email string) (*User, error)
	GetById(id string) (*User, error)
	GetByActivationCode(activationCode string) (*User, error)
	GetByNewEmailConfirmationCode(confirmationCode string) (*User, error)
	GetByResetPwdCode(resetPwdCode string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}
