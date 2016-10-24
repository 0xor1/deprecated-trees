package user

type CoreApi interface {
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
	Get(id string) (*User, error)
}

type Store interface {
	GetByEmail(email string) (*fullUserInfo, error)
	GetById(id string) (*fullUserInfo, error)
	GetByActivationCode(activationCode string) (*fullUserInfo, error)
	GetByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error)
	GetByResetPwdCode(resetPwdCode string) (*fullUserInfo, error)
	Create(user *fullUserInfo) error
	Update(user *fullUserInfo) error
	Delete(id string) error
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}
