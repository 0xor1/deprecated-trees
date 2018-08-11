package centralaccount

import (
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/crypt"
	"github.com/0xor1/trees/server/util/ctx"
	"github.com/0xor1/trees/server/util/endpoint"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/private"
	t "github.com/0xor1/trees/server/util/time"
	"github.com/0xor1/trees/server/util/validate"
	"bytes"
	"github.com/0xor1/panic"
	"github.com/nfnt/resize"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var spacesRegex = regexp.MustCompile(`\s+`)

//endpoints

type registerArgs struct {
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Pwd         string      `json:"pwd"`
	Region      cnst.Region `json:"region"`
	Language    string      `json:"language"`
	DisplayName *string     `json:"displayName"`
	Theme       cnst.Theme  `json:"theme"`
}

var register = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/register",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &registerArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*registerArgs)
		args.Name = strings.Trim(args.Name, " ")
		validate.StringArg("name", args.Name, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())
		args.Email = strings.Trim(args.Email, " ")
		validate.Email(args.Email)
		validate.StringArg("pwd", args.Pwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())
		args.Language = strings.Trim(args.Language, " ") // may need more validation than this at some point to check it is a language we support and not a junk value, but it isnt critical right now
		args.Theme.Validate()
		if args.DisplayName != nil {
			*args.DisplayName = spacesRegex.ReplaceAllString(*args.DisplayName, " ") //replace any multiple spaces with a single space
			*args.DisplayName = strings.Trim(*args.DisplayName, " ")
			if *args.DisplayName == "" {
				args.DisplayName = nil
			}
			if args.DisplayName != nil {
				validate.StringArg("displayName", *args.DisplayName, ctx.DisplayNameMinRuneCount(), ctx.DisplayNameMaxRuneCount(), ctx.DisplayNameRegexMatchers())
			}
		}

		args.Region.ValidateForDataRegions()
		ctx.ReturnNowIf(dbAccountWithCiNameExists(ctx, args.Name), http.StatusBadRequest, "name already in use")

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
				if e, ok := r.(error); ok {
					panic.IfNotNil(e)
				}
				panic.If(true, "%v", r)
			}
		}()
		var e error
		acc.Shard, e = ctx.RegionalV1PrivateClient().CreateAccount(acc.Region, acc.Id, acc.Id, acc.Name, acc.DisplayName, acc.HasAvatar)
		panic.IfNotNil(e)
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

var resendActivationEmail = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/resendActivationEmail",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &resendActivationEmailArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
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

var activate = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/activate",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &activateArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*activateArgs)
		args.ActivationCode = strings.Trim(args.ActivationCode, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		ctx.ReturnBadRequestNowIf(acc == nil || acc.activationCode == nil || args.ActivationCode != *acc.activationCode, "invalid activation attempt")
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

type AuthenticateResult struct {
	Me         *Me                  `json:"me"`
	MyAccounts *GetMyAccountsResult `json:"myAccounts"`
}

func (ar *AuthenticateResult) Id() id.Id {
	return ar.Me.Id
}

var authenticate = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/authenticate",
	RequiresSession:          false,
	ExampleResponseStructure: &Me{},
	IsAuthentication:         true,
	GetArgsStruct: func() interface{} {
		return &authenticateArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*authenticateArgs)
		args.Email = strings.Trim(args.Email, " ")
		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		ctx.ReturnBadRequestNowIf(acc == nil, "invalid name or password")

		pwdInfo := dbGetPwdInfo(ctx, acc.Id)
		scryptPwdTry := crypt.ScryptKey([]byte(args.PwdTry), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
		ctx.ReturnNowIf(!pwdsMatch(pwdInfo.pwd, scryptPwdTry), http.StatusBadRequest, "invalid name or password")

		//must do this after checking the acc has the correct pwd otherwise it allows anyone to fish for valid emails on the system
		ctx.ReturnNowIf(!acc.isActivated(), http.StatusBadRequest, "account is not activated, confirm email address")

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

		myAccounts, more := dbGetGroupAccounts(ctx, acc.Id, nil, 100)
		return &AuthenticateResult{
			Me: &acc.Me,
			MyAccounts: &GetMyAccountsResult{
				Accounts: myAccounts,
				More:     more,
			},
		}
	},
}

type confirmNewEmailArgs struct {
	CurrentEmail     string `json:"currentEmail"`
	NewEmail         string `json:"newEmail"`
	ConfirmationCode string `json:"confirmationCode"`
}

var confirmNewEmail = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/confirmNewEmail",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &confirmNewEmailArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*confirmNewEmailArgs)
		acc := dbGetPersonalAccountByEmail(ctx, args.CurrentEmail)
		ctx.ReturnBadRequestNowIf(acc == nil || acc.NewEmail == nil || args.NewEmail != *acc.NewEmail || acc.newEmailConfirmationCode == nil || args.ConfirmationCode != *acc.newEmailConfirmationCode, "invalid email confirmation attempt")

		newAcc := dbGetPersonalAccountByEmail(ctx, args.NewEmail)
		ctx.ReturnBadRequestNowIf(newAcc != nil, "email already in use")

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

var resetPwd = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/resetPwd",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &resetPwdArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
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

var setNewPwdFromPwdReset = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/setNewPwdFromPwdReset",
	RequiresSession: false,
	GetArgsStruct: func() interface{} {
		return &setNewPwdFromPwdResetArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setNewPwdFromPwdResetArgs)
		validate.StringArg("pwd", args.NewPwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())

		acc := dbGetPersonalAccountByEmail(ctx, args.Email)
		ctx.ReturnBadRequestNowIf(acc == nil || acc.resetPwdCode == nil || args.ResetPwdCode != *acc.resetPwdCode, "invalid reset password attempt")

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

var getAccount = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/getAccount",
	RequiresSession:          false,
	ExampleResponseStructure: &Account{},
	GetArgsStruct: func() interface{} {
		return &getAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getAccountArgs)
		return dbGetAccountByCiName(ctx, strings.Trim(args.Name, " "))
	},
}

type getAccountsArgs struct {
	Accounts []id.Id `json:"accounts"`
}

var getAccounts = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/getAccounts",
	RequiresSession:          false,
	ExampleResponseStructure: []*Account{{}},
	GetArgsStruct: func() interface{} {
		return &getAccountsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getAccountsArgs)
		validate.EntityCount(len(args.Accounts), ctx.MaxProcessEntityCount())

		return dbGetAccounts(ctx, args.Accounts)
	},
}

type searchAccountsArgs struct {
	NameOrDisplayNamePrefix string `json:"nameOrDisplayNamePrefix"`
}

var searchAccounts = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/searchAccounts",
	RequiresSession:          false,
	ExampleResponseStructure: []*Account{{}},
	GetArgsStruct: func() interface{} {
		return &searchAccountsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*searchAccountsArgs)
		return dbSearchAccounts(ctx, args.NameOrDisplayNamePrefix)
	},
}

type searchPersonalAccountsArgs struct {
	NameOrDisplayNamePrefix string `json:"nameOrDisplayNamePrefix"`
}

var searchPersonalAccounts = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/searchPersonalAccounts",
	RequiresSession:          false,
	ExampleResponseStructure: []*Account{{}},
	GetArgsStruct: func() interface{} {
		return &searchPersonalAccountsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*searchPersonalAccountsArgs)
		return dbSearchPersonalAccounts(ctx, args.NameOrDisplayNamePrefix)
	},
}

var getMe = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/getMe",
	RequiresSession:          true,
	ExampleResponseStructure: &Me{},
	CtxHandler: func(ctx ctx.Ctx, _ interface{}) interface{} {
		acc := dbGetPersonalAccountById(ctx, ctx.Me())
		ctx.ReturnNowIf(acc == nil, http.StatusNotFound, "no such account")
		return &acc.Me
	},
}

type setMyPwdArgs struct {
	NewPwd string `json:"newPwd"`
	OldPwd string `json:"oldPwd"`
}

var setMyPwd = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/setMyPwd",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyPwdArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMyPwdArgs)
		validate.StringArg("pwd", args.NewPwd, ctx.PwdMinRuneCount(), ctx.PwdMaxRuneCount(), ctx.PwdRegexMatchers())

		pwdInfo := dbGetPwdInfo(ctx, ctx.Me())
		panic.If(pwdInfo == nil, "no such account")

		scryptPwdTry := crypt.ScryptKey([]byte(args.OldPwd), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)

		ctx.ReturnNowIf(!pwdsMatch(pwdInfo.pwd, scryptPwdTry), http.StatusBadRequest, "password mismatch")

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

var setMyEmail = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/setMyEmail",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMyEmailArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMyEmailArgs)
		args.NewEmail = strings.Trim(args.NewEmail, " ")
		validate.Email(args.NewEmail)

		if acc := dbGetPersonalAccountByEmail(ctx, args.NewEmail); acc != nil {
			emailSendMultipleAccountPolicyNotice(ctx, acc.Email)
		}

		acc := dbGetPersonalAccountById(ctx, ctx.Me())
		panic.If(acc == nil, "no such account")

		confirmationCode := crypt.UrlSafeString(ctx.CryptCodeLen())

		acc.NewEmail = &args.NewEmail
		acc.newEmailConfirmationCode = &confirmationCode
		dbUpdatePersonalAccount(ctx, acc)
		emailSendNewEmailConfirmationLink(ctx, acc.Email, args.NewEmail, confirmationCode)
		return nil
	},
}

var resendMyNewEmailConfirmationEmail = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/resendMyNewEmailConfirmationEmail",
	RequiresSession: true,
	CtxHandler: func(ctx ctx.Ctx, _ interface{}) interface{} {
		acc := dbGetPersonalAccountById(ctx, ctx.Me())
		panic.If(acc == nil, "no such account")

		// check the acc has actually registered a new email
		ctx.ReturnBadRequestNowIf(acc.NewEmail == nil, "no new email registered")

		// just in case something has gone crazy wrong
		panic.If(acc.newEmailConfirmationCode == nil, "new email confirmation code is nil when new email was not nil")

		emailSendNewEmailConfirmationLink(ctx, acc.Email, *acc.NewEmail, *acc.newEmailConfirmationCode)
		return nil
	},
}

type setAccountNameArgs struct {
	Account id.Id  `json:"account"`
	NewName string `json:"newName"`
}

var setAccountName = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/setAccountName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountNameArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setAccountNameArgs)
		args.NewName = strings.Trim(args.NewName, " ")
		validate.StringArg("name", args.NewName, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())

		ctx.ReturnNowIf(dbAccountWithCiNameExists(ctx, args.NewName), http.StatusBadRequest, "name already in use")

		acc := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(acc == nil, "no such account")

		if !ctx.Me().Equal(args.Account) {
			ctx.ReturnUnauthorizedNowIf(acc.IsPersonal) // can't rename someone else's personal account

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
			panic.IfNotNil(e)
			ctx.ReturnUnauthorizedNowIf(!isAccountOwner)
		}

		acc.Name = args.NewName
		dbUpdateAccount(ctx, acc)

		if ctx.Me().Equal(args.Account) { // if i did rename my personal account, i need to update all the stored names in all the accounts Im a member of
			var after *id.Id
			privateClientCallBatch := make([]func(), 0, 10)
			privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
				return func() {
					ctx.RegionalV1PrivateClient().SetMemberName(a.Region, a.Shard, a.Id, ctx.Me(), args.NewName)
				}
			}(acc))
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
						return func() {
							ctx.RegionalV1PrivateClient().SetMemberName(a.Region, a.Shard, a.Id, ctx.Me(), args.NewName)
						}
					}(acc))
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
			panic.IfNotNil(panic.SafeGoGroup(privateClientCallBatch...))
		}
		return nil
	},
}

type setAccountDisplayNameArgs struct {
	Account        id.Id   `json:"account"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setAccountDisplayName = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/setAccountDisplayName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setAccountDisplayNameArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setAccountDisplayNameArgs)
		if args.NewDisplayName != nil {
			*args.NewDisplayName = spacesRegex.ReplaceAllString(*args.NewDisplayName, " ") //replace any multiple spaces with a single space
			*args.NewDisplayName = strings.Trim(*args.NewDisplayName, " ")
			if *args.NewDisplayName == "" {
				args.NewDisplayName = nil
			}
			if args.NewDisplayName != nil {
				validate.StringArg("displayName", *args.NewDisplayName, ctx.DisplayNameMinRuneCount(), ctx.DisplayNameMaxRuneCount(), ctx.DisplayNameRegexMatchers())
			}
		}

		acc := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(acc == nil, "no such account")

		if !ctx.Me().Equal(args.Account) {
			ctx.ReturnUnauthorizedNowIf(acc.IsPersonal) // can't rename someone else's personal account

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
			panic.IfNotNil(e)
			ctx.ReturnUnauthorizedNowIf(!isAccountOwner)
		}

		if (acc.DisplayName == nil && args.NewDisplayName == nil) || (acc.DisplayName != nil && args.NewDisplayName != nil && *acc.DisplayName == *args.NewDisplayName) {
			return nil //if there is no change, dont do any redundant work
		}

		acc.DisplayName = args.NewDisplayName
		dbUpdateAccount(ctx, acc)

		if ctx.Me().Equal(args.Account) { // if i did set my personal account displayName, i need to update all the stored displayNames in all the accounts Im a member of
			var after *id.Id
			privateClientCallBatch := make([]func(), 0, 10)
			privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
				return func() {
					ctx.RegionalV1PrivateClient().SetMemberDisplayName(a.Region, a.Shard, a.Id, ctx.Me(), args.NewDisplayName)
				}
			}(acc))
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
						return func() {
							ctx.RegionalV1PrivateClient().SetMemberDisplayName(a.Region, a.Shard, a.Id, ctx.Me(), args.NewDisplayName)
						}
					}(acc))
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
			panic.IfNotNil(panic.SafeGoGroup(privateClientCallBatch...))
		}
		return nil
	},
}

type setAccountAvatarArgs struct {
	Account id.Id         `json:"account"`
	Avatar  io.ReadCloser `json:"avatar"`
}

var setAccountAvatar = &endpoint.Endpoint{
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
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setAccountAvatarArgs)
		if args.Avatar != nil {
			defer args.Avatar.Close()
		}

		account := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(account == nil, "no such account")

		if !ctx.Me().Equal(args.Account) {
			ctx.ReturnUnauthorizedNowIf(account.IsPersonal) // can't set avatar on someone else's personal account

			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(account.Region, account.Shard, args.Account, ctx.Me())
			panic.IfNotNil(e)
			ctx.ReturnUnauthorizedNowIf(!isAccountOwner)
		}

		hasAvatarStatusChanged := false
		if args.Avatar != nil {
			avatarImage, _, e := image.Decode(args.Avatar)
			panic.IfNotNil(e)
			bounds := avatarImage.Bounds()
			ctx.ReturnBadRequestNowIf(bounds.Max.X-bounds.Min.X != bounds.Max.Y-bounds.Min.Y, "invalid avatar shape, must be square") //if it  isn't square, then error
			if uint(bounds.Max.X-bounds.Min.X) > ctx.AvatarClient().MaxAvatarDim() {                                                  // if it is larger than allowed then resize
				avatarImage = resize.Resize(ctx.AvatarClient().MaxAvatarDim(), ctx.AvatarClient().MaxAvatarDim(), avatarImage, resize.NearestNeighbor)
			}
			buff := &bytes.Buffer{}
			panic.IfNotNil(png.Encode(buff, avatarImage))
			data := buff.Bytes()
			reader := bytes.NewReader(data)
			ctx.AvatarClient().Save(ctx.Me().String(), "image/png", reader)
			hasAvatarStatusChanged = !account.HasAvatar
		} else {
			ctx.AvatarClient().Delete(ctx.Me().String())
			hasAvatarStatusChanged = account.HasAvatar
		}
		if hasAvatarStatusChanged {
			account.HasAvatar = !account.HasAvatar
			dbUpdateAccount(ctx, account)
			if account.Id.Equal(ctx.Me()) {
				var after *id.Id
				privateClientCallBatch := make([]func(), 0, 10)
				privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
					return func() {
						ctx.RegionalV1PrivateClient().SetMemberHasAvatar(a.Region, a.Shard, a.Id, ctx.Me(), account.HasAvatar)
					}
				}(account))
				for {
					accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
					for _, acc := range accs {
						privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
							return func() {
								ctx.RegionalV1PrivateClient().SetMemberHasAvatar(a.Region, a.Shard, a.Id, ctx.Me(), account.HasAvatar)
							}
						}(acc))
					}
					if more {
						after = &accs[len(accs)-1].Id
					} else {
						break
					}
				}
				panic.IfNotNil(panic.SafeGoGroup(privateClientCallBatch...))
			}
		}
		return nil
	},
}

type migrateAccountArgs struct {
	Account   id.Id       `json:"account"`
	NewRegion cnst.Region `json:"newRegion"`
}

var migrateAccount = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/migrateAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &migrateAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, _ interface{}) interface{} {
		ctx.ReturnBadRequestNowIf(true, "not implemented")
		return nil
	},
}

type createAccountArgs struct {
	Name        string      `json:"name"`
	Region      cnst.Region `json:"region"`
	DisplayName *string     `json:"displayName"`
}

var createAccount = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/createAccount",
	RequiresSession:          true,
	ExampleResponseStructure: &Account{},
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		args.Name = strings.Trim(args.Name, " ")
		validate.StringArg("name", args.Name, ctx.NameMinRuneCount(), ctx.NameMaxRuneCount(), ctx.NameRegexMatchers())

		args.Region.ValidateForDataRegions()
		ctx.ReturnNowIf(dbAccountWithCiNameExists(ctx, args.Name), http.StatusBadRequest, "name already in use")

		account := &Account{}
		account.Id = id.New()
		account.Name = args.Name
		account.DisplayName = args.DisplayName
		account.CreatedOn = t.Now()
		account.Region = args.Region
		account.Shard = -1
		account.IsPersonal = false
		dbCreateGroupAccountAndMembership(ctx, account, ctx.Me())

		owner := dbGetPersonalAccountById(ctx, ctx.Me())
		ctx.ReturnBadRequestNowIf(owner == nil, "no such account")

		defer func() {
			r := recover()
			if r != nil {
				dbDeleteAccountAndAllAssociatedMemberships(ctx, account.Id)
				if e, ok := r.(error); ok {
					panic.IfNotNil(e)
				}
				panic.If(true, "%v", r)
			}
		}()
		shard, e := ctx.RegionalV1PrivateClient().CreateAccount(args.Region, account.Id, ctx.Me(), owner.Name, owner.DisplayName, owner.HasAvatar)
		panic.IfNotNil(e)

		account.Shard = shard
		dbUpdateAccount(ctx, account)
		return account
	},
}

type getMyAccountsArgs struct {
	After *id.Id `json:"after"`
	Limit int    `json:"limit"`
}

type GetMyAccountsResult struct {
	Accounts []*Account `json:"accounts"`
	More     bool       `json:"more"`
}

var getMyAccounts = &endpoint.Endpoint{
	Path:                     "/api/v1/centralAccount/getMyAccounts",
	RequiresSession:          true,
	ExampleResponseStructure: &GetMyAccountsResult{Accounts: []*Account{{}}},
	GetArgsStruct: func() interface{} {
		return &getMyAccountsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMyAccountsArgs)
		res := &GetMyAccountsResult{}
		res.Accounts, res.More = dbGetGroupAccounts(ctx, ctx.Me(), args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
		return res
	},
}

type deleteAccountArgs struct {
	Account id.Id `json:"account"`
}

var deleteAccount = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/deleteAccount",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		acc := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(acc == nil, "no such account")

		if !ctx.Me().Equal(args.Account) {
			ctx.ReturnUnauthorizedNowIf(acc.IsPersonal) // can't delete someone else's personal account

			//otherwise attempting to delete a group account
			isAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsAccountOwner(acc.Region, acc.Shard, args.Account, ctx.Me())
			panic.IfNotNil(e)
			ctx.ReturnUnauthorizedNowIf(!isAccountOwner)

			ctx.RegionalV1PrivateClient().DeleteAccount(acc.Region, acc.Shard, args.Account, ctx.Me())
		} else {
			var after *id.Id
			privateClientCallBatch := make([]func(), 0, 10)
			privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
				return func() {
					panic.IfNotNil(ctx.RegionalV1PrivateClient().DeleteAccount(a.Region, a.Shard, a.Id, ctx.Me()))
				}
			}(acc))
			for {
				accs, more := dbGetGroupAccounts(ctx, ctx.Me(), after, 100)
				for _, acc := range accs {
					privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
						return func() {
							isOnlyAccountOwner, e := ctx.RegionalV1PrivateClient().MemberIsOnlyAccountOwner(a.Region, a.Shard, a.Id, ctx.Me())
							panic.IfNotNil(e)
							panic.If(isOnlyAccountOwner, "only account owner on account %s, please delete the group account or add another account owner before deleting your personal account", acc.Id)
						}
					}(acc))
				}
				errors := panic.SafeGoGroup(privateClientCallBatch...)
				if errors != nil {

				}
				privateClientCallBatch = privateClientCallBatch[:0]
				for _, acc := range accs {
					privateClientCallBatch = append(privateClientCallBatch, func(a *Account) func() {
						return func() {
							panic.IfNotNil(ctx.RegionalV1PrivateClient().DeleteAccount(a.Region, a.Shard, a.Id, ctx.Me()))
						}
					}(acc))
				}
				if more {
					after = &accs[len(accs)-1].Id
				} else {
					break
				}
			}
			panic.IfNotNil(panic.SafeGoGroup(privateClientCallBatch...))
		}
		dbDeleteAccountAndAllAssociatedMemberships(ctx, args.Account)
		return nil
	},
}

type addMembersArgs struct {
	Account    id.Id        `json:"account"`
	NewMembers []*AddMember `json:"newMembers"`
}

var addMembers = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/addMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		ctx.ReturnBadRequestNowIf(args.Account.Equal(ctx.Me()), "can't add/remove members to/from a personal account")
		validate.EntityCount(len(args.NewMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(account == nil, "no such account")

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
			ami.HasAvatar = acc.HasAvatar
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

var removeMembers = &endpoint.Endpoint{
	Path:            "/api/v1/centralAccount/removeMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		ctx.ReturnBadRequestNowIf(args.Account.Equal(ctx.Me()), "can't add/remove members to/from a personal account")
		validate.EntityCount(len(args.ExistingMembers), ctx.MaxProcessEntityCount())

		account := dbGetAccount(ctx, args.Account)
		ctx.ReturnBadRequestNowIf(account == nil, "no such account")

		ctx.RegionalV1PrivateClient().RemoveMembers(account.Region, account.Shard, args.Account, ctx.Me(), args.ExistingMembers)
		dbDeleteMemberships(ctx, args.Account, args.ExistingMembers)
		return nil
	},
}

var Endpoints = []*endpoint.Endpoint{
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
type Account struct {
	Id          id.Id       `json:"id"`
	Name        string      `json:"name"`
	DisplayName *string     `json:"displayName"`
	CreatedOn   time.Time   `json:"createdOn"`
	Region      cnst.Region `json:"region"`
	NewRegion   *string     `json:"newRegion,omitempty"`
	Shard       int         `json:"shard"`
	HasAvatar   bool        `json:"hasAvatar"`
	IsPersonal  bool        `json:"isPersonal"`
}

func (a *Account) isMigrating() bool {
	return a.NewRegion != nil
}

type Me struct {
	Account
	Email    string     `json:"email"`
	Language string     `json:"language"`
	Theme    cnst.Theme `json:"theme"`
	NewEmail *string    `json:"newEmail,omitempty"`
}

type fullPersonalAccountInfo struct {
	Me
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
