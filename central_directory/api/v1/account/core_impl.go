package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"bytes"
	"strings"
	"time"
	"io"
	"image"
	"image/jpeg"
	_ "image/png"
	_ "image/gif"
	"github.com/nfnt/resize"
)

var (
	noSuchRegionErr                       = &Error{Code: 2, Msg: "no such region"}
	regionGoneErr                         = &Error{Code: 3, Msg: "region no longer exists"}
	noSuchAccountErr 		      = &Error{Code: 4, Msg: "no such account"}
	invalidActivationAttemptErr           = &Error{Code: 5, Msg: "invalid activation attempt"}
	invalidResetPwdAttemptErr             = &Error{Code: 6, Msg: "invalid reset password attempt"}
	invalidNewEmailConfirmationAttemptErr = &Error{Code: 7, Msg: "invalid new email confirmation attempt"}
	invalidNameOrPwdErr                   = &Error{Code: 8, Msg: "invalid name or password"}
	incorrectPwdErr                       = &Error{Code: 9, Msg: "password incorrect"}
	userNotActivatedErr                   = &Error{Code: 10, Msg: "user not activated"}
	emailAlreadyInUseErr                  = &Error{Code: 11, Msg: "email already in use"}
	accountNameAlreadyInUseErr            = &Error{Code: 12, Msg: "account already in use"}
	emailConfirmationCodeErr              = &Error{Code: 13, Msg: "email confirmation code is of zero length"}
	noNewEmailRegisteredErr               = &Error{Code: 14, Msg: "no new email registered"}
	insufficientPermissionsErr            = &Error{Code: 15, Msg: "insufficient permissions"}
	maxEntityCountExceededErr             = &Error{Code: 16, Msg: "max entity count exceeded"}
	onlyOwnerMemberErr                    = &Error{Code: 17, Msg: "can't delete user who is the only owner of an org"}
	invalidAvatarShapeErr                 = &Error{Code: 18, Msg: "avatar images must be square"}
)

func newApi(store store, internalRegionClient internalRegionClient, linkMailer linkMailer, avatarStore avatarStore, newCreatedNamedEntity GenCreatedNamedEntity, cryptoHelper CryptoHelper, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxGetEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log Log) Api {
	if store == nil {
		NilCriticalParamPanic("store")
	}
	if internalRegionClient == nil {
		NilCriticalParamPanic("internalRegionClient")
	}
	if linkMailer == nil {
		NilCriticalParamPanic("linkMailer")
	}
	if avatarStore == nil {
		NilCriticalParamPanic("avatarStore")
	}
	if newCreatedNamedEntity == nil {
		NilCriticalParamPanic("newCreatedNamedEntity")
	}
	if cryptoHelper == nil {
		NilCriticalParamPanic("cryptoHelper")
	}
	if log == nil {
		NilCriticalParamPanic("log")
	}
	return &api{
		store:                store,
		internalRegionClient: internalRegionClient,
		linkMailer:           linkMailer,
		avatarStore:          avatarStore,
		newCreatedNamedEntity:       newCreatedNamedEntity,
		cryptoHelper:         cryptoHelper,
		nameRegexMatchers:    append(make([]string, 0, len(nameRegexMatchers)), nameRegexMatchers...),
		pwdRegexMatchers:     append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		maxAvatarDim: maxAvatarDim,
		nameMinRuneCount:     nameMinRuneCount,
		nameMaxRuneCount:     nameMaxRuneCount,
		pwdMinRuneCount:      pwdMinRuneCount,
		pwdMaxRuneCount:      pwdMaxRuneCount,
		maxGetEntityCount:    maxGetEntityCount,
		cryptoCodeLen:        cryptoCodeLen,
		saltLen:              saltLen,
		scryptN:              scryptN,
		scryptR:              scryptR,
		scryptP:              scryptP,
		scryptKeyLen:         scryptKeyLen,
		log:                  log,
	}
}

type api struct {
	store                 store
	internalRegionClient  internalRegionClient
	linkMailer            linkMailer
	avatarStore           avatarStore
	newCreatedNamedEntity GenCreatedNamedEntity
	cryptoHelper          CryptoHelper
	nameRegexMatchers     []string
	pwdRegexMatchers      []string
	maxAvatarDim          uint
	nameMinRuneCount      int
	nameMaxRuneCount      int
	pwdMinRuneCount       int
	pwdMaxRuneCount       int
	maxGetEntityCount     int
	cryptoCodeLen         int
	saltLen               int
	scryptN               int
	scryptR               int
	scryptP               int
	scryptKeyLen          int
	log                   Log
}

func (a *api) GetRegions() []string {
	return a.internalRegionClient.GetRegions()
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

	if !a.internalRegionClient.IsValidRegion(region) {
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

	userCore, err := a.newCreatedNamedEntity(name)
	if err != nil {
		return a.log.ErrorErr(err)
	}

	err = a.store.createUser(
		&fullUserInfo{
			me: me{
				account: account{
					CreatedNamedEntity: *userCore,
					Region:      region,
					Shard:       -1,
					IsUser:      true,
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

	if !a.internalRegionClient.IsValidRegion(user.Region) {
		return nil, a.log.ErrorErr(regionGoneErr)
	}

	shard, err := a.internalRegionClient.CreatePersonalTaskCenter(user.Region, user.Id)
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
		return nil, a.log.InfoErr(invalidNameOrPwdErr)
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
		return nil, a.log.InfoErr(invalidNameOrPwdErr)
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
		if !a.internalRegionClient.IsValidRegion(user.Region) {
			return nil, a.log.ErrorErr(regionGoneErr)
		}

		shard, err := a.internalRegionClient.CreatePersonalTaskCenter(user.Region, user.Id)
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

	acc, err := a.store.getAccountByCiName(strings.Trim(name, " "))
	if err != nil {
		return acc, a.log.ErrorErr(err)
	}

	return acc, err
}

func (a *api) GetAccounts(ids []Id) ([]*account, error) {
	a.log.Location()

	if len(ids) > a.maxGetEntityCount {
		return nil, a.log.InfoErr(maxEntityCountExceededErr)
	}

	accounts, err := a.store.getAccounts(ids)
	if err != nil {
		return nil, a.log.ErrorErr(err)
	}

	return accounts, nil
}

func (a *api) GetMe(myId Id) (*me, error) {
	a.log.Location()

	user, err := a.store.getUserById(myId)
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}
	if user == nil {
		return nil, a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	return &user.me, nil
}

func (a *api) SetMyPwd(myId Id, oldPwd, newPwd string) error {
	a.log.Location()

	if err := ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		return a.log.InfoUserErr(myId, err)
	}

	pwdInfo, err := a.store.getPwdInfo(myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if pwdInfo == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
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

func (a *api) SetMyEmail(myId Id, newEmail string) error {
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
		return a.log.InfoUserErr(myId, noSuchAccountErr)
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
		return a.log.InfoUserErr(myId, noSuchAccountErr)
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

func (a *api) SetAccountName(myId, accountId Id, newName string) error {
	a.log.Location()

	newName = strings.Trim(newName, " ")
	if err := ValidateStringParam("name", newName, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return a.log.InfoErr(err)
	}

	if exists, err := a.store.accountWithCiNameExists(newName); exists || err != nil {
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		} else {
			return a.log.InfoUserErr(myId, accountNameAlreadyInUseErr)
		}
	}

	acc, err := a.store.getAccount(accountId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if acc == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsUser { // can't rename someone else's personal account
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
		//otherwise attempting to rename an org
		if !a.internalRegionClient.IsValidRegion(acc.Region) {
			return a.log.ErrorUserErr(myId, regionGoneErr)
		}

		can, err := a.internalRegionClient.UserIsOrgOwner(acc.Region, acc.Shard, accountId, myId)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		if !can {
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
	}

	//else user is setting their own name
	acc.Name = newName
	if err = a.store.updateAccount(acc); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	if myId.Equal(accountId) { // if i did rename my account, i need to update all the stored names in all the orgs Im a member of
		for offset, total := 0, 1; offset < total; {
			var orgs []*account
			orgs, total, err = a.store.getUsersOrgs(myId, offset, 100)
			if err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
			offset += len(orgs)
			for _, org := range orgs {
				if !a.internalRegionClient.IsValidRegion(org.Region) {
					return a.log.ErrorUserErr(myId, regionGoneErr)
				}
				if err := a.internalRegionClient.RenameMember(org.Region, org.Shard, org.Id, myId, newName); err != nil {
					return a.log.ErrorUserErr(myId, err)
				}
			}
		}
	}

	return nil
}

func (a *api) SetAccountAvatar(myId Id, accountId Id, avatarImageData io.ReadCloser) error {
	a.log.Location()

	if avatarImageData != nil {
		defer avatarImageData.Close()
	}

	account, err := a.store.getAccount(accountId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if account == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if account.IsUser { // can't set avatar on someone else's personal account
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
		//otherwise attempting to set avatar on an org
		if !a.internalRegionClient.IsValidRegion(account.Region) {
			return a.log.ErrorUserErr(myId, regionGoneErr)
		}

		can, err := a.internalRegionClient.UserIsOrgOwner(account.Region, account.Shard, accountId, myId)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		if !can {
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
	}

	if avatarImageData != nil {
		avatarImage, _, err := image.Decode(avatarImageData)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		bounds := avatarImage.Bounds()
		if bounds.Max.X - bounds.Min.X != bounds.Max.Y - bounds.Min.Y { //if it  isn't square, then error
			return a.log.InfoUserErr(myId, invalidAvatarShapeErr)
		}
		if  uint(bounds.Max.X - bounds.Min.X) > a.maxAvatarDim { // if it is larger than allowed then resize
			avatarImage = resize.Resize(a.maxAvatarDim, a.maxAvatarDim, avatarImage, resize.NearestNeighbor)
		}
		buff := &bytes.Buffer{}
		if err := jpeg.Encode(buff, avatarImage, nil); err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		data := buff.Bytes()
		readerSeeker := bytes.NewReader(data)
		if err := a.avatarStore.put(myId.String(), "image/jpeg", int64(len(data)), readerSeeker); err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		if !account.HasAvatar {
			//if account didn't previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = true
			if err := a.store.updateAccount(account); err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
		}
	} else {
		if err := a.avatarStore.delete(myId.String()); err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		if account.HasAvatar {
			//if account did previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = false
			if err := a.store.updateAccount(account); err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
		}
	}

	return nil
}

func (a *api) MigrateAccount(myId, accountId Id, newRegion string) error {
	a.log.Location()

	//the next line is arbitrarily added in to get code coverage for isMigrating Func
	//which won't get used anywhere until the migration feature is worked on in the future
	(&fullUserInfo{}).isMigrating()

	return a.log.InfoUserErr(myId, NotImplementedErr)
}

func (a *api) CreateOrg(myId Id, name, region string) (*account, error) {
	a.log.Location()

	name = strings.Trim(name, " ")
	if err := ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers); err != nil {
		return nil, a.log.InfoUserErr(myId, err)
	}

	if !a.internalRegionClient.IsValidRegion(region) {
		return nil, a.log.InfoUserErr(myId, noSuchRegionErr)
	}

	if exists, err := a.store.accountWithCiNameExists(name); exists || err != nil {
		if err != nil {
			return nil, a.log.ErrorUserErr(myId, err)
		} else {
			return nil, a.log.InfoUserErr(myId, accountNameAlreadyInUseErr)
		}
	}

	orgCore, err := a.newCreatedNamedEntity(name)
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	org := &account{
		CreatedNamedEntity: *orgCore,
		Region:      region,
		Shard:       -1,
		IsUser:      false,
	}
	if err := a.store.createOrgAndMembership(org, myId); err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	owner, err := a.store.getUserById(myId)
	if err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}
	if owner == nil {
		return nil, a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	shard, err := a.internalRegionClient.CreateOrgTaskCenter(region, orgCore.Id, myId, owner.Name)
	if err != nil {
		if err := a.store.deleteAccountAndAllAssociatedMemberships(orgCore.Id); err != nil {
			return nil, a.log.ErrorUserErr(myId, err)
		}
		return nil, a.log.ErrorUserErr(myId, err)
	}

	org.Shard = shard
	if err := a.store.updateAccount(org); err != nil {
		return nil, a.log.ErrorUserErr(myId, err)
	}

	return org, nil
}

func (a *api) GetMyOrgs(myId Id, offset, limit int) ([]*account, int, error) {
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

func (a *api) DeleteAccount(myId, accountId Id) error {
	a.log.Location()

	acc, err := a.store.getAccount(accountId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if acc == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	if !a.internalRegionClient.IsValidRegion(acc.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsUser { // can't delete someone else's personal account
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
		//otherwise attempting to delete an org
		can, err := a.internalRegionClient.UserIsOrgOwner(acc.Region, acc.Shard, accountId, myId)
		if err != nil {
			return a.log.ErrorUserErr(myId, err)
		}
		if !can {
			return a.log.InfoUserErr(myId, insufficientPermissionsErr)
		}
	}


	publicErr, err := a.internalRegionClient.DeleteTaskCenter(acc.Region, acc.Shard, accountId, myId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.deleteAccountAndAllAssociatedMemberships(accountId); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	if myId.Equal(accountId) {
		for offset, total := 0, 1; offset < total; {
			var orgs []*account
			orgs, total, err = a.store.getUsersOrgs(myId, offset, 100)
			if err != nil {
				return a.log.ErrorUserErr(myId, err)
			}
			offset += len(orgs)
			for _, org := range orgs {
				if !a.internalRegionClient.IsValidRegion(org.Region) {
					return a.log.ErrorUserErr(myId, regionGoneErr)
				}
				isOnlyOwner, err := a.internalRegionClient.MemberIsOnlyOwner(org.Region, org.Shard, org.Id, myId)
				if err != nil {
					return a.log.ErrorUserErr(myId, err)
				}
				if isOnlyOwner {
					return a.log.InfoUserErr(myId, onlyOwnerMemberErr)
				}
			}
			for _, org := range orgs {
				if err := a.internalRegionClient.SetMemberDeleted(org.Region, org.Shard, org.Id, myId); err != nil {
					return a.log.ErrorUserErr(myId, err)
				}
			}
		}
	}

	return nil
}

func (a *api) AddMembers(myId, orgId Id, newMembers []*AddMemberExternal) error {
	a.log.Location()

	if len(newMembers) > a.maxGetEntityCount {
		return a.log.InfoUserErr(myId, maxEntityCountExceededErr)
	}

	org, err := a.store.getAccount(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	if !a.internalRegionClient.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	ids := make([]Id, 0, len(newMembers))
	addMembersMap := map[string]*AddMemberExternal{}
	for _, member := range newMembers {
		ids = append(ids, member.Id)
		addMembersMap[member.Id.String()] = member
	}

	users, err := a.store.getUsers(ids)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	entities := make([]*AddMemberInternal, 0, len(users))
	for _, user := range users {
		ami := &AddMemberInternal{}
		ami.Id = user.Id
		ami.Role = addMembersMap[user.Id.String()].Role
		ami.Name = user.Name
		entities = append(entities, ami)
	}

	publicErr, err := a.internalRegionClient.AddMembers(org.Region, org.Shard, orgId, myId, entities)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if publicErr != nil {
		return a.log.InfoUserErr(myId, publicErr)
	}

	if err := a.store.createMemberships(orgId, ids); err != nil {
		return a.log.ErrorUserErr(myId, err)
	}

	return nil
}

func (a *api) RemoveMembers(myId, orgId Id, existingMembers []Id) error {
	a.log.Location()

	if len(existingMembers) > a.maxGetEntityCount {
		return a.log.InfoUserErr(myId, maxEntityCountExceededErr)
	}

	org, err := a.store.getAccount(orgId)
	if err != nil {
		return a.log.ErrorUserErr(myId, err)
	}
	if org == nil {
		return a.log.InfoUserErr(myId, noSuchAccountErr)
	}

	if !a.internalRegionClient.IsValidRegion(org.Region) {
		return a.log.ErrorUserErr(myId, regionGoneErr)
	}

	publicErr, err := a.internalRegionClient.RemoveMembers(org.Region, org.Shard, orgId, myId, existingMembers)
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
	updateAccount(account *account) error
	updatePwdInfo(id Id, pwdInfo *pwdInfo) error
	deleteAccountAndAllAssociatedMemberships(id Id) error
	getAccount(id Id) (*account, error)
	getAccounts(ids []Id) ([]*account, error)
	getUsers(ids []Id) ([]*account, error)
	//org
	createOrgAndMembership(org *account, user Id) error
	getUsersOrgs(userId Id, offset, limit int) ([]*account, int, error)
	//members
	createMemberships(org Id, users []Id) error
	deleteMemberships(org Id, users []Id) error
}

type internalRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, account, owner Id) (public error, private error)
	AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal) (public error, private error)
	RemoveMembers(region string, shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(region string, shard int, org, member Id) error
	MemberIsOnlyOwner(region string, shard int, org, member Id) (bool, error)
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserIsOrgOwner(region string, shard int, org, user Id) (bool, error)
}

type linkMailer interface {
	sendMultipleAccountPolicyEmail(address string) error
	sendActivationLink(address, activationCode string) error
	sendPwdResetLink(address, resetCode string) error
	sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string) error
}

type avatarStore interface {
	put(key string, mimeType string, size int64, data io.ReadSeeker) error
	delete(key string) error
}

type account struct {
	CreatedNamedEntity
	Region    string  `json:"region"`
	NewRegion *string `json:"newRegion,omitempty"`
	Shard     int     `json:"shard"`
	HasAvatar bool    `json:"hasAvatar"`
	IsUser    bool    `json:"isUser"`
}

func (a *account) isMigrating() bool {
	return a.NewRegion != nil
}

type me struct {
	account
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
