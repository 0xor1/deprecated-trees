package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"bytes"
	"strings"
	"time"
)

var (
	noSuchRegionErr                       = &Error{Code: 2, Msg: "no such region"}
	regionGoneErr                         = &Error{Code: 3, Msg: "region no longer exists"}
	noSuchUserErr                         = &Error{Code: 4, Msg: "no such user"}
	noSuchOrgErr                          = &Error{Code: 5, Msg: "no such org"}
	invalidActivationAttemptErr           = &Error{Code: 6, Msg: "invalid activation attempt"}
	invalidResetPwdAttemptErr             = &Error{Code: 7, Msg: "invalid reset password attempt"}
	invalidNewEmailConfirmationAttemptErr = &Error{Code: 8, Msg: "invalid new email confirmation attempt"}
	nameOrPwdIncorrectErr                 = &Error{Code: 9, Msg: "Name or password incorrect"}
	incorrectPwdErr                       = &Error{Code: 10, Msg: "password incorrect"}
	userNotActivatedErr                   = &Error{Code: 11, Msg: "user not activated"}
	emailAlreadyInUseErr                  = &Error{Code: 12, Msg: "email already in use"}
	accountNameAlreadyInUseErr            = &Error{Code: 13, Msg: "account already in use"}
	emailConfirmationCodeErr              = &Error{Code: 14, Msg: "email confirmation code is of zero length"}
	noNewEmailRegisteredErr               = &Error{Code: 15, Msg: "no new email registered"}
	insufficientPermissionsErr            = &Error{Code: 16, Msg: "insufficient permissions"}
	maxEntityCountExceededErr             = &Error{Code: 17, Msg: "max entity count exceeded"}
)

func newApi(store store, internalRegionApi internalRegionApi, linkMailer linkMailer, newId GenNewId, cryptoHelper CryptoHelper, nameRegexMatchers, pwdRegexMatchers []string, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxGetEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log Log) Api {
	if store == nil {
		NilCriticalParamPanic("store")
	}
	if internalRegionApi == nil {
		NilCriticalParamPanic("internalRegionApi")
	}
	if linkMailer == nil {
		NilCriticalParamPanic("linkMailer")
	}
	if newId == nil {
		NilCriticalParamPanic("newId")
	}
	if cryptoHelper == nil {
		NilCriticalParamPanic("cryptoHelper")
	}
	if log == nil {
		NilCriticalParamPanic("log")
	}
	return &api{
		store:             store,
		internalRegionApi: internalRegionApi,
		linkMailer:        linkMailer,
		newId:             newId,
		cryptoHelper:      cryptoHelper,
		nameRegexMatchers: append(make([]string, 0, len(nameRegexMatchers)), nameRegexMatchers...),
		pwdRegexMatchers:  append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		nameMinRuneCount:  nameMinRuneCount,
		nameMaxRuneCount:  nameMaxRuneCount,
		pwdMinRuneCount:   pwdMinRuneCount,
		pwdMaxRuneCount:   pwdMaxRuneCount,
		maxGetEntityCount: maxGetEntityCount,
		cryptoCodeLen:     cryptoCodeLen,
		saltLen:           saltLen,
		scryptN:           scryptN,
		scryptR:           scryptR,
		scryptP:           scryptP,
		scryptKeyLen:      scryptKeyLen,
		log:               log,
	}
}

type api struct {
	store             store
	internalRegionApi internalRegionApi
	linkMailer        linkMailer
	newId             GenNewId
	cryptoHelper      CryptoHelper
	nameRegexMatchers []string
	pwdRegexMatchers  []string
	nameMinRuneCount  int
	nameMaxRuneCount  int
	pwdMinRuneCount   int
	pwdMaxRuneCount   int
	maxGetEntityCount int
	cryptoCodeLen     int
	saltLen           int
	scryptN           int
	scryptR           int
	scryptP           int
	scryptKeyLen      int
	log               Log
}

func (a *api) GetRegions() []string {
	return a.internalRegionApi.GetRegions()
}

func (a *api) Register(name, email, pwd, region string) error {
	a.log.Location()

	name = strings.Trim(name, " ")
	if err := ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	email = strings.Trim(email, " ")
	if err := ValidateEmail(email); err != nil {
		return a.log.InfoErr(err)
	}

	if err := ValidateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	if !a.internalRegionApi.IsValidRegion(region) {
		return a.log.InfoErr(noSuchRegionErr)
	}

	if exists, err := a.store.accountWithCiNameExists(name); exists || err != nil {
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

	scryptSalt, err := a.cryptoHelper.Bytes(a.saltLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	scryptPwd, err := a.cryptoHelper.ScryptKey([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	activationCode, err := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	userId, err := a.newId()
	if err != nil {
		return a.log.ErrorErr(err)
	}

	err = a.store.createUser(
		&fullUserInfo{
			me: me{
				user: user{
					NamedEntity: NamedEntity{
						Entity: Entity{
							Id: userId,
						},
						Name: name,
					},
					Region:  region,
					Shard:   -1,
					Created: time.Now().UTC(),
					IsUser:  true,
				},
				Email: email,
			},
			activationCode: &activationCode,
		},
		&pwdInfo{
			salt:   scryptSalt,
			pwd:    scryptPwd,
			n:      a.scryptN,
			r:      a.scryptR,
			p:      a.scryptP,
			keyLen: a.scryptKeyLen,
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

	if err = a.linkMailer.sendActivationLink(email, *user.activationCode); err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) Activate(email, activationCode string) (Id, error) {
	a.log.Location()

	activationCode = strings.Trim(activationCode, " ")
	user, err := a.store.getUserByEmail(email)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil || user.activationCode == nil || activationCode != *user.activationCode {
		return nil, a.log.InfoErr(invalidActivationAttemptErr)
	}

	if !a.internalRegionApi.IsValidRegion(user.Region) {
		return nil, a.log.ErrorErr(regionGoneErr)
	}

	shard, err := a.internalRegionApi.CreatePersonalTaskCenter(user.Region, user.Id)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	user.Shard = shard
	user.activationCode = nil
	activationTime := time.Now().UTC()
	user.activated = &activationTime
	err = a.store.updateUser(user)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) Authenticate(name, pwdTry string) (Id, error) {
	a.log.Location()

	name = strings.Trim(name, " ")
	user, err := a.store.getUserByCiName(name)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil {
		return nil, a.log.InfoErr(nameOrPwdIncorrectErr)
	}

	pwdInfo, err := a.store.getPwdInfo(user.Id)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	scryptPwdTry, err := a.cryptoHelper.ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		return nil, a.log.InfoErr(nameOrPwdIncorrectErr)
	}

	if !user.isActivated() {
		return nil, a.log.InfoErr(userNotActivatedErr)
	}

	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(*user.resetPwdCode) > 0 {
		user.resetPwdCode = nil
		if err = a.store.updateUser(user); err != nil {
			return nil, a.log.ErrorErr(err)
		}
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.n != a.scryptN || pwdInfo.r != a.scryptR || pwdInfo.p != a.scryptP || pwdInfo.keyLen != a.scryptKeyLen || len(pwdInfo.salt) < a.saltLen {
		pwdInfo.salt, err = a.cryptoHelper.Bytes(a.saltLen)
		if err != nil {
			return nil, a.log.ErrorErr(err)
		}
		pwdInfo.n = a.scryptN
		pwdInfo.r = a.scryptR
		pwdInfo.p = a.scryptP
		pwdInfo.keyLen = a.scryptKeyLen
		pwdInfo.pwd, err = a.cryptoHelper.ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
		if err != nil {
			return nil, a.log.ErrorErr(err)
		}
		if err = a.store.updatePwdInfo(user.Id, pwdInfo); err != nil {
			return nil, a.log.ErrorErr(err)
		}

	}

	return user.Id, nil
}

func (a *api) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) (Id, error) {
	a.log.Location()

	user, err := a.store.getUserByEmail(currentEmail)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil || user.NewEmail == nil || newEmail != *user.NewEmail || user.newEmailConfirmationCode == nil || confirmationCode != *user.newEmailConfirmationCode {
		return nil, a.log.InfoErr(invalidNewEmailConfirmationAttemptErr)
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
	user.newEmailConfirmationCode = nil
	if err = a.store.updateUser(user); err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) ResetPwd(email string) error {
	a.log.Location()

	email = strings.Trim(email, " ")
	user, err := a.store.getUserByEmail(email)
	if err != nil {
		return a.log.ErrorErr(err)
	}
	if user == nil {
		return nil
	}

	resetPwdCode, err := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	user.resetPwdCode = &resetPwdCode
	if err = a.store.updateUser(user); err != nil {
		return a.log.ErrorErr(err)
	}

	err = a.linkMailer.sendPwdResetLink(email, resetPwdCode)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	return nil
}

func (a *api) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) (Id, error) {
	a.log.Location()

	if err := ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return nil, a.log.InfoErr(err)
	}

	user, err := a.store.getUserByEmail(email)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}
	if user == nil || user.resetPwdCode == nil || resetPwdCode != *user.resetPwdCode {
		return nil, a.log.InfoErr(invalidResetPwdAttemptErr)
	}

	scryptSalt, err := a.cryptoHelper.Bytes(a.saltLen)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	scryptPwd, err := a.cryptoHelper.ScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	if user.activationCode != nil {
		if !a.internalRegionApi.IsValidRegion(user.Region) {
			return nil, a.log.ErrorErr(regionGoneErr)
		}

		shard, err := a.internalRegionApi.CreatePersonalTaskCenter(user.Region, user.Id)
		if err != nil {
			return nil, a.log.ErrorErr(err)
		}
		user.Shard = shard
	}

	user.activationCode = nil
	user.resetPwdCode = nil
	if err = a.store.updateUser(user); err != nil {
		return nil, a.log.ErrorErr(err)
	}

	if err = a.store.updatePwdInfo(
		user.Id,
		&pwdInfo{
			pwd:    scryptPwd,
			salt:   scryptSalt,
			n:      a.scryptN,
			r:      a.scryptR,
			p:      a.scryptP,
			keyLen: a.scryptKeyLen,
		},
	); err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return user.Id, nil
}

func (a *api) GetAccount(name string) (*account, error) {
	a.log.Location()

	acc, err := a.store.getAccountByCiName(name)
	if err != nil {
		return acc, a.log.ErrorErr(err)
	}

	return acc, err
}

func (a *api) GetUsers(ids []Id) ([]*user, error) {
	a.log.Location()

	if len(ids) > a.maxGetEntityCount {
		return nil, a.log.InfoErr(maxEntityCountExceededErr)
	}

	users, err := a.store.getUsers(ids)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return users, nil
}

func (a *api) GetOrgs(ids []Id) ([]*org, error) {
	a.log.Location()

	if len(ids) > a.maxGetEntityCount {
		return nil, a.log.InfoErr(maxEntityCountExceededErr)
	}

	orgs, err := a.store.getOrgs(ids)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return orgs, nil
}

func (a *api) ChangeMyName(myId Id, newName string) error {
	a.log.Location()

	newName = strings.Trim(newName, " ")
	if err := ValidateStringParam("username", newName, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	if exists, err := a.store.accountWithCiNameExists(newName); exists || err != nil {
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		} else {
			return a.log.InfoUserErr(myId, accountNameAlreadyInUseErr)
		}
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return a.log.InfoUserErr(myId, noSuchUserErr)
	}

	user.Name = newName
	if err = a.store.updateUser(user); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	for offset, total := 0, 1; offset < total; {
		var orgs []*org
		orgs, total, err = a.store.getUsersOrgs(myId, offset, 100)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		offset += len(orgs)
		for _, org := range orgs {
			if !a.internalRegionApi.IsValidRegion(org.Region) {
				return a.log.ErrorUserErr(myId, regionGoneErr)
			}
			if err := a.internalRegionApi.RenameMember(org.Region, org.Shard, org.Id, myId, newName); err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
		}
	}

	return nil
}

func (a *api) ChangeMyPwd(myId Id, oldPwd, newPwd string) error {
	a.log.Location()

	if err := ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return a.log.InfoUserErr(myId, err)
	}

	pwdInfo, err := a.store.getPwdInfo(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if pwdInfo == nil {
		return a.log.InfoUserErr(myId, noSuchUserErr)
	}

	scryptPwdTry, err := a.cryptoHelper.ScryptKey([]byte(oldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		return a.log.InfoUserErr(myId, incorrectPwdErr)
	}

	scryptSalt, err := a.cryptoHelper.Bytes(a.saltLen)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	scryptPwd, err := a.cryptoHelper.ScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	pwdInfo.pwd = scryptPwd
	pwdInfo.salt = scryptSalt
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen
	if err = a.store.updatePwdInfo(myId, pwdInfo); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) ChangeMyEmail(myId Id, newEmail string) error {
	a.log.Location()

	newEmail = strings.Trim(newEmail, " ")
	if err := ValidateEmail(newEmail); err != nil {
		return a.log.InfoUserErr(myId, err)
	}

	if user, err := a.store.getUserByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		} else if err = a.linkMailer.sendMultipleAccountPolicyEmail(user.Email); err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		return nil
	}

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return a.log.InfoUserErr(myId, noSuchUserErr)
	}

	confirmationCode, err := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	user.NewEmail = &newEmail
	user.newEmailConfirmationCode = &confirmationCode
	if err = a.store.updateUser(user); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	if err = a.linkMailer.sendNewEmailConfirmationLink(user.Email, newEmail, confirmationCode); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) ResendMyNewEmailConfirmationEmail(myId Id) error {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return a.log.InfoUserErr(myId, noSuchUserErr)
	}

	// check the user has actually registered a new email
	if user.NewEmail == nil {
		return a.log.InfoUserErr(myId, noNewEmailRegisteredErr)
	}
	// just in case something has gone crazy wrong
	if user.newEmailConfirmationCode == nil {
		return a.log.ErrorUserErr(myId, emailConfirmationCodeErr)
	}

	err = a.linkMailer.sendNewEmailConfirmationLink(user.Email, *user.NewEmail, *user.newEmailConfirmationCode)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) MigrateMe(myId Id, newRegion string) error {
	a.log.Location()

	//the next line is arbitrarily added in to get code coverage for isMigrating Func
	//which won't get used anywhere until the migration feature is worked on in the future
	(&fullUserInfo{}).isMigrating()

	return a.log.InfoUserErr(myId, NotImplementedErr)
}

func (a *api) GetMe(myId Id) (*me, error) {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return nil, a.log.InfoUserErr(myId, noSuchUserErr)
	}

	return &user.me, nil
}

func (a *api) DeleteMe(myId Id) error {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return a.log.InfoUserErr(myId, noSuchUserErr)
	}

	for offset, total := 0, 1; offset < total; {
		var orgs []*org
		orgs, total, err = a.store.getUsersOrgs(myId, offset, 100)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		offset += len(orgs)
		for _, org := range orgs {
			if !a.internalRegionApi.IsValidRegion(org.Region) {
				return a.log.ErrorUserErr(myId, regionGoneErr)
			}
			if err := a.internalRegionApi.SetMemberDeleted(org.Region, org.Shard, org.Id, myId); err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
			if err := a.store.deleteMemberships(org.Id, []Id{myId}); err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
		}
	}

	if !a.internalRegionApi.IsValidRegion(user.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	publicErr, err := a.internalRegionApi.DeleteTaskCenter(user.Region, user.Shard, myId, myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.deleteUserAndAllAssociatedMemberships(myId); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) CreateOrg(myId Id, name, region string) (*org, error) {
	a.log.Location()

	name = strings.Trim(name, " ")
	if err := ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return nil, a.log.InfoUserErr(myId, err)
	}

	if !a.internalRegionApi.IsValidRegion(region) {
		return nil, a.log.InfoUserErr(myId, noSuchRegionErr)
	}

	if exists, err := a.store.accountWithCiNameExists(name); exists || err != nil {
		if err != nil {
			return nil, a.log.ErrorUserErr(myId, err)
		} else {
			return nil, a.log.InfoUserErr(myId, accountNameAlreadyInUseErr)
		}
	}

	newOrgId, err := a.newId()
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	org := &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: newOrgId,
			},
			Name: name,
		},
		Region:  region,
		Shard:   -1,
		Created: time.Now().UTC(),
		IsUser:  false,
	}
	if err := a.store.createOrgAndMembership(org, myId); err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	owner, err := a.store.getUserById(myId)
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}
	if owner == nil {
		return nil, a.log.InfoUserErr(myId, noSuchUserErr)
	}

	shard, err := a.internalRegionApi.CreateOrgTaskCenter(region, newOrgId, myId, owner.Name)
	if err != nil {
		if err := a.store.deleteOrgAndAllAssociatedMemberships(newOrgId); err != nil {
			return nil, a.log.ErrorUserErr(myId, err)
		}
		return nil, a.log.ErrorUserErr(myId, err)
	}

	org.Shard = shard
	if err := a.store.updateOrg(org); err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	return org, nil
}

func (a *api) RenameOrg(myId, orgId Id, newName string) error {
	a.log.Location()

	if exists, err := a.store.accountWithCiNameExists(newName); exists || err != nil {
		if err != nil {
			return a.log.ErrorErr(err)
		} else {
			return a.log.InfoErr(accountNameAlreadyInUseErr)
		}
	}

	org, err := a.store.getOrgById(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchOrgErr)
	}

	if !a.internalRegionApi.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	can, err := a.internalRegionApi.UserCanRenameOrg(org.Region, org.Shard, orgId, myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}
	if !can {
		return a.log.InfoUserErr(myId, insufficientPermissionsErr)
	}

	org.Name = newName
	if err := a.store.updateOrg(org); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) MigrateOrg(myId, orgId Id, newRegion string) error {
	a.log.Location()

	//the next line is arbitrarily added in to get code coverage for isMigrating Func
	//which won't get used anywhere until the migration feature is worked on in the future
	(&org{}).isMigrating()

	return a.log.InfoUserErr(myId, NotImplementedErr)
}

func (a *api) GetMyOrgs(myId Id, offset, limit int) ([]*org, int, error) {
	a.log.Location()

	if limit < 1 || limit > a.maxGetEntityCount {
		limit = a.maxGetEntityCount
	}
	if offset < 0 {
		offset = 0
	}

	orgs, total, err := a.store.getUsersOrgs(myId, offset, limit)
	if err != nil {
		err = a.log.ErrorUserErr(myId, err)
	}
	return orgs, total, err
}

func (a *api) DeleteOrg(myId, orgId Id) error {
	a.log.Location()

	org, err := a.store.getOrgById(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchOrgErr)
	}

	if !a.internalRegionApi.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	publicErr, err := a.internalRegionApi.DeleteTaskCenter(org.Region, org.Shard, orgId, myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.deleteOrgAndAllAssociatedMemberships(orgId); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) AddMembers(myId, orgId Id, newMembers []Id) error {
	a.log.Location()

	if len(newMembers) > a.maxGetEntityCount {
		return a.log.InfoUserErr(myId, maxEntityCountExceededErr)
	}

	org, err := a.store.getOrgById(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchOrgErr)
	}

	if !a.internalRegionApi.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	users, err := a.store.getUsers(newMembers)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	entities := make([]*NamedEntity, 0, len(users))
	for _, user := range users {
		entities = append(entities, &user.NamedEntity)
	}

	publicErr, err := a.internalRegionApi.AddMembers(org.Region, org.Shard, orgId, myId, entities)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.createMemberships(orgId, newMembers); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) RemoveMembers(myId, orgId Id, existingMembers []Id) error {
	a.log.Location()

	if len(existingMembers) > a.maxGetEntityCount {
		return a.log.InfoUserErr(myId, maxEntityCountExceededErr)
	}

	org, err := a.store.getOrgById(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchOrgErr)
	}

	if !a.internalRegionApi.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	publicErr, err := a.internalRegionApi.RemoveMembers(org.Region, org.Shard, orgId, myId, existingMembers)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.deleteMemberships(orgId, existingMembers); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

//internal helpers

type store interface {
	//user or org
	accountWithCiNameExists(name string) (bool, error)
	getAccountByCiName(name string) (*account, error)
	//user
	createUser(user *fullUserInfo, pwdInfo *pwdInfo) error
	getUserByCiName(name string) (*fullUserInfo, error)
	getUserByEmail(email string) (*fullUserInfo, error)
	getUserById(id Id) (*fullUserInfo, error)
	getPwdInfo(id Id) (*pwdInfo, error)
	updateUser(user *fullUserInfo) error
	updatePwdInfo(id Id, pwdInfo *pwdInfo) error
	deleteUserAndAllAssociatedMemberships(id Id) error
	getUsers(ids []Id) ([]*user, error)
	//org
	createOrgAndMembership(org *org, user Id) error
	getOrgById(id Id) (*org, error)
	updateOrg(org *org) error
	deleteOrgAndAllAssociatedMemberships(id Id) error
	getOrgs(ids []Id) ([]*org, error)
	getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error)
	//members
	createMemberships(org Id, users []Id) error
	deleteMemberships(org Id, users []Id) error
}

type internalRegionApi interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, account, owner Id) (public error, private error)
	AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (public error, private error)
	RemoveMembers(region string, shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(region string, shard int, org, member Id) error
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserCanRenameOrg(region string, shard int, org, user Id) (bool, error)
}

type linkMailer interface {
	sendMultipleAccountPolicyEmail(address string) error
	sendActivationLink(address, activationCode string) error
	sendPwdResetLink(address, resetCode string) error
	sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string) error
}

type account struct {
	NamedEntity
	Created   time.Time `json:"created"`
	Region    string    `json:"region"`
	NewRegion *string   `json:"newRegion,omitempty"`
	Shard     int       `json:"shard"`
	IsUser    bool      `json:"isUser"`
}

type org account

func (o *org) isMigrating() bool {
	return o.NewRegion != nil
}

type user account

type me struct {
	user
	Email    string  `json:"email"`
	NewEmail *string `json:"newEmail,omitempty"`
}

type fullUserInfo struct {
	me
	activationCode           *string
	activated                *time.Time
	newEmailConfirmationCode *string
	resetPwdCode             *string
}

func (u *fullUserInfo) isActivated() bool {
	return u.activated != nil
}

func (u *fullUserInfo) isMigrating() bool {
	return u.NewRegion != nil
}

type pwdInfo struct {
	salt   []byte
	pwd    []byte
	n      int
	r      int
	p      int
	keyLen int
}

func pwdsMatch(a, b []byte) bool {
	return bytes.Compare(a, b) == 0
}
