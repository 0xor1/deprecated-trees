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
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// errors
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

//endpoints

var getRegions = &Endpoint{
	Method: GET,
	ValueDlmKeys: func(ctx Ctx, _ interface{}) []string {
		return []string{ctx.Cache().DlmKeyForSystem()}
	},
	Path: "/api/v1/account/getRegions",
	CentralHandler: func(ctx CentralCtx, _ interface{}) interface{} {
		return ctx.RegionalV1PrivateClient().GetRegions()
	},
}

type registerArgs struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Pwd         string  `json:"pwd"`
	Region      string  `json:"region"`
	Language    string  `json:"language"`
	DisplayName *string `json:"displayName"`
	Theme       Theme   `json:"theme"`
}

var register = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/register",
	GetArgsStruct: func() interface{} {
		return &registerArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*registerArgs)
		args.Name = strings.Trim(args.Name, " ")
		ctx.Validate().Name(args.Name)
		args.Email = strings.Trim(args.Email, " ")
		ctx.Validate().Email(args.Email)
		ctx.Validate().Pwd(args.Pwd)
		args.Language = strings.Trim(args.Language, " ") // may need more validation than this at some point to check it is a language we support and not a junk value, but it isnt critical right now
		args.Theme.Validate()
		if args.DisplayName != nil {
			*args.DisplayName = strings.Trim(*args.DisplayName, " ")
			if *args.DisplayName == "" {
				args.DisplayName = nil
			}
		}

		if !ctx.RegionalV1PrivateClient().IsValidRegion(args.Region) {
			noSuchRegionErr.Panic()
		}

		if exists := dbAccountWithCiNameExists(ctx, args.Name); exists {
			nameAlreadyInUseErr.Panic()
		}

		if acc := dbGetPersonalAccountByEmail(ctx, args.Email); acc != nil {
			emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
		}

		activationCode := ctx.Crypt().CreateUrlSafeString()
		acc := &fullPersonalAccountInfo{}
		acc.Id = NewId()
		acc.Name = args.Name
		acc.DisplayName = args.DisplayName
		acc.CreatedOn = ctx.Time().Now()
		acc.Region = args.Region

		defer func() {
			r := recover()
			if r != nil {
				dbDeleteAccountAndAllAssociatedMemberships(ctx, acc.Id)
				panic(r)
			}
		}()
		acc.Shard = ctx.RegionalV1PrivateClient().CreateAccount(acc.Region, acc.Id, acc.Id, acc.Name, acc.DisplayName)
		acc.IsPersonal = true
		acc.Email = args.Email
		acc.Language = args.Language
		acc.Theme = args.Theme
		acc.activationCode = &activationCode

		pwdInfo := &pwdInfo{}
		pwdInfo.salt = ctx.Crypt().CreatePwdSalt()
		pwdInfo.pwd = ctx.Crypt().ScryptKey([]byte(args.Pwd), pwdInfo.salt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())
		pwdInfo.n = ctx.Crypt().ScryptN()
		pwdInfo.r = ctx.Crypt().ScryptR()
		pwdInfo.p = ctx.Crypt().ScryptP()
		pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()

		dbCreatePersonalAccount(ctx, acc, pwdInfo)

		emailSendActivationLink(ctx, args.Email, *acc.activationCode)
		return nil
	},
}

type resendActivationEmailArgs struct {
	Email string `json:"email"`
}

var resendActivationEmail = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/resendActivationEmail",
	GetArgsStruct: func() interface{} {
		return &resendActivationEmailArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*resendActivationEmailArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil || acc.isActivated() {
			return nil
		}
		emailSendActivationLink(ctx, args.Email, *acc.activationCode)
		return nil
	},
}

type activateArgs struct {
	Email          string `json:"email"`
	ActivationCode string `json:"activationCode"`
}

var activate = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/activate",
	GetArgsStruct: func() interface{} {
		return &activateArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*activateArgs)
		args.ActivationCode = strings.Trim(args.ActivationCode, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil || acc.activationCode == nil || args.ActivationCode != *acc.activationCode {
			invalidActivationAttemptErr.Panic()
		}

		acc.activationCode = nil
		activationTime := time.Now().UTC()
		acc.activatedOn = &activationTime
		dbUpdatePersonalAccount(ctx, acc)
		return nil
	},
}

type authenticateArgs struct {
	Email  string `json:"email"`
	PwdTry string `json:"pwdTry"`
}

var authenticate = &Endpoint{
	Method:           POST,
	Path:             "/api/v1/account/authenticate",
	IsAuthentication: true,
	GetArgsStruct: func() interface{} {
		return &authenticateArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*authenticateArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil {
			invalidNameOrPwdErr.Panic()
		}

		pwdInfo := dbGetPwdInfo(ctx, acc.Id)
		scryptPwdTry := ctx.Crypt().ScryptKey([]byte(args.PwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
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
			pwdInfo.pwd = ctx.Crypt().ScryptKey([]byte(args.PwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
			dbUpdatePwdInfo(ctx, acc.Id, pwdInfo)
		}

		return acc.Id
	},
}

type confirmNewEmailArgs struct {
	CurrentEmail     string `json:"currentEmail"`
	NewEmail         string `json:"newEmail"`
	ConfirmationCode string `json:"confirmationCode"`
}

var confirmNewEmail = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/confirmNewEmail",
	GetArgsStruct: func() interface{} {
		return &confirmNewEmailArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*confirmNewEmailArgs)
		acc := dbGetPersonalAccountByEmail(ctx, args.CurrentEmail)
		if acc == nil || acc.NewEmail == nil || args.NewEmail != *acc.NewEmail || acc.newEmailConfirmationCode == nil || args.ConfirmationCode != *acc.newEmailConfirmationCode {
			invalidNewEmailConfirmationAttemptErr.Panic()
		}

		if acc := dbGetPersonalAccountByEmail(ctx, args.NewEmail); acc != nil {
			emailAlreadyInUseErr.Panic()
		}

		acc.Email = args.NewEmail
		acc.NewEmail = nil
		acc.newEmailConfirmationCode = nil
		dbUpdatePersonalAccount(ctx, acc)
		return nil
	},
}

type resetPwdArgs struct {
	Email string `json:"email"`
}

var resetPwd = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/resetPwd",
	GetArgsStruct: func() interface{} {
		return &resetPwdArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*resetPwdArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil {
			return nil
		}

		resetPwdCode := ctx.Crypt().CreateUrlSafeString()

		acc.resetPwdCode = &resetPwdCode
		dbUpdatePersonalAccount(ctx, acc)

		emailSendPwdResetLink(ctx, args.Email, resetPwdCode)
		return nil
	},
}

type setNewPwdFromPwdResetArgs struct {
	Email        string `json:"email"`
	ResetPwdCode string `json:"resetCode"`
	NewPwd       string `json:"newPwd"`
}

var setNewPwdFromPwdReset = &Endpoint{
	Method: POST,
	Path:   "/api/v1/account/setNewPwdFromPwdReset",
	GetArgsStruct: func() interface{} {
		return &setNewPwdFromPwdResetArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setNewPwdFromPwdResetArgs)
		ctx.Validate().Pwd(args.NewPwd)

		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil || acc.resetPwdCode == nil || args.ResetPwdCode != *acc.resetPwdCode {
			invalidResetPwdAttemptErr.Panic()
		}

		scryptSalt := ctx.Crypt().CreatePwdSalt()
		scryptPwd := ctx.Crypt().ScryptKey([]byte(args.NewPwd), scryptSalt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())

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
		return nil
	},
}

type getAccountArgs struct {
	Name string `json:"name"`
}

var getAccount = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/getAccount",
	ResponseStructure: &account{},
	GetArgsStruct: func() interface{} {
		return &getAccountArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*getAccountArgs)
		return dbGetAccountByCiName(ctx, strings.Trim(args.Name, " "))
	},
}

type getAccountsArgs struct {
	Ids []Id `json:"ids"`
}

var getAccounts = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/getAccounts",
	ResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &getAccountsArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*getAccountsArgs)
		ctx.Validate().EntityCount(len(args.Ids))

		return dbGetAccounts(ctx, args.Ids)
	},
}

type searchAccountsArgs struct {
	NameOrDisplayNameStartsWith string `json:"nameOrDisplayNameStartsWith"`
}

var searchAccounts = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/searchAccounts",
	ResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &searchAccountsArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*searchAccountsArgs)
		args.NameOrDisplayNameStartsWith = strings.Trim(args.NameOrDisplayNameStartsWith, " ")
		if utf8.RuneCountInString(args.NameOrDisplayNameStartsWith) < 3 || strings.Contains(args.NameOrDisplayNameStartsWith, "%") {
			InvalidArgumentsErr.Panic()
		}
		return dbSearchAccounts(ctx, args.NameOrDisplayNameStartsWith)
	},
}

type searchPersonalAccountsArgs struct {
	NameOrDisplayNameOrEmailStartsWith string `json:"nameOrDisplayNameOrEmailStartsWith"`
}

var searchPersonalAccounts = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/searchPersonalAccounts",
	ResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &searchPersonalAccountsArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*searchPersonalAccountsArgs)
		args.NameOrDisplayNameOrEmailStartsWith = strings.Trim(args.NameOrDisplayNameOrEmailStartsWith, " ")
		if utf8.RuneCountInString(args.NameOrDisplayNameOrEmailStartsWith) < 3 || strings.Contains(args.NameOrDisplayNameOrEmailStartsWith, "%") {
			InvalidArgumentsErr.Panic()
		}
		return dbSearchPersonalAccounts(ctx, args.NameOrDisplayNameOrEmailStartsWith)
	},
}

var getMe = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/getMe",
	ResponseStructure: &me{},
	RequiresSession:   true,
	CentralHandler: func(ctx CentralCtx, _ interface{}) interface{} {
		acc := dbGetPersonalAccountById(ctx, ctx.MyId())
		if acc == nil {
			noSuchAccountErr.Panic()
		}
		return &acc.me
	},
}

type setMyPwdArgs struct {
	NewPwd string `json:"newPwd"`
	OldPwd string `json:"oldPwd"`
}

var setMyPwd = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/setMyPwd",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyPwdArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setMyPwdArgs)
		ctx.Validate().Pwd(args.NewPwd)

		pwdInfo := dbGetPwdInfo(ctx, ctx.MyId())
		if pwdInfo == nil {
			noSuchAccountErr.Panic()
		}

		scryptPwdTry := ctx.Crypt().ScryptKey([]byte(args.OldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

		if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
			incorrectPwdErr.Panic()
		}

		pwdInfo.salt = ctx.Crypt().CreatePwdSalt()
		pwdInfo.pwd = ctx.Crypt().ScryptKey([]byte(args.NewPwd), pwdInfo.salt, ctx.Crypt().ScryptN(), ctx.Crypt().ScryptR(), ctx.Crypt().ScryptP(), ctx.Crypt().ScryptKeyLen())
		pwdInfo.n = ctx.Crypt().ScryptN()
		pwdInfo.r = ctx.Crypt().ScryptR()
		pwdInfo.p = ctx.Crypt().ScryptP()
		pwdInfo.keyLen = ctx.Crypt().ScryptKeyLen()
		dbUpdatePwdInfo(ctx, ctx.MyId(), pwdInfo)
		return nil
	},
}

type setMyEmailArgs struct {
	NewEmail string `json:"newEmail"`
}

var setMyEmail = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/setMyEmail",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyEmailArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setMyEmailArgs)
		args.NewEmail = strings.Trim(args.NewEmail, " ")
		ctx.Validate().Email(args.NewEmail)

		if acc := dbGetPersonalAccountByEmail(ctx, args.NewEmail); acc != nil {
			emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
		}

		acc := dbGetPersonalAccountById(ctx, ctx.MyId())
		if acc == nil {
			noSuchAccountErr.Panic()
		}

		confirmationCode := ctx.Crypt().CreateUrlSafeString()

		acc.NewEmail = &args.NewEmail
		acc.newEmailConfirmationCode = &confirmationCode
		dbUpdatePersonalAccount(ctx, acc)
		emailSendNewEmailConfirmationLink(ctx, acc.Email, args.NewEmail, confirmationCode)
		return nil
	},
}

var resendMyNewEmailConfirmationEmail = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/resendMyNewEmailConfirmationEmail",
	RequiresSession: true,
	CentralHandler: func(ctx CentralCtx, _ interface{}) interface{} {
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
		return nil
	},
}

type setAccountNameArgs struct {
	AccountId Id     `json:"accountId"`
	NewName   string `json:"newName"`
}

var setAccountName = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/setAccountName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountNameArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setAccountNameArgs)
		args.NewName = strings.Trim(args.NewName, " ")
		ctx.Validate().Name(args.NewName)

		if exists := dbAccountWithCiNameExists(ctx, args.NewName); exists {
			nameAlreadyInUseErr.Panic()
		}

		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			noSuchAccountErr.Panic()
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't rename someone else's personal account
				InsufficientPermissionErr.Panic()
			}

			if !ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId()) {
				InsufficientPermissionErr.Panic()
			}
		}

		acc.Name = args.NewName
		dbUpdateAccount(ctx, acc)

		if ctx.MyId().Equal(args.AccountId) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewName) //first rename myself in my personal org
			var after *Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewName)
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
		}
		return nil
	},
}

type setAccountDisplayNameArgs struct {
	AccountId      Id      `json:"accountId"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setAccountDisplayName = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/setAccountDisplayName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountDisplayNameArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setAccountDisplayNameArgs)
		if args.NewDisplayName != nil {
			*args.NewDisplayName = strings.Trim(*args.NewDisplayName, " ")
			if *args.NewDisplayName == "" {
				args.NewDisplayName = nil
			}
		}

		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			noSuchAccountErr.Panic()
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't rename someone else's personal account
				InsufficientPermissionErr.Panic()
			}

			if !ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId()) {
				InsufficientPermissionErr.Panic()
			}
		}

		if (acc.DisplayName == nil && args.NewDisplayName == nil) || (acc.DisplayName != nil && args.NewDisplayName != nil && *acc.DisplayName == *args.NewDisplayName) {
			return nil //if there is no change, dont do any redundant work
		}

		acc.DisplayName = args.NewDisplayName
		dbUpdateAccount(ctx, acc)

		if ctx.MyId().Equal(args.AccountId) { // if i did set my personal account displayName, i need to update all the stored displayNames in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewDisplayName) //first set my display name in my personal org
			var after *Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewDisplayName)
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
		}
		return nil
	},
}

type setAccountAvatarArgs struct {
	AccountId       Id
	AvatarImageData io.ReadCloser
}

var setAccountAvatar = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/setAccountAvatar",
	RequiresSession: true,
	ProcessForm: func(w http.ResponseWriter, r *http.Request) interface{} {
		r.Body = http.MaxBytesReader(w, r.Body, 600000) //limit to 6kb
		f, _, err := r.FormFile("avatar")
		if err != nil {
			panic(err)
		}
		return &setAccountAvatarArgs{
			AccountId:       ParseId(r.FormValue("accountId")),
			AvatarImageData: f,
		}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*setAccountAvatarArgs)
		if args.AvatarImageData != nil {
			defer args.AvatarImageData.Close()
		}

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			noSuchAccountErr.Panic()
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if account.IsPersonal { // can't set avatar on someone else's personal account
				InsufficientPermissionErr.Panic()
			}

			if !ctx.RegionalV1PrivateClient().MemberIsAccountOwner(account.Region, account.Shard, args.AccountId, ctx.MyId()) {
				InsufficientPermissionErr.Panic()
			}
		}

		if args.AvatarImageData != nil {
			avatarImage, _, err := image.Decode(args.AvatarImageData)
			ctx.Error().PanicIf(err)
			bounds := avatarImage.Bounds()
			if bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y { //if it  isn't square, then error
				invalidAvatarShapeErr.Panic()
			}
			if uint(bounds.Max.X-bounds.Min.X) > ctx.Avatar().MaxAvatarDim() { // if it is larger than allowed then resize
				avatarImage = resize.Resize(ctx.Avatar().MaxAvatarDim(), ctx.Avatar().MaxAvatarDim(), avatarImage, resize.NearestNeighbor)
			}
			buff := &bytes.Buffer{}
			ctx.Error().PanicIf(png.Encode(buff, avatarImage))
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
		return nil
	},
}

type migrateAccountArgs struct {
	AccountId Id     `json:"accountId"`
	NewRegion string `json:"newRegion"`
}

var migrateAccount = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/migrateAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &migrateAccountArgs{}
	},
	CentralHandler: func(ctx CentralCtx, _ interface{}) interface{} {
		NotImplementedErr.Panic()
		return nil
	},
}

type createAccountArgs struct {
	Name        string  `json:"name"`
	Region      string  `json:"region"`
	DisplayName *string `json:"displayName"`
}

var createAccount = &Endpoint{
	Method:            POST,
	Path:              "/api/v1/account/createAccount",
	RequiresSession:   true,
	ResponseStructure: &account{},
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		args.Name = strings.Trim(args.Name, " ")
		ctx.Validate().Name(args.Name)

		if !ctx.RegionalV1PrivateClient().IsValidRegion(args.Region) {
			noSuchRegionErr.Panic()
		}

		if exists := dbAccountWithCiNameExists(ctx, args.Name); exists {
			nameAlreadyInUseErr.Panic()
		}

		account := &account{}
		account.Id = NewId()
		account.Name = args.Name
		account.DisplayName = args.DisplayName
		account.CreatedOn = ctx.Time().Now()
		account.Region = args.Region
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
		shard := ctx.RegionalV1PrivateClient().CreateAccount(args.Region, account.Id, ctx.MyId(), owner.Name, owner.DisplayName)

		account.Shard = shard
		dbUpdateAccount(ctx, account)
		return account
	},
}

type getMyAccountsArgs struct {
	After *Id `json:"after"`
	Limit int `json:"limit"`
}

type getMyAccountsResp struct {
	Accounts []*account `json:"accounts"`
	More     bool       `json:"more"`
}

var getMyAccounts = &Endpoint{
	Method:            GET,
	Path:              "/api/v1/account/getMyAccounts",
	ResponseStructure: &getMyAccountsResp{},
	RequiresSession:   true,
	GetArgsStruct: func() interface{} {
		return &getMyAccountsArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*getMyAccountsArgs)
		res := &getMyAccountsResp{}
		res.Accounts, res.More = dbGetGroupAccounts(ctx, ctx.MyId(), args.After, ctx.Validate().Limit(args.Limit))
		return res
	},
}

type deleteAccountArgs struct {
	AccountId Id `json:"accountId"`
}

var deleteAccount = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/deleteAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			noSuchAccountErr.Panic()
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't delete someone else's personal account
				InsufficientPermissionErr.Panic()
			}
			//otherwise attempting to delete a group account
			if !ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId()) {
				InsufficientPermissionErr.Panic()
			}
		}

		ctx.RegionalV1PrivateClient().DeleteAccount(acc.Region, acc.Shard, args.AccountId, ctx.MyId())
		dbDeleteAccountAndAllAssociatedMemberships(ctx, args.AccountId)

		if ctx.MyId().Equal(args.AccountId) {
			var after *Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
				for _, acc := range accs {
					if ctx.RegionalV1PrivateClient().MemberIsOnlyAccountOwner(acc.Region, acc.Shard, acc.Id, ctx.MyId()) {
						onlyOwnerMemberErr.Panic()
					}
				}
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().RemoveMembers(acc.Region, acc.Shard, acc.Id, ctx.MyId(), []Id{ctx.MyId()})
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
		}
		return nil
	},
}

type addMembersArgs struct {
	AccountId  Id                 `json:"accountId"`
	NewMembers []*AddMemberPublic `json:"newMembers"`
}

var addMembers = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/addMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		if args.AccountId.Equal(ctx.MyId()) {
			InvalidOperationErr.Panic()
		}
		ctx.Validate().EntityCount(len(args.NewMembers))

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			noSuchAccountErr.Panic()
		}

		ids := make([]Id, 0, len(args.NewMembers))
		addMembersMap := map[string]*AddMemberPublic{}
		for _, member := range args.NewMembers {
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

		ctx.RegionalV1PrivateClient().AddMembers(account.Region, account.Shard, args.AccountId, ctx.MyId(), members)
		dbCreateMemberships(ctx, args.AccountId, ids)
		return nil
	},
}

type removeMembersArgs struct {
	AccountId       Id   `json:"accountId"`
	ExistingMembers []Id `json:"existingMembers"`
}

var removeMembers = &Endpoint{
	Method:          POST,
	Path:            "/api/v1/account/removeMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CentralHandler: func(ctx CentralCtx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		if args.AccountId.Equal(ctx.MyId()) {
			InvalidOperationErr.Panic()
		}
		ctx.Validate().EntityCount(len(args.ExistingMembers))

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			noSuchAccountErr.Panic()
		}

		ctx.RegionalV1PrivateClient().RemoveMembers(account.Region, account.Shard, args.AccountId, ctx.MyId(), args.ExistingMembers)
		dbDeleteMemberships(ctx, args.AccountId, args.ExistingMembers)
		return nil
	},
}

var Endpoints = []*Endpoint{
	getRegions,
	register,
	resendActivationEmail,
	activate,
	authenticate,
	confirmNewEmail,
	resetPwd,
	setNewPwdFromPwdReset,
	getAccount,
	getAccounts,
	searchAccounts,
	searchPersonalAccounts,
	getMe,
	setMyPwd,
	setMyEmail,
	resendMyNewEmailConfirmationEmail,
	setAccountName,
	setAccountDisplayName,
	setAccountAvatar,
	migrateAccount,
	createAccount,
	getMyAccounts,
	deleteAccount,
	addMembers,
	removeMembers,
}

// The main account Client interface
type Client interface {
	//accessible outside of active session
	GetRegions() ([]string, error)
	Register(name, email, pwd, region, language string, displayName *string, theme Theme) error
	ResendActivationEmail(email string) error
	Activate(email, activationCode string) error
	Authenticate(email, pwd string) (Id, error)
	ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error
	ResetPwd(email string) error
	SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error
	GetAccount(name string) (*account, error)
	GetAccounts(ids []Id) ([]*account, error)
	SearchAccounts(nameOrDisplayNameStartsWith string) ([]*account, error)
	SearchPersonalAccounts(nameOrDisplayNameOrEmailStartsWith string) ([]*account, error)
	//requires active session to access
	GetMe() (*me, error)
	SetMyPwd(oldPwd, newPwd string) error
	SetMyEmail(newEmail string) error
	ResendMyNewEmailConfirmationEmail() error
	SetAccountName(accountId Id, newName string) error
	SetAccountDisplayName(accountId Id, newDisplayName *string) error
	SetAccountAvatar(accountId Id, avatarImage io.ReadCloser) error
	MigrateAccount(accountId Id, newRegion string) error
	CreateAccount(name, region string, displayName *string) (*account, error)
	GetMyAccounts(after *Id, limit int) (*getMyAccountsResp, error)
	DeleteAccount(accountId Id) error
	//member centric - must be an owner or admin
	AddMembers(accountId Id, newMembers []*AddMemberPublic) error
	RemoveMembers(accountId Id, existingMembers []Id) error
}

type client struct {
	schema string
	host   string
}

func (c *client) GetRegions() ([]string, error) {
	val, err := getRegions.DoRequest(c.schema, c.host, nil, nil, []string{})
	if val != nil {
		return val.([]string), err
	}
	return nil, err
}

func (c *client) Register(name, email, pwd, region, language string, displayName *string, theme Theme) error {
	_, err := register.DoRequest(c.schema, c.host, &registerArgs{
		Name:        name,
		Email:       email,
		Pwd:         pwd,
		Region:      region,
		Language:    language,
		DisplayName: displayName,
		Theme:       theme,
	}, nil, nil)
	return err
}

func (c *client) ResendActivationEmail(email string) error {
	_, err := resendActivationEmail.DoRequest(c.schema, c.host, &resendActivationEmailArgs{
		Email: email,
	}, nil, nil)
	return err
}

func (c *client) Activate(email, activationCode string) error {
	_, err := activate.DoRequest(c.schema, c.host, &activateArgs{
		Email:          email,
		ActivationCode: activationCode,
	}, nil, nil)
	return err
}

func (c *client) Authenticate(email, pwdTry string) (Id, error) {
	val, err := authenticate.DoRequest(c.schema, c.host, &authenticateArgs{
		Email:  email,
		PwdTry: pwdTry,
	}, nil, &Id{})
	if val != nil {
		return *val.(*Id), err
	}
	return nil, err
}

func (c *client) ConfirmNewEmail(currentEmail, newEmail, confirmationCode string) error {
	_, err := confirmNewEmail.DoRequest(c.schema, c.host, &confirmNewEmailArgs{
		CurrentEmail:     currentEmail,
		NewEmail:         newEmail,
		ConfirmationCode: confirmationCode,
	}, nil, nil)
	return err
}

func (c *client) ResetPwd(email string) error {
	_, err := resetPwd.DoRequest(c.schema, c.host, &resetPwdArgs{
		Email: email,
	}, nil, nil)
	return err
}

func (c *client) SetNewPwdFromPwdReset(newPwd, email, resetPwdCode string) error {
	_, err := setNewPwdFromPwdReset.DoRequest(c.schema, c.host, &setNewPwdFromPwdResetArgs{
		NewPwd:       newPwd,
		Email:        email,
		ResetPwdCode: resetPwdCode,
	}, nil, nil)
	return err
}

func (c *client) GetAccount(name string) (*account, error) {
	val, err := getAccount.DoRequest(c.schema, c.host, &getAccountArgs{
		Name: name,
	}, nil, &account{})
	if val != nil {
		return val.(*account), err
	}
	return nil, err
}

func (c *client) GetAccounts(ids []Id) ([]*account, error) {
	val, err := getAccounts.DoRequest(c.schema, c.host, &getAccountsArgs{
		Ids: ids,
	}, nil, []*account{})
	if val != nil {
		return val.([]*account), err
	}
	return nil, err
}

func (c *client) SearchAccounts(nameOrDisplayNameStartsWith string) ([]*account, error) {
	val, err := searchAccounts.DoRequest(c.schema, c.host, &searchAccountsArgs{
		NameOrDisplayNameStartsWith: nameOrDisplayNameStartsWith,
	}, nil, []*account{})
	if val != nil {
		return val.([]*account), err
	}
	return nil, err
}

func (c *client) SearchPersonalAccounts(nameOrDisplayNameOrEmailStartsWith string) ([]*account, error) {
	val, err := searchPersonalAccounts.DoRequest(c.schema, c.host, &searchPersonalAccountsArgs{
		NameOrDisplayNameOrEmailStartsWith: nameOrDisplayNameOrEmailStartsWith,
	}, nil, []*account{})
	if val != nil {
		return val.([]*account), err
	}
	return nil, err
}

func (c *client) GetMe() (*me, error) {
	val, err := getMe.DoRequest(c.schema, c.host, nil, nil, &me{})
	if val != nil {
		return val.(*me), err
	}
	return nil, err
}

func (c *client) SetMyPwd(oldPwd, newPwd string) error {
	_, err := setMyPwd.DoRequest(c.schema, c.host, &setMyPwdArgs{
		OldPwd: oldPwd,
		NewPwd: newPwd,
	}, nil, nil)
	return err
}

func (c *client) SetMyEmail(newEmail string) error {
	_, err := setMyEmail.DoRequest(c.schema, c.host, &setMyEmailArgs{
		NewEmail: newEmail,
	}, nil, nil)
	return err
}

func (c *client) ResendMyNewEmailConfirmationEmail(email string) error {
	_, err := resendActivationEmail.DoRequest(c.schema, c.host, &resendActivationEmailArgs{
		Email: email,
	}, nil, nil)
	return err
}

func (c *client) SetAccountName(accountId Id, newName string) error {
	_, err := setAccountName.DoRequest(c.schema, c.host, &setAccountNameArgs{
		AccountId: accountId,
		NewName:   newName,
	}, nil, nil)
	return err
}

func (c *client) SetAccountDisplayName(accountId Id, newDisplayName *string) error {
	_, err := setAccountDisplayName.DoRequest(c.schema, c.host, &setAccountDisplayNameArgs{
		AccountId:      accountId,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return err
}

func (c *client) SetAccountAvatar(accountId Id, avatarImageData io.ReadCloser) error {
	defer avatarImageData.Close()
	_, err := setAccountAvatar.DoRequest(c.schema, c.host, &setAccountAvatarArgs{
		AccountId:       accountId,
		AvatarImageData: avatarImageData,
	}, nil, nil)
	return err
}

func (c *client) MigrateAccount(accountId Id, newRegion string) error {
	_, err := migrateAccount.DoRequest(c.schema, c.host, &migrateAccountArgs{
		AccountId: accountId,
		NewRegion: newRegion,
	}, nil, nil)
	return err
}

func (c *client) CreateAccount(name, region string, displayName *string) (*account, error) {
	val, err := createAccount.DoRequest(c.schema, c.host, &createAccountArgs{
		Name:        name,
		Region:      region,
		DisplayName: displayName,
	}, nil, &account{})
	if val != nil {
		return val.(*account), err
	}
	return nil, err
}

func (c *client) GetMyAccounts(after *Id, limit int) (*getMyAccountsResp, error) {
	val, err := getMyAccounts.DoRequest(c.schema, c.host, &getMyAccountsArgs{
		After: after,
		Limit: limit,
	}, nil, &getMyAccountsResp{})
	if val != nil {
		return val.(*getMyAccountsResp), err
	}
	return nil, err
}

func (c *client) DeleteAccount(accountId Id) error {
	_, err := deleteAccount.DoRequest(c.schema, c.host, &deleteAccountArgs{
		AccountId: accountId,
	}, nil, nil)
	return err
}

func (c *client) AddMembers(accountId Id, newMembers []*AddMemberPublic) error {
	_, err := addMembers.DoRequest(c.schema, c.host, &addMembersArgs{
		AccountId:  accountId,
		NewMembers: newMembers,
	}, nil, nil)
	return err
}

func (c *client) RemoveMembers(accountId Id, existingMembers []Id) error {
	_, err := removeMembers.DoRequest(c.schema, c.host, &removeMembersArgs{
		AccountId:       accountId,
		ExistingMembers: existingMembers,
	}, nil, nil)
	return err
}

//internal helpers

//db helpers
func dbAccountWithCiNameExists(ctx CentralCtx, name string) bool {
	row := ctx.CentralDb().Account().QueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?`, name)
	count := 0
	ctx.Error().PanicIf(row.Scan(&count))
	return count != 0
}

func dbGetAccountByCiName(ctx CentralCtx, name string) *account {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?`, name)
	acc := account{}
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)) {
		return nil
	}
	return &acc
}

func dbCreatePersonalAccount(ctx CentralCtx, account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	id := []byte(account.Id)
	_, err := ctx.CentralDb().Account().Exec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode)
	ctx.Error().PanicIf(err)
	_, err = ctx.CentralDb().Pwd().Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	ctx.Error().PanicIf(err)
}

func dbGetPersonalAccountByEmail(ctx CentralCtx, email string) *fullPersonalAccountInfo {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPersonalAccountById(ctx CentralCtx, id Id) *fullPersonalAccountInfo {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, []byte(id))
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPwdInfo(ctx CentralCtx, id Id) *pwdInfo {
	row := ctx.CentralDb().Pwd().QueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?`, []byte(id))
	pwd := pwdInfo{}
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)) {
		return nil
	}
	return &pwd
}

func dbUpdatePersonalAccount(ctx CentralCtx, personalAccountInfo *fullPersonalAccountInfo) {
	_, err := ctx.CentralDb().Account().Exec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(personalAccountInfo.Id), personalAccountInfo.Name, personalAccountInfo.DisplayName, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode)
	ctx.Error().PanicIf(err)
}

func dbUpdateAccount(ctx CentralCtx, account *account) {
	_, err := ctx.CentralDb().Account().Exec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal)
	ctx.Error().PanicIf(err)
}

func dbUpdatePwdInfo(ctx CentralCtx, id Id, pwdInfo *pwdInfo) {
	_, err := ctx.CentralDb().Pwd().Exec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id))
	ctx.Error().PanicIf(err)
}

func dbDeleteAccountAndAllAssociatedMemberships(ctx CentralCtx, id Id) {
	castId := []byte(id)
	_, err := ctx.CentralDb().Account().Exec(`CALL deleteAccountAndAllAssociatedMemberships(?)`, castId)
	ctx.Error().PanicIf(err)
	_, err = ctx.CentralDb().Pwd().Exec(`DELETE FROM pwds WHERE id = ?`, castId)
	ctx.Error().PanicIf(err)
}

func dbGetAccount(ctx CentralCtx, id Id) *account {
	row := ctx.CentralDb().Account().QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?`, []byte(id))
	a := account{}
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)) {
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
	ctx.Error().PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		ctx.Error().PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
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
	ctx.Error().PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		ctx.Error().PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal))
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
	ctx.Error().PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		ctx.Error().PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
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
	ctx.Error().PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		ctx.Error().PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbCreateGroupAccountAndMembership(ctx CentralCtx, account *account, memberId Id) {
	_, err := ctx.CentralDb().Account().Exec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, []byte(memberId))
	ctx.Error().PanicIf(err)
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
	ctx.Error().PanicIf(err)
	res := make([]*account, 0, limit+1)
	for rows.Next() {
		a := account{}
		ctx.Error().PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
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
	ctx.Error().PanicIf(err)
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
	ctx.Error().PanicIf(err)
}

//email helpers

func emailSendMultipleAccountPolicyNotice(ctx CentralCtx, address string) {
	ctx.Mail().Send([]string{address}, "sendMultipleAccountPolicyNotice")
}

func emailSendActivationLink(ctx CentralCtx, address, activationCode string) {
	ctx.Mail().Send([]string{address}, fmt.Sprintf("sendActivationLink: activationCode: %s", activationCode))
}

func emailSendPwdResetLink(ctx CentralCtx, address, resetCode string) {
	ctx.Mail().Send([]string{address}, fmt.Sprintf("sendPwdResetLink: resetCode: %s", resetCode))
}

func emailSendNewEmailConfirmationLink(ctx CentralCtx, currentAddress, newAddress, confirmationCode string) {
	ctx.Mail().Send([]string{newAddress}, fmt.Sprintf("sendNewEmailConfirmationLink: currentAddress: %s newAddress: %s confirmationCode: %s", currentAddress, newAddress, confirmationCode))
}

//avatar storage helpers
//TODO delete after moving to lcl ctx implementation

func avatarSave(ctx CentralCtx, key string, mimeType string, data io.Reader) {
	/*
		s.mtx.Lock()
		defer s.mtx.Unlock()
		avatarBytes, err := ioutil.ReadAll(data)
		ctx.Error().PanicIf(err)
		ctx.Error().PanicIf(ioutil.WriteFile(path.Join(s.absDirPath, key), avatarBytes, os.ModePerm))
	*/
	ctx.Avatar().Save(key, mimeType, data)
}

func avatarDelete(ctx CentralCtx, key string) {
	/*
		s.mtx.Lock()
		defer s.mtx.Unlock()
		ctx.Error().PanicIf(os.Remove(path.Join(s.absDirPath, key)))
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
