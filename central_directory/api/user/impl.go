package user

import (
	"bitbucket.org/robsix/task_center/misc"
	"bytes"
	"fmt"
	. "github.com/pborman/uuid"
	"github.com/uber-go/zap"
	"golang.org/x/crypto/scrypt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	redactedInfo                            = "******"
	registerFnLogMsg                        = "user.api.Register"
	resendActivationEmailFnLogMsg           = "user.api.ResendActivationEmail"
	activateFnLogMsg                        = "user.api.Activate"
	authenticateFnLogMsg                    = "user.api.Authenticate"
	changeUsernameFnLogMsg                  = "user.api.ChangeUsername"
	changeEmailFnLogMsg                     = "user.api.ChangeEmail"
	resendNewEmailConfirmationEmailFnLogMsg = "user.api.ResendNewEmailConfirmationEmail"
	ConfirmNewEmailFnLogMsg                 = "user.api.ConfirmNewEmail"
	resetPwdFnLogMsg                        = "user.api.ResetPwd"
	setNewPwdFromPwdResetFnLogMsg           = "user.api.SetNewPwdFromPwdReset"
	changePwdFnLogMsg                       = "user.api.ChangePwd"
	getMeFnLogMsg                           = "user.api.GetMe"
	deleteFnLogMsg                          = "user.api.Delete"
	getFnLogMsg                             = "user.api.Get"
	searchFnLogMsg                          = "user.api.Search"
	subcall                                 = "subcall"
	userStoreCreate                         = "userStore.Create"
	userStoreGetByUsername                  = "userStore.GetByUsername"
	userStoreGetByEmail                     = "userStore.GetByEmail"
	userStoreGetById                        = "userStore.GetById"
	userStoreGetByActivationCode            = "userStore.GetByActivationCode"
	userStoreGetByNewEmailConfirmationCode  = "userStore.GetByNewEmailConfirmationCode"
	userStoreGetByResetPwdCode              = "userStore.GetByResetPwdCode"
	userStoreGetByIds                       = "userStore.GetByIds"
	userStoreSearch                         = "userStore.Search"
	userStoreUpdate                         = "userStore.Update"
	userStoreDelete                         = "userStore.Delete"
	pwdStoreCreate                          = "pwdStore.Create"
	pwdStoreGet                             = "pwdStore.Get"
	pwdStoreUpdate                          = "pwdStore.Update"
	pwdStoreDelete                          = "pwdStore.Delete"
	linkMailerSendActivationLink            = "linkMailer.SendActivationLink"
	linkMailerSendPwdResetLink              = "linkMailer.SendPwdResetLink"
	linkMailerSendNewEmailConfirmationLink  = "linkMailer.SendNewEmailConfirmationLink"
	miscGenerateCryptoBytes                 = "misc.GenerateCryptoBytes"
	miscGenerateCryptoString                = "misc.GenerateCryptoString"
	scryptKey                               = "scryptKey"
)

func NewApi(userStore UserStore, pwdStore PwdStore, linkMailer LinkMailer, usernameRegexMatchers, pwdRegexMatchers []string, usernameMinRuneCount, usernameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, minSearchTermRuneCount, maxSearchLimitResults, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log zap.Logger) (Api, error) {
	if userStore == nil {
		return nil, NilUserStoreErr
	}
	if pwdStore == nil {
		return nil, NilPwdStoreErr
	}
	if linkMailer == nil {
		return nil, NilLinkMailerErr
	}
	if log == nil {
		return nil, NilLogErr
	}
	return &api{
		userStore:              userStore,
		pwdStore:               pwdStore,
		linkMailer:             linkMailer,
		usernameRegexMatchers:  append(make([]string, 0, len(usernameRegexMatchers)), usernameRegexMatchers...),
		pwdRegexMatchers:       append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		usernameMinRuneCount:   usernameMinRuneCount,
		usernameMaxRuneCount:   usernameMaxRuneCount,
		pwdMinRuneCount:        pwdMinRuneCount,
		pwdMaxRuneCount:        pwdMaxRuneCount,
		minSearchTermRuneCount: minSearchTermRuneCount,
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
	userStore              UserStore
	pwdStore               PwdStore
	linkMailer             LinkMailer
	usernameRegexMatchers  []string
	pwdRegexMatchers       []string
	usernameMinRuneCount   int
	usernameMaxRuneCount   int
	pwdMinRuneCount        int
	pwdMaxRuneCount        int
	minSearchTermRuneCount int
	maxSearchLimitResults  int
	cryptoCodeLen          int
	saltLen                int
	scryptN                int
	scryptR                int
	scryptP                int
	scryptKeyLen           int
	log                    zap.Logger
}

func (a *api) Register(username, region, email, pwd string) error {
	a.log.Debug(registerFnLogMsg, zap.String("username", username), zap.String("region", region), zap.String("email", redactedInfo), zap.String("pwd", redactedInfo))

	username = strings.Trim(username, " ")
	if err := validateStringParam("username", username, a.usernameMinRuneCount, a.usernameMaxRuneCount, a.usernameRegexMatchers); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	email = strings.Trim(email, " ")
	if err := validateEmail(email); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	if err := validateStringParam("password", pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.userStore.GetByUsername(username); user != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, userStoreGetByUsername), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(UsernameAlreadyInUseErr))
			return UsernameAlreadyInUseErr
		}
	}

	if user, err := a.userStore.GetByEmail(email); user != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, userStoreGetByEmail), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, miscGenerateCryptoBytes), zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	activationCode, err := misc.GenerateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, miscGenerateCryptoString), zap.Error(err))
		return err
	}

	userId, err := misc.NewId()
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.Error(err))
		return err
	}

	err = a.userStore.Create(
		&FullUserInfo{
			Me: Me{
				User: User{
					CentralEntity: misc.CentralEntity{
						Entity: misc.Entity{
							Id: userId,
						},
						Region: region,
						Shard:  -1,
					},
					Username: username,
				},
				Email: email,
			},
			RegistrationTime: time.Now().UTC(),
			ActivationCode:   &activationCode,
		},
	)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, userStoreCreate), zap.Error(err))
		return err
	}

	err = a.pwdStore.Create(
		userId,
		&PwdInfo{
			ScryptSalt:   scryptSalt,
			ScryptPwd:    scryptPwd,
			ScryptN:      a.scryptN,
			ScryptR:      a.scryptR,
			ScryptP:      a.scryptP,
			ScryptKeyLen: a.scryptKeyLen,
		},
	)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, pwdStoreCreate), zap.Error(err))
		return err
	}

	err = a.linkMailer.SendActivationLink(email, activationCode)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, linkMailerSendActivationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendActivationEmail(email string) error {
	a.log.Debug(resendActivationEmailFnLogMsg, zap.String("email", redactedInfo))

	email = strings.Trim(email, " ")
	user, err := a.userStore.GetByEmail(email)
	if err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, userStoreGetByEmail), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resendActivationEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	if user.isActivated() {
		a.log.Info(resendActivationEmailFnLogMsg, zap.Error(UserAlreadyActivatedErr))
		return UserAlreadyActivatedErr
	}

	if err = a.linkMailer.SendActivationLink(email, *user.ActivationCode); err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, linkMailerSendActivationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) Activate(activationCode string) (UUID, error) {
	a.log.Debug(activateFnLogMsg, zap.String("activationCode", redactedInfo))

	activationCode = strings.Trim(activationCode, " ")
	user, err := a.userStore.GetByActivationCode(activationCode)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, userStoreGetByActivationCode), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(activateFnLogMsg, zap.Error(NoSuchUserErr))
		return nil, NoSuchUserErr
	}

	user.ActivationCode = nil
	activationTime := time.Now().UTC()
	user.ActivationTime = &activationTime
	err = a.userStore.Update(user)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
		return nil, err
	}

	return user.Id, nil
}

func (a *api) Authenticate(username, pwdTry string) (UUID, error) {
	a.log.Debug(authenticateFnLogMsg, zap.String("username", username), zap.String("pwdTry", redactedInfo))

	username = strings.Trim(username, " ")
	user, err := a.userStore.GetByUsername(username)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, userStoreGetByUsername), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(authenticateFnLogMsg, zap.Error(NoSuchUserErr))
		return nil, NoSuchUserErr
	}
	if !user.isActivated() {
		a.log.Info(authenticateFnLogMsg, zap.Error(UserNotActivated))
		return nil, UserNotActivated
	}

	pwdInfo, err := a.pwdStore.Get(user.Id)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, pwdStoreGet), zap.Error(err))
		return nil, err
	}

	scryptPwdTry, err := scrypt.Key([]byte(pwdTry), pwdInfo.ScryptSalt, pwdInfo.ScryptN, pwdInfo.ScryptR, pwdInfo.ScryptP, pwdInfo.ScryptKeyLen)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return nil, err
	}

	if !pwdsMatch(pwdInfo.ScryptPwd, scryptPwdTry) {
		a.log.Info(authenticateFnLogMsg, zap.Error(IncorrectPwdErr))
		return nil, IncorrectPwdErr
	}

	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(user.ResetPwdCode) > 0 {
		user.ResetPwdCode = nil
		if err = a.userStore.Update(user); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
			return nil, err
		}
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if pwdInfo.ScryptN != a.scryptN || pwdInfo.ScryptR != a.scryptR || pwdInfo.ScryptP != a.scryptP || pwdInfo.ScryptKeyLen != a.scryptKeyLen || len(pwdInfo.ScryptSalt) < a.saltLen {
		pwdInfo.ScryptSalt, err = misc.GenerateCryptoBytes(a.saltLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, miscGenerateCryptoBytes), zap.Error(err))
			return nil, err
		}
		pwdInfo.ScryptN = a.scryptN
		pwdInfo.ScryptR = a.scryptR
		pwdInfo.ScryptP = a.scryptP
		pwdInfo.ScryptKeyLen = a.scryptKeyLen
		pwdInfo.ScryptPwd, err = scrypt.Key([]byte(pwdTry), pwdInfo.ScryptSalt, pwdInfo.ScryptN, pwdInfo.ScryptR, pwdInfo.ScryptP, pwdInfo.ScryptKeyLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
			return nil, err
		}
		if err = a.pwdStore.Update(user.Id, pwdInfo); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, pwdStoreUpdate), zap.Error(err))
			return nil, err
		}

	}

	return user.Id, nil
}

func (a *api) ChangeUsername(id UUID, newUsername string) error {
	a.log.Debug(changeUsernameFnLogMsg, zap.Base64("id", id), zap.String("newUsername", newUsername))

	newUsername = strings.Trim(newUsername, " ")
	if err := validateStringParam("username", newUsername, a.usernameMinRuneCount, a.usernameMaxRuneCount, a.usernameRegexMatchers); err != nil {
		a.log.Info(changeUsernameFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.userStore.GetByUsername(newUsername); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, userStoreGetByUsername), zap.Error(err))
			return err
		} else {
			a.log.Info(changeUsernameFnLogMsg, zap.Error(UsernameAlreadyInUseErr))
			return UsernameAlreadyInUseErr
		}
	}

	user, err := a.userStore.GetById(id)
	if err != nil {
		a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, userStoreGetById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeUsernameFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	user.Username = newUsername
	if err = a.userStore.Update(user); err != nil {
		a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ChangeEmail(id UUID, newEmail string) error {
	a.log.Debug(changeEmailFnLogMsg, zap.Base64("id", id), zap.String("newEmail", redactedInfo))

	newEmail = strings.Trim(newEmail, " ")
	if err := validateEmail(newEmail); err != nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.userStore.GetByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeEmailFnLogMsg, zap.String(subcall, userStoreGetByEmail), zap.Error(err))
			return err
		} else {
			a.log.Info(changeEmailFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	user, err := a.userStore.GetById(id)
	if err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, userStoreGetById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	confirmationCode, err := misc.GenerateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, miscGenerateCryptoString), zap.Error(err))
		return err
	}

	user.NewEmail = &newEmail
	user.NewEmailConfirmationCode = &confirmationCode
	if err = a.userStore.Update(user); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
		return err
	}

	if err = a.linkMailer.SendNewEmailConfirmationLink(newEmail, confirmationCode); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, linkMailerSendNewEmailConfirmationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendNewEmailConfirmationEmail(id UUID) error {
	a.log.Debug(resendNewEmailConfirmationEmailFnLogMsg, zap.Base64("id", id))

	user, err := a.userStore.GetById(id)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, userStoreGetById), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	// check the user has actually registered a new email
	if len(*user.NewEmail) == 0 {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(NewEmailErr))
		return NewEmailErr
	}
	// just in case something has gone crazy wrong
	if len(*user.NewEmailConfirmationCode) == 0 {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(EmailConfirmationCodeErr))
		return EmailConfirmationCodeErr
	}

	err = a.linkMailer.SendNewEmailConfirmationLink(user.NewEmail, user.NewEmailConfirmationCode)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, linkMailerSendNewEmailConfirmationLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ConfirmNewEmail(newEmail string, confirmationCode string) (UUID, error) {
	a.log.Debug(ConfirmNewEmailFnLogMsg, zap.String("newEmail", redactedInfo), zap.String("newConfirmationCode", redactedInfo))

	user, err := a.userStore.GetByNewEmailConfirmationCode(confirmationCode)
	if err != nil {
		a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, userStoreGetByNewEmailConfirmationCode), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	if *user.NewEmail != newEmail || *user.NewEmailConfirmationCode != confirmationCode {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(NewEmailConfirmationErr))
		return NewEmailConfirmationErr
	}

	if user, err := a.userStore.GetByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, userStoreGetByEmail), zap.Error(err))
			return err
		} else {
			a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	user.Email = newEmail
	user.NewEmail = nil
	user.NewEmailConfirmationCode = nil
	if err = a.userStore.Update(user); err != nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
		return err
	}

	return user.Id, nil
}

func (a *api) ResetPwd(email string) error {
	a.log.Debug(resetPwdFnLogMsg, zap.String("email", redactedInfo))

	user, err := a.userStore.GetByEmail(email)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, userStoreGetByEmail), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	resetPwdCode, err := misc.GenerateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.String(subcall, miscGenerateCryptoString), zap.Error(err))
		return err
	}

	user.ResetPwdCode = resetPwdCode
	if err = a.userStore.Update(user); err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	err = a.linkMailer.SendPwdResetLink(email, resetPwdCode)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, linkMailerSendPwdResetLink), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (UUID, error) {
	a.log.Debug(setNewPwdFromPwdResetFnLogMsg, zap.String("newPwd", redactedInfo), zap.String("resetPwdCode", redactedInfo))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return nil, err
	}

	user, err := a.userStore.GetByResetPwdCode(resetPwdCode)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, userStoreGetByResetPwdCode), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(NoSuchUserErr))
		return nil, NoSuchUserErr
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return nil, err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return nil, err
	}

	user.ActivationCode = nil
	user.ResetPwdCode = nil
	if err = a.userStore.Update(user); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, userStoreUpdate), zap.Error(err))
		return nil, err
	}

	if err = a.pwdStore.Update(
		user.Id,
		&PwdInfo{
			ScryptPwd:    scryptPwd,
			ScryptSalt:   scryptSalt,
			ScryptN:      a.scryptN,
			ScryptR:      a.scryptR,
			ScryptP:      a.scryptP,
			ScryptKeyLen: a.scryptKeyLen,
		},
	); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, pwdStoreUpdate), zap.Error(err))
		return nil, err
	}

	return user.Id, nil
}

func (a *api) ChangePwd(id UUID, oldPwd, newPwd string) error {
	a.log.Debug(changePwdFnLogMsg, zap.Base64("id", id), zap.String("oldPwd", redactedInfo), zap.String("newPwd", redactedInfo))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(changePwdFnLogMsg, zap.Error(err))
		return err
	}

	pwdInfo, err := a.pwdStore.Get(id)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, pwdStoreGet), zap.Error(err))
		return err
	}
	if pwdInfo == nil {
		a.log.Info(changePwdFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	scryptPwdTry, err := scrypt.Key([]byte(oldPwd), pwdInfo.ScryptSalt, pwdInfo.ScryptN, pwdInfo.ScryptR, pwdInfo.ScryptP, pwdInfo.ScryptKeyLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	if !pwdsMatch(pwdInfo.ScryptPwd, scryptPwdTry) {
		a.log.Info(changePwdFnLogMsg, zap.Error(IncorrectPwdErr))
		return IncorrectPwdErr
	}

	scryptSalt, err := misc.GenerateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, scryptKey), zap.Error(err))
		return err
	}

	pwdInfo.ScryptPwd = scryptPwd
	pwdInfo.ScryptSalt = scryptSalt
	pwdInfo.ScryptN = a.scryptN
	pwdInfo.ScryptR = a.scryptR
	pwdInfo.ScryptP = a.scryptP
	pwdInfo.ScryptKeyLen = a.scryptKeyLen
	if err = a.pwdStore.Update(id, pwdInfo); err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, pwdStoreUpdate), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) GetMe(id UUID) (*Me, error) {
	a.log.Debug(getMeFnLogMsg, zap.Base64("id", id))

	user, err := a.userStore.GetById(id)
	if err != nil {
		a.log.Error(getMeFnLogMsg, zap.String(subcall, userStoreGetById), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(getMeFnLogMsg, zap.Error(NoSuchUserErr))
		return nil, NoSuchUserErr
	}

	return &user.Me, nil
}

func (a *api) Delete(id UUID) error {
	a.log.Debug(deleteFnLogMsg, zap.Base64("id", id))

	if err := a.userStore.Delete(id); err != nil {
		a.log.Error(deleteFnLogMsg, zap.String(subcall, userStoreDelete), zap.Error(err))
		return err
	}

	if err := a.pwdStore.Delete(id); err != nil {
		a.log.Error(deleteFnLogMsg, zap.String(subcall, pwdStoreDelete), zap.Error(err))
		return nil
	}

	return nil
}

func (a *api) Get(ids []UUID) ([]*User, error) {
	a.log.Debug(getFnLogMsg, zap.String("ids", fmt.Sprintf("%v", ids)))

	users, err := a.userStore.GetByIds(ids)
	if err != nil {
		a.log.Error(getFnLogMsg, zap.String(subcall, userStoreGetByIds), zap.Error(err))
	}

	return users, err
}

func (a *api) Search(search string, limit int) ([]*User, error) {
	a.log.Debug(searchFnLogMsg, zap.String("search", search), zap.Int("limit", limit))

	if limit < 1 || limit > a.maxSearchLimitResults {
		limit = a.maxSearchLimitResults
	}

	search = strings.Trim(search, " ")
	if utf8.RuneCountInString(search) < a.minSearchTermRuneCount {
		return nil, &SearchTermTooShortErr{
			MinRuneCount: a.minSearchTermRuneCount,
		}
	}

	users, err := a.userStore.Search(search, limit)
	if err != nil {
		a.log.Error(searchFnLogMsg, zap.String(subcall, userStoreSearch), zap.Error(err))
	}

	return users, err
}

//helpers

func newInvalidStringParamErr(paramPurpose string, minRuneCount, maxRuneCount int, regexMatchers []string) *InvalidStringParamErr {
	return &InvalidStringParamErr{
		ParamPurpose:  paramPurpose,
		MinRuneCount:  minRuneCount,
		MaxRuneCount:  maxRuneCount,
		RegexMatchers: append(make([]string, 0, len(regexMatchers)), regexMatchers...),
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
	matches, err := regexp.MatchString(`.+@.+\..+`, email)
	if err == nil && !matches {
		err = InvalidEmailErr
	}

	return err
}

func pwdsMatch(a, b []byte) bool {
	return bytes.Compare(a, b) == 0
}
