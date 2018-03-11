package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/core"
	"bitbucket.org/0xor1/task/server/util/crypt"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// errors
	noSuchAccountErr                      = &err.Err{Code: "c_v1_a_nsa", Message: "no such account"}
	invalidActivationAttemptErr           = &err.Err{Code: "c_v1_a_iaa", Message: "invalid activation attempt"}
	invalidResetPwdAttemptErr             = &err.Err{Code: "c_v1_a_irpa", Message: "invalid reset password attempt"}
	invalidNewEmailConfirmationAttemptErr = &err.Err{Code: "c_v1_a_ineca", Message: "invalid new email confirmation attempt"}
	invalidNameOrPwdErr                   = &err.Err{Code: "c_v1_a_inop", Message: "invalid name or password"}
	incorrectPwdErr                       = &err.Err{Code: "c_v1_a_ip", Message: "password incorrect"}
	accountNotActivatedErr                = &err.Err{Code: "c_v1_a_ana", Message: "account not activated"}
	emailAlreadyInUseErr                  = &err.Err{Code: "c_v1_a_eaiu", Message: "email already in use"}
	nameAlreadyInUseErr                   = &err.Err{Code: "c_v1_a_naiu", Message: "name already in use"}
	emailConfirmationCodeErr              = &err.Err{Code: "c_v1_a_ecc", Message: "email confirmation code is of zero length"}
	noNewEmailRegisteredErr               = &err.Err{Code: "c_v1_a_nner", Message: "no new email registered"}
	onlyOwnerMemberErr                    = &err.Err{Code: "c_v1_a_oom", Message: "can't delete member who is the only owner of an account"}
	invalidAvatarShapeErr                 = &err.Err{Code: "c_v1_a_ias", Message: "avatar images must be square"}
)

//endpoints

var getRegions = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getRegions",
	ValueDlmKeys: func(ctx *core.Ctx, _ interface{}) []string {
		return []string{ctx.DlmKeyForSystem()}
	},
	CtxHandler: func(ctx *core.Ctx, _ interface{}) interface{} {
		return ctx.RegionalV1PrivateClient().GetRegions()
	},
}

type registerArgs struct {
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Pwd         string     `json:"pwd"`
	Region      string     `json:"region"`
	Language    string     `json:"language"`
	DisplayName *string    `json:"displayName"`
	Theme       cnst.Theme `json:"theme"`
}

var register = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/register",
	GetArgsStruct: func() interface{} {
		return &registerArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*registerArgs)
		args.Name = strings.Trim(args.Name, " ")
		validate.StringArg("name", args.Name, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())
		args.Email = strings.Trim(args.Email, " ")
		validate.Email(args.Email)
		validate.StringArg("pwd", args.Pwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())
		args.Language = strings.Trim(args.Language, " ") // may need more validation than this at some point to check it is a language we support and not a junk value, but it isnt critical right now
		args.Theme.Validate()
		if args.DisplayName != nil {
			*args.DisplayName = strings.Trim(*args.DisplayName, " ")
			if *args.DisplayName == "" {
				args.DisplayName = nil
			}
		}

		if !ctx.RegionalV1PrivateClient().IsValidRegion(args.Region) {
			panic(err.NoSuchRegion)
		}

		if exists := dbAccountWithCiNameExists(ctx, args.Name); exists {
			panic(nameAlreadyInUseErr)
		}

		if acc := dbGetPersonalAccountByEmail(ctx, args.Email); acc != nil {
			emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
		}

		activationCode := crypt.UrlSafeString(ctx.CryptCodeLen())
		acc := &fullPersonalAccountInfo{}
		acc.Id = id.New()
		acc.Name = args.Name
		acc.DisplayName = args.DisplayName
		acc.CreatedOn = t.Now()
		acc.Region = args.Region

		defer func() {
			r := recover()
			if r != nil {
				dbDeleteAccountAndAllAssociatedMemberships(ctx, acc.Id)
				panic(r)
			}
		}()
		var e error
		acc.Shard, e = ctx.RegionalV1PrivateClient().CreateAccount(acc.Region, acc.Id, acc.Id, acc.Name, acc.DisplayName)
		err.PanicIf(e)
		acc.IsPersonal = true
		acc.Email = args.Email
		acc.Language = args.Language
		acc.Theme = args.Theme
		acc.activationCode = &activationCode

		pwdInfo := &pwdInfo{}
		pwdInfo.salt = crypt.Bytes(ctx.SaltLen())
		pwdInfo.pwd = crypt.ScryptKey([]byte(args.Pwd), pwdInfo.salt, ctx.ScryptN(), ctx.ScryptR(), ctx.ScryptP(), ctx.ScryptKeyLen())
		pwdInfo.n = ctx.ScryptN()
		pwdInfo.r = ctx.ScryptR()
		pwdInfo.p = ctx.ScryptP()
		pwdInfo.keyLen = ctx.ScryptKeyLen()

		dbCreatePersonalAccount(ctx, acc, pwdInfo)

		emailSendActivationLink(ctx, args.Email, *acc.activationCode)
		return nil
	},
}

type resendActivationEmailArgs struct {
	Email string `json:"email"`
}

var resendActivationEmail = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/resendActivationEmail",
	GetArgsStruct: func() interface{} {
		return &resendActivationEmailArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
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

var activate = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/activate",
	GetArgsStruct: func() interface{} {
		return &activateArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*activateArgs)
		args.ActivationCode = strings.Trim(args.ActivationCode, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil || acc.activationCode == nil || args.ActivationCode != *acc.activationCode {
			panic(invalidActivationAttemptErr)
		}

		acc.activationCode = nil
		activationTime := t.Now()
		acc.activatedOn = &activationTime
		dbUpdatePersonalAccount(ctx, acc)
		return nil
	},
}

type authenticateArgs struct {
	Email  string `json:"email"`
	PwdTry string `json:"pwdTry"`
}

var authenticate = &core.Endpoint{
	Method:           cnst.POST,
	Path:             "/api/v1/centralAccount/authenticate",
	IsAuthentication: true,
	GetArgsStruct: func() interface{} {
		return &authenticateArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*authenticateArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil {
			panic(invalidNameOrPwdErr)
		}

		pwdInfo := dbGetPwdInfo(ctx, acc.Id)
		scryptPwdTry := crypt.ScryptKey([]byte(args.PwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
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
			dbUpdatePersonalAccount(ctx, acc)
		}
		// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
		if pwdInfo.n != ctx.ScryptN() || pwdInfo.r != ctx.ScryptR() || pwdInfo.p != ctx.ScryptP() || pwdInfo.keyLen != ctx.ScryptKeyLen() || len(pwdInfo.salt) != ctx.SaltLen() {
			pwdInfo.salt = crypt.Bytes(ctx.SaltLen())
			pwdInfo.n = ctx.ScryptN()
			pwdInfo.r = ctx.ScryptR()
			pwdInfo.p = ctx.ScryptP()
			pwdInfo.keyLen = ctx.ScryptKeyLen()
			pwdInfo.pwd = crypt.ScryptKey([]byte(args.PwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
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

var confirmNewEmail = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/confirmNewEmail",
	GetArgsStruct: func() interface{} {
		return &confirmNewEmailArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*confirmNewEmailArgs)
		acc := dbGetPersonalAccountByEmail(ctx, args.CurrentEmail)
		if acc == nil || acc.NewEmail == nil || args.NewEmail != *acc.NewEmail || acc.newEmailConfirmationCode == nil || args.ConfirmationCode != *acc.newEmailConfirmationCode {
			panic(invalidNewEmailConfirmationAttemptErr)
		}

		if acc := dbGetPersonalAccountByEmail(ctx, args.NewEmail); acc != nil {
			panic(emailAlreadyInUseErr)
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

var resetPwd = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/resetPwd",
	GetArgsStruct: func() interface{} {
		return &resetPwdArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*resetPwdArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil {
			return nil
		}

		resetPwdCode := crypt.UrlSafeString(ctx.CryptCodeLen())

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

var setNewPwdFromPwdReset = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/setNewPwdFromPwdReset",
	GetArgsStruct: func() interface{} {
		return &setNewPwdFromPwdResetArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setNewPwdFromPwdResetArgs)
		validate.StringArg("pwd", args.NewPwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())

		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		if acc == nil || acc.resetPwdCode == nil || args.ResetPwdCode != *acc.resetPwdCode {
			panic(invalidResetPwdAttemptErr)
		}

		scryptSalt := crypt.Bytes(ctx.SaltLen())
		scryptPwd := crypt.ScryptKey([]byte(args.NewPwd), scryptSalt, ctx.ScryptN(), ctx.ScryptR(), ctx.ScryptP(), ctx.ScryptKeyLen())

		acc.activationCode = nil
		acc.resetPwdCode = nil
		dbUpdatePersonalAccount(ctx, acc)

		pwdInfo := &pwdInfo{}
		pwdInfo.pwd = scryptPwd
		pwdInfo.salt = scryptSalt
		pwdInfo.n = ctx.ScryptN()
		pwdInfo.r = ctx.ScryptR()
		pwdInfo.p = ctx.ScryptP()
		pwdInfo.keyLen = ctx.ScryptKeyLen()
		dbUpdatePwdInfo(ctx, acc.Id, pwdInfo)
		return nil
	},
}

type getAccountArgs struct {
	Name string `json:"name"`
}

var getAccount = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getAccount",
	ExampleResponseStructure: &account{},
	GetArgsStruct: func() interface{} {
		return &getAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*getAccountArgs)
		return dbGetAccountByCiName(ctx, strings.Trim(args.Name, " "))
	},
}

type getAccountsArgs struct {
	Accounts []id.Id `json:"accounts"`
}

var getAccounts = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getAccounts",
	ExampleResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &getAccountsArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*getAccountsArgs)
		validate.EntityCount(len(args.Accounts), ctx.MaxProcessEntityCount())

		return dbGetAccounts(ctx, args.Accounts)
	},
}

type searchAccountsArgs struct {
	NameOrDisplayNameStartsWith string `json:"nameOrDisplayNameStartsWith"`
}

var searchAccounts = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/searchAccounts",
	ExampleResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &searchAccountsArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*searchAccountsArgs)
		args.NameOrDisplayNameStartsWith = strings.Trim(args.NameOrDisplayNameStartsWith, " ")
		if utf8.RuneCountInString(args.NameOrDisplayNameStartsWith) < 3 || strings.Contains(args.NameOrDisplayNameStartsWith, "%") {
			panic(err.InvalidArguments)
		}
		return dbSearchAccounts(ctx, args.NameOrDisplayNameStartsWith)
	},
}

type searchPersonalAccountsArgs struct {
	NameOrDisplayNameOrEmailStartsWith string `json:"nameOrDisplayNameOrEmailStartsWith"`
}

var searchPersonalAccounts = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/searchPersonalAccounts",
	ExampleResponseStructure: []*account{{}},
	GetArgsStruct: func() interface{} {
		return &searchPersonalAccountsArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*searchPersonalAccountsArgs)
		args.NameOrDisplayNameOrEmailStartsWith = strings.Trim(args.NameOrDisplayNameOrEmailStartsWith, " ")
		if utf8.RuneCountInString(args.NameOrDisplayNameOrEmailStartsWith) < 3 || strings.Contains(args.NameOrDisplayNameOrEmailStartsWith, "%") {
			panic(err.InvalidArguments)
		}
		return dbSearchPersonalAccounts(ctx, args.NameOrDisplayNameOrEmailStartsWith)
	},
}

var getMe = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getMe",
	ExampleResponseStructure: &me{},
	RequiresSession:          true,
	CtxHandler: func(ctx *core.Ctx, _ interface{}) interface{} {
		acc := dbGetPersonalAccountById(ctx, ctx.MyId())
		if acc == nil {
			panic(noSuchAccountErr)
		}
		return &acc.me
	},
}

type setMyPwdArgs struct {
	NewPwd string `json:"newPwd"`
	OldPwd string `json:"oldPwd"`
}

var setMyPwd = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/setMyPwd",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyPwdArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setMyPwdArgs)
		validate.StringArg("pwd", args.NewPwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())

		pwdInfo := dbGetPwdInfo(ctx, ctx.MyId())
		if pwdInfo == nil {
			panic(noSuchAccountErr)
		}

		scryptPwdTry := crypt.ScryptKey([]byte(args.OldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

		if !pwdsMatch(pwdInfo.pwd, scryptPwdTry) {
			panic(incorrectPwdErr)
		}

		pwdInfo.salt = crypt.Bytes(ctx.SaltLen())
		pwdInfo.pwd = crypt.ScryptKey([]byte(args.NewPwd), pwdInfo.salt, ctx.ScryptN(), ctx.ScryptR(), ctx.ScryptP(), ctx.ScryptKeyLen())
		pwdInfo.n = ctx.ScryptN()
		pwdInfo.r = ctx.ScryptR()
		pwdInfo.p = ctx.ScryptP()
		pwdInfo.keyLen = ctx.ScryptKeyLen()
		dbUpdatePwdInfo(ctx, ctx.MyId(), pwdInfo)
		return nil
	},
}

type setMyEmailArgs struct {
	NewEmail string `json:"newEmail"`
}

var setMyEmail = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/setMyEmail",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyEmailArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setMyEmailArgs)
		args.NewEmail = strings.Trim(args.NewEmail, " ")
		validate.Email(args.NewEmail)

		if acc := dbGetPersonalAccountByEmail(ctx, args.NewEmail); acc != nil {
			emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
		}

		acc := dbGetPersonalAccountById(ctx, ctx.MyId())
		if acc == nil {
			panic(noSuchAccountErr)
		}

		confirmationCode := crypt.UrlSafeString(ctx.CryptCodeLen())

		acc.NewEmail = &args.NewEmail
		acc.newEmailConfirmationCode = &confirmationCode
		dbUpdatePersonalAccount(ctx, acc)
		emailSendNewEmailConfirmationLink(ctx, acc.Email, args.NewEmail, confirmationCode)
		return nil
	},
}

var resendMyNewEmailConfirmationEmail = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/resendMyNewEmailConfirmationEmail",
	RequiresSession: true,
	CtxHandler: func(ctx *core.Ctx, _ interface{}) interface{} {
		acc := dbGetPersonalAccountById(ctx, ctx.MyId())
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

		emailSendNewEmailConfirmationLink(ctx, acc.Email, *acc.NewEmail, *acc.newEmailConfirmationCode)
		return nil
	},
}

type setAccountNameArgs struct {
	AccountId id.Id  `json:"accountId"`
	NewName   string `json:"newName"`
}

var setAccountName = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/setAccountName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountNameArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setAccountNameArgs)
		args.NewName = strings.Trim(args.NewName, " ")
		validate.StringArg("name", args.NewName, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())

		if exists := dbAccountWithCiNameExists(ctx, args.NewName); exists {
			panic(nameAlreadyInUseErr)
		}

		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't rename someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		acc.Name = args.NewName
		dbUpdateAccount(ctx, acc)

		if ctx.MyId().Equal(args.AccountId) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewName) //first rename myself in my personal org
			var after *id.Id
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
	AccountId      id.Id   `json:"accountId"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setAccountDisplayName = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/setAccountDisplayName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountDisplayNameArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setAccountDisplayNameArgs)
		if args.NewDisplayName != nil {
			*args.NewDisplayName = strings.Trim(*args.NewDisplayName, " ")
			if *args.NewDisplayName == "" {
				args.NewDisplayName = nil
			}
		}

		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't rename someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		if (acc.DisplayName == nil && args.NewDisplayName == nil) || (acc.DisplayName != nil && args.NewDisplayName != nil && *acc.DisplayName == *args.NewDisplayName) {
			return nil //if there is no change, dont do any redundant work
		}

		acc.DisplayName = args.NewDisplayName
		dbUpdateAccount(ctx, acc)

		if ctx.MyId().Equal(args.AccountId) { // if i did set my personal account displayName, i need to update all the stored displayNames in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.MyId(), args.NewDisplayName) //first set my display name in my personal org
			var after *id.Id
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
	AccountId       id.Id
	AvatarImageData io.ReadCloser
}

var setAccountAvatar = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/setAccountAvatar",
	RequiresSession: true,
	FormStruct: map[string]string{
		"account": "Id",
		"avatar":  "file (png, jpeg, gif)",
	},
	ProcessForm: func(w http.ResponseWriter, r *http.Request) interface{} {
		r.Body = http.MaxBytesReader(w, r.Body, 600000) //limit to 6kb
		f, _, err := r.FormFile("avatar")
		if err != nil {
			f = nil
		}
		return &setAccountAvatarArgs{
			AccountId:       id.Parse(r.FormValue("account")),
			AvatarImageData: f,
		}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setAccountAvatarArgs)
		if args.AvatarImageData != nil {
			defer args.AvatarImageData.Close()
		}

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if account.IsPersonal { // can't set avatar on someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(account.Region, account.Shard, args.AccountId, ctx.MyId())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		if args.AvatarImageData != nil {
			avatarImage, _, e := image.Decode(args.AvatarImageData)
			err.PanicIf(e)
			bounds := avatarImage.Bounds()
			if bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y { //if it  isn't square, then error
				panic(invalidAvatarShapeErr)
			}
			if uint(bounds.Max.X-bounds.Min.X) > ctx.AvatarClient().MaxAvatarDim() { // if it is larger than allowed then resize
				avatarImage = resize.Resize(ctx.AvatarClient().MaxAvatarDim(), ctx.AvatarClient().MaxAvatarDim(), avatarImage, resize.NearestNeighbor)
			}
			buff := &bytes.Buffer{}
			err.PanicIf(png.Encode(buff, avatarImage))
			data := buff.Bytes()
			reader := bytes.NewReader(data)
			ctx.AvatarClient().Save(ctx.MyId().String(), "image/png", reader)
			if !account.HasAvatar {
				//if account didn't previously have an avatar then lets update the store to reflect it's new state
				account.HasAvatar = true
				dbUpdateAccount(ctx, account)
			}
		} else {
			ctx.AvatarClient().Delete(ctx.MyId().String())
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
	AccountId id.Id  `json:"accountId"`
	NewRegion string `json:"newRegion"`
}

var migrateAccount = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/migrateAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &migrateAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, _ interface{}) interface{} {
		panic(err.NotImplemented)
		return nil
	},
}

type createAccountArgs struct {
	Name        string  `json:"name"`
	Region      string  `json:"region"`
	DisplayName *string `json:"displayName"`
}

var createAccount = &core.Endpoint{
	Method:                   cnst.POST,
	Path:                     "/api/v1/centralAccount/createAccount",
	RequiresSession:          true,
	ExampleResponseStructure: &account{},
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		args.Name = strings.Trim(args.Name, " ")
		validate.StringArg("name", args.Name, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())

		if !ctx.RegionalV1PrivateClient().IsValidRegion(args.Region) {
			panic(err.NoSuchRegion)
		}

		if exists := dbAccountWithCiNameExists(ctx, args.Name); exists {
			panic(nameAlreadyInUseErr)
		}

		account := &account{}
		account.Id = id.New()
		account.Name = args.Name
		account.DisplayName = args.DisplayName
		account.CreatedOn = t.Now()
		account.Region = args.Region
		account.Shard = -1
		account.IsPersonal = false
		dbCreateGroupAccountAndMembership(ctx, account, ctx.MyId())

		owner := dbGetPersonalAccountById(ctx, ctx.MyId())
		if owner == nil {
			panic(noSuchAccountErr)
		}

		defer func() {
			r := recover()
			if r != nil {
				dbDeleteAccountAndAllAssociatedMemberships(ctx, account.Id)
				panic(r)
			}
		}()
		shard, e := ctx.RegionalV1PrivateClient().CreateAccount(args.Region, account.Id, ctx.MyId(), owner.Name, owner.DisplayName)
		err.PanicIf(e)

		account.Shard = shard
		dbUpdateAccount(ctx, account)
		return account
	},
}

type getMyAccountsArgs struct {
	After *id.Id `json:"after"`
	Limit int    `json:"limit"`
}

type getMyAccountsResp struct {
	Accounts []*account `json:"accounts"`
	More     bool       `json:"more"`
}

var getMyAccounts = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getMyAccounts",
	ExampleResponseStructure: &getMyAccountsResp{},
	RequiresSession:          true,
	GetArgsStruct: func() interface{} {
		return &getMyAccountsArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*getMyAccountsArgs)
		res := &getMyAccountsResp{}
		res.Accounts, res.More = dbGetGroupAccounts(ctx, ctx.MyId(), args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
		return res
	},
}

type deleteAccountArgs struct {
	AccountId id.Id `json:"accountId"`
}

var deleteAccount = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/deleteAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		acc := dbGetAccount(ctx, args.AccountId)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.MyId().Equal(args.AccountId) {
			if acc.IsPersonal { // can't delete someone else's personal account
				panic(err.InsufficientPermission)
			}
			//otherwise attempting to delete a group account
			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.AccountId, ctx.MyId())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		ctx.RegionalV1PrivateClient().DeleteAccount(acc.Region, acc.Shard, args.AccountId, ctx.MyId())
		dbDeleteAccountAndAllAssociatedMemberships(ctx, args.AccountId)

		if ctx.MyId().Equal(args.AccountId) {
			var after *id.Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.MyId(), after, 100)
				for _, acc := range accs {
					isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, acc.Id, ctx.MyId())
					err.PanicIf(e)
					if isAccountOwner {
						panic(onlyOwnerMemberErr)
					}
				}
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().RemoveMembers(acc.Region, acc.Shard, acc.Id, ctx.MyId(), []id.Id{ctx.MyId()})
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
	AccountId  id.Id        `json:"accountId"`
	NewMembers []*AddMember `json:"newMembers"`
}

var addMembers = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/addMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		if args.AccountId.Equal(ctx.MyId()) {
			panic(err.InvalidOperation)
		}
		validate.EntityCount(len(args.NewMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			panic(noSuchAccountErr)
		}

		ids := make([]id.Id, 0, len(args.NewMembers))
		addMembersMap := map[string]*AddMember{}
		for _, member := range args.NewMembers {
			ids = append(ids, member.Id)
			addMembersMap[member.Id.String()] = member
		}

		accs := dbGetPersonalAccounts(ctx, ids)

		members := make([]*private.AddMember, 0, len(accs))
		for _, acc := range accs {
			role := addMembersMap[acc.Id.String()].Role
			role.Validate()
			ami := &private.AddMember{}
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
	AccountId       id.Id   `json:"accountId"`
	ExistingMembers []id.Id `json:"existingMembers"`
}

var removeMembers = &core.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/centralAccount/removeMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		if args.AccountId.Equal(ctx.MyId()) {
			panic(err.InvalidOperation)
		}
		validate.EntityCount(len(args.ExistingMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.AccountId)
		if account == nil {
			panic(noSuchAccountErr)
		}

		ctx.RegionalV1PrivateClient().RemoveMembers(account.Region, account.Shard, args.AccountId, ctx.MyId(), args.ExistingMembers)
		dbDeleteMemberships(ctx, args.AccountId, args.ExistingMembers)
		return nil
	},
}

var Endpoints = []*core.Endpoint{
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

// The main account client interface
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
	GetAccounts(ids []id.Id) ([]*account, error)
	SearchAccounts(nameOrDisplayNameStartsWith string) ([]*account, error)
	SearchPersonalAccounts(nameOrDisplayNameOrEmailStartsWith string) ([]*account, error)
	//requires active session to access
	GetMe(css *clientsession.Store) (*me, error)
	SetMyPwd(css *clientsession.Store, oldPwd, newPwd string) error
	SetMyEmail(css *clientsession.Store, newEmail string) error
	ResendMyNewEmailConfirmationEmail(css *clientsession.Store) error
	SetAccountName(css *clientsession.Store, accountId id.Id, newName string) error
	SetAccountDisplayName(css *clientsession.Store, accountId id.Id, newDisplayName *string) error
	SetAccountAvatar(css *clientsession.Store, accountId id.Id, avatarImage io.ReadCloser) error
	MigrateAccount(css *clientsession.Store, accountId id.Id, newRegion string) error
	CreateAccount(css *clientsession.Store, name, region string, displayName *string) (*account, error)
	GetMyAccounts(css *clientsession.Store, after *id.Id, limit int) (*getMyAccountsResp, error)
	DeleteAccount(css *clientsession.Store, accountId id.Id) error
	//member centric - must be an owner or admin
	AddMembers(css *clientsession.Store, accountId id.Id, newMembers []*AddMember) error
	RemoveMembers(css *clientsession.Store, accountId id.Id, existingMembers []id.Id) error
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

func (c *client) GetAccounts(ids []id.Id) ([]*account, error) {
	val, e := getAccounts.DoRequest(nil, c.host, &getAccountsArgs{
		Accounts: ids,
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

func (c *client) SearchPersonalAccounts(nameOrDisplayNameOrEmailStartsWith string) ([]*account, error) {
	val, e := searchPersonalAccounts.DoRequest(nil, c.host, &searchPersonalAccountsArgs{
		NameOrDisplayNameOrEmailStartsWith: nameOrDisplayNameOrEmailStartsWith,
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

func (c *client) SetAccountName(css *clientsession.Store, accountId id.Id, newName string) error {
	_, e := setAccountName.DoRequest(css, c.host, &setAccountNameArgs{
		AccountId: accountId,
		NewName:   newName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountDisplayName(css *clientsession.Store, accountId id.Id, newDisplayName *string) error {
	_, e := setAccountDisplayName.DoRequest(css, c.host, &setAccountDisplayNameArgs{
		AccountId:      accountId,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func (c *client) SetAccountAvatar(css *clientsession.Store, accountId id.Id, avatarImageData io.ReadCloser) error {
	defer avatarImageData.Close()
	_, e := setAccountAvatar.DoRequest(css, c.host, &setAccountAvatarArgs{
		AccountId:       accountId,
		AvatarImageData: avatarImageData,
	}, func() (io.ReadCloser, string) {
		body := bytes.NewBuffer([]byte{})
		writer := multipart.NewWriter(body)
		part, e := writer.CreateFormFile("avatar", "avatar")
		err.PanicIf(e)
		_, e = io.Copy(part, avatarImageData)
		err.PanicIf(e)
		err.PanicIf(writer.WriteField("account", accountId.String()))
		err.PanicIf(writer.Close())
		return ioutil.NopCloser(body), writer.FormDataContentType()
	}, nil)
	return e
}

func (c *client) MigrateAccount(css *clientsession.Store, accountId id.Id, newRegion string) error {
	_, e := migrateAccount.DoRequest(css, c.host, &migrateAccountArgs{
		AccountId: accountId,
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

func (c *client) DeleteAccount(css *clientsession.Store, accountId id.Id) error {
	_, e := deleteAccount.DoRequest(css, c.host, &deleteAccountArgs{
		AccountId: accountId,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, accountId id.Id, newMembers []*AddMember) error {
	_, e := addMembers.DoRequest(css, c.host, &addMembersArgs{
		AccountId:  accountId,
		NewMembers: newMembers,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, accountId id.Id, existingMembers []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, &removeMembersArgs{
		AccountId:       accountId,
		ExistingMembers: existingMembers,
	}, nil, nil)
	return e
}

//internal helpers

//db helpers
func dbAccountWithCiNameExists(ctx *core.Ctx, name string) bool {
	row := ctx.AccountQueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?`, name)
	count := 0
	err.PanicIf(row.Scan(&count))
	return count != 0
}

func dbGetAccountByCiName(ctx *core.Ctx, name string) *account {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?`, name)
	acc := account{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)) {
		return nil
	}
	return &acc
}

func dbCreatePersonalAccount(ctx *core.Ctx, account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	_, e := ctx.AccountExec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode)
	err.PanicIf(e)
	_, e = ctx.PwdExec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)`, account.Id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	err.PanicIf(e)
}

func dbGetPersonalAccountByEmail(ctx *core.Ctx, email string) *fullPersonalAccountInfo {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPersonalAccountById(ctx *core.Ctx, id id.Id) *fullPersonalAccountInfo {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, id)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPwdInfo(ctx *core.Ctx, id id.Id) *pwdInfo {
	row := ctx.PwdQueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?`, id)
	pwd := pwdInfo{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)) {
		return nil
	}
	return &pwd
}

func dbUpdatePersonalAccount(ctx *core.Ctx, personalAccountInfo *fullPersonalAccountInfo) {
	_, e := ctx.AccountExec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, personalAccountInfo.Id, personalAccountInfo.Name, personalAccountInfo.DisplayName, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode)
	err.PanicIf(e)
}

func dbUpdateAccount(ctx *core.Ctx, account *account) {
	_, e := ctx.AccountExec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal)
	err.PanicIf(e)
}

func dbUpdatePwdInfo(ctx *core.Ctx, id id.Id, pwdInfo *pwdInfo) {
	_, e := ctx.PwdExec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, id)
	err.PanicIf(e)
}

func dbDeleteAccountAndAllAssociatedMemberships(ctx *core.Ctx, id id.Id) {
	_, e := ctx.AccountExec(`CALL deleteAccountAndAllAssociatedMemberships(?)`, id)
	err.PanicIf(e)
	_, e = ctx.PwdExec(`DELETE FROM pwds WHERE id = ?`, id)
	err.PanicIf(e)
}

func dbGetAccount(ctx *core.Ctx, id id.Id) *account {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?`, id)
	a := account{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)) {
		return nil
	}
	return &a
}

func dbGetAccounts(ctx *core.Ctx, ids []id.Id) []*account {
	args := make([]interface{}, 0, len(ids))
	args = append(args, ids[0])
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (?`)
	for _, i := range ids[1:] {
		query.WriteString(`,?`)
		args = append(args, i)
	}
	query.WriteString(`)`)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		err.PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	return res
}

func dbSearchAccounts(ctx *core.Ctx, nameOrDisplayNameStartsWith string) []*account {
	searchTerm := nameOrDisplayNameStartsWith + "%"
	//rows, err := ctx.AccountQuery(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isPersonal FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	rows, e := ctx.AccountQuery(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? OR displayName LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal))
		res = append(res, &acc)
	}
	return res
}

func dbSearchPersonalAccounts(ctx *core.Ctx, nameOrDisplayNameOrEmailStartsWith string) []*account {
	searchTerm := nameOrDisplayNameOrEmailStartsWith + "%"
	//rows, err := ctx.AccountQuery(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	rows, e := ctx.AccountQuery(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? OR displayName LIKE ? OR email LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, searchTerm, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbGetPersonalAccounts(ctx *core.Ctx, ids []id.Id) []*account {
	args := make([]interface{}, 0, len(ids))
	args = append(args, ids[0])
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE activatedOn IS NOT NULL AND id IN (?`)
	for _, i := range ids[1:] {
		query.WriteString(`,?`)
		args = append(args, i)
	}
	query.WriteString(`)`)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbCreateGroupAccountAndMembership(ctx *core.Ctx, account *account, memberId id.Id) {
	_, e := ctx.AccountExec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, memberId)
	err.PanicIf(e)
}

func dbGetGroupAccounts(ctx *core.Ctx, memberId id.Id, after *id.Id, limit int) ([]*account, bool) {
	args := make([]interface{}, 0, 3)
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (SELECT account FROM memberships WHERE member = ?)`)
	args = append(args, memberId)
	if after != nil {
		query.WriteString(` AND name > (SELECT name FROM accounts WHERE id = ?)`)
		args = append(args, *after)
	}
	query.WriteString(` ORDER BY name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, limit+1)
	for rows.Next() {
		a := account{}
		err.PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func dbCreateMemberships(ctx *core.Ctx, accountId id.Id, members []id.Id) {
	args := make([]interface{}, 0, len(members)*2)
	args = append(args, accountId, members[0])
	query := bytes.NewBufferString(`INSERT INTO memberships (account, member) VALUES (?,?)`)
	for _, member := range members[1:] {
		query.WriteString(`,(?,?)`)
		args = append(args, accountId, member)
	}
	_, e := ctx.AccountExec(query.String(), args...)
	err.PanicIf(e)
}

func dbDeleteMemberships(ctx *core.Ctx, accountId id.Id, members []id.Id) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, accountId, members[0])
	query := bytes.NewBufferString(`DELETE FROM memberships WHERE account=? AND member IN (?`)
	for _, member := range members[1:] {
		query.WriteString(`,?`)
		args = append(args, member)
	}
	query.WriteString(`)`)
	_, e := ctx.AccountExec(query.String(), args...)
	err.PanicIf(e)
}

//email helpers

func emailSendMultipleAccountPolicyNotice(ctx *core.Ctx, address string) {
	ctx.MailClient().Send([]string{address}, "sendMultipleAccountPolicyNotice")
}

func emailSendActivationLink(ctx *core.Ctx, address, activationCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendActivationLink: activationCode: %s", activationCode))
}

func emailSendPwdResetLink(ctx *core.Ctx, address, resetCode string) {
	ctx.MailClient().Send([]string{address}, fmt.Sprintf("sendPwdResetLink: resetCode: %s", resetCode))
}

func emailSendNewEmailConfirmationLink(ctx *core.Ctx, currentAddress, newAddress, confirmationCode string) {
	ctx.MailClient().Send([]string{newAddress}, fmt.Sprintf("sendNewEmailConfirmationLink: currentAddress: %s newAddress: %s confirmationCode: %s", currentAddress, newAddress, confirmationCode))
}

//structs

type account struct {
	Id          id.Id     `json:"id"`
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
	Email    string     `json:"email"`
	Language string     `json:"language"`
	Theme    cnst.Theme `json:"theme"`
	NewEmail *string    `json:"newEmail,omitempty"`
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

type AddMember struct {
	Id   id.Id            `json:"id"`
	Role cnst.AccountRole `json:"role"`
}
