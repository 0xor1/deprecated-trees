package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
	"io"
)

// The main account Api interface
type Api interface {
	//accessible outside of active session
	GetRegions() []string
	Register(name, email, pwd, region string) error
	ResendActivationEmail(email string) error
	Activate(email, activationCode string) (Id, error)
	Authenticate(username, pwd string) (Id, error)
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) (Id, error)
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) (Id, error)
	GetAccount(name string) (*account, error)
	GetAccounts(ids []Id) ([]*account, error)
	//requires active session to access
	GetMe(myId Id) (*me, error)
	SetMyPwd(myId Id, oldPwd, newPwd string) error
	SetMyEmail(myId Id, newEmail string) error
	ResendMyNewEmailConfirmationEmail(myId Id) error
	SetAccountName(myId, accountId Id, newName string) error
	SetAccountAvatar(myId, accountId Id, avatarImage io.ReadCloser) error
	MigrateAccount(myId, accountId Id, newRegion string) error
	CreateOrg(myId Id, name, region string) (*account, error)
	GetMyOrgs(myId Id, offset, limit int) ([]*account, int, error)
	DeleteAccount(myId, accountId Id) error
	//member centric - must be an owner or admin
	AddMembers(myId, orgId Id, newMembers []Id) error
	RemoveMembers(myId, orgId Id, existingMembers []Id) error
}

// Return a new account Api backed by sql storage and sending link emails via an email service
func NewApi(internalRegionClient internalRegionClient, linkMailer linkMailer, avatarStore avatarStore, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, accountsDb, pwdsDb isql.DB, log Log) Api {
	return newApi(newSqlStore(accountsDb, pwdsDb), internalRegionClient, linkMailer, avatarStore, NewNamedEntity, NewCryptoHelper(), nameRegexMatchers, pwdRegexMatchers, maxAvatarDim, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen, log)
}

func NewLogLinkMailer(log Log) linkMailer {
	if log == nil {
		NilCriticalParamPanic("log")
	}
	return &logLinkMailer{
		log: log,
	}
}

func NewSparkPostLinkMailer() linkMailer {
	panic(NotImplementedErr)
}
