package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
	"io"
	"os"
	"path"
	"sync"
)

// The main account Api interface
type Api interface {
	//accessible outside of active session
	GetRegions() []string
	Register(name, email, pwd, region, language string, theme Theme)
	ResendActivationEmail(email string)
	Activate(email, activationCode string)
	Authenticate(email, pwd string) Id
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string)
	ResetPwd(email string)
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string)
	GetAccount(name string) *account
	GetAccounts(ids []Id) []*account
	SearchAccounts(nameStartsWith string) []*account
	SearchPersonalAccounts(nameOrEmailStartsWith string) []*account
	//requires active session to access
	GetMe(myId Id) *me
	SetMyPwd(myId Id, oldPwd, newPwd string)
	SetMyEmail(myId Id, newEmail string)
	ResendMyNewEmailConfirmationEmail(myId Id)
	SetAccountName(myId, accountId Id, newName string)
	SetAccountAvatar(myId, accountId Id, avatarImage io.ReadCloser)
	MigrateAccount(myId, accountId Id, newRegion string)
	CreateAccount(myId Id, name, region string) *account
	GetMyAccounts(myId Id, offset, limit int) ([]*account, int)
	DeleteAccount(myId, accountId Id)
	//member centric - must be an owner or admin
	AddMembers(myId, accountId Id, newMembers []*AddMemberPublic)
	RemoveMembers(myId, accountId Id, existingMembers []Id)
}

// Return a new account Api backed by sql storage and sending link emails via an email service
func New(accountsDb, pwdsDb isql.ReplicaSet, internalRegionClient PrivateRegionClient, linkMailer linkMailer, avatarStore avatarStore, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxProcessEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int) Api {
	return newApi(newSqlStore(accountsDb, pwdsDb), internalRegionClient, linkMailer, avatarStore, nameRegexMatchers, pwdRegexMatchers, maxAvatarDim, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxProcessEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen)
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

func NewLocalAvatarStore(relDirPath string) avatarStore {
	if relDirPath == "" {
		panic(InvalidArgumentsErr)
	}
	wd, err := os.Getwd()
	PanicIf(err)
	absDirPath := path.Join(wd, relDirPath)
	os.MkdirAll(absDirPath, os.ModeDir)
	return &localAvatarStore{
		mtx:        &sync.Mutex{},
		absDirPath: absDirPath,
	}
}

func NewS3AvatarStore() avatarStore {
	panic(NotImplementedErr)
}