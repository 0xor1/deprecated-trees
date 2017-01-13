package account

import (
	. "bitbucket.org/robsix/task_center/misc"
)

type Api interface {
	//accessible outside of active session
	Register(name, email, pwd, region string) error
	ResendActivationEmail(email string) error
	Activate(email, activationCode string) (Id, error)
	Authenticate(username, pwd string) (Id, error)
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) (Id, error)
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) (Id, error)
	GetUsers(ids []Id) ([]*user, error)
	SearchUsers(search string, offset, limit int) ([]*user, int, error)
	GetOrgs(ids []Id) ([]*org, error)
	SearchOrgs(search string, offset, limit int) ([]*org, int, error)
	//requires active session to access
	//user centric
	ChangeMyName(myId Id, newUsername string) error
	ChangeMyPwd(myId Id, oldPwd, newPwd string) error
	ChangeMyEmail(myId Id, newEmail string) error
	ResendMyNewEmailConfirmationEmail(myId Id) error
	MigrateMe(myId Id, newRegion string) error
	GetMe(myId Id) (*me, error)
	DeleteMe(myId Id) error
	//org centric - must be an owner member
	CreateOrg(myId Id, name, region string) (*org, error)
	RenameOrg(myId, orgId Id, newName string) error
	MigrateOrg(myId, orgId Id, newRegion string) error
	GetMyOrgs(myId Id, offset, limit int) ([]*org, int, error)
	DeleteOrg(myId, orgId Id) error
	//member centric - must be an owner or admin
	AddMembers(myId, orgId Id, newMembers []Id) error
	RemoveMembers(myId, orgId Id, existingMembers []Id) error
}

func NewLogLinkMailer(log Log) (linkMailer, error) {
	if log == nil {
		return nil, nilLogErr
	}
	return &logLinkMailer{
		log: log,
	}, nil
}