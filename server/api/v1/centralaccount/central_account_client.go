package centralaccount

import (
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
	"bytes"
	"github.com/0xor1/panic"
	"io"
	"io/ioutil"
	"mime/multipart"
)

type Client interface {
	//accessible outside of active session
	Register(region cnst.Region, name, email, pwd, language string, displayName *string, theme cnst.Theme) error
	ResendActivationEmail(email string) error
	Activate(email, activationCode string) error
	Authenticate(css *clientsession.Store, email, pwd string) (*AuthenticateResult, error)
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error
	GetAccount(name string) (*Account, error)
	GetAccounts(accounts []id.Id) ([]*Account, error)
	SearchAccounts(nameOrDisplayNamePrefix string) ([]*Account, error)
	SearchPersonalAccounts(nameOrDisplayNamePrefix string) ([]*Account, error)
	//requires active session to access
	GetMe(css *clientsession.Store) (*Me, error)
	SetMyPwd(css *clientsession.Store, oldPwd, newPwd string) error
	SetMyEmail(css *clientsession.Store, newEmail string) error
	ResendMyNewEmailConfirmationEmail(css *clientsession.Store) error
	SetAccountName(css *clientsession.Store, account id.Id, newName string) error
	SetAccountDisplayName(css *clientsession.Store, account id.Id, newDisplayName *string) error
	SetAccountAvatar(css *clientsession.Store, account id.Id, avatar io.ReadCloser) error
	MigrateAccount(css *clientsession.Store, account id.Id, newRegion cnst.Region) error
	CreateAccount(css *clientsession.Store, region cnst.Region, name string, displayName *string) (*Account, error)
	GetMyAccounts(css *clientsession.Store, after *id.Id, limit int) (*GetMyAccountsResult, error)
	DeleteAccount(css *clientsession.Store, account id.Id) error
	//member centric - must be an owner or admin
	AddMembers(css *clientsession.Store, account id.Id, newMembers []*AddMember) error
	RemoveMembers(css *clientsession.Store, account id.Id, existingMembers []id.Id) error
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Register(region cnst.Region, name, email, pwd, language string, displayName *string, theme cnst.Theme) error {
	_, e := register.DoRequest(nil, c.host, cnst.CentralRegion, &registerArgs{
		Name:        name,
		Email:       email,
		Pwd:         pwd,
		Region:      region,
		Language:    language,
		DisplayName: displayName,
		Theme:       theme,
	}, nil, nil)
	return e
}

func (c *client) ResendActivationEmail(email string) error {
	_, e := resendActivationEmail.DoRequest(nil, c.host, cnst.CentralRegion, &resendActivationEmailArgs{
		Email: email,
	}, nil, nil)
	return e
}

func (c *client) Activate(email, activationCode string) error {
	_, e := activate.DoRequest(nil, c.host, cnst.CentralRegion, &activateArgs{
		Email:          email,
		ActivationCode: activationCode,
	}, nil, nil)
	return e
}

func (c *client) Authenticate(css *clientsession.Store, email, pwdTry string) (*AuthenticateResult, error) {
	val, e := authenticate.DoRequest(css, c.host, cnst.CentralRegion, &authenticateArgs{
		Email:  email,
		PwdTry: pwdTry,
	}, nil, &AuthenticateResult{})
	if val != nil {
		return val.(*AuthenticateResult), e
	}
	return nil, e
}

func (c *client) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error {
	_, e := confirmNewEmail.DoRequest(nil, c.host, cnst.CentralRegion, &confirmNewEmailArgs{
		CurrentEmail:     currentEmail,
		NewEmail:         newEmail,
		ConfirmationCode: confirmationCode,
	}, nil, nil)
	return e
}

func (c *client) ResetPwd(email string) error {
	_, e := resetPwd.DoRequest(nil, c.host, cnst.CentralRegion, &resetPwdArgs{
		Email: email,
	}, nil, nil)
	return e
}

func (c *client) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error {
	_, e := setNewPwdFromPwdReset.DoRequest(nil, c.host, cnst.CentralRegion, &setNewPwdFromPwdResetArgs{
		NewPwd:       newPwd,
		Email:        email,
		ResetPwdCode: resetPwdCode,
	}, nil, nil)
	return e
}

func (c *client) GetAccount(name string) (*Account, error) {
	val, e := getAccount.DoRequest(nil, c.host, cnst.CentralRegion, &getAccountArgs{
		Name: name,
	}, nil, &Account{})
	if val != nil {
		return val.(*Account), e
	}
	return nil, e
}

func (c *client) GetAccounts(accounts []id.Id) ([]*Account, error) {
	val, e := getAccounts.DoRequest(nil, c.host, cnst.CentralRegion, &getAccountsArgs{
		Accounts: accounts,
	}, nil, &[]*Account{})
	if val != nil {
		return *val.(*[]*Account), e
	}
	return nil, e
}

func (c *client) SearchAccounts(nameOrDisplayNamePrefix string) ([]*Account, error) {
	val, e := searchAccounts.DoRequest(nil, c.host, cnst.CentralRegion, &searchAccountsArgs{
		NameOrDisplayNamePrefix: nameOrDisplayNamePrefix,
	}, nil, &[]*Account{})
	if val != nil {
		return *val.(*[]*Account), e
	}
	return nil, e
}

func (c *client) SearchPersonalAccounts(nameOrDisplayNamePrefix string) ([]*Account, error) {
	val, e := searchPersonalAccounts.DoRequest(nil, c.host, cnst.CentralRegion, &searchPersonalAccountsArgs{
		NameOrDisplayNamePrefix: nameOrDisplayNamePrefix,
	}, nil, &[]*Account{})
	if val != nil {
		return *val.(*[]*Account), e
	}
	return nil, e
}

func (c *client) GetMe(css *clientsession.Store) (*Me, error) {
	val, e := getMe.DoRequest(css, c.host, cnst.CentralRegion, nil, nil, &Me{})
	if val != nil {
		return val.(*Me), e
	}
	return nil, e
}

func (c *client) SetMyPwd(css *clientsession.Store, oldPwd, newPwd string) error {
	_, e := setMyPwd.DoRequest(css, c.host, cnst.CentralRegion, &setMyPwdArgs{
		OldPwd: oldPwd,
		NewPwd: newPwd,
	}, nil, nil)
	return e
}

func (c *client) SetMyEmail(css *clientsession.Store, newEmail string) error {
	_, e := setMyEmail.DoRequest(css, c.host, cnst.CentralRegion, &setMyEmailArgs{
		NewEmail: newEmail,
	}, nil, nil)
	return e
}

func (c *client) ResendMyNewEmailConfirmationEmail(css *clientsession.Store) error {
	_, e := resendMyNewEmailConfirmationEmail.DoRequest(css, c.host, cnst.CentralRegion, nil, nil, nil)
	return e
}

func (c *client) SetAccountName(css *clientsession.Store, account id.Id, newName string) error {
	_, e := setAccountName.DoRequest(css, c.host, cnst.CentralRegion, &setAccountNameArgs{
		Account: account,
		NewName: newName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountDisplayName(css *clientsession.Store, account id.Id, newDisplayName *string) error {
	_, e := setAccountDisplayName.DoRequest(css, c.host, cnst.CentralRegion, &setAccountDisplayNameArgs{
		Account:        account,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountAvatar(css *clientsession.Store, account id.Id, avatar io.ReadCloser) error {
	defer avatar.Close()
	_, e := setAccountAvatar.DoRequest(css, c.host, cnst.CentralRegion, &setAccountAvatarArgs{
		Account: account,
		Avatar:  avatar,
	}, func() (io.ReadCloser, string) {
		body := bytes.NewBuffer([]byte{})
		writer := multipart.NewWriter(body)
		part, e := writer.CreateFormFile("avatar", "avatar")
		panic.IfNotNil(e)
		_, e = io.Copy(part, avatar)
		panic.IfNotNil(e)
		panic.IfNotNil(writer.WriteField("account", account.String()))
		panic.IfNotNil(writer.Close())
		return ioutil.NopCloser(body), writer.FormDataContentType()
	}, nil)
	return e
}

func (c *client) MigrateAccount(css *clientsession.Store, account id.Id, newRegion cnst.Region) error {
	_, e := migrateAccount.DoRequest(css, c.host, cnst.CentralRegion, &migrateAccountArgs{
		Account:   account,
		NewRegion: newRegion,
	}, nil, nil)
	return e
}

func (c *client) CreateAccount(css *clientsession.Store, region cnst.Region, name string, displayName *string) (*Account, error) {
	val, e := createAccount.DoRequest(css, c.host, cnst.CentralRegion, &createAccountArgs{
		Name:        name,
		Region:      region,
		DisplayName: displayName,
	}, nil, &Account{})
	if val != nil {
		return val.(*Account), e
	}
	return nil, e
}

func (c *client) GetMyAccounts(css *clientsession.Store, after *id.Id, limit int) (*GetMyAccountsResult, error) {
	val, e := getMyAccounts.DoRequest(css, c.host, cnst.CentralRegion, &getMyAccountsArgs{
		After: after,
		Limit: limit,
	}, nil, &GetMyAccountsResult{})
	if val != nil {
		return val.(*GetMyAccountsResult), e
	}
	return nil, e
}

func (c *client) DeleteAccount(css *clientsession.Store, account id.Id) error {
	_, e := deleteAccount.DoRequest(css, c.host, cnst.CentralRegion, &deleteAccountArgs{
		Account: account,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, account id.Id, newMembers []*AddMember) error {
	_, e := addMembers.DoRequest(css, c.host, cnst.CentralRegion, &addMembersArgs{
		Account:    account,
		NewMembers: newMembers,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, account id.Id, existingMembers []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, cnst.CentralRegion, &removeMembersArgs{
		Account:         account,
		ExistingMembers: existingMembers,
	}, nil, nil)
	return e
}
