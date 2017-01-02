package user

import (
	. "github.com/pborman/uuid"
	"github.com/uber-go/zap"
)

type Org account

type User account

type Me struct {
	User
	Email    string  `json:"email"`
	NewEmail *string `json:"newEmail,omitempty"`
}

type Api interface {
	//accessible outside of active session
	Register(name, region, email, pwd string) error
	ResendActivationEmail(email string) error
	Activate(activationCode string) (UUID, error)
	Authenticate(username, pwd string) (UUID, error)
	ConfirmNewEmail(email, confirmationCode string) (UUID, error)
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error)
	GetUsers(ids []UUID) ([]*User, error)
	SearchUsers(search string, limit int) ([]*User, error)
	GetOrgs(ids []UUID) ([]*Org, error)
	SearchOrgs(search string, limit int) ([]*Org, error)
	//requires active session to access
	//user centric
	ChangeMyName(myId UUID, newUsername string) error
	ChangeMyEmail(myId UUID, newEmail string) error
	ResendMyNewEmailConfirmationEmail(myId UUID) error
	ChangeMyPwd(myId UUID, oldPwd, newPwd string) error
	MigrateMe(myId UUID, newRegion string) error
	GetMe(myId UUID) (*Me, error)
	DeleteMe(myId UUID) error
	//org centric - must be an owner member
	CreateOrg(myId UUID, name, region string) (*Org, error)
	RenameOrg(myId, orgId UUID, newName string) error
	MigrateOrg(myId, orgId UUID, newRegion string) error
	GetMyOrgs(myId UUID, limit int) ([]*Org, error)
	DeleteOrg(myId, orgId UUID) error
	//member centric - must be an owner or admin
	AddMembers(myId, orgId UUID, newMembers []UUID) error
	RemoveMembers(myId, orgId UUID, existingMembers []UUID) error
}

type InternalRegionalApiProvider interface {
	Exists(string) bool
	Get(string) (InternalRegionalApi, error)
}

type InternalRegionalApi interface {
	CreatePersonalTaskCenter(userId UUID) (int, error)
	CreateOrgTaskCenter(ownerId, orgId UUID) (int, error)
	RenameMember(memberId, orgId UUID, newName string) error
}

func NewLogLinkMailer(log zap.Logger) (linkMailer, error) {
	if log == nil {
		return nil, nilLogErr
	}
	return &logLinkMailer{
		log: log,
	}, nil
}
