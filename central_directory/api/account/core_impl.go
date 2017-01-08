package account

import (
	"bitbucket.org/robsix/task_center/misc"
	"bytes"
	"errors"
	"fmt"
	. "github.com/pborman/uuid"
	"github.com/uber-go/zap"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	nilStoreErr                       = errors.New("nil store")
	nilInternalRegionApisErr          = errors.New("nil internalRegionApis")
	nilLinkMailerErr                  = errors.New("nil linkMailer")
	nilGenNewIdErr                    = errors.New("nil genNewId")
	nilGenCryptoBytesErr              = errors.New("nil genCryptoBytes")
	nilGenCryptoUrlSafeStringErr      = errors.New("nil nilGenCryptoUrlSafeString")
	nilGenScryptKeyErr                = errors.New("nil nilGenScryptKey")
	nilLogErr                         = errors.New("nil log")
	noSuchRegionErr                   = errors.New("no such region")
	userRegionGoneErr                 = errors.New("user registered region no longer exists at activation time")
	noSuchUserErr                     = errors.New("no such user")
	noSuchActivationCodeErr           = errors.New("no such activation code")
	noSuchNewEmailConfirmationCodeErr = errors.New("no such new email confirmation code")
	incorrectPwdErr                   = errors.New("password incorrect")
	userNotActivated                  = errors.New("user not activated")
	emailAlreadyInUseErr              = errors.New("email already in use")
	accountNameAlreadyInUseErr        = errors.New("account already in use")
	emailConfirmationCodeErr          = errors.New("email confirmation code is of zero length")
	newEmailErr                       = errors.New("newEmail is of zero length")
	newEmailConfirmationErr           = errors.New("new email and confirmation code do not match those recorded")
)

type invalidStringParamErr struct {
	paramPurpose  string
	minRuneCount  int
	maxRuneCount  int
	regexMatchers []string
}

func (e *invalidStringParamErr) Error() string {
	return fmt.Sprintf("%s must be between %d and %d utf8 characters long and match all regexs %v", e.paramPurpose, e.minRuneCount, e.maxRuneCount, e.regexMatchers)
}

func newApi(store store, internalRegionApis map[string]internalRegionApi, linkMailer linkMailer, genNewId misc.GenNewId, genCryptoBytes misc.GenCryptoBytes, genCryptoUrlSafeString misc.GenCryptoUrlSafeString, genScryptKey misc.GenScryptKey, nameRegexMatchers, pwdRegexMatchers []string, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log misc.Log) (Api, error) {
	if store == nil {
		return nil, nilStoreErr
	}
	if internalRegionApis == nil {
		return nil, nilInternalRegionApisErr
	}
	if linkMailer == nil {
		return nil, nilLinkMailerErr
	}
	if genNewId == nil {
		return nil, nilGenNewIdErr
	}
	if genCryptoBytes == nil {
		return nil, nilGenCryptoBytesErr
	}
	if genCryptoUrlSafeString == nil {
		return nil, nilGenCryptoUrlSafeStringErr
	}
	if genScryptKey == nil {
		return nil, nilGenScryptKeyErr
	}
	if log == nil {
		return nil, nilLogErr
	}
	return &api{
		store:                  store,
		internalRegionApis:     internalRegionApis,
		linkMailer:             linkMailer,
		genNewId:               genNewId,
		genCryptoBytes:         genCryptoBytes,
		genCryptoUrlSafeString: genCryptoUrlSafeString,
		genScryptKey:           genScryptKey,
		nameRegexMatchers:      append(make([]string, 0, len(nameRegexMatchers)), nameRegexMatchers...),
		pwdRegexMatchers:       append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		nameMinRuneCount:       nameMinRuneCount,
		nameMaxRuneCount:       nameMaxRuneCount,
		pwdMinRuneCount:        pwdMinRuneCount,
		pwdMaxRuneCount:        pwdMaxRuneCount,
		maxSearchLimitResults:  maxSearchLimitResults,
		cryptoCodeLen:          cryptoCodeLen,
		saltLen:                saltLen,
		scryptN:                scryptN,
		scryptR:                scryptR,
		scryptP:                scryptP,
		scryptKeyLen:           scryptKeyLen,
		log:                    log,
	}, nil
}

type api struct {
	store                  store
	internalRegionApis     map[string]internalRegionApi
	linkMailer             linkMailer
	genNewId               misc.GenNewId
	genCryptoBytes         misc.GenCryptoBytes
	genCryptoUrlSafeString misc.GenCryptoUrlSafeString
	genScryptKey           misc.GenScryptKey
	nameRegexMatchers      []string
	pwdRegexMatchers       []string
	nameMinRuneCount       int
	nameMaxRuneCount       int
	pwdMinRuneCount        int
	pwdMaxRuneCount        int
	maxSearchLimitResults  int
	cryptoCodeLen          int
	saltLen                int
	scryptN                int
	scryptR                int
	scryptP                int
	scryptKeyLen           int
	log                    misc.Log
}

func (a *api) Register(name, email, pwd, region string) error {
	a.log.Location()

	name = strings.Trim(name, " ")
	if err := validateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	email = strings.Trim(email, " ")
	if err := validateEmail(email); err != nil {
		return a.log.InfoErr(err)
	}

	if err := validateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	if _, exists := a.internalRegionApis[region]; !exists {
		return a.log.InfoErr(noSuchRegionErr)
	}

	if account, err := a.store.getAccountByName(name); account != nil || err != nil {
		if err != nil {
			return a.log.ErrorErr(err)
		} else {
			return a.log.InfoErr(accountNameAlreadyInUseErr)
		}
	}

	if user, err := a.store.getUserByEmail(email); user != nil || err != nil {
		if err != nil {
			return a.log.ErrorErr(err)
		} else if err = a.linkMailer.sendMultipleAccountPolicyEmail(user.Email); err != nil {
			return a.log.ErrorErr(err)
		}
		return nil
	}

	scryptSalt, err := a.genCryptoBytes(a.saltLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	scryptPwd, err := a.genScryptKey([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	activationCode, err := a.genCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	userId, err := a.genNewId()
	if err != nil {
		return a.log.ErrorErr(err)
	}

	err = a.store.createUser(
		&fullUserInfo{
			me: me{
				user: user{
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
		return a.log.ErrorErr(err)
	}

	if err = a.linkMailer.sendActivationLink(email, activationCode); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) ResendActivationEmail(email string) error {
	a.log.Location()

	email = strings.Trim(email, " ")
	user, err := a.store.getUserByEmail(email)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil || user.isActivated() {
		return nil
	}

	if err = a.linkMailer.sendActivationLink(email, *user.ActivationCode); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) Activate(activationCode string) (UUID, error) {
	a.log.Location()

	activationCode = strings.Trim(activationCode, " ")
	user, err := a.store.getUserByActivationCode(activationCode)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(noSuchActivationCodeErr)
	}

	internalRegionApi, exists := a.internalRegionApis[user.Region]
	if !exists {
		return nil, a.log.ErrorErr(userRegionGoneErr)
	}

	shard, err := internalRegionApi.CreatePersonalTaskCenter(user.Id)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	user.Shard = shard
	user.ActivationCode = nil
	activationTime := time.Now().UTC()
	user.Activated = &activationTime
	err = a.store.updateUser(user)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) Authenticate(name, pwdTry string) (UUID, error) {
	a.log.Location()

	name = strings.Trim(name, " ")
	user, err := a.store.getUserByName(name)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(noSuchUserErr)
	}
	if !user.isActivated() {
		return nil, a.log.InfoErr(userNotActivated)
	}

	pwdInfo, err := a.store.getPwdInfo(user.Id)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	scryptPwdTry, err := a.genScryptKey([]byte(pwdTry), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	if !pwdsMatch(pwdInfo.Pwd, scryptPwdTry) {
		return nil, a.log.InfoErr(incorrectPwdErr)
	}

	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(*user.ResetPwdCode) > 0 {
		user.ResetPwdCode = nil
		if err = a.store.updateUser(user); err != nil {
			return nil, a.log.ErrorErr(err)
		}
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.N != a.scryptN || pwdInfo.R != a.scryptR || pwdInfo.P != a.scryptP || pwdInfo.KeyLen != a.scryptKeyLen || len(pwdInfo.Salt) < a.saltLen {
		pwdInfo.Salt, err = a.genCryptoBytes(a.saltLen)
		if err != nil {
			return nil, a.log.ErrorErr(err)
		}
		pwdInfo.N = a.scryptN
		pwdInfo.R = a.scryptR
		pwdInfo.P = a.scryptP
		pwdInfo.KeyLen = a.scryptKeyLen
		pwdInfo.Pwd, err = a.genScryptKey([]byte(pwdTry), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
		if err != nil {
			return nil, a.log.ErrorErr(err)
		}
		if err = a.store.updatePwdInfo(user.Id, pwdInfo); err != nil {
			return nil, a.log.ErrorErr(err)
		}

	}

	return user.Id, nil
}

func (a *api) ConfirmNewEmail(newEmail string, confirmationCode string) (UUID, error) {
	a.log.Location()

	user, err := a.store.getUserByNewEmailConfirmationCode(confirmationCode)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(noSuchNewEmailConfirmationCodeErr)
	}

	if *user.NewEmail != newEmail || *user.NewEmailConfirmationCode != confirmationCode {
		return nil, a.log.InfoErr(newEmailConfirmationErr)
	}

	if user, err := a.store.getUserByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			return nil, a.log.ErrorErr(err)
		} else {
			return nil, a.log.InfoErr(emailAlreadyInUseErr)
		}
	}

	user.Email = newEmail
	user.NewEmail = nil
	user.NewEmailConfirmationCode = nil
	if err = a.store.updateUser(user); err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) ResetPwd(email string) error {
	a.log.Location()

	user, err := a.store.getUserByEmail(email)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil {
		return a.log.InfoErr(noSuchUserErr)
	}

	resetPwdCode, err := a.genCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.InfoErr(err)
	}

	user.ResetPwdCode = &resetPwdCode
	if err = a.store.updateUser(user); err != nil {
		return a.log.InfoErr(err)
	}

	err = a.linkMailer.sendPwdResetLink(email, resetPwdCode)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error) {
	a.log.Location()

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return nil, a.log.InfoErr(err)
	}

	user, err := a.store.getUserByResetPwdCode(resetPwdCode)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(noSuchUserErr)
	}

	scryptSalt, err := a.genCryptoBytes(a.saltLen)
	if err != nil {
		return nil, a.log.InfoErr(err)
	}

	scryptPwd, err := a.genScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	user.ActivationCode = nil
	user.ResetPwdCode = nil
	if err = a.store.updateUser(user); err != nil {
		return nil, a.log.ErrorErr(err)
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
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) GetUsers(ids []UUID) ([]*user, error) {
	a.log.Location()

	users, err := a.store.getUsers(ids)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return users, err
}

func (a *api) SearchUsers(search string, limit int) ([]*user, error) {
	a.log.Location()

	if limit < 1 || limit > a.maxSearchLimitResults {
		limit = a.maxSearchLimitResults
	}

	search = strings.Trim(search, " ")
	if err := validateStringParam("search", search, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return nil, a.log.InfoErr(err)
	}

	users, err := a.store.searchUsers(search, limit)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return users, nil
}

func (a *api) GetOrgs(ids []UUID) ([]*org, error) {
	a.log.Location()

	orgs, err := a.store.getOrgs(ids)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return orgs, nil
}

func (a *api) SearchOrgs(search string, limit int) ([]*org, error) {
	a.log.Location()

	if limit < 1 || limit > a.maxSearchLimitResults {
		limit = a.maxSearchLimitResults
	}

	search = strings.Trim(search, " ")
	if err := validateStringParam("search", search, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return nil, a.log.InfoErr(err)
	}

	orgs, err := a.store.searchOrgs(search, limit)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return orgs, nil
}

func (a *api) ChangeMyName(myId UUID, newUsername string) error {
	a.log.Location()

	newUsername = strings.Trim(newUsername, " ")
	if err := validateStringParam("username", newUsername, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	if user, err := a.store.getUserByName(newUsername); user != nil || err != nil {
		if err != nil {
			return a.log.ErrorErr(err)
		} else {
			return a.log.InfoErr(accountNameAlreadyInUseErr)
		}
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil {
		return a.log.InfoErr(noSuchUserErr)
	}

	user.Name = newUsername
	if err = a.store.updateUser(user); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) ChangeMyEmail(myId UUID, newEmail string) error {
	a.log.Location()

	newEmail = strings.Trim(newEmail, " ")
	if err := validateEmail(newEmail); err != nil {
		return a.log.InfoErr(err)
	}

	if user, err := a.store.getUserByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			return a.log.ErrorErr(err)
		} else if err = a.linkMailer.sendMultipleAccountPolicyEmail(user.Email); err != nil {
			return a.log.ErrorErr(err)
		}
		return nil
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil {
		return a.log.InfoErr(noSuchUserErr)
	}

	confirmationCode, err := a.genCryptoUrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	user.NewEmail = &newEmail
	user.NewEmailConfirmationCode = &confirmationCode
	if err = a.store.updateUser(user); err != nil {
		return a.log.ErrorErr(err)
	}

	if err = a.linkMailer.sendNewEmailConfirmationLink(newEmail, confirmationCode); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) ResendMyNewEmailConfirmationEmail(myId UUID) error {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil {
		return a.log.InfoErr(noSuchUserErr)
	}

	// check the user has actually registered a new email
	if len(*user.NewEmail) == 0 {
		return a.log.ErrorErr(newEmailErr)
	}
	// just in case something has gone crazy wrong
	if len(*user.NewEmailConfirmationCode) == 0 {
		return a.log.ErrorErr(emailConfirmationCodeErr)
	}

	err = a.linkMailer.sendNewEmailConfirmationLink(*user.NewEmail, *user.NewEmailConfirmationCode)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) ChangeMyPwd(myId UUID, oldPwd, newPwd string) error {
	a.log.Location()

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	pwdInfo, err := a.store.getPwdInfo(myId)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if pwdInfo == nil {
		return a.log.InfoErr(noSuchUserErr)
	}

	scryptPwdTry, err := a.genScryptKey([]byte(oldPwd), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	if !pwdsMatch(pwdInfo.Pwd, scryptPwdTry) {
		return a.log.InfoErr(incorrectPwdErr)
	}

	scryptSalt, err := a.genCryptoBytes(a.saltLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	scryptPwd, err := a.genScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	pwdInfo.Pwd = scryptPwd
	pwdInfo.Salt = scryptSalt
	pwdInfo.N = a.scryptN
	pwdInfo.R = a.scryptR
	pwdInfo.P = a.scryptP
	pwdInfo.KeyLen = a.scryptKeyLen
	if err = a.store.updatePwdInfo(myId, pwdInfo); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) MigrateMe(myId UUID, newRegion string) error {
	a.log.Location()

	//TODO

	return nil
}

func (a *api) GetMe(myId UUID) (*me, error) {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(noSuchUserErr)
	}

	return &user.me, nil
}

func (a *api) DeleteMe(id UUID) error {
	a.log.Location()

	if err := a.store.deleteUser(id); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) CreateOrg(myId UUID, name, region string) (*org, error) {
	a.log.Location()

	//TODO

	return nil, nil
}

func (a *api) RenameOrg(myId, orgId UUID, newName string) error {
	a.log.Location()

	//TODO

	return nil
}

func (a *api) MigrateOrg(myId, orgId UUID, newRegion string) error {
	a.log.Location()

	//TODO

	return nil
}

func (a *api) GetMyOrgs(myId UUID, limit int) ([]*org, error) {
	a.log.Location()

	//TODO

	return nil, nil
}

func (a *api) DeleteOrg(myId, orgId UUID) error {
	a.log.Location()

	//TODO

	return nil
}

func (a *api) AddMembers(myId, orgId UUID, newMembers []UUID) error {
	a.log.Location()

	//TODO

	return nil
}

func (a *api) RemoveMembers(myId, orgId UUID, existingMembers []UUID) error {
	a.log.Location()

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
	getUsers(ids []UUID) ([]*user, error)
	searchUsers(search string, limit int) ([]*user, error)
	//org
	createOrg(org *org) error
	getOrgById(id UUID) (*org, error)
	getOrgByName(name string) (*org, error)
	updateOrg(org *org) error
	deleteOrg(id UUID) error
	getOrgs(ids []UUID) ([]*org, error)
	searchOrgs(search string, limit int) ([]*org, error)
	getUsersOrgs(userId UUID, limit int) ([]*org, error)
}

type internalRegionApi interface {
	CreatePersonalTaskCenter(userId UUID) (int, error)
	CreateOrgTaskCenter(ownerId, orgId UUID) (int, error)
	RenameMember(memberId, orgId UUID, newName string) error
}

type linkMailer interface {
	sendMultipleAccountPolicyEmail(address string) error
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

type org account

type user account

type me struct {
	user
	Email    string  `json:"email"`
	NewEmail *string `json:"newEmail,omitempty"`
}

func (a *account) isMigrating() bool {
	return len(a.NewRegion) != 0
}

type fullUserInfo struct {
	me
	ActivationCode           *string
	Activated                *time.Time
	NewEmailConfirmationCode *string
	ResetPwdCode             *string
}

func (u *fullUserInfo) isActivated() bool {
	return u.Activated != nil
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
	log misc.Log
}

func (l *logLinkMailer) sendMultipleAccountPolicyEmail(address string) error {
	l.log.Info(zap.String("address", address))
	return nil
}

func (l *logLinkMailer) sendActivationLink(address, activationCode string) error {
	l.log.Info(zap.String("address", address), zap.String("activationCode", activationCode))
	return nil
}

func (l *logLinkMailer) sendPwdResetLink(address, resetCode string) error {
	l.log.Info(zap.String("address", address), zap.String("resetCode", resetCode))
	return nil
}

func (l *logLinkMailer) sendNewEmailConfirmationLink(address, confirmationCode string) error {
	l.log.Info(zap.String("address", address), zap.String("confirmationCode", confirmationCode))
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
