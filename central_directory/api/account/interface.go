package account

import (
	"bitbucket.org/robsix/task_center/misc"
	. "github.com/pborman/uuid"
)

type Api interface {
	//accessible outside of active session
	Register(name, email, pwd, region string) error
	ResendActivationEmail(email string) error
	Activate(activationCode string) (UUID, error)
	Authenticate(username, pwd string) (UUID, error)
	ConfirmNewEmail(email, confirmationCode string) (UUID, error)
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error)
	GetUsers(ids []UUID) ([]*user, error)
	SearchUsers(search string, limit int) ([]*user, error)
	GetOrgs(ids []UUID) ([]*org, error)
	SearchOrgs(search string, limit int) ([]*org, error)
	//requires active session to access
	//user centric
	ChangeMyName(myId UUID, newUsername string) error
	ChangeMyEmail(myId UUID, newEmail string) error
	ResendMyNewEmailConfirmationEmail(myId UUID) error
	ChangeMyPwd(myId UUID, oldPwd, newPwd string) error
	MigrateMe(myId UUID, newRegion string) error
	GetMe(myId UUID) (*me, error)
	DeleteMe(myId UUID) error
	//org centric - must be an owner member
	CreateOrg(myId UUID, name, region string) (*org, error)
	RenameOrg(myId, orgId UUID, newName string) error
	MigrateOrg(myId, orgId UUID, newRegion string) error
	GetMyOrgs(myId UUID, limit int) ([]*org, error)
	DeleteOrg(myId, orgId UUID) error
	//member centric - must be an owner or admin
	AddMembers(myId, orgId UUID, newMembers []UUID) error
	RemoveMembers(myId, orgId UUID, existingMembers []UUID) error
}

func NewLogLinkMailer(log misc.Log) (linkMailer, error) {
	if log == nil {
		return nil, nilLogErr
	}
	return &logLinkMailer{
		log: log,
	}, nil
}
