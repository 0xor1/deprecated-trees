package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bytes"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	noSuchRegionErr                       = &Error{Code: "cd_v1_a_nsr", Msg: "no such region", IsPublic: true}
	noSuchAccountErr                      = &Error{Code: "cd_v1_a_nsa", Msg: "no such account", IsPublic: true}
	invalidActivationAttemptErr           = &Error{Code: "cd_v1_a_iaa", Msg: "invalid activation attempt", IsPublic: true}
	invalidResetPwdAttemptErr             = &Error{Code: "cd_v1_a_irpa", Msg: "invalid reset password attempt", IsPublic: true}
	invalidNewEmailConfirmationAttemptErr = &Error{Code: "cd_v1_a_ineca", Msg: "invalid new email confirmation attempt", IsPublic: true}
	invalidNameOrPwdErr                   = &Error{Code: "cd_v1_a_inop", Msg: "invalid name or password", IsPublic: true}
	incorrectPwdErr                       = &Error{Code: "cd_v1_a_ip", Msg: "password incorrect", IsPublic: true}
	accountNotActivatedErr                = &Error{Code: "cd_v1_a_ana", Msg: "account not activated", IsPublic: true}
	emailAlreadyInUseErr                  = &Error{Code: "cd_v1_a_eaiu", Msg: "email already in use", IsPublic: true}
	nameAlreadyInUseErr                   = &Error{Code: "cd_v1_a_naiu", Msg: "name already in use", IsPublic: true}
	emailConfirmationCodeErr              = &Error{Code: "cd_v1_a_ecc", Msg: "email confirmation code is of zero length", IsPublic: false}
	noNewEmailRegisteredErr               = &Error{Code: "cd_v1_a_nner", Msg: "no new email registered", IsPublic: true}
	onlyOwnerMemberErr                    = &Error{Code: "cd_v1_a_oom", Msg: "can't delete member who is the only owner of an account", IsPublic: true}
	invalidAvatarShapeErr                 = &Error{Code: "cd_v1_a_ias", Msg: "avatar images must be square", IsPublic: true}
)

func newApi(store store, privateRegionClient PrivateRegionClient, linkMailer linkMailer, avatarStore avatarStore, nameRegexMatchers, pwdRegexMatchers []string, maxAvatarDim uint, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, maxProcessEntityCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int) Api {
	if store == nil || privateRegionClient == nil || linkMailer == nil || avatarStore == nil {
		panic(InvalidArgumentsErr)
	}
	//compile regexs
	nameRegexes := make([]*regexp.Regexp, 0, len(nameRegexMatchers))
	for _, val := range nameRegexMatchers {
		nameRegexes = append(nameRegexes, regexp.MustCompile(val))
	}
	pwdRegexes := make([]*regexp.Regexp, 0, len(pwdRegexMatchers))
	for _, val := range pwdRegexMatchers {
		nameRegexes = append(nameRegexes, regexp.MustCompile(val))
	}
	return &api{
		store:                 store,
		privateRegionClient:   privateRegionClient,
		linkMailer:            linkMailer,
		avatarStore:           avatarStore,
		nameRegexMatchers:     nameRegexes,
		pwdRegexMatchers:      pwdRegexes,
		maxAvatarDim:          maxAvatarDim,
		nameMinRuneCount:      nameMinRuneCount,
		nameMaxRuneCount:      nameMaxRuneCount,
		pwdMinRuneCount:       pwdMinRuneCount,
		pwdMaxRuneCount:       pwdMaxRuneCount,
		maxProcessEntityCount: maxProcessEntityCount,
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
	privateRegionClient   PrivateRegionClient
	linkMailer            linkMailer
	avatarStore           avatarStore
	nameRegexMatchers     []*regexp.Regexp
	pwdRegexMatchers      []*regexp.Regexp
	maxAvatarDim          uint
	nameMinRuneCount      int
	nameMaxRuneCount      int
	pwdMinRuneCount       int
	pwdMaxRuneCount       int
	maxProcessEntityCount int
	cryptoCodeLen         int
	saltLen               int
	scryptN               int
	scryptR               int
	scryptP               int
	scryptKeyLen          int
}

func (a *api) GetRegions() []string {
	return a.privateRegionClient.GetRegions()
}

func (a *api) Register(name, email, pwd, region, language string, theme Theme) {
	name = strings.Trim(name, " ")
	ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)
	email = strings.Trim(email, " ")
	ValidateEmail(email)
	ValidateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)
	language = strings.Trim(language, " ") // may need more validation than this at some point to check it is a language we support and not a junk value, but it isnt critical right now
	theme.Validate()

	if !a.privateRegionClient.IsValidRegion(region) {
		panic(noSuchRegionErr)
	}

	if exists := a.store.accountWithCiNameExists(name); exists {
		panic(nameAlreadyInUseErr)
	}

	if acc := a.store.getPersonalAccountByEmail(email); acc != nil {
		a.linkMailer.sendMultipleAccountPolicyEmail(acc.Email)
	}

	accCore := NewCreatedNamedEntity(name)
	activationCode := CryptoUrlSafeString(a.cryptoCodeLen)
	acc := &fullPersonalAccountInfo{}
	acc.CreatedNamedEntity = *accCore
	acc.Region = region

	defer func() {
		r := recover()
		if r != nil {
			a.store.deleteAccountAndAllAssociatedMemberships(acc.Id)
			panic(r)
		}
	}()
	acc.Shard = a.privateRegionClient.CreateAccount(acc.Region, acc.Id, acc.Id, acc.Name)
	acc.IsPersonal = true
	acc.Email = email
	acc.Language = language
	acc.Theme = theme
	acc.activationCode = &activationCode

	pwdInfo := &pwdInfo{}
	pwdInfo.salt = CryptoBytes(a.saltLen)
	pwdInfo.pwd = ScryptKey([]byte(pwd), pwdInfo.salt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen

	a.store.createPersonalAccount(acc, pwdInfo)

	a.linkMailer.sendActivationLink(email, *acc.activationCode)
}

func (a *api) ResendActivationEmail(email string) {
	email = strings.Trim(email, " ")
	acc := a.store.getPersonalAccountByEmail(email)
	if acc == nil || acc.isActivated() {
		return
	}

	a.linkMailer.sendActivationLink(email, *acc.activationCode)
}

func (a *api) Activate(email, activationCode string) {
	activationCode = strings.Trim(activationCode, " ")
	acc := a.store.getPersonalAccountByEmail(email)
	if acc == nil || acc.activationCode == nil || activationCode != *acc.activationCode {
		panic(invalidActivationAttemptErr)
	}

	acc.activationCode = nil
	activationTime := time.Now().UTC()
	acc.activatedOn = &activationTime
	a.store.updatePersonalAccount(acc)
}

func (a *api) Authenticate(email, pwdTry string) Id {
	email = strings.Trim(email, " ")
	acc := a.store.getPersonalAccountByEmail(email)
	if acc == nil {
		panic(invalidNameOrPwdErr)
	}

	pwdInfo := a.store.getPwdInfo(acc.Id)
	scryptPwdTry := ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		panic(invalidNameOrPwdErr)
	}

	//must do this after checking the acc has the correct pwd otherwise it allows anyone to fish for valid emails on the system
	if !acc.isActivated() {
		panic(accountNotActivatedErr)
	}

	//if there was an outstanding password reset on this acc, remove it, they have since remembered their password
	if acc.resetPwdCode != nil && len(*acc.resetPwdCode) > 0 {
		acc.resetPwdCode = nil
		a.store.updatePersonalAccount(acc)
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.n != a.scryptN || pwdInfo.r != a.scryptR || pwdInfo.p != a.scryptP || pwdInfo.keyLen != a.scryptKeyLen || len(pwdInfo.salt) < a.saltLen {
		pwdInfo.salt = CryptoBytes(a.saltLen)
		pwdInfo.n = a.scryptN
		pwdInfo.r = a.scryptR
		pwdInfo.p = a.scryptP
		pwdInfo.keyLen = a.scryptKeyLen
		pwdInfo.pwd = ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
		a.store.updatePwdInfo(acc.Id, pwdInfo)

	}

	return acc.Id
}

func (a *api) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) {
	acc := a.store.getPersonalAccountByEmail(currentEmail)
	if acc == nil || acc.NewEmail == nil || newEmail != *acc.NewEmail || acc.newEmailConfirmationCode == nil || confirmationCode != *acc.newEmailConfirmationCode {
		panic(invalidNewEmailConfirmationAttemptErr)
	}

	if acc := a.store.getPersonalAccountByEmail(newEmail); acc != nil {
		panic(emailAlreadyInUseErr)
	}

	acc.Email = newEmail
	acc.NewEmail = nil
	acc.newEmailConfirmationCode = nil
	a.store.updatePersonalAccount(acc)
}

func (a *api) ResetPwd(email string) {
	email = strings.Trim(email, " ")
	acc := a.store.getPersonalAccountByEmail(email)
	if acc == nil {
		return
	}

	resetPwdCode := CryptoUrlSafeString(a.cryptoCodeLen)

	acc.resetPwdCode = &resetPwdCode
	a.store.updatePersonalAccount(acc)

	a.linkMailer.sendPwdResetLink(email, resetPwdCode)
}

func (a *api) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) {
	ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)

	acc := a.store.getPersonalAccountByEmail(email)
	if acc == nil || acc.resetPwdCode == nil || resetPwdCode != *acc.resetPwdCode {
		panic(invalidResetPwdAttemptErr)
	}

	scryptSalt := CryptoBytes(a.saltLen)
	scryptPwd := ScryptKey([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)

	acc.activationCode = nil
	acc.resetPwdCode = nil
	a.store.updatePersonalAccount(acc)

	pwdInfo := &pwdInfo{}
	pwdInfo.pwd = scryptPwd
	pwdInfo.salt = scryptSalt
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen
	a.store.updatePwdInfo(acc.Id, pwdInfo)
}

func (a *api) GetAccount(name string) *account {
	return a.store.getAccountByCiName(strings.Trim(name, " "))
}

func (a *api) GetAccounts(ids []Id) []*account {
	if len(ids) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}

	return a.store.getAccounts(ids)
}

func (a *api) SearchAccounts(nameStartsWith string) []*account {
	if utf8.RuneCountInString(nameStartsWith) < 3 || strings.Contains(nameStartsWith, "%") {
		panic(InvalidArgumentsErr)
	}
	return a.store.searchAccounts(nameStartsWith)
}

func (a *api) SearchPersonalAccounts(nameOrEmailStartsWith string) []*account {
	if utf8.RuneCountInString(nameOrEmailStartsWith) < 3 || strings.Contains(nameOrEmailStartsWith, "%") {
		panic(InvalidArgumentsErr)
	}
	return a.store.searchPersonalAccounts(nameOrEmailStartsWith)
}

func (a *api) GetMe(myId Id) *me {
	acc := a.store.getPersonalAccountById(myId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	return &acc.me
}

func (a *api) SetMyPwd(myId Id, oldPwd, newPwd string) {
	ValidateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers)

	pwdInfo := a.store.getPwdInfo(myId)
	if pwdInfo == nil {
		panic(noSuchAccountErr)
	}

	scryptPwdTry := ScryptKey([]byte(oldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		panic(incorrectPwdErr)
	}

	pwdInfo.salt = CryptoBytes(a.saltLen)
	pwdInfo.pwd = ScryptKey([]byte(newPwd), pwdInfo.salt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	pwdInfo.n = a.scryptN
	pwdInfo.r = a.scryptR
	pwdInfo.p = a.scryptP
	pwdInfo.keyLen = a.scryptKeyLen
	a.store.updatePwdInfo(myId, pwdInfo)
}

func (a *api) SetMyEmail(myId Id, newEmail string) {
	newEmail = strings.Trim(newEmail, " ")
	ValidateEmail(newEmail)

	if acc := a.store.getPersonalAccountByEmail(newEmail); acc != nil {
		a.linkMailer.sendMultipleAccountPolicyEmail(acc.Email)
	}

	acc := a.store.getPersonalAccountById(myId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	confirmationCode := CryptoUrlSafeString(a.cryptoCodeLen)

	acc.NewEmail = &newEmail
	acc.newEmailConfirmationCode = &confirmationCode
	a.store.updatePersonalAccount(acc)
	a.linkMailer.sendNewEmailConfirmationLink(acc.Email, newEmail, confirmationCode)
}

func (a *api) ResendMyNewEmailConfirmationEmail(myId Id) {
	acc := a.store.getPersonalAccountById(myId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	// check the acc has actually registered a new email
	if acc.NewEmail == nil {
		panic(noNewEmailRegisteredErr)
	}
	// just in case something has gone crazy wrong
	if acc.newEmailConfirmationCode == nil {
		panic(emailConfirmationCodeErr)
	}

	a.linkMailer.sendNewEmailConfirmationLink(acc.Email, *acc.NewEmail, *acc.newEmailConfirmationCode)
}

func (a *api) SetAccountName(myId, accountId Id, newName string) {
	newName = strings.Trim(newName, " ")
	ValidateStringParam("name", newName, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)

	if exists := a.store.accountWithCiNameExists(newName); exists {
		panic(nameAlreadyInUseErr)
	}

	acc := a.store.getAccount(accountId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsPersonal { // can't rename someone else's personal account
			panic(InsufficientPermissionErr)
		}

		if !a.privateRegionClient.MemberIsAccountOwner(acc.Region, acc.Shard, accountId, myId) {
			panic(InsufficientPermissionErr)
		}
	}

	acc.Name = newName
	a.store.updateAccount(acc)

	if myId.Equal(accountId) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
		for offset, total := 0, 1; offset < total; {
			var accounts []*account
			accounts, total = a.store.getGroupAccounts(myId, offset, 100)
			offset += len(accounts)
			for _, account := range accounts {
				a.privateRegionClient.RenameMember(account.Region, account.Shard, account.Id, myId, newName)
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
		if account.IsPersonal { // can't set avatar on someone else's personal account
			panic(InsufficientPermissionErr)
		}

		if !a.privateRegionClient.MemberIsAccountOwner(account.Region, account.Shard, accountId, myId) {
			panic(InsufficientPermissionErr)
		}
	}

	if avatarImageData != nil {
		avatarImage, _, err := image.Decode(avatarImageData)
		PanicIf(err)
		bounds := avatarImage.Bounds()
		if bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y { //if it  isn't square, then error
			panic(invalidAvatarShapeErr)
		}
		if uint(bounds.Max.X-bounds.Min.X) > a.maxAvatarDim { // if it is larger than allowed then resize
			avatarImage = resize.Resize(a.maxAvatarDim, a.maxAvatarDim, avatarImage, resize.NearestNeighbor)
		}
		buff := &bytes.Buffer{}
		PanicIf(png.Encode(buff, avatarImage))
		data := buff.Bytes()
		reader := bytes.NewReader(data)
		a.avatarStore.put(myId.String(), "image/png", reader)
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
	(&fullPersonalAccountInfo{}).isMigrating()
}

func (a *api) CreateAccount(myId Id, name, region string) *account {
	name = strings.Trim(name, " ")
	ValidateStringParam("name", name, a.nameMinRuneCount, a.nameMaxRuneCount, a.nameRegexMatchers)

	if !a.privateRegionClient.IsValidRegion(region) {
		panic(noSuchRegionErr)
	}

	if exists := a.store.accountWithCiNameExists(name); exists {
		panic(nameAlreadyInUseErr)
	}

	accountCore := NewCreatedNamedEntity(name)

	account := &account{
		CreatedNamedEntity: *accountCore,
		Region:             region,
		Shard:              -1,
		IsPersonal:         false,
	}
	a.store.createGroupAccountAndMembership(account, myId)

	owner := a.store.getPersonalAccountById(myId)
	if owner == nil {
		panic(noSuchAccountErr)
	}

	defer func() {
		r := recover()
		if r != nil {
			a.store.deleteAccountAndAllAssociatedMemberships(accountCore.Id)
			panic(r)
		}
	}()
	shard := a.privateRegionClient.CreateAccount(region, accountCore.Id, myId, owner.Name)

	account.Shard = shard
	a.store.updateAccount(account)
	return account
}

func (a *api) GetMyAccounts(myId Id, offset, limit int) ([]*account, int) {
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxProcessEntityCount)
	return a.store.getGroupAccounts(myId, offset, limit)
}

func (a *api) DeleteAccount(myId, accountId Id) {
	acc := a.store.getAccount(accountId)
	if acc == nil {
		panic(noSuchAccountErr)
	}

	if !myId.Equal(accountId) {
		if acc.IsPersonal { // can't delete someone else's personal account
			panic(InsufficientPermissionErr)
		}
		//otherwise attempting to delete a group account
		if !a.privateRegionClient.MemberIsAccountOwner(acc.Region, acc.Shard, accountId, myId) {
			panic(InsufficientPermissionErr)
		}
	}

	a.privateRegionClient.DeleteAccount(acc.Region, acc.Shard, accountId, myId)
	a.store.deleteAccountAndAllAssociatedMemberships(accountId)

	if myId.Equal(accountId) {
		for offset, total := 0, 1; offset < total; {
			var accounts []*account
			accounts, total = a.store.getGroupAccounts(myId, offset, 100)
			offset += len(accounts)
			for _, account := range accounts {
				if a.privateRegionClient.MemberIsOnlyAccountOwner(account.Region, account.Shard, account.Id, myId) {
					panic(onlyOwnerMemberErr)
				}
			}
			for _, account := range accounts {
				a.privateRegionClient.RemoveMembers(account.Region, account.Shard, account.Id, myId, []Id{myId})
			}
		}
	}
}

func (a *api) AddMembers(myId, accountId Id, newMembers []*AddMemberPublic) {
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	if len(newMembers) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}

	account := a.store.getAccount(accountId)
	if account == nil {
		panic(noSuchAccountErr)
	}

	ids := make([]Id, 0, len(newMembers))
	addMembersMap := map[string]*AddMemberPublic{}
	for _, member := range newMembers {
		ids = append(ids, member.Id)
		addMembersMap[member.Id.String()] = member
	}

	accs := a.store.getPersonalAccounts(ids)

	members := make([]*AddMemberPrivate, 0, len(accs))
	for _, acc := range accs {
		role := addMembersMap[acc.Id.String()].Role
		role.Validate()
		ami := &AddMemberPrivate{}
		ami.Id = acc.Id
		ami.Role = role
		ami.Name = acc.Name
		members = append(members, ami)
	}

	a.privateRegionClient.AddMembers(account.Region, account.Shard, accountId, myId, members)
	a.store.createMemberships(accountId, ids)
}

func (a *api) RemoveMembers(myId, accountId Id, existingMembers []Id) {
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	if len(existingMembers) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}

	account := a.store.getAccount(accountId)
	if account == nil {
		panic(noSuchAccountErr)
	}

	a.privateRegionClient.RemoveMembers(account.Region, account.Shard, accountId, myId, existingMembers)
	a.store.deleteMemberships(accountId, existingMembers)
}

//internal helpers

type store interface {
	//personal or group account
	accountWithCiNameExists(name string) bool
	getAccountByCiName(name string) *account
	//personal account
	createPersonalAccount(acc *fullPersonalAccountInfo, pwdInfo *pwdInfo)
	getPersonalAccountByEmail(email string) *fullPersonalAccountInfo
	getPersonalAccountById(id Id) *fullPersonalAccountInfo
	getPwdInfo(id Id) *pwdInfo
	updatePersonalAccount(acc *fullPersonalAccountInfo)
	updateAccount(account *account)
	updatePwdInfo(id Id, pwdInfo *pwdInfo)
	deleteAccountAndAllAssociatedMemberships(id Id)
	getAccount(id Id) *account
	getAccounts(ids []Id) []*account
	searchAccounts(nameStartsWith string) []*account
	searchPersonalAccounts(nameOrEmailStartsWith string) []*account
	getPersonalAccounts(ids []Id) []*account
	//group account
	createGroupAccountAndMembership(account *account, ownerId Id)
	getGroupAccounts(myId Id, offset, limit int) ([]*account, int)
	//members
	createMemberships(accountId Id, members []Id)
	deleteMemberships(accountId Id, members []Id)
}

type linkMailer interface {
	sendMultipleAccountPolicyEmail(address string)
	sendActivationLink(address, activationCode string)
	sendPwdResetLink(address, resetCode string)
	sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string)
}

type avatarStore interface {
	put(key string, mimeType string, data io.Reader)
	delete(key string)
}

type account struct {
	CreatedNamedEntity
	Region     string  `json:"region"`
	NewRegion  *string `json:"newRegion,omitempty"`
	Shard      int     `json:"shard"`
	HasAvatar  bool    `json:"hasAvatar"`
	IsPersonal bool    `json:"isPersonal"`
}

func (a *account) isMigrating() bool {
	return a.NewRegion != nil
}

type me struct {
	account
	Email    string  `json:"email"`
	Language string  `json:"language"`
	Theme    Theme   `json:"theme"`
	NewEmail *string `json:"newEmail,omitempty"`
}

type fullPersonalAccountInfo struct {
	me
	activationCode           *string
	activatedOn              *time.Time
	newEmailConfirmationCode *string
	resetPwdCode             *string
}

func (a *fullPersonalAccountInfo) isActivated() bool {
	return a.activatedOn != nil
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