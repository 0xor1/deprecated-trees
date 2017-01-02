package user

import (
	"bitbucket.org/robsix/task_center/misc"
	"bytes"
	"errors"
	"fmt"
	. "github.com/pborman/uuid"
	"github.com/uber-go/zap"
	"golang.org/x/crypto/scrypt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	redactedInfo                                = "******"
	registerFnLogMsg                            = "account.api.Register"
	resendActivationEmailFnLogMsg               = "account.api.ResendActivationEmail"
	activateFnLogMsg                            = "account.api.Activate"
	authenticateFnLogMsg                        = "account.api.Authenticate"
	ConfirmNewEmailFnLogMsg                     = "account.api.ConfirmNewEmail"
	resetPwdFnLogMsg                            = "account.api.ResetPwd"
	setNewPwdFromPwdResetFnLogMsg               = "account.api.SetNewPwdFromPwdReset"
	getUsersFnLogMsg                            = "account.api.GetUsers"
	searchUsersFnLogMsg                         = "account.api.SearchUsers"
	getOrgsFnLogMsg                             = "account.api.GetOrgs"
	searchOrgsFnLogMsg                          = "account.api.SearchOrgs"
	changeMyNameFnLogMsg                        = "account.api.ChangeMyName"
	changeMyEmailFnLogMsg                       = "account.api.ChangeMyEmail"
	resendMyNewEmailConfirmationEmailFnLogMsg   = "account.api.ResendMyNewEmailConfirmationEmail"
	changeMyPwdFnLogMsg                         = "account.api.ChangeMyPwd"
	migrateMeFnLogMsg                           = "account.api.MigrateMe"
	getMeFnLogMsg                               = "account.api.GetMe"
	deleteMeFnLogMsg                            = "account.api.DeleteMe"
	createOrgFnLogMsg                           = "account.api.CreateOrg"
	renameOrgFnLogMsg                           = "account.api.RenameOrg"
	migrateOrgFnLogMsg                          = "account.api.MigrateOrg"
	getMyOrgsFnLogMsg                           = "account.api.GetMyOrgs"
	deleteOrgFnLogMsg                           = "account.api.DeleteOrg"
	addMembersFnLogMsg                          = "account.api.AddMembers"
	removeMembersFnLogMsg                       = "account.api.RemoveMembers"
	subcall                                     = "subcall"
	storeGetAccountByName                       = "store.getAccountByName"
	storeCreateUser                             = "store.createUser"
	storeGetUserByName                          = "store.getUserByName"
	storeGetUserByEmail                         = "store.getUserByEmail"
	storeGetUserById                            = "store.getUserById"
	storeGetUserByActivationCode                = "store.getUserByActivationCode"
	storeGetUserByNewEmailConfirmationCode      = "store.getUserByNewEmailConfirmationCode"
	storeGetUserByResetPwdCode                  = "store.getUserByResetPwdCode"
	storeGetPwdInfo                             = "store.getPwdInfo"
	storeUpdateUser                             = "store.updateUser"
	storeUpdatePwdInfo                          = "store.updatePwdInfo"
	storeDeleteUser                             = "store.deleteUser"
	storeGetUsers                               = "store.getUsers"
	storeSearchUsers                            = "store.searchUsers"
	storeCreateOrg                              = "store.createOrg"
	storeGetOrgById                             = "store.getOrgById"
	storeGetOrgByName                           = "store.getOrgByName"
	storeUpdateOrg                              = "store.updateOrg"
	storeDeleteOrg                              = "store.deleteOrg"
	storeGetOrgs                                = "store.getOrgs"
	storeSearchOrgs                             = "store.searchOrgs"
	storeGetUsersOrgs                           = "store.getUsersOrgs"
	internalRegionalApiProviderGet              = "internalRegionalApiProvider.Get"
	internalRegionalApiCreatePersonalTaskCenter = "internalRegionalApi.CreatePersonalTaskCenter"
	internalRegionalApiCreateOrgTaskCenter      = "internalRegionalApi.CreateOrgTaskCenter"
	linkMailerSendActivationLink                = "linkMailer.sendActivationLink"
	linkMailerSendPwdResetLink                  = "linkMailer.sendPwdResetLink"
	linkMailerSendNewEmailConfirmationLink      = "linkMailer.sendNewEmailConfirmationLink"
	miscGenerateCryptoBytes                     = "misc.GenerateCryptoBytes"
	miscGenerateCryptoUrlSafeString             = "misc.GenerateCryptoUrlSafeString"
	scryptKey                                   = "scrypt.Key"
)

var (
	nilStoreErr                = errors.New("nil store")
	nilLinkMailerErr           = errors.New("nil linkMailer")
	nilLogErr                  = errors.New("nil log")
	noSuchRegionErr            = errors.New("no such region")
	noSuchUserErr              = errors.New("no such user")
	incorrectPwdErr            = errors.New("password incorrect")
	userNotActivated           = errors.New("user not activated")
	emailAlreadyInUseErr       = errors.New("email already in use")
	accountNameAlreadyInUseErr = errors.New("account already in use")
	userAlreadyActivatedErr    = errors.New("user already activated")
	emailConfirmationCodeErr   = errors.New("email confirmation code is of zero length")
	newEmailErr                = errors.New("newEmail is of zero length")
	newEmailConfirmationErr    = errors.New("new email and confirmation code do not match those recorded")
)

type invalidStringParamErr struct {
	paramPurpose  string
	minRuneCount  int
	maxRuneCount  int
	regexMatchers []string
}

func (e *invalidStringParamErr) Error() string {
	return fmt.Sprintf(fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.paramPurpose, e.minRuneCount, e.maxRuneCount, e.regexMatchers))
}

func newApi(store store, internalRegionalApiProvider InternalRegionalApiProvider, linkMailer linkMailer, nameRegexMatchers, pwdRegexMatchers []string, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log zap.Logger) (Api, error) {
	if store == nil {
		return nil, nilStoreErr
	}
	if linkMailer == nil {
		return nil, nilLinkMailerErr
	}
	if log == nil {
		return nil, nilLogErr
	}
	return &api{
		store: store,
		internalRegionalApiProvider: internalRegionalApiProvider,
		linkMailer:                  linkMailer,
		nameRegexMatchers:           append(make([]string, 0, len(nameRegexMatchers)), nameRegexMatchers...),
		pwdRegexMatchers:            append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		nameMinRuneCount:            nameMinRuneCount,
		nameMaxRuneCount:            nameMaxRuneCount,
		pwdMinRuneCount:             pwdMinRuneCount,
		pwdMaxRuneCount:             pwdMaxRuneCount,
		maxSearchLimitResults:       maxSearchLimitResults,
		cryptoCodeLen:               cryptoCodeLen,
		saltLen:                     saltLen,
		scryptN:                     scryptN,
		scryptR:                     scryptR,
		scryptP:                     scryptP,
		scryptKeyLen:                scryptKeyLen,
		log:                         log,
	}, nil
}

type api struct {
	store                       store
	internalRegionalApiProvider InternalRegionalApiProvider
	linkMailer                  linkMailer
	nameRegexMatchers           []string
	pwdRegexMatchers            []string
	nameMinRuneCount            int
	nameMaxRuneCount            int
	pwdMinRuneCount             int
	pwdMaxRuneCount             int
	maxSearchLimitResults       int
	cryptoCodeLen               int
	saltLen                     int
	scryptN                     int
	scryptR                     int
	scryptP                     int
	scryptKeyLen                int
	log                         zap.Logger
}

func (a *api) Register(name, region, email, pwd string) error {
	a.log.Debug(registerFnLogMsg, zap.String("name", name), zap.String("region", region), zap.String("email", redactedInfo), zap.String("pwd", redactedInfo))

	name = strings.Trim(name, " ")
	if err := validateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	email = strings.Trim(email, " ")
	if err := validateEmail(email); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	if err := validateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	if !a.internalRegionalApiProvider.Exists(region) {
		a.log.Info(registerFnLogMsg, zap.Error(noSuchRegionErr))
		return noSuchRegionErr
	}

	if account, err := a.store.getAccountByName(name); account != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, storeGetAccountByName), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(accountNameAlreadyInUseErr))
			return accountNameAlreadyInUseErr
		}
	}

	if user, err := a.store.getUserByEmail(email); user != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, storeGetUserByEmail), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(emailAlreadyInUseErr))
			return emailAlreadyInUseErr
		}
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, miscGenerateCryptoBytes), zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	activationCode, err := misc.GenerateCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, miscGenerateCryptoUrlSafeString), zap.Error(err))
		return err
	}

	userId, err := misc.NewId()
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.Error(err))
		return err
	}

	err = a.store.createUser(
		&fullUserInfo{
			Me: Me{
				User: User{
					Entity: misc.Entity{
						Id: userId,
					},
					Name:    name,
					Region:  region,
					Shard:   -1,
					Created: time.Now().UTC(),
				},
				Email: email,
			},
			ActivationCode: &activationCode,
		},
		&pwdInfo{
			Salt:   scryptSalt,
			Pwd:    scryptPwd,
			N:      a.scryptN,
			R:      a.scryptR,
			P:      a.scryptP,
			KeyLen: a.scryptKeyLen,
		},
	)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, storeCreateUser), zap.Error(err))
		return err
	}

	err = a.linkMailer.sendActivationLink(email, activationCode)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, linkMailerSendActivationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendActivationEmail(email string) error {
	a.log.Debug(resendActivationEmailFnLogMsg, zap.String("email", redactedInfo))

	email = strings.Trim(email, " ")
	user, err := a.store.getUserByEmail(email)
	if err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, storeGetUserByEmail), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resendActivationEmailFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	if user.isActivated() {
		a.log.Info(resendActivationEmailFnLogMsg, zap.Error(userAlreadyActivatedErr))
		return userAlreadyActivatedErr
	}

	if err = a.linkMailer.sendActivationLink(email, *user.ActivationCode); err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, linkMailerSendActivationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) Activate(activationCode string) (UUID, error) {
	a.log.Debug(activateFnLogMsg, zap.String("activationCode", redactedInfo))

	activationCode = strings.Trim(activationCode, " ")
	user, err := a.store.getUserByActivationCode(activationCode)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, storeGetUserByActivationCode), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(activateFnLogMsg, zap.Error(noSuchUserErr))
		return nil, noSuchUserErr
	}

	internalRegionalApi, err := a.internalRegionalApiProvider.Get(user.Region)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, internalRegionalApiProviderGet), zap.Error(err))
		return nil, err
	}

	shard, err := internalRegionalApi.CreatePersonalTaskCenter(user.Id)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, internalRegionalApiCreatePersonalTaskCenter), zap.Error(err))
		return nil, err
	}

	user.Shard = shard
	user.ActivationCode = nil
	activationTime := time.Now().UTC()
	user.Activated = &activationTime
	err = a.store.updateUser(user)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
		return nil, err
	}

	return user.Id, nil
}

func (a *api) Authenticate(name, pwdTry string) (UUID, error) {
	a.log.Debug(authenticateFnLogMsg, zap.String("name", name), zap.String("pwdTry", redactedInfo))

	name = strings.Trim(name, " ")
	user, err := a.store.getUserByName(name)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, storeGetUserByName), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(authenticateFnLogMsg, zap.Error(noSuchUserErr))
		return nil, noSuchUserErr
	}
	if !user.isActivated() {
		a.log.Info(authenticateFnLogMsg, zap.Error(userNotActivated))
		return nil, userNotActivated
	}

	pwdInfo, err := a.store.getPwdInfo(user.Id)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, storeGetPwdInfo), zap.Error(err))
		return nil, err
	}

	scryptPwdTry, err := scrypt.Key([]byte(pwdTry), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return nil, err
	}

	if !pwdsMatch(pwdInfo.Pwd, scryptPwdTry) {
		a.log.Info(authenticateFnLogMsg, zap.Error(incorrectPwdErr))
		return nil, incorrectPwdErr
	}

	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(*user.ResetPwdCode) > 0 {
		user.ResetPwdCode = nil
		if err = a.store.updateUser(user); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
			return nil, err
		}
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.N != a.scryptN || pwdInfo.R != a.scryptR || pwdInfo.P != a.scryptP || pwdInfo.KeyLen != a.scryptKeyLen || len(pwdInfo.Salt) < a.saltLen {
		pwdInfo.Salt, err = misc.GenerateCryptoBytes(a.saltLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, miscGenerateCryptoBytes), zap.Error(err))
			return nil, err
		}
		pwdInfo.N = a.scryptN
		pwdInfo.R = a.scryptR
		pwdInfo.P = a.scryptP
		pwdInfo.KeyLen = a.scryptKeyLen
		pwdInfo.Pwd, err = scrypt.Key([]byte(pwdTry), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
			return nil, err
		}
		if err = a.store.updatePwdInfo(user.Id, pwdInfo); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, storeUpdatePwdInfo), zap.Error(err))
			return nil, err
		}

	}

	return user.Id, nil
}

func (a *api) ConfirmNewEmail(newEmail string, confirmationCode string) (UUID, error) {
	a.log.Debug(ConfirmNewEmailFnLogMsg, zap.String("newEmail", redactedInfo), zap.String("newConfirmationCode", redactedInfo))

	user, err := a.store.getUserByNewEmailConfirmationCode(confirmationCode)
	if err != nil {
		a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, storeGetUserByNewEmailConfirmationCode), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(noSuchUserErr))
		return nil, noSuchUserErr
	}

	if *user.NewEmail != newEmail || *user.NewEmailConfirmationCode != confirmationCode {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(newEmailConfirmationErr))
		return nil, newEmailConfirmationErr
	}

	if user, err := a.store.getUserByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, storeGetUserByEmail), zap.Error(err))
			return nil, err
		} else {
			a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(emailAlreadyInUseErr))
			return nil, emailAlreadyInUseErr
		}
	}

	user.Email = newEmail
	user.NewEmail = nil
	user.NewEmailConfirmationCode = nil
	if err = a.store.updateUser(user); err != nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
		return nil, err
	}

	return user.Id, nil
}

func (a *api) ResetPwd(email string) error {
	a.log.Debug(resetPwdFnLogMsg, zap.String("email", redactedInfo))

	user, err := a.store.getUserByEmail(email)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, storeGetUserByEmail), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	resetPwdCode, err := misc.GenerateCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.String(subcall, miscGenerateCryptoUrlSafeString), zap.Error(err))
		return err
	}

	user.ResetPwdCode = &resetPwdCode
	if err = a.store.updateUser(user); err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	err = a.linkMailer.sendPwdResetLink(email, resetPwdCode)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, linkMailerSendPwdResetLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error) {
	a.log.Debug(setNewPwdFromPwdResetFnLogMsg, zap.String("newPwd", redactedInfo), zap.String("resetPwdCode", redactedInfo))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return nil, err
	}

	user, err := a.store.getUserByResetPwdCode(resetPwdCode)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, storeGetUserByResetPwdCode), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(noSuchUserErr))
		return nil, noSuchUserErr
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return nil, err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return nil, err
	}

	user.ActivationCode = nil
	user.ResetPwdCode = nil
	if err = a.store.updateUser(user); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
		return nil, err
	}

	if err = a.store.updatePwdInfo(
		user.Id,
		&pwdInfo{
			Pwd:    scryptPwd,
			Salt:   scryptSalt,
			N:      a.scryptN,
			R:      a.scryptR,
			P:      a.scryptP,
			KeyLen: a.scryptKeyLen,
		},
	); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, storeGetPwdInfo), zap.Error(err))
		return nil, err
	}

	return user.Id, nil
}

func (a *api) GetUsers(ids []UUID) ([]*User, error) {
	a.log.Debug(getUsersFnLogMsg, zap.String("ids", fmt.Sprintf("%v", ids)))

	users, err := a.store.getUsers(ids)
	if err != nil {
		a.log.Error(getUsersFnLogMsg, zap.String(subcall, storeGetUsers), zap.Error(err))
	}

	return users, err
}

func (a *api) SearchUsers(search string, limit int) ([]*User, error) {
	a.log.Debug(searchUsersFnLogMsg, zap.String("search", search), zap.Int("limit", limit))

	if limit < 1 || limit > a.maxSearchLimitResults {
		limit = a.maxSearchLimitResults
	}

	search = strings.Trim(search, " ")
	if err := validateStringParam("search", search, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		a.log.Info(searchUsersFnLogMsg, zap.Error(err))
		return err
	}

	users, err := a.store.searchUsers(search, limit)
	if err != nil {
		a.log.Error(searchUsersFnLogMsg, zap.String(subcall, storeSearchUsers), zap.Error(err))
	}

	return users, err
}

func (a *api) GetOrgs(ids []UUID) ([]*Org, error) {
	a.log.Debug(getOrgsFnLogMsg, zap.String("ids", fmt.Sprintf("%v", ids)))

	orgs, err := a.store.getOrgs(ids)
	if err != nil {
		a.log.Error(getOrgsFnLogMsg, zap.String(subcall, storeGetOrgs), zap.Error(err))
	}

	return orgs, err
}

func (a *api) SearchOrgs(search string, limit int) ([]*Org, error) {
	a.log.Debug(searchOrgsFnLogMsg, zap.String("search", search), zap.Int("limit", limit))

	if limit < 1 || limit > a.maxSearchLimitResults {
		limit = a.maxSearchLimitResults
	}

	search = strings.Trim(search, " ")
	if err := validateStringParam("search", search, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		a.log.Info(searchOrgsFnLogMsg, zap.Error(err))
		return err
	}

	orgs, err := a.store.searchOrgs(search, limit)
	if err != nil {
		a.log.Error(searchOrgsFnLogMsg, zap.String(subcall, storeSearchOrgs), zap.Error(err))
	}

	return orgs, err
}

func (a *api) ChangeMyName(myId UUID, newUsername string) error {
	a.log.Debug(changeMyNameFnLogMsg, zap.Base64("myId", myId), zap.String("newUsername", newUsername))

	newUsername = strings.Trim(newUsername, " ")
	if err := validateStringParam("username", newUsername, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		a.log.Info(changeMyNameFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.store.getUserByName(newUsername); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeMyNameFnLogMsg, zap.String(subcall, storeGetUserByName), zap.Error(err))
			return err
		} else {
			a.log.Info(changeMyNameFnLogMsg, zap.Error(accountNameAlreadyInUseErr))
			return accountNameAlreadyInUseErr
		}
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		a.log.Error(changeMyNameFnLogMsg, zap.String(subcall, storeGetUserById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeMyNameFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	user.Name = newUsername
	if err = a.store.updateUser(user); err != nil {
		a.log.Error(changeMyNameFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ChangeMyEmail(myId UUID, newEmail string) error {
	a.log.Debug(changeMyEmailFnLogMsg, zap.Base64("myId", myId), zap.String("newEmail", redactedInfo))

	newEmail = strings.Trim(newEmail, " ")
	if err := validateEmail(newEmail); err != nil {
		a.log.Info(changeMyEmailFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.store.getUserByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeMyEmailFnLogMsg, zap.String(subcall, storeGetUserByEmail), zap.Error(err))
			return err
		} else {
			a.log.Info(changeMyEmailFnLogMsg, zap.Error(emailAlreadyInUseErr))
			return emailAlreadyInUseErr
		}
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		a.log.Error(changeMyEmailFnLogMsg, zap.String(subcall, storeGetUserById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeMyEmailFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	confirmationCode, err := misc.GenerateCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(changeMyEmailFnLogMsg, zap.String(subcall, miscGenerateCryptoUrlSafeString), zap.Error(err))
		return err
	}

	user.NewEmail = &newEmail
	user.NewEmailConfirmationCode = &confirmationCode
	if err = a.store.updateUser(user); err != nil {
		a.log.Error(changeMyEmailFnLogMsg, zap.String(subcall, storeUpdateUser), zap.Error(err))
		return err
	}

	if err = a.linkMailer.sendNewEmailConfirmationLink(newEmail, confirmationCode); err != nil {
		a.log.Error(changeMyEmailFnLogMsg, zap.String(subcall, linkMailerSendNewEmailConfirmationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendMyNewEmailConfirmationEmail(myId UUID) error {
	a.log.Debug(resendMyNewEmailConfirmationEmailFnLogMsg, zap.Base64("myId", myId))

	user, err := a.store.getUserById(myId)
	if err != nil {
		a.log.Error(resendMyNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, storeGetUserById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resendMyNewEmailConfirmationEmailFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	// check the user has actually registered a new email
	if len(user.NewEmail) == 0 {
		a.log.Error(resendMyNewEmailConfirmationEmailFnLogMsg, zap.Error(newEmailErr))
		return newEmailErr
	}
	// just in case something has gone crazy wrong
	if len(user.NewEmailConfirmationCode) == 0 {
		a.log.Error(resendMyNewEmailConfirmationEmailFnLogMsg, zap.Error(emailConfirmationCodeErr))
		return emailConfirmationCodeErr
	}

	err = a.linkMailer.sendNewEmailConfirmationLink(*user.NewEmail, *user.NewEmailConfirmationCode)
	if err != nil {
		a.log.Error(resendMyNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, linkMailerSendNewEmailConfirmationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ChangeMyPwd(myId UUID, oldPwd, newPwd string) error {
	a.log.Debug(changeMyPwdFnLogMsg, zap.Base64("myId", myId), zap.String("oldPwd", redactedInfo), zap.String("newPwd", redactedInfo))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(changeMyPwdFnLogMsg, zap.Error(err))
		return err
	}

	pwdInfo, err := a.store.getPwdInfo(myId)
	if err != nil {
		a.log.Error(changeMyPwdFnLogMsg, zap.String(subcall, storeGetPwdInfo), zap.Error(err))
		return err
	}
	if pwdInfo == nil {
		a.log.Info(changeMyPwdFnLogMsg, zap.Error(noSuchUserErr))
		return noSuchUserErr
	}

	scryptPwdTry, err := scrypt.Key([]byte(oldPwd), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
	if err != nil {
		a.log.Error(changeMyPwdFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	if !pwdsMatch(pwdInfo.Pwd, scryptPwdTry) {
		a.log.Info(changeMyPwdFnLogMsg, zap.Error(incorrectPwdErr))
		return incorrectPwdErr
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(changeMyPwdFnLogMsg, zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(changeMyPwdFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	pwdInfo.Pwd = scryptPwd
	pwdInfo.Salt = scryptSalt
	pwdInfo.N = a.scryptN
	pwdInfo.R = a.scryptR
	pwdInfo.P = a.scryptP
	pwdInfo.KeyLen = a.scryptKeyLen
	if err = a.store.updatePwdInfo(myId, pwdInfo); err != nil {
		a.log.Error(changeMyPwdFnLogMsg, zap.String(subcall, storeUpdatePwdInfo), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) MigrateMe(myId UUID, newRegion string) error {
	a.log.Debug(migrateMeFnLogMsg, zap.Base64("myId", myId))

	//TODO

	return nil
}

func (a *api) GetMe(myId UUID) (*Me, error) {
	a.log.Debug(getMeFnLogMsg, zap.Base64("myId", myId))

	user, err := a.store.getUserById(myId)
	if err != nil {
		a.log.Error(getMeFnLogMsg, zap.String(subcall, storeGetUserById), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(getMeFnLogMsg, zap.Error(noSuchUserErr))
		return nil, noSuchUserErr
	}

	return &user.Me, nil
}

func (a *api) DeleteMe(id UUID) error {
	a.log.Debug(deleteMeFnLogMsg, zap.Base64("id", id))

	if err := a.store.deleteUser(id); err != nil {
		a.log.Error(deleteMeFnLogMsg, zap.String(subcall, storeDeleteUser), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) CreateOrg(myId UUID, name, region string) (*Org, error) {
	a.log.Debug(createOrgFnLogMsg, zap.Base64("myId", myId), zap.String("name", name), zap.String("region", region))

	//TODO

	return nil, nil
}

func (a *api) RenameOrg(myId, orgId UUID, newName string) error {
	a.log.Debug(renameOrgFnLogMsg, zap.Base64("myId", myId), zap.Base64("orgId", orgId), zap.String("newName", newName))

	//TODO

	return nil
}

func (a *api) MigrateOrg(myId, orgId UUID, newRegion string) error {
	a.log.Debug(migrateOrgFnLogMsg, zap.Base64("myId", myId), zap.Base64("orgId", orgId), zap.String("newRegion", newRegion))

	//TODO

	return nil
}

func (a *api) GetMyOrgs(myId UUID, limit int) ([]*Org, error) {
	a.log.Debug(getMyOrgsFnLogMsg, zap.Base64("myId", myId), zap.Int("limit", limit))

	//TODO

	return nil, nil
}

func (a *api) DeleteOrg(myId, orgId UUID) error {
	a.log.Debug(deleteOrgFnLogMsg, zap.Base64("myId", myId), zap.Base64("orgId", orgId))

	//TODO

	return nil
}

func (a *api) AddMembers(myId, orgId UUID, newMembers []UUID) error {
	a.log.Debug(addMembersFnLogMsg, zap.Base64("myId", myId), zap.Base64("orgId", orgId), zap.String("newMembers", fmt.Sprintf("%v", newMembers)))

	//TODO

	return nil
}

func (a *api) RemoveMembers(myId, orgId UUID, existingMembers []UUID) error {
	a.log.Debug(removeMembersFnLogMsg, zap.Base64("myId", myId), zap.Base64("orgId", orgId), zap.String("existingMembers", fmt.Sprintf("%v", existingMembers)))

	//TODO

	return nil
}

//internal helpers

type store interface {
	//user or org
	getAccountByName(name string) (*account, error)
	//user
	createUser(user *fullUserInfo, pwdInfo *pwdInfo) error
	getUserByName(name string) (*fullUserInfo, error)
	getUserByEmail(email string) (*fullUserInfo, error)
	getUserById(id UUID) (*fullUserInfo, error)
	getUserByActivationCode(activationCode string) (*fullUserInfo, error)
	getUserByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error)
	getUserByResetPwdCode(resetPwdCode string) (*fullUserInfo, error)
	getPwdInfo(id UUID) (*pwdInfo, error)
	updateUser(user *fullUserInfo) error
	updatePwdInfo(id UUID, pwdInfo *pwdInfo) error
	deleteUser(id UUID) error
	getUsers(ids []UUID) ([]*User, error)
	searchUsers(search string, limit int) ([]*User, error)
	//org
	createOrg(org *Org) error
	getOrgById(id UUID) (*Org, error)
	getOrgByName(name string) (*Org, error)
	updateOrg(org *Org) error
	deleteOrg(id UUID) error
	getOrgs(ids []UUID) ([]*Org, error)
	searchOrgs(search string, limit int) ([]*Org, error)
	getUsersOrgs(userId UUID, limit int) ([]*Org, error)
}

type linkMailer interface {
	sendActivationLink(address, activationCode string) error
	sendPwdResetLink(address, resetCode string) error
	sendNewEmailConfirmationLink(address, confirmationCode string) error
}

type account struct {
	misc.Entity
	Created   time.Time `json:"created"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	NewRegion string    `json:"newRegion"`
	Shard     int       `json:"shard"`
	IsUser    bool      `json:"isUser"`
}

func (a *account) isMigrating() bool {
	return len(a.NewRegion) != 0
}

type fullUserInfo struct {
	Me
	ActivationCode           *string
	Activated                *time.Time
	NewEmailConfirmationCode *string
	ResetPwdCode             *string
}

func (u *fullUserInfo) isActivated() bool {
	return u.Activated == nil
}

type pwdInfo struct {
	Salt   []byte
	Pwd    []byte
	N      int
	R      int
	P      int
	KeyLen int
}

type logLinkMailer struct {
	log zap.Logger
}

func (l *logLinkMailer) SendActivationLink(address, activationCode string) error {
	l.log.Info("logLinkMailer.SendActivationLink", zap.String("address", address), zap.String("activationCode", activationCode))
	return nil
}

func (l *logLinkMailer) SendPwdResetLink(address, resetCode string) error {
	l.log.Info("logLinkMailer.SendPwdResetLink", zap.String("address", address), zap.String("resetCode", resetCode))
	return nil
}

func (l *logLinkMailer) SendNewEmailConfirmationLink(address, confirmationCode string) error {
	l.log.Info("logLinkMailer.SendNewEmailConfirmationLink", zap.String("address", address), zap.String("confirmationCode", confirmationCode))
	return nil
}

func newInvalidStringParamErr(paramPurpose string, minRuneCount, maxRuneCount int, regexMatchers []string) *invalidStringParamErr {
	return &invalidStringParamErr{
		paramPurpose:  paramPurpose,
		minRuneCount:  minRuneCount,
		maxRuneCount:  maxRuneCount,
		regexMatchers: append(make([]string, 0, len(regexMatchers)), regexMatchers...),
	}
}

func validateStringParam(paramPurpose, param string, minRuneCount, maxRuneCount int, regexMatchers []string) error {
	valRuneCount := utf8.RuneCountInString(param)
	if valRuneCount < minRuneCount || valRuneCount > maxRuneCount {
		return newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers)
	}
	for _, regex := range regexMatchers {
		if matches, err := regexp.MatchString(regex, param); !matches || err != nil {
			if err != nil {
				return err
			}
			return newInvalidStringParamErr(paramPurpose, minRuneCount, maxRuneCount, regexMatchers)
		}
	}
	return nil
}

func validateEmail(email string) error {
	return validateStringParam("email", email, 6, 254, []string{`.+@.+\..+`})
}

func pwdsMatch(a, b []byte) bool {
	return bytes.Compare(a, b) == 0
}
