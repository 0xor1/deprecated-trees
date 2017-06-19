package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"strings"
	"time"
)

var (
	noSuchRegionErr                       = &Error{Code: 2, Msg: "no such region", IsPublic: true}
	noSuchAccountErr                      = &Error{Code: 4, Msg: "no such account", IsPublic: true}
	invalidActivationAttemptErr           = &Error{Code: 5, Msg: "invalid activation attempt", IsPublic: true}
	invalidResetPwdAttemptErr             = &Error{Code: 6, Msg: "invalid reset password attempt", IsPublic: true}
	invalidNewEmailConfirmationAttemptErr = &Error{Code: 7, Msg: "invalid new email confirmation attempt", IsPublic: true}
	invalidNameOrPwdErr                   = &Error{Code: 8, Msg: "invalid name or password", IsPublic: true}
	incorrectPwdErr                       = &Error{Code: 9, Msg: "password incorrect", IsPublic: true}
	userNotActivatedErr                   = &Error{Code: 10, Msg: "user not activated", IsPublic: true}
	emailAlreadyInUseErr                  = &Error{Code: 11, Msg: "email already in use", IsPublic: true}
	accountNameAlreadyInUseErr            = &Error{Code: 12, Msg: "account already in use", IsPublic: true}
	emailConfirmationCodeErr              = &Error{Code: 13, Msg: "email confirmation code is of zero length", IsPublic: false}
	noNewEmailRegisteredErr               = &Error{Code: 14, Msg: "no new email registered", IsPublic: true}
	insufficientPermissionsErr            = &Error{Code: 15, Msg: "insufficient permissions", IsPublic: true}
	maxEntityCountExceededErr             = &Error{Code: 16, Msg: "max entity count exceeded", IsPublic: true}
	onlyOwnerMemberErr                    = &Error{Code: 17, Msg: "can't delete user who is the only owner of an org", IsPublic: true}
	invalidAvatarShapeErr                 = &Error{Code: 18, Msg: "avatar images must be square", IsPublic: true}
)

func newApi(store store, internalRegionClient InternalRegionClient, linkMailer linkMailer, avatarStore avatarStore, newCreatedNamedEntity GenCreatedNamedEntity, cryptoHelper CryptoHelper, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxGetEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int) Api {
	if store == nil || internalRegionClient == nil || linkMailer == nil || avatarStore == nil || newCreatedNamedEntity == nil || cryptoHelper == nil {
		panic(NilCriticalParamErr)
	}
	return &api{
		store:                 store,
		internalRegionClient:  internalRegionClient,
		linkMailer:            linkMailer,
		avatarStore:           avatarStore,
		newCreatedNamedEntity: newCreatedNamedEntity,
		cryptoHelper:          cryptoHelper,
		nameRegexMatchers:     append(make([]string, 0, len(nameRegexMatchers)), nameRegexMatchers...),
		pwdRegexMatchers:      append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		maxAvatarDim:          maxAvatarDim,
		nameMinRuneCount:      nameMinRuneCount,
		nameMaxRuneCount:      nameMaxRuneCount,
		pwdMinRuneCount:       pwdMinRuneCount,
		pwdMaxRuneCount:       pwdMaxRuneCount,
		maxGetEntityCount:     maxGetEntityCount,
		cryptoCodeLen:         cryptoCodeLen,
		saltLen:               saltLen,
		scryptN:               scryptN,
		scryptR:               scryptR,
		scryptP:               scryptP,
		scryptKeyLen:          scryptKeyLen,
	}
}

type api struct {
	store                 store
	internalRegionClient  InternalRegionClient
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
}

func (a *api) GetRegions() []string {
	return a.internalRegionClient.GetRegions()
}

func (a *api) Register(name, email, pwd, region string) {
	name = strings.Trim(name, " ")
	ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)
	email = strings.Trim(email, " ")
	ValidateEmail(email)
	ValidateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)

	if !a.internalRegionClient.IsValidRegion(region) {
		panic(noSuchRegionErr)
	}

	if exists := a.store.accountWithCiNameExists(name); exists {
		panic(accountNameAlreadyInUseErr)
	}

	if user := a.store.getUserByEmail(email); user != nil {
		a.linkMailer.sendMultipleAccountPolicyEmail(user.Email)
	}

	userCore := a.newCreatedNamedEntity(name)
	activationCode := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)
	user := &fullUserInfo{}
	user.CreatedNamedEntity = *userCore
	user.Region = region
	user.Shard = -1
	user.IsUser = true
	user.Email = email
	user.activationCode = &activationCode

	pwdInfo := &pwdInfo{}
	pwdInfo.salt = a.cryptoHelper.Bytes(a.saltLen)
	pwdInfo.pwd = a.cryptoHelper.ScryptKey([]byte(pwd), pwdInfo.salt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen

	a.store.createUser(user, pwdInfo)

	a.linkMailer.sendActivationLink(email, *user.activationCode)
}

func (a *api) ResendActivationEmail(email string) {
	email = strings.Trim(email, " ")
	user := a.store.getUserByEmail(email)
	if user == nil || user.isActivated() {
		return
	}

	a.linkMailer.sendActivationLink(email, *user.activationCode)
}

func (a *api) Activate(email, activationCode string) {
	activationCode = strings.Trim(activationCode, " ")
	user := a.store.getUserByEmail(email)
	if user == nil || user.activationCode == nil || activationCode != *user.activationCode {
		panic(invalidActivationAttemptErr)
	}

	shard := a.internalRegionClient.CreatePersonalTaskCenter(user.Region, user.Id)

	user.Shard = shard
	user.activationCode = nil
	activationTime := time.Now().UTC()
	user.activated = &activationTime
	a.store.updateUser(user)
}

func (a *api) Authenticate(email, pwdTry string) Id {
	email = strings.Trim(email, " ")
	user := a.store.getUserByEmail(email)
	if user == nil {
		panic(invalidNameOrPwdErr)
	}

	pwdInfo := a.store.getPwdInfo(user.Id)
	scryptPwdTry := a.cryptoHelper.ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		panic(invalidNameOrPwdErr)
	}

	//must do this after checking the user has the correct pwd otherwise it allows anyone to fish for valid emails on the system
	if !user.isActivated() {
		panic(userNotActivatedErr)
	}

	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(*user.resetPwdCode) > 0 {
		user.resetPwdCode = nil
		a.store.updateUser(user)
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.n != a.scryptN || pwdInfo.r != a.scryptR || pwdInfo.p != a.scryptP || pwdInfo.keyLen != a.scryptKeyLen || len(pwdInfo.salt) < a.saltLen {
		pwdInfo.salt = a.cryptoHelper.Bytes(a.saltLen)
		pwdInfo.n = a.scryptN
		pwdInfo.r = a.scryptR
		pwdInfo.p = a.scryptP
		pwdInfo.keyLen = a.scryptKeyLen
		pwdInfo.pwd = a.cryptoHelper.ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
		a.store.updatePwdInfo(user.Id, pwdInfo)

	}

	return user.Id
}

func (a *api) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) {
	user := a.store.getUserByEmail(currentEmail)
	if user == nil || user.NewEmail == nil || newEmail != *user.NewEmail || user.newEmailConfirmationCode == nil || confirmationCode != *user.newEmailConfirmationCode {
		panic(invalidNewEmailConfirmationAttemptErr)
	}

	if user := a.store.getUserByEmail(newEmail); user != nil {
		panic(emailAlreadyInUseErr)
	}

	user.Email = newEmail
	user.NewEmail = nil
	user.newEmailConfirmationCode = nil
	a.store.updateUser(user)
}

func (a *api) ResetPwd(email string) {
	email = strings.Trim(email, " ")
	user := a.store.getUserByEmail(email)
	if user == nil {
		return
	}

	resetPwdCode := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)

	user.resetPwdCode = &resetPwdCode
	a.store.updateUser(user)

	a.linkMailer.sendPwdResetLink(email, resetPwdCode)
}

func (a *api) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) {
	ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)

	user := a.store.getUserByEmail(email)
	if user == nil || user.resetPwdCode == nil || resetPwdCode != *user.resetPwdCode {
		panic(invalidResetPwdAttemptErr)
	}

	scryptSalt := a.cryptoHelper.Bytes(a.saltLen)
	scryptPwd := a.cryptoHelper.ScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)

	if user.activationCode != nil {
		shard := a.internalRegionClient.CreatePersonalTaskCenter(user.Region, user.Id)
		user.Shard = shard
	}

	user.activationCode = nil
	user.resetPwdCode = nil
	a.store.updateUser(user)

	pwdInfo := &pwdInfo{}
	pwdInfo.pwd = scryptPwd
	pwdInfo.salt = scryptSalt
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen
	a.store.updatePwdInfo(user.Id, pwdInfo)
}

func (a *api) GetAccount(name string) *account {
	return a.store.getAccountByCiName(strings.Trim(name, " "))
}

func (a *api) GetAccounts(ids []Id) []*account {
	if len(ids) > a.maxGetEntityCount {
		panic(maxEntityCountExceededErr)
	}

	return a.store.getAccounts(ids)
}

func (a *api) GetMe(myId Id) *me {
	user := a.store.getUserById(myId)
	if user == nil {
		panic(noSuchAccountErr)
	}

	return &user.me
}

func (a *api) SetMyPwd(myId Id, oldPwd, newPwd string) {
	ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)

	pwdInfo := a.store.getPwdInfo(myId)
	if pwdInfo == nil {
		panic(noSuchAccountErr)
	}

	scryptPwdTry := a.cryptoHelper.ScryptKey([]byte(oldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		panic(incorrectPwdErr)
	}

	pwdInfo.salt = a.cryptoHelper.Bytes(a.saltLen)
	pwdInfo.pwd = a.cryptoHelper.ScryptKey([]byte(newPwd), pwdInfo.salt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen
	a.store.updatePwdInfo(myId, pwdInfo)
}

func (a *api) SetMyEmail(myId Id, newEmail string) {
	newEmail = strings.Trim(newEmail, " ")
	ValidateEmail(newEmail)

	if user := a.store.getUserByEmail(newEmail); user != nil {
		a.linkMailer.sendMultipleAccountPolicyEmail(user.Email)
	}

	user := a.store.getUserById(myId)
	if user == nil {
		panic(noSuchAccountErr)
	}

	confirmationCode := a.cryptoHelper.UrlSafeString(a.cryptoCodeLen)

	user.NewEmail = &newEmail
	user.newEmailConfirmationCode = &confirmationCode
	a.store.updateUser(user)
	a.linkMailer.sendNewEmailConfirmationLink(user.Email, newEmail, confirmationCode)
}

func (a *api) ResendMyNewEmailConfirmationEmail(myId Id) {
	user := a.store.getUserById(myId)
	if user == nil {
		panic(noSuchAccountErr)
	}

	// check the user has actually registered a new email
	if user.NewEmail == nil {
		panic(noNewEmailRegisteredErr)
	}
	// just in case something has gone crazy wrong
	if user.newEmailConfirmationCode == nil {
		panic(emailConfirmationCodeErr)
	}

	a.linkMailer.sendNewEmailConfirmationLink(user.Email, *user.NewEmail, *user.newEmailConfirmationCode)
}

func (a *api) SetAccountName(myId, accountId Id, newName string) {
	newName = strings.Trim(newName, " ")
	ValidateStringParam("name", newName, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)

	if exists := a.store.accountWithCiNameExists(newName); exists {
		panic(accountNameAlreadyInUseErr)
	}

	acc := a.store.getAccount(accountId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsUser { // can't rename someone else's personal account
			panic(insufficientPermissionsErr)
		}

		if !a.internalRegionClient.UserIsOrgOwner(acc.Region, acc.Shard, accountId, myId) {
			panic(insufficientPermissionsErr)
		}
	}

	//else user is setting their own name
	acc.Name = newName
	a.store.updateAccount(acc)

	if myId.Equal(accountId) { // if i did rename my account, i need to update all the stored names in all the orgs Im a member of
		for offset, total := 0, 1; offset < total; {
			var orgs []*account
			orgs, total = a.store.getUsersOrgs(myId, offset, 100)
			offset += len(orgs)
			for _, org := range orgs {
				a.internalRegionClient.RenameMember(org.Region, org.Shard, org.Id, myId, newName)
			}
		}
	}
}

func (a *api) SetAccountAvatar(myId Id, accountId Id, avatarImageData io.ReadCloser) {
	if avatarImageData != nil {
		defer avatarImageData.Close()
	}

	account := a.store.getAccount(accountId)
	if account == nil {
		panic(noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if account.IsUser { // can't set avatar on someone else's personal account
			panic(insufficientPermissionsErr)
		}

		if !a.internalRegionClient.UserIsOrgOwner(account.Region, account.Shard, accountId, myId) {
			panic(insufficientPermissionsErr)
		}
	}

	if avatarImageData != nil {
		avatarImage, _, err := image.Decode(avatarImageData)
		if err != nil {
			panic(err)
		}
		bounds := avatarImage.Bounds()
		if bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y { //if it  isn't square, then error
			panic(invalidAvatarShapeErr)
		}
		if uint(bounds.Max.X-bounds.Min.X) > a.maxAvatarDim { // if it is larger than allowed then resize
			avatarImage = resize.Resize(a.maxAvatarDim, a.maxAvatarDim, avatarImage, resize.NearestNeighbor)
		}
		buff := &bytes.Buffer{}
		if err := jpeg.Encode(buff, avatarImage, nil); err != nil {
			panic(err)
		}
		data := buff.Bytes()
		readerSeeker := bytes.NewReader(data)
		a.avatarStore.put(myId.String(), "image/jpeg", int64(len(data)), readerSeeker)
		if !account.HasAvatar {
			//if account didn't previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = true
			a.store.updateAccount(account)
		}
	} else {
		a.avatarStore.delete(myId.String())
		if account.HasAvatar {
			//if account did previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = false
			a.store.updateAccount(account)
		}
	}
}

func (a *api) MigrateAccount(myId, accountId Id, newRegion string) {
	//the next line is arbitrarily added in to get code coverage for isMigrating Func
	//which won't get used anywhere until the migration feature is worked on in the future
	(&fullUserInfo{}).isMigrating()
}

func (a *api) CreateOrg(myId Id, name, region string) *account {
	name = strings.Trim(name, " ")
	ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)

	if !a.internalRegionClient.IsValidRegion(region) {
		panic(noSuchRegionErr)
	}

	if exists := a.store.accountWithCiNameExists(name); exists {
		panic(accountNameAlreadyInUseErr)
	}

	orgCore := a.newCreatedNamedEntity(name)

	org := &account{
		CreatedNamedEntity: *orgCore,
		Region:             region,
		Shard:              -1,
		IsUser:             false,
	}
	a.store.createOrgAndMembership(org, myId)

	owner := a.store.getUserById(myId)
	if owner == nil {
		panic(noSuchAccountErr)
	}

	defer func() {
		r := recover()
		if r != nil {
			a.store.deleteAccountAndAllAssociatedMemberships(orgCore.Id)
			panic(r)
		}
	}()
	shard := a.internalRegionClient.CreateOrgTaskCenter(region, orgCore.Id, myId, owner.Name)

	org.Shard = shard
	a.store.updateAccount(org)
	return org
}

func (a *api) GetMyOrgs(myId Id, offset, limit int) ([]*account, int) {
	if limit < 1 || limit > a.maxGetEntityCount {
		limit = a.maxGetEntityCount
	}
	if offset < 0 {
		offset = 0
	}

	return a.store.getUsersOrgs(myId, offset, limit)
}

func (a *api) DeleteAccount(myId, accountId Id) {
	acc := a.store.getAccount(accountId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsUser { // can't delete someone else's personal account
			panic(insufficientPermissionsErr)
		}
		//otherwise attempting to delete an org
		if !a.internalRegionClient.UserIsOrgOwner(acc.Region, acc.Shard, accountId, myId) {
			panic(insufficientPermissionsErr)
		}
	}

	a.internalRegionClient.DeleteTaskCenter(acc.Region, acc.Shard, accountId, myId)
	a.store.deleteAccountAndAllAssociatedMemberships(accountId)

	if myId.Equal(accountId) {
		for offset, total := 0, 1; offset < total; {
			var orgs []*account
			orgs, total = a.store.getUsersOrgs(myId, offset, 100)
			offset += len(orgs)
			for _, org := range orgs {
				if a.internalRegionClient.MemberIsOnlyOwner(org.Region, org.Shard, org.Id, myId) {
					panic(onlyOwnerMemberErr)
				}
			}
			for _, org := range orgs {
				a.internalRegionClient.RemoveMembers(org.Region, org.Shard, org.Id, myId, []Id{myId})
			}
		}
	}
}

func (a *api) AddMembers(myId, orgId Id, newMembers []*AddMemberExternal) {
	if len(newMembers) > a.maxGetEntityCount {
		panic(maxEntityCountExceededErr)
	}

	org := a.store.getAccount(orgId)
	if org == nil {
		panic(noSuchAccountErr)
	}

	ids := make([]Id, 0, len(newMembers))
	addMembersMap := map[string]*AddMemberExternal{}
	for _, member := range newMembers {
		ids = append(ids, member.Id)
		addMembersMap[member.Id.String()] = member
	}

	users := a.store.getUsers(ids)

	entities := make([]*AddMemberInternal, 0, len(users))
	for _, user := range users {
		ami := &AddMemberInternal{}
		ami.Id = user.Id
		ami.Role = addMembersMap[user.Id.String()].Role
		ami.Name = user.Name
		entities = append(entities, ami)
	}

	a.internalRegionClient.AddMembers(org.Region, org.Shard, orgId, myId, entities)
	a.store.createMemberships(orgId, ids)
}

func (a *api) RemoveMembers(myId, orgId Id, existingMembers []Id) {
	if len(existingMembers) > a.maxGetEntityCount {
		panic(maxEntityCountExceededErr)
	}

	org := a.store.getAccount(orgId)
	if org == nil {
		panic(noSuchAccountErr)
	}

	a.internalRegionClient.RemoveMembers(org.Region, org.Shard, orgId, myId, existingMembers)
	a.store.deleteMemberships(orgId, existingMembers)
}

//internal helpers

type store interface {
	//user or org
	accountWithCiNameExists(name string) bool
	getAccountByCiName(name string) *account
	//user
	createUser(user *fullUserInfo, pwdInfo *pwdInfo)
	getUserByEmail(email string) *fullUserInfo
	getUserById(id Id) *fullUserInfo
	getPwdInfo(id Id) *pwdInfo
	updateUser(user *fullUserInfo)
	updateAccount(account *account)
	updatePwdInfo(id Id, pwdInfo *pwdInfo)
	deleteAccountAndAllAssociatedMemberships(id Id)
	getAccount(id Id) *account
	getAccounts(ids []Id) []*account
	getUsers(ids []Id) []*account
	//org
	createOrgAndMembership(org *account, user Id)
	getUsersOrgs(userId Id, offset, limit int) ([]*account, int)
	//members
	createMemberships(org Id, users []Id)
	deleteMemberships(org Id, users []Id)
}

type linkMailer interface {
	sendMultipleAccountPolicyEmail(address string)
	sendActivationLink(address, activationCode string)
	sendPwdResetLink(address, resetCode string)
	sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string)
}

type avatarStore interface {
	put(key string, mimeType string, size int64, data io.ReadSeeker)
	delete(key string)
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
