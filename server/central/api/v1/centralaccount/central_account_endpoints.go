package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/core"
	"bitbucket.org/0xor1/task/server/util/crypt"
	"bitbucket.org/0xor1/task/server/util/dlm"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
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
	noSuchAccountErr                      = &err.Err{Code: "c_v1_ca_nsa", Message: "no such account"}
	invalidActivationAttemptErr           = &err.Err{Code: "c_v1_ca_iaa", Message: "invalid activation attempt"}
	invalidResetPwdAttemptErr             = &err.Err{Code: "c_v1_ca_irpa", Message: "invalid reset password attempt"}
	invalidNewEmailConfirmationAttemptErr = &err.Err{Code: "c_v1_ca_ineca", Message: "invalid new email confirmation attempt"}
	invalidNameOrPwdErr                   = &err.Err{Code: "c_v1_ca_inop", Message: "invalid name or password"}
	incorrectPwdErr                       = &err.Err{Code: "c_v1_ca_ip", Message: "password incorrect"}
	accountNotActivatedErr                = &err.Err{Code: "c_v1_ca_ana", Message: "account not activated"}
	emailAlreadyInUseErr                  = &err.Err{Code: "c_v1_ca_eaiu", Message: "email already in use"}
	nameAlreadyInUseErr                   = &err.Err{Code: "c_v1_ca_naiu", Message: "name already in use"}
	emailConfirmationCodeErr              = &err.Err{Code: "c_v1_ca_ecc", Message: "email confirmation code is of zero length"}
	noNewEmailRegisteredErr               = &err.Err{Code: "c_v1_ca_nner", Message: "no new email registered"}
	onlyOwnerMemberErr                    = &err.Err{Code: "c_v1_ca_oom", Message: "can't delete member who is the only owner of an account"}
	invalidAvatarShapeErr                 = &err.Err{Code: "c_v1_ca_ias", Message: "avatar images must be square"}
)

//endpoints

var getRegions = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/centralAccount/getRegions",
	ExampleResponseStructure: []string{"use", "usw", "eu"},
	ValueDlmKeys: func(ctx *core.Ctx, _ interface{}) []string {
		return []string{dlm.ForSystem()}
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
	Method: cnst.POST,
	Path:   "/api/v1/centralAccount/authenticate",
	ExampleResponseStructure: id.New(),
	IsAuthentication:         true,
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
		acc := dbGetPersonalAccountById(ctx, ctx.Me())
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

		pwdInfo := dbGetPwdInfo(ctx, ctx.Me())
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
		dbUpdatePwdInfo(ctx, ctx.Me(), pwdInfo)
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

		acc := dbGetPersonalAccountById(ctx, ctx.Me())
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
		acc := dbGetPersonalAccountById(ctx, ctx.Me())
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
	Account id.Id  `json:"account"`
	NewName string `json:"newName"`
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

		acc := dbGetAccount(ctx, args.Account)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.Me().Equal(args.Account) {
			if acc.IsPersonal { // can't rename someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		acc.Name = args.NewName
		dbUpdateAccount(ctx, acc)

		if ctx.Me().Equal(args.Account) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.Me(), args.NewName) //first rename myself in my personal org
			var after *id.Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().SetMemberName(acc.Region, acc.Shard, acc.Id, ctx.Me(), args.NewName)
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
	Account        id.Id   `json:"account"`
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

		acc := dbGetAccount(ctx, args.Account)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.Me().Equal(args.Account) {
			if acc.IsPersonal { // can't rename someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
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

		if ctx.Me().Equal(args.Account) { // if i did set my personal account displayName, i need to update all the stored displayNames in all the accounts Im a member of
			ctx.RegionalV1PrivateClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.Me(), args.NewDisplayName) //first set my display name in my personal org
			var after *id.Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().SetMemberDisplayName(acc.Region, acc.Shard, acc.Id, ctx.Me(), args.NewDisplayName)
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
	Account id.Id         `json:"account"`
	Avatar  io.ReadCloser `json:"avatar"`
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
			Account: id.Parse(r.FormValue("account")),
			Avatar:  f,
		}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setAccountAvatarArgs)
		if args.Avatar != nil {
			defer args.Avatar.Close()
		}

		account := dbGetAccount(ctx, args.Account)
		if account == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.Me().Equal(args.Account) {
			if account.IsPersonal { // can't set avatar on someone else's personal account
				panic(err.InsufficientPermission)
			}

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(account.Region, account.Shard, args.Account, ctx.Me())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		if args.Avatar != nil {
			avatarImage, _, e := image.Decode(args.Avatar)
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
			ctx.AvatarClient().Save(ctx.Me().String(), "image/png", reader)
			if !account.HasAvatar {
				//if account didn't previously have an avatar then lets update the store to reflect it's new state
				account.HasAvatar = true
				dbUpdateAccount(ctx, account)
			}
		} else {
			ctx.AvatarClient().Delete(ctx.Me().String())
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
	Account   id.Id  `json:"account"`
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
		dbCreateGroupAccountAndMembership(ctx, account, ctx.Me())

		owner := dbGetPersonalAccountById(ctx, ctx.Me())
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
		shard, e := ctx.RegionalV1PrivateClient().CreateAccount(args.Region, account.Id, ctx.Me(), owner.Name, owner.DisplayName)
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
		res.Accounts, res.More = dbGetGroupAccounts(ctx, ctx.Me(), args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
		return res
	},
}

type deleteAccountArgs struct {
	Account id.Id `json:"account"`
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
		acc := dbGetAccount(ctx, args.Account)
		if acc == nil {
			panic(noSuchAccountErr)
		}

		if !ctx.Me().Equal(args.Account) {
			if acc.IsPersonal { // can't delete someone else's personal account
				panic(err.InsufficientPermission)
			}
			//otherwise attempting to delete a group account
			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
			err.PanicIf(e)
			if !isAccountOwner {
				panic(err.InsufficientPermission)
			}
		}

		ctx.RegionalV1PrivateClient().DeleteAccount(acc.Region, acc.Shard, args.Account, ctx.Me())
		dbDeleteAccountAndAllAssociatedMemberships(ctx, args.Account)

		if ctx.Me().Equal(args.Account) {
			var after *id.Id
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, acc.Id, ctx.Me())
					err.PanicIf(e)
					if isAccountOwner {
						panic(onlyOwnerMemberErr)
					}
				}
				for _, acc := range accs {
					ctx.RegionalV1PrivateClient().RemoveMembers(acc.Region, acc.Shard, acc.Id, ctx.Me(), []id.Id{ctx.Me()})
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
	Account    id.Id        `json:"account"`
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
		if args.Account.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		validate.EntityCount(len(args.NewMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.Account)
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

		ctx.RegionalV1PrivateClient().AddMembers(account.Region, account.Shard, args.Account, ctx.Me(), members)
		dbCreateMemberships(ctx, args.Account, ids)
		return nil
	},
}

type removeMembersArgs struct {
	Account         id.Id   `json:"account"`
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
		if args.Account.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		validate.EntityCount(len(args.ExistingMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.Account)
		if account == nil {
			panic(noSuchAccountErr)
		}

		ctx.RegionalV1PrivateClient().RemoveMembers(account.Region, account.Shard, args.Account, ctx.Me(), args.ExistingMembers)
		dbDeleteMemberships(ctx, args.Account, args.ExistingMembers)
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
