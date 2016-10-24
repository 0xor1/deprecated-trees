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
	GetMe(id string) (*Me, error)
}

type Store interface {
	GetByEmail(email string) (*FullUserInfo, error)
	GetById(id string) (*FullUserInfo, error)
	GetByActivationCode(activationCode string) (*FullUserInfo, error)
	GetByNewEmailConfirmationCode(confirmationCode string) (*FullUserInfo, error)
	GetByResetPwdCode(resetPwdCode string) (*FullUserInfo, error)
	Create(user *FullUserInfo) error
	Update(user *FullUserInfo) error
	Delete(id string) error
}

type LinkMailer interface {
	SendActivationLink(address, activationCode string) error
	SendPwdResetLink(address, resetCode string) error
	SendNewEmailConfirmationLink(address, confirmationCode string) error
}
