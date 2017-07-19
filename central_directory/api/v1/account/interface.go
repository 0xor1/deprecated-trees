package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"io"
)

// The main account Api interface
type Api interface {
	//accessible outside of active session
	GetRegions() []string
	Register(name, email, pwd, region, language string, theme Theme)
	ResendActivationEmail(email string)
	Activate(email, activationCode string)
	Authenticate(username, pwd string) Id
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string)
	ResetPwd(email string)
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string)
	GetAccount(name string) *account
	GetAccounts(ids []Id) []*account
	//requires active session to access
	GetMe(myId Id) *me
	SetMyPwd(myId Id, oldPwd, newPwd string)
	SetMyEmail(myId Id, newEmail string)
	ResendMyNewEmailConfirmationEmail(myId Id)
	SetAccountName(myId, accountId Id, newName string)
	SetAccountAvatar(myId, accountId Id, avatarImage io.ReadCloser)
	MigrateAccount(myId, accountId Id, newRegion string)
	CreateOrg(myId Id, name, region string) *account
	GetMyOrgs(myId Id, offset, limit int) ([]*account, int)
	DeleteAccount(myId, accountId Id)
	//member centric - must be an owner or admin
	AddMembers(myId, orgId Id, newMembers []*AddMemberExternal)
	RemoveMembers(myId, orgId Id, existingMembers []Id)
}

// Return a new account Api backed by sql storage and sending link emails via an email service
func NewApi(internalRegionClient InternalRegionClient, linkMailer linkMailer, avatarStore avatarStore, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, accountsDb, pwdsDb isql.DB) Api {
	return newApi(newSqlStore(accountsDb, pwdsDb), internalRegionClient, linkMailer, avatarStore, NewCreatedNamedEntity, NewCryptoHelper(), nameRegexMatchers, pwdRegexMatchers, maxAvatarDim, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen)
}

func NewLogLinkMailer(log Log) linkMailer {
	if log == nil {
		panic(InvalidArgumentsErr)
	}
	return &logLinkMailer{
		log: log,
	}
}

func NewSparkPostLinkMailer() linkMailer {
	panic(NotImplementedErr)
}
