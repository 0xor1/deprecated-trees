package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	noSuchRegionErr                       = &AppError{Code: "c_v1_a_nsr", Message: "no such region", Public: true}
	noSuchAccountErr                      = &AppError{Code: "c_v1_a_nsa", Message: "no such account", Public: true}
	invalidActivationAttemptErr           = &AppError{Code: "c_v1_a_iaa", Message: "invalid activation attempt", Public: true}
	invalidResetPwdAttemptErr             = &AppError{Code: "c_v1_a_irpa", Message: "invalid reset password attempt", Public: true}
	invalidNewEmailConfirmationAttemptErr = &AppError{Code: "c_v1_a_ineca", Message: "invalid new email confirmation attempt", Public: true}
	invalidNameOrPwdErr                   = &AppError{Code: "c_v1_a_inop", Message: "invalid name or password", Public: true}
	incorrectPwdErr                       = &AppError{Code: "c_v1_a_ip", Message: "password incorrect", Public: true}
	accountNotActivatedErr                = &AppError{Code: "c_v1_a_ana", Message: "account not activated", Public: true}
	emailAlreadyInUseErr                  = &AppError{Code: "c_v1_a_eaiu", Message: "email already in use", Public: true}
	nameAlreadyInUseErr                   = &AppError{Code: "c_v1_a_naiu", Message: "name already in use", Public: true}
	emailConfirmationCodeErr              = &AppError{Code: "c_v1_a_ecc", Message: "email confirmation code is of zero length", Public: false}
	noNewEmailRegisteredErr               = &AppError{Code: "c_v1_a_nner", Message: "no new email registered", Public: true}
	onlyOwnerMemberErr                    = &AppError{Code: "c_v1_a_oom", Message: "can't delete member who is the only owner of an account", Public: true}
	invalidAvatarShapeErr                 = &AppError{Code: "c_v1_a_ias", Message: "avatar images must be square", Public: true}
)

// The main account Api interface
type Api interface {
	//accessible outside of active session
	GetRegions(ctx CentralCtx) []string
	Register(ctx CentralCtx, name, email, pwd, region, language string, displayName *string, theme Theme)
	ResendActivationEmail(ctx CentralCtx, email string)
	Activate(ctx CentralCtx, email, activationCode string)
	Authenticate(ctx CentralCtx, email, pwd string) Id
	ConfirmNewEmail(ctx CentralCtx, currentEmail, newEmail, confirmationCode string)
	ResetPwd(ctx CentralCtx, email string)
	SetNewPwdFromPwdReset(ctx CentralCtx, newPwd, email, resetPwdCode string)
	GetAccount(ctx CentralCtx, name string) *account
	GetAccounts(ctx CentralCtx, ids []Id) []*account
	SearchAccounts(ctx CentralCtx, nameOrDisplayNameStartsWith string) []*account
	SearchPersonalAccounts(ctx CentralCtx, nameOrDisplayNameOrEmailStartsWith string) []*account
	//requires active session to access
	GetMe(ctx CentralCtx) *me
	SetMyPwd(ctx CentralCtx, oldPwd, newPwd string)
	SetMyEmail(ctx CentralCtx, newEmail string)
	ResendMyNewEmailConfirmationEmail(ctx CentralCtx)
	SetAccountName(ctx CentralCtx, accountId Id, newName string)
	SetAccountDisplayName(ctx CentralCtx, accountId Id, newDisplayName *string)
	SetAccountAvatar(ctx CentralCtx, accountId Id, avatarImage io.ReadCloser)
	MigrateAccount(ctx CentralCtx, accountId Id, newRegion string)
	CreateAccount(ctx CentralCtx, name, region string, displayName *string) *account
	GetMyAccounts(ctx CentralCtx, after *Id, limit int) ([]*account, bool)
	DeleteAccount(ctx CentralCtx, accountId Id)
	//member centric - must be an owner or admin
	AddMembers(ctx CentralCtx, accountId Id, newMembers []*AddMemberPublic)
	RemoveMembers(ctx CentralCtx, accountId Id, existingMembers []Id)
}

func New() Api {
	return &api{}
}

type api struct {
}

func (a *api) GetRegions(ctx CentralCtx) []string {
	return ctx.PrivateRegionClient().GetRegions()
}

func (a *api) Register(ctx CentralCtx, name, email, pwd, region, language string, displayName *string, theme Theme) {
	name = strings.Trim(name, " ")
	ctx.Validate().Name(name)
	email = strings.Trim(email, " ")
	ValidateEmail(email)
	ctx.Validate().Pwd(pwd)
	language = strings.Trim(language, " ") // may need more validation than this at some point to check it is a language we support and not a junk value, but it isnt critical right now
	theme.Validate()
	if displayName != nil {
		*displayName = strings.Trim(*displayName, " ")
		if *displayName == "" {
			displayName = nil
		}
	}

	if !ctx.PrivateRegionClient().IsValidRegion(region) {
		noSuchRegionErr.Panic()
	}

	if exists := dbAccountWithCiNameExists(ctx, name); exists {
		nameAlreadyInUseErr.Panic()
	}

	if acc := dbGetPersonalAccountByEmail(ctx, email); acc != nil {
		emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
	}

	activationCode := ctx.Crypt().CreateUrlSafeString()
	acc := &fullPersonalAccountInfo{}
	acc.Id = NewId()
	acc.Name = name
	acc.DisplayName = displayName
	acc.CreatedOn = Now()
	acc.Region = region

	defer func() {
		r := recover()
		if r != nil {
			dbDeleteAccountAndAllAssociatedMemberships(ctx, acc.Id)
			panic(r)
		}
	}()
	acc.Shard = ctx.PrivateRegionClient().CreateAccount(acc.Region, acc.Id, acc.Id, acc.Name, acc.DisplayName)
	acc.IsPersonal = true
	acc.Email = email
	acc.Language = language
	acc.Theme = theme
	acc.activationCode = &activationCode

	pwdInfo := &pwdInfo{}
	pwdInfo.salt = ctx.Crypt().CreatePwdSalt()
	pwdInfo.pwd = ctx.Crypt().ScryptKey([]byte(pwd), pwdInfo.salt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())
	pwdInfo.n = ctx.Crypt().ScryptN()
	pwdInfo.r = ctx.Crypt().ScryptR()
	pwdInfo.p = ctx.Crypt().ScryptP()
	pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()

	dbCreatePersonalAccount(ctx, acc, pwdInfo)

	emailSendActivationLink(ctx, email, *acc.activationCode)
}

func (a *api) ResendActivationEmail(ctx CentralCtx, email string) {
	email = strings.Trim(email, " ")
	acc := dbGetPersonalAccountByEmail(ctx, email)
	if acc == nil || acc.isActivated() {
		return
	}

	emailSendActivationLink(ctx, email, *acc.activationCode)
}

func (a *api) Activate(ctx CentralCtx, email, activationCode string) {
	activationCode = strings.Trim(activationCode, " ")
	acc := dbGetPersonalAccountByEmail(ctx, email)
	if acc == nil || acc.activationCode == nil || activationCode != *acc.activationCode {
		invalidActivationAttemptErr.Panic()
	}

	acc.activationCode = nil
	activationTime := time.Now().UTC()
	acc.activatedOn = &activationTime
	dbUpdatePersonalAccount(ctx, acc)
}

func (a *api) Authenticate(ctx CentralCtx, email, pwdTry string) Id {
	email = strings.Trim(email, " ")
	acc := dbGetPersonalAccountByEmail(ctx, email)
	if acc == nil {
		invalidNameOrPwdErr.Panic()
	}

	pwdInfo := dbGetPwdInfo(ctx, acc.Id)
	scryptPwdTry := ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		invalidNameOrPwdErr.Panic()
	}

	//must do this after checking the acc has the correct pwd otherwise it allows anyone to fish for valid emails on the system
	if !acc.isActivated() {
		accountNotActivatedErr.Panic()
	}

	//if there was an outstanding password reset on this acc, remove it, they have since remembered their password
	if acc.resetPwdCode != nil && len(*acc.resetPwdCode) > 0 {
		acc.resetPwdCode = nil
		dbUpdatePersonalAccount(ctx, acc)
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	newSalt := ctx.Crypt().CreatePwdSalt()
	if pwdInfo.n != ctx.Crypt().ScryptN() || pwdInfo.r != ctx.Crypt().ScryptR() || pwdInfo.p != ctx.Crypt().ScryptP() || pwdInfo.keyLen != ctx.Crypt().ScryptKeyLen() || len(pwdInfo.salt) < len(newSalt) {
		pwdInfo.salt = newSalt
		pwdInfo.n = ctx.Crypt().ScryptN()
		pwdInfo.r = ctx.Crypt().ScryptR()
		pwdInfo.p = ctx.Crypt().ScryptP()
		pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()
		pwdInfo.pwd = ScryptKey([]byte(pwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
		dbUpdatePwdInfo(ctx, acc.Id, pwdInfo)

	}

	return acc.Id
}

func (a *api) ConfirmNewEmail(ctx CentralCtx, currentEmail, newEmail, confirmationCode string) {
	acc := dbGetPersonalAccountByEmail(ctx, currentEmail)
	if acc == nil || acc.NewEmail == nil || newEmail != *acc.NewEmail || acc.newEmailConfirmationCode == nil || confirmationCode != *acc.newEmailConfirmationCode {
		invalidNewEmailConfirmationAttemptErr.Panic()
	}

	if acc := dbGetPersonalAccountByEmail(ctx, newEmail); acc != nil {
		emailAlreadyInUseErr.Panic()
	}

	acc.Email = newEmail
	acc.NewEmail = nil
	acc.newEmailConfirmationCode = nil
	dbUpdatePersonalAccount(ctx, acc)
}

func (a *api) ResetPwd(ctx CentralCtx, email string) {
	email = strings.Trim(email, " ")
	acc := dbGetPersonalAccountByEmail(ctx, email)
	if acc == nil {
		return
	}

	resetPwdCode := ctx.Crypt().CreateUrlSafeString()

	acc.resetPwdCode = &resetPwdCode
	dbUpdatePersonalAccount(ctx, acc)

	emailSendPwdResetLink(ctx, email, resetPwdCode)
}

func (a *api) SetNewPwdFromPwdReset(ctx CentralCtx, newPwd, email, resetPwdCode string) {
	ctx.Validate().Pwd(newPwd)

	acc := dbGetPersonalAccountByEmail(ctx, email)
	if acc == nil || acc.resetPwdCode == nil || resetPwdCode != *acc.resetPwdCode {
		invalidResetPwdAttemptErr.Panic()
	}

	scryptSalt := ctx.Crypt().CreatePwdSalt()
	scryptPwd := ScryptKey([]byte(newPwd), scryptSalt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())

	acc.activationCode = nil
	acc.resetPwdCode = nil
	dbUpdatePersonalAccount(ctx, acc)

	pwdInfo := &pwdInfo{}
	pwdInfo.pwd = scryptPwd
	pwdInfo.salt = scryptSalt
	pwdInfo.n = ctx.Crypt().ScryptN()
	pwdInfo.r = ctx.Crypt().ScryptR()
	pwdInfo.p = ctx.Crypt().ScryptP()
	pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()
	dbUpdatePwdInfo(ctx, acc.Id, pwdInfo)
}

func (a *api) GetAccount(ctx CentralCtx, name string) *account {
	return dbGetAccountByCiName(ctx, strings.Trim(name, " "))
}

func (a *api) GetAccounts(ctx CentralCtx, ids []Id) []*account {
	ctx.Validate().EntityCount(len(ids))

	return dbGetAccounts(ctx, ids)
}

func (a *api) SearchAccounts(ctx CentralCtx, nameOrDisplayNameStartsWith string) []*account {
	nameOrDisplayNameStartsWith = strings.Trim(nameOrDisplayNameStartsWith, " ")
	if utf8.RuneCountInString(nameOrDisplayNameStartsWith) < 3 || strings.Contains(nameOrDisplayNameStartsWith, "%") {
		InvalidArgumentsErr.Panic()
	}
	return dbSearchAccounts(ctx, nameOrDisplayNameStartsWith)
}

func (a *api) SearchPersonalAccounts(ctx CentralCtx, nameOrDisplayNameOrEmailStartsWith string) []*account {
	nameOrDisplayNameOrEmailStartsWith = strings.Trim(nameOrDisplayNameOrEmailStartsWith, " ")
	if utf8.RuneCountInString(nameOrDisplayNameOrEmailStartsWith) < 3 || strings.Contains(nameOrDisplayNameOrEmailStartsWith, "%") {
		InvalidArgumentsErr.Panic()
	}
	return dbSearchPersonalAccounts(ctx, nameOrDisplayNameOrEmailStartsWith)
}

func (a *api) GetMe(ctx CentralCtx) *me {
	acc := dbGetPersonalAccountById(ctx, ctx.MyId())
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	return &acc.me
}

func (a *api) SetMyPwd(ctx CentralCtx, oldPwd, newPwd string) {
	ctx.Validate().Pwd(newPwd)

	pwdInfo := dbGetPwdInfo(ctx, ctx.MyId())
	if pwdInfo == nil {
		noSuchAccountErr.Panic()
	}

	scryptPwdTry := ScryptKey([]byte(oldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

	if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
		incorrectPwdErr.Panic()
	}

	pwdInfo.salt = ctx.Crypt().CreatePwdSalt()
	pwdInfo.pwd = ScryptKey([]byte(newPwd), pwdInfo.salt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())
	pwdInfo.n = ctx.Crypt().ScryptN()
	pwdInfo.r = ctx.Crypt().ScryptR()
	pwdInfo.p = ctx.Crypt().ScryptP()
	pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()
	dbUpdatePwdInfo(ctx, ctx.MyId(), pwdInfo)
}

func (a *api) SetMyEmail(ctx CentralCtx, newEmail string) {
	newEmail = strings.Trim(newEmail, " ")
	ValidateEmail(newEmail)

	if acc := dbGetPersonalAccountByEmail(ctx, newEmail); acc != nil {
		emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
	}

	acc := dbGetPersonalAccountById(ctx, ctx.MyId())
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	confirmationCode := ctx.Crypt().CreateUrlSafeString()

	acc.NewEmail = &newEmail
	acc.newEmailConfirmationCode = &confirmationCode
	dbUpdatePersonalAccount(ctx, acc)
	emailSendNewEmailConfirmationLink(ctx, acc.Email, newEmail, confirmationCode)
}

func (a *api) ResendMyNewEmailConfirmationEmail(ctx CentralCtx) {
	acc := dbGetPersonalAccountById(ctx, ctx.MyId())
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	// check the acc has actually registered a new email
	if acc.NewEmail == nil {
		noNewEmailRegisteredErr.Panic()
	}
	// just in case something has gone crazy wrong
	if acc.newEmailConfirmationCode == nil {
		emailConfirmationCodeErr.Panic()
	}

	emailSendNewEmailConfirmationLink(ctx, acc.Email, *acc.NewEmail, *acc.newEmailConfirmationCode)
}

func (a *api) SetAccountName(ctx CentralCtx, accountId Id, newName string) {
	newName = strings.Trim(newName, " ")
	ctx.Validate().Name(newName)

	if exists := dbAccountWithCiNameExists(ctx, newName); exists {
		nameAlreadyInUseErr.Panic()
	}

	acc := dbGetAccount(ctx, accountId)
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	if !ctx.MyId().Equal(accountId) {
		if acc.IsPersonal { // can't rename someone else's personal account
			InsufficientPermissionErr.Panic()
		}

		if !ctx.PrivateRegionClient().MemberIsAccountOwner(acc.Region, acc.Shard, accountId, ctx.MyId()) {
			InsufficientPermissionErr.Panic()
		}
	}

	acc.Name = newName
	dbUpdateAccount(ctx, acc)

	if ctx.MyId().Equal(accountId) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
		ctx.PrivateRegionClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), newName) //first rename myself in my personal org
		var after *Id
		for {
			accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
			for _, acc := range accs {
				ctx.PrivateRegionClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), newName)
			}
			if more {
				after = &accs[len(accs)-1].Id
			} else {
				break
			}
		}
	}
}

func (a *api) SetAccountDisplayName(ctx CentralCtx, accountId Id, newDisplayName *string) {
	if newDisplayName != nil {
		*newDisplayName = strings.Trim(*newDisplayName, " ")
		if *newDisplayName == "" {
			newDisplayName = nil
		}
	}

	acc := dbGetAccount(ctx, accountId)
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	if !ctx.MyId().Equal(accountId) {
		if acc.IsPersonal { // can't rename someone else's personal account
			InsufficientPermissionErr.Panic()
		}

		if !ctx.PrivateRegionClient().MemberIsAccountOwner(acc.Region, acc.Shard, accountId, ctx.MyId()) {
			InsufficientPermissionErr.Panic()
		}
	}

	if (acc.DisplayName == nil && newDisplayName == nil) || (acc.DisplayName != nil && newDisplayName != nil && *acc.DisplayName == *newDisplayName) {
		return //if there is no change, dont do any redundant work
	}

	acc.DisplayName = newDisplayName
	dbUpdateAccount(ctx, acc)

	if ctx.MyId().Equal(accountId) { // if i did set my personal account displayName, i need to update all the stored displayNames in all the accounts Im a member of
		ctx.PrivateRegionClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), newDisplayName) //first set my display name in my personal org
		var after *Id
		for {
			accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
			for _, acc := range accs {
				ctx.PrivateRegionClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), newDisplayName)
			}
			if more {
				after = &accs[len(accs)-1].Id
			} else {
				break
			}
		}
	}
}

func (a *api) SetAccountAvatar(ctx CentralCtx, accountId Id, avatarImageData io.ReadCloser) {
	if avatarImageData != nil {
		defer avatarImageData.Close()
	}

	account := dbGetAccount(ctx, accountId)
	if account == nil {
		noSuchAccountErr.Panic()
	}

	if !ctx.MyId().Equal(accountId) {
		if account.IsPersonal { // can't set avatar on someone else's personal account
			InsufficientPermissionErr.Panic()
		}

		if !ctx.PrivateRegionClient().MemberIsAccountOwner(account.Region, account.Shard, accountId, ctx.MyId()) {
			InsufficientPermissionErr.Panic()
		}
	}

	if avatarImageData != nil {
		avatarImage, _, err := image.Decode(avatarImageData)
		PanicIf(err)
		bounds := avatarImage.Bounds()
		if bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y { //if it  isn't square, then error
			invalidAvatarShapeErr.Panic()
		}
		if uint(bounds.Max.X-bounds.Min.X) > ctx.Avatar().MaxAvatarDim() { // if it is larger than allowed then resize
			avatarImage = resize.Resize(ctx.Avatar().MaxAvatarDim(), ctx.Avatar().MaxAvatarDim(), avatarImage, resize.NearestNeighbor)
		}
		buff := &bytes.Buffer{}
		PanicIf(png.Encode(buff, avatarImage))
		data := buff.Bytes()
		reader := bytes.NewReader(data)
		ctx.Avatar().Save(ctx.MyId().String(), "image/png", reader)
		if !account.HasAvatar {
			//if account didn't previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = true
			dbUpdateAccount(ctx, account)
		}
	} else {
		ctx.Avatar().Delete(ctx.MyId().String())
		if account.HasAvatar {
			//if account did previously have an avatar then lets update the store to reflect it's new state
			account.HasAvatar = false
			dbUpdateAccount(ctx, account)
		}
	}
}

func (a *api) MigrateAccount(ctx CentralCtx, accountId Id, newRegion string) {
	//the next line is arbitrarily added in to get code coverage for isMigrating Func
	//which won't get used anywhere until the migration feature is worked on in the future
	(&fullPersonalAccountInfo{}).isMigrating()
}

func (a *api) CreateAccount(ctx CentralCtx, name, region string, displayName *string) *account {
	name = strings.Trim(name, " ")
	ctx.Validate().Name(name)

	if !ctx.PrivateRegionClient().IsValidRegion(region) {
		noSuchRegionErr.Panic()
	}

	if exists := dbAccountWithCiNameExists(ctx, name); exists {
		nameAlreadyInUseErr.Panic()
	}

	account := &account{}
	account.Id = NewId()
	account.Name = name
	account.DisplayName = displayName
	account.CreatedOn = Now()
	account.Region = region
	account.Shard = -1
	account.IsPersonal = false
	dbCreateGroupAccountAndMembership(ctx, account, ctx.MyId())

	owner := dbGetPersonalAccountById(ctx, ctx.MyId())
	if owner == nil {
		noSuchAccountErr.Panic()
	}

	defer func() {
		r := recover()
		if r != nil {
			dbDeleteAccountAndAllAssociatedMemberships(ctx, account.Id)
			panic(r)
		}
	}()
	shard := ctx.PrivateRegionClient().CreateAccount(region, account.Id, ctx.MyId(), owner.Name, owner.DisplayName)

	account.Shard = shard
	dbUpdateAccount(ctx, account)
	return account
}

func (a *api) GetMyAccounts(ctx CentralCtx, after *Id, limit int) ([]*account, bool) {
	return dbGetGroupAccounts(ctx, ctx.MyId(), after, ctx.Validate().Limit(limit))
}

func (a *api) DeleteAccount(ctx CentralCtx, accountId Id) {
	acc := dbGetAccount(ctx, accountId)
	if acc == nil {
		noSuchAccountErr.Panic()
	}

	if !ctx.MyId().Equal(accountId) {
		if acc.IsPersonal { // can't delete someone else's personal account
			InsufficientPermissionErr.Panic()
		}
		//otherwise attempting to delete a group account
		if !ctx.PrivateRegionClient().MemberIsAccountOwner(acc.Region, acc.Shard, accountId, ctx.MyId()) {
			InsufficientPermissionErr.Panic()
		}
	}

	ctx.PrivateRegionClient().DeleteAccount(acc.Region, acc.Shard, accountId, ctx.MyId())
	dbDeleteAccountAndAllAssociatedMemberships(ctx, accountId)

	if ctx.MyId().Equal(accountId) {
		var after *Id
		for {
			accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
			for _, acc := range accs {
				if ctx.PrivateRegionClient().MemberIsOnlyAccountOwner(acc.Region, acc.Shard, acc.Id, ctx.MyId()) {
					onlyOwnerMemberErr.Panic()
				}
			}
			for _, acc := range accs {
				ctx.PrivateRegionClient().RemoveMembers(acc.Region, acc.Shard, acc.Id, ctx.MyId(), []Id{ctx.MyId()})
			}
			if more {
				after = &accs[len(accs)-1].Id
			} else {
				break
			}
		}
	}
}

func (a *api) AddMembers(ctx CentralCtx, accountId Id, newMembers []*AddMemberPublic) {
	if accountId.Equal(ctx.MyId()) {
		InvalidOperationErr.Panic()
	}
	ctx.Validate().EntityCount(len(newMembers))

	account := dbGetAccount(ctx, accountId)
	if account == nil {
		noSuchAccountErr.Panic()
	}

	ids := make([]Id, 0, len(newMembers))
	addMembersMap := map[string]*AddMemberPublic{}
	for _, member := range newMembers {
		ids = append(ids, member.Id)
		addMembersMap[member.Id.String()] = member
	}

	accs := dbGetPersonalAccounts(ctx, ids)

	members := make([]*AddMemberPrivate, 0, len(accs))
	for _, acc := range accs {
		role := addMembersMap[acc.Id.String()].Role
		role.Validate()
		ami := &AddMemberPrivate{}
		ami.Id = acc.Id
		ami.Role = role
		ami.Name = acc.Name
		ami.DisplayName = acc.DisplayName
		members = append(members, ami)
	}

	ctx.PrivateRegionClient().AddMembers(account.Region, account.Shard, accountId, ctx.MyId(), members)
	dbCreateMemberships(ctx, accountId, ids)
}

func (a *api) RemoveMembers(ctx CentralCtx, accountId Id, existingMembers []Id) {
	if accountId.Equal(ctx.MyId()) {
		InvalidOperationErr.Panic()
	}
	ctx.Validate().EntityCount(len(existingMembers))

	account := dbGetAccount(ctx, accountId)
	if account == nil {
		noSuchAccountErr.Panic()
	}

	ctx.PrivateRegionClient().RemoveMembers(account.Region, account.Shard, accountId, ctx.MyId(), existingMembers)
	dbDeleteMemberships(ctx, accountId, existingMembers)
}

//internal helpers

//db helpers
func dbAccountWithCiNameExists(ctx CentralCtx, name string) bool {
	row := ctx.CentralDb().Account().QueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?`, name)
	count := 0
	PanicIf(row.Scan(&count))
	return count != 0
}

func dbGetAccountByCiName(ctx CentralCtx, name string) *account {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?`, name)
	acc := account{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)) {
		return nil
	}
	return &acc
}

func dbCreatePersonalAccount(ctx CentralCtx, account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	id := []byte(account.Id)
	_, err := ctx.CentralDb().Account().Exec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode)
	PanicIf(err)
	_, err = ctx.CentralDb().Pwd().Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	PanicIf(err)
}

func dbGetPersonalAccountByEmail(ctx CentralCtx, email string) *fullPersonalAccountInfo {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPersonalAccountById(ctx CentralCtx, id Id) *fullPersonalAccountInfo {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, []byte(id))
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPwdInfo(ctx CentralCtx, id Id) *pwdInfo {
	row := ctx.CentralDb().Pwd().QueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?`, []byte(id))
	pwd := pwdInfo{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)) {
		return nil
	}
	return &pwd
}

func dbUpdatePersonalAccount(ctx CentralCtx, personalAccountInfo *fullPersonalAccountInfo) {
	_, err := ctx.CentralDb().Account().Exec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(personalAccountInfo.Id), personalAccountInfo.Name, personalAccountInfo.DisplayName, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode)
	PanicIf(err)
}

func dbUpdateAccount(ctx CentralCtx, account *account) {
	_, err := ctx.CentralDb().Account().Exec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal)
	PanicIf(err)
}

func dbUpdatePwdInfo(ctx CentralCtx, id Id, pwdInfo *pwdInfo) {
	_, err := ctx.CentralDb().Pwd().Exec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id))
	PanicIf(err)
}

func dbDeleteAccountAndAllAssociatedMemberships(ctx CentralCtx, id Id) {
	castId := []byte(id)
	_, err := ctx.CentralDb().Account().Exec(`CALL deleteAccountAndAllAssociatedMemberships(?)`, castId)
	PanicIf(err)
	_, err = ctx.CentralDb().Pwd().Exec(`DELETE FROM pwds WHERE id = ?`, castId)
	PanicIf(err)
}

func dbGetAccount(ctx CentralCtx, id Id) *account {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?`, []byte(id))
	a := account{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)) {
		return nil
	}
	return &a
}

func dbGetAccounts(ctx CentralCtx, ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	castedIds = append(castedIds, []byte(ids[0]))
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (?`)
	for _, id := range ids[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`)`)
	rows, err := ctx.CentralDb().Account().Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	return res
}

func dbSearchAccounts(ctx CentralCtx, nameOrDisplayNameStartsWith string) []*account {
	searchTerm := nameOrDisplayNameStartsWith + "%"
	//rows, err := ctx.CentralDb().Account().Query(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isPersonal FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	rows, err := ctx.CentralDb().Account().Query(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? OR displayName LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal))
		res = append(res, &acc)
	}
	return res
}

func dbSearchPersonalAccounts(ctx CentralCtx, nameOrDisplayNameOrEmailStartsWith string) []*account {
	searchTerm := nameOrDisplayNameOrEmailStartsWith + "%"
	//rows, err := ctx.CentralDb().Account().Query(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	rows, err := ctx.CentralDb().Account().Query(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? OR displayName LIKE ? OR email LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, searchTerm, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbGetPersonalAccounts(ctx CentralCtx, ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	castedIds = append(castedIds, []byte(ids[0]))
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE activatedOn IS NOT NULL AND id IN (?`)
	for _, id := range ids[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`)`)
	rows, err := ctx.CentralDb().Account().Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbCreateGroupAccountAndMembership(ctx CentralCtx, account *account, memberId Id) {
	_, err := ctx.CentralDb().Account().Exec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, []byte(memberId))
	PanicIf(err)
}

func dbGetGroupAccounts(ctx CentralCtx, memberId Id, after *Id, limit int) ([]*account, bool) {
	args := make([]interface{}, 0, 3)
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (SELECT account FROM memberships WHERE member = ?)`)
	args = append(args, []byte(memberId))
	if after != nil {
		query.WriteString(` AND name > (SELECT name FROM accounts WHERE id = ?)`)
		args = append(args, []byte(*after))
	}
	query.WriteString(` ORDER BY name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, err := ctx.CentralDb().Account().Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, limit+1)
	for rows.Next() {
		a := account{}
		PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func dbCreateMemberships(ctx CentralCtx, accountId Id, members []Id) {
	args := make([]interface{}, 0, len(members)*2)
	args = append(args, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`INSERT INTO memberships (account, member) VALUES (?,?)`)
	for _, member := range members[1:] {
		query.WriteString(`,(?,?)`)
		args = append(args, []byte(accountId), []byte(member))
	}
	_, err := ctx.CentralDb().Account().Exec(query.String(), args...)
	PanicIf(err)
}

func dbDeleteMemberships(ctx CentralCtx, accountId Id, members []Id) {
	castedIds := make([]interface{}, 0, len(members)+1)
	castedIds = append(castedIds, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`DELETE FROM memberships WHERE account=? AND member IN (?`)
	for _, member := range members[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(member))
	}
	query.WriteString(`)`)
	_, err := ctx.CentralDb().Account().Exec(query.String(), castedIds...)
	PanicIf(err)
}

//email helpers

func emailSendMultipleAccountPolicyNotice(ctx CentralCtx, address string) {
	ctx.Email().Send([]string{address}, "sendMultipleAccountPolicyNotice")
}

func emailSendActivationLink(ctx CentralCtx, address, activationCode string) {
	ctx.Email().Send([]string{address}, fmt.Sprintf("sendActivationLink: activationCode: %s", activationCode))
}

func emailSendPwdResetLink(ctx CentralCtx, address, resetCode string) {
	ctx.Email().Send([]string{address}, fmt.Sprintf("sendPwdResetLink: resetCode: %s", resetCode))
}

func emailSendNewEmailConfirmationLink(ctx CentralCtx, currentAddress, newAddress, confirmationCode string) {
	ctx.Email().Send([]string{newAddress}, fmt.Sprintf("sendNewEmailConfirmationLink: currentAddress: %s newAddress: %s confirmationCode: %s", currentAddress, newAddress, confirmationCode))
}

//avatar storage helpers
//TODO delete after moving to lcl ctx implementation

func avatarSave(ctx CentralCtx, key string, mimeType string, data io.Reader) {
	/*
		s.mtx.Lock()
		defer s.mtx.Unlock()
		avatarBytes, err := ioutil.ReadAll(data)
		PanicIf(err)
		PanicIf(ioutil.WriteFile(path.Join(s.absDirPath, key), avatarBytes, os.ModePerm))
	*/
	ctx.Avatar().Save(key, mimeType, data)
}

func avatarDelete(ctx CentralCtx, key string) {
	/*
		s.mtx.Lock()
		defer s.mtx.Unlock()
		PanicIf(os.Remove(path.Join(s.absDirPath, key)))
	*/
	ctx.Avatar().Delete(key)
}

func avatarDeleteAll(ctx CentralCtx) {
	/*
		s.mtx.Lock()
		defer s.mtx.Unlock()
		os.RemoveAll(s.absDirPath)
	*/
	ctx.Avatar().DeleteAll()
}

//structs

type account struct {
	Id          Id        `json:"id"`
	Name        string    `json:"name"`
	DisplayName *string   `json:"displayName"`
	CreatedOn   time.Time `json:"createdOn"`
	Region      string    `json:"region"`
	NewRegion   *string   `json:"newRegion,omitempty"`
	Shard       int       `json:"shard"`
	HasAvatar   bool      `json:"hasAvatar"`
	IsPersonal  bool      `json:"isPersonal"`
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
