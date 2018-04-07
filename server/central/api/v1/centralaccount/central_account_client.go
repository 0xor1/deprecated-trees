package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
)

type Client interface {
	//accessible outside of active session
	GetRegions() ([]string, error)
	Register(name, email, pwd, region, language string, displayName *string, theme cnst.Theme) error
	ResendActivationEmail(email string) error
	Activate(email, activationCode string) error
	Authenticate(css *clientsession.Store, email, pwd string) (id.Id, error)
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error
	GetAccount(name string) (*account, error)
	GetAccounts(accounts []id.Id) ([]*account, error)
	SearchAccounts(nameOrDisplayNameStartsWith string) ([]*account, error)
	SearchPersonalAccounts(nameOrDisplayNameStartsWith string) ([]*account, error)
	//requires active session to access
	GetMe(css *clientsession.Store) (*me, error)
	SetMyPwd(css *clientsession.Store, oldPwd, newPwd string) error
	SetMyEmail(css *clientsession.Store, newEmail string) error
	ResendMyNewEmailConfirmationEmail(css *clientsession.Store) error
	SetAccountName(css *clientsession.Store, account id.Id, newName string) error
	SetAccountDisplayName(css *clientsession.Store, account id.Id, newDisplayName *string) error
	SetAccountAvatar(css *clientsession.Store, account id.Id, avatar io.ReadCloser) error
	MigrateAccount(css *clientsession.Store, account id.Id, newRegion string) error
	CreateAccount(css *clientsession.Store, name, region string, displayName *string) (*account, error)
	GetMyAccounts(css *clientsession.Store, after *id.Id, limit int) (*getMyAccountsResp, error)
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

func (c *client) GetRegions() ([]string, error) {
	val, e := getRegions.DoRequest(nil, c.host, nil, nil, &[]string{})
	if val != nil {
		return *(val.(*[]string)), e
	}
	return nil, e
}

func (c *client) Register(name, email, pwd, region, language string, displayName *string, theme cnst.Theme) error {
	_, e := register.DoRequest(nil, c.host, &registerArgs{
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
	_, e := resendActivationEmail.DoRequest(nil, c.host, &resendActivationEmailArgs{
		Email: email,
	}, nil, nil)
	return e
}

func (c *client) Activate(email, activationCode string) error {
	_, e := activate.DoRequest(nil, c.host, &activateArgs{
		Email:          email,
		ActivationCode: activationCode,
	}, nil, nil)
	return e
}

func (c *client) Authenticate(css *clientsession.Store, email, pwdTry string) (id.Id, error) {
	val, e := authenticate.DoRequest(css, c.host, &authenticateArgs{
		Email:  email,
		PwdTry: pwdTry,
	}, nil, &id.Id{})
	if val != nil {
		return *val.(*id.Id), e
	}
	return nil, e
}

func (c *client) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error {
	_, e := confirmNewEmail.DoRequest(nil, c.host, &confirmNewEmailArgs{
		CurrentEmail:     currentEmail,
		NewEmail:         newEmail,
		ConfirmationCode: confirmationCode,
	}, nil, nil)
	return e
}

func (c *client) ResetPwd(email string) error {
	_, e := resetPwd.DoRequest(nil, c.host, &resetPwdArgs{
		Email: email,
	}, nil, nil)
	return e
}

func (c *client) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error {
	_, e := setNewPwdFromPwdReset.DoRequest(nil, c.host, &setNewPwdFromPwdResetArgs{
		NewPwd:       newPwd,
		Email:        email,
		ResetPwdCode: resetPwdCode,
	}, nil, nil)
	return e
}

func (c *client) GetAccount(name string) (*account, error) {
	val, e := getAccount.DoRequest(nil, c.host, &getAccountArgs{
		Name: name,
	}, nil, &account{})
	if val != nil {
		return val.(*account), e
	}
	return nil, e
}

func (c *client) GetAccounts(accounts []id.Id) ([]*account, error) {
	val, e := getAccounts.DoRequest(nil, c.host, &getAccountsArgs{
		Accounts: accounts,
	}, nil, &[]*account{})
	if val != nil {
		return *val.(*[]*account), e
	}
	return nil, e
}

func (c *client) SearchAccounts(nameOrDisplayNameStartsWith string) ([]*account, error) {
	val, e := searchAccounts.DoRequest(nil, c.host, &searchAccountsArgs{
		NameOrDisplayNameStartsWith: nameOrDisplayNameStartsWith,
	}, nil, &[]*account{})
	if val != nil {
		return *val.(*[]*account), e
	}
	return nil, e
}

func (c *client) SearchPersonalAccounts(nameOrDisplayNameStartsWith string) ([]*account, error) {
	val, e := searchPersonalAccounts.DoRequest(nil, c.host, &searchPersonalAccountsArgs{
		NameOrDisplayNameStartsWith: nameOrDisplayNameStartsWith,
	}, nil, &[]*account{})
	if val != nil {
		return *val.(*[]*account), e
	}
	return nil, e
}

func (c *client) GetMe(css *clientsession.Store) (*me, error) {
	val, e := getMe.DoRequest(css, c.host, nil, nil, &me{})
	if val != nil {
		return val.(*me), e
	}
	return nil, e
}

func (c *client) SetMyPwd(css *clientsession.Store, oldPwd, newPwd string) error {
	_, e := setMyPwd.DoRequest(css, c.host, &setMyPwdArgs{
		OldPwd: oldPwd,
		NewPwd: newPwd,
	}, nil, nil)
	return e
}

func (c *client) SetMyEmail(css *clientsession.Store, newEmail string) error {
	_, e := setMyEmail.DoRequest(css, c.host, &setMyEmailArgs{
		NewEmail: newEmail,
	}, nil, nil)
	return e
}

func (c *client) ResendMyNewEmailConfirmationEmail(css *clientsession.Store) error {
	_, e := resendMyNewEmailConfirmationEmail.DoRequest(css, c.host, nil, nil, nil)
	return e
}

func (c *client) SetAccountName(css *clientsession.Store, account id.Id, newName string) error {
	_, e := setAccountName.DoRequest(css, c.host, &setAccountNameArgs{
		Account: account,
		NewName: newName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountDisplayName(css *clientsession.Store, account id.Id, newDisplayName *string) error {
	_, e := setAccountDisplayName.DoRequest(css, c.host, &setAccountDisplayNameArgs{
		Account:        account,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountAvatar(css *clientsession.Store, account id.Id, avatar io.ReadCloser) error {
	defer avatar.Close()
	_, e := setAccountAvatar.DoRequest(css, c.host, &setAccountAvatarArgs{
		Account: account,
		Avatar:  avatar,
	}, func() (io.ReadCloser, string) {
		body := bytes.NewBuffer([]byte{})
		writer := multipart.NewWriter(body)
		part, e := writer.CreateFormFile("avatar", "avatar")
		err.PanicIf(e)
		_, e = io.Copy(part, avatar)
		err.PanicIf(e)
		err.PanicIf(writer.WriteField("account", account.String()))
		err.PanicIf(writer.Close())
		return ioutil.NopCloser(body), writer.FormDataContentType()
	}, nil)
	return e
}

func (c *client) MigrateAccount(css *clientsession.Store, account id.Id, newRegion string) error {
	_, e := migrateAccount.DoRequest(css, c.host, &migrateAccountArgs{
		Account:   account,
		NewRegion: newRegion,
	}, nil, nil)
	return e
}

func (c *client) CreateAccount(css *clientsession.Store, name, region string, displayName *string) (*account, error) {
	val, e := createAccount.DoRequest(css, c.host, &createAccountArgs{
		Name:        name,
		Region:      region,
		DisplayName: displayName,
	}, nil, &account{})
	if val != nil {
		return val.(*account), e
	}
	return nil, e
}

func (c *client) GetMyAccounts(css *clientsession.Store, after *id.Id, limit int) (*getMyAccountsResp, error) {
	val, e := getMyAccounts.DoRequest(css, c.host, &getMyAccountsArgs{
		After: after,
		Limit: limit,
	}, nil, &getMyAccountsResp{})
	if val != nil {
		return val.(*getMyAccountsResp), e
	}
	return nil, e
}

func (c *client) DeleteAccount(css *clientsession.Store, account id.Id) error {
	_, e := deleteAccount.DoRequest(css, c.host, &deleteAccountArgs{
		Account: account,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, account id.Id, newMembers []*AddMember) error {
	_, e := addMembers.DoRequest(css, c.host, &addMembersArgs{
		Account:    account,
		NewMembers: newMembers,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, account id.Id, existingMembers []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, &removeMembersArgs{
		Account:         account,
		ExistingMembers: existingMembers,
	}, nil, nil)
	return e
}
