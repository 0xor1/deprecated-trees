package user

import (
	"bitbucket.org/robsix/core/helper"
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/uber-go/zap"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
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
)

func newApi(store store, linkMailer helper.LinkMailer, usernameRegexMatchers, pwdRegexMatchers []string, minSearchTermRuneCount, maxSearchLimitResults, usernameMinRuneCount, usernameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log zap.Logger) (Api, error) {
	if linkMailer == nil {
		return nil, NilLinkMailerErr
	}
	if log == nil {
		return nil, NilLogErr
	}
	return &api{
		store:                  store,
		linkMailer:             linkMailer,
		usernameRegexMatchers:  append(make([]string, 0, len(usernameRegexMatchers)), usernameRegexMatchers...),
		pwdRegexMatchers:       append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		minSearchTermRuneCount: minSearchTermRuneCount,
		maxSearchLimitResults:  maxSearchLimitResults,
		usernameMinRuneCount:   usernameMinRuneCount,
		usernameMaxRuneCount:   usernameMaxRuneCount,
		pwdMinRuneCount:        pwdMinRuneCount,
		pwdMaxRuneCount:        pwdMaxRuneCount,
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
	store                  store
	linkMailer             helper.LinkMailer
	usernameRegexMatchers  []string
	pwdRegexMatchers       []string
	minSearchTermRuneCount int
	maxSearchLimitResults  int
	usernameMinRuneCount   int
	usernameMaxRuneCount   int
	pwdMinRuneCount        int
	pwdMaxRuneCount        int
	cryptoCodeLen          int
	saltLen                int
	scryptN                int
	scryptR                int
	scryptP                int
	scryptKeyLen           int
	log                    zap.Logger
}

func (a *api) Register(username, email, pwd string) error {
	a.log.Debug(registerFnLogMsg, zap.String("username", username), zap.String("email", email), zap.String("pwd", "******"))

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

	if user, err := a.store.getByUsername(username); user != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, "store.getByUsername"), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(UsernameAlreadyInUseErr))
			return UsernameAlreadyInUseErr
		}
	}

	if user, err := a.store.getByEmail(email); user != nil || err != nil {
		if err != nil {
			a.log.Error(registerFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
			return err
		} else {
			a.log.Info(registerFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	scryptSalt, err := generateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return err
	}

	activationCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.Error(err))
		return err
	}

	id, err := helper.NewId()
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.Error(err))
		return err
	}

	err = a.store.create(&fullUserInfo{
		Me: Me{
			User: User{
				Entity: helper.Entity{
					Id: id,
				},
				Username: username,
			},
			Email: email,
		},
		RegistrationTime: time.Now().UTC(),
		ActivationCode:   activationCode,
		ScryptSalt:       scryptSalt,
		ScryptPwd:        scryptPwd,
		ScryptN:          a.scryptN,
		ScryptR:          a.scryptR,
		ScryptP:          a.scryptP,
		ScryptKeyLen:     a.scryptKeyLen,
	})
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, "store.create"), zap.Error(err))
		return err
	}

	err = a.linkMailer.SendActivationLink(email, activationCode)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, "linkMailer.SendActivationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendActivationEmail(email string) error {
	a.log.Debug(resendActivationEmailFnLogMsg, zap.String("email", email))

	email = strings.Trim(email, " ")
	user, err := a.store.getByEmail(email)
	if err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
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

	if err = a.linkMailer.SendActivationLink(email, user.ActivationCode); err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, "linkMailer.SendActivationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) Activate(activationCode string) (string, error) {
	a.log.Debug(activateFnLogMsg, zap.String("activationCode", "******"))

	activationCode = strings.Trim(activationCode, " ")
	user, err := a.store.getByActivationCode(activationCode)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, "store.getByActivationCode"), zap.Error(err))
		return "", err
	}
	if user == nil {
		a.log.Info(activateFnLogMsg, zap.Error(NoSuchUserErr))
		return "", NoSuchUserErr
	}

	user.ActivationCode = ""
	err = a.store.update(user)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return "", err
	}

	return user.Id, nil
}

func (a *api) Authenticate(username, pwdTry string) (string, error) {
	a.log.Debug(authenticateFnLogMsg, zap.String("username", username), zap.String("pwdTry", "******"))

	username = strings.Trim(username, " ")
	user, err := a.store.getByUsername(username)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, "store.getByUsername"), zap.Error(err))
		return "", err
	}
	if user == nil {
		a.log.Info(authenticateFnLogMsg, zap.Error(NoSuchUserErr))
		return "", NoSuchUserErr
	}

	scryptPwdTry, err := scrypt.Key([]byte(pwdTry), user.ScryptSalt, user.ScryptN, user.ScryptR, user.ScryptP, user.ScryptKeyLen)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return "", err
	}

	if !pwdsMatch(user.ScryptPwd, scryptPwdTry) {
		a.log.Info(authenticateFnLogMsg, zap.Error(IncorrectPwdErr))
		return "", IncorrectPwdErr
	}

	if !user.isActivated() {
		a.log.Info(authenticateFnLogMsg, zap.Error(UserNotActivated))
		return "", UserNotActivated
	}

	userUpdated := false
	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if len(user.ResetPwdCode) > 0 {
		user.ResetPwdCode = ""
		userUpdated = true
	}
	// check that the password is encrypted with the latest scrypt settings, if not, encrypt again using the latest settings
	if user.ScryptN != a.scryptN || user.ScryptR != a.scryptR || user.ScryptP != a.scryptP || user.ScryptKeyLen != a.scryptKeyLen || len(user.ScryptSalt) < a.saltLen {
		user.ScryptSalt, err = generateCryptoBytes(a.saltLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.Error(err))
			return "", err
		}
		user.ScryptN = a.scryptN
		user.ScryptR = a.scryptR
		user.ScryptP = a.scryptP
		user.ScryptKeyLen = a.scryptKeyLen
		user.ScryptPwd, err = scrypt.Key([]byte(pwdTry), user.ScryptSalt, user.ScryptN, user.ScryptR, user.ScryptP, user.ScryptKeyLen)
		if err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
			return "", err
		}
		userUpdated = true
	}
	if userUpdated {
		if err = a.store.update(user); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
			return "", err
		}
	}

	return user.Id, nil
}

func (a *api) ChangeUsername(id, newUsername string) error {
	a.log.Debug(changeUsernameFnLogMsg, zap.String("id", id), zap.String("newUsername", newUsername))

	newUsername = strings.Trim(newUsername, " ")
	if err := validateStringParam("username", newUsername, a.usernameMinRuneCount, a.usernameMaxRuneCount, a.usernameRegexMatchers); err != nil {
		a.log.Info(changeUsernameFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.store.getByUsername(newUsername); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, "store.getByUsername"), zap.Error(err))
			return err
		} else {
			a.log.Info(changeUsernameFnLogMsg, zap.Error(UsernameAlreadyInUseErr))
			return UsernameAlreadyInUseErr
		}
	}

	user, err := a.store.getById(id)
	if err != nil {
		a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeUsernameFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	user.Username = newUsername
	if err = a.store.update(user); err != nil {
		a.log.Error(changeUsernameFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ChangeEmail(id, newEmail string) error {
	a.log.Debug(changeEmailFnLogMsg, zap.String("id", id), zap.String("newEmail", newEmail))

	newEmail = strings.Trim(newEmail, " ")
	if err := validateEmail(newEmail); err != nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(err))
		return err
	}

	if user, err := a.store.getByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
			return err
		} else {
			a.log.Info(changeEmailFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	user, err := a.store.getById(id)
	if err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	confirmationCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.Error(err))
		return err
	}

	user.NewEmail = newEmail
	user.NewEmailConfirmationCode = confirmationCode
	if err = a.store.update(user); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return err
	}

	if err = a.linkMailer.SendNewEmailConfirmationLink(newEmail, confirmationCode); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "linkMailer.SendNewEmailConfirmationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResendNewEmailConfirmationEmail(id string) error {
	a.log.Debug(resendNewEmailConfirmationEmailFnLogMsg, zap.String("id", id))

	user, err := a.store.getById(id)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	// check the user has actually registered a new email
	if len(user.NewEmail) == 0 {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(NewEmailErr))
		return NewEmailErr
	}
	// just in case something has gone crazy wrong
	if len(user.NewEmailConfirmationCode) == 0 {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(EmailConfirmationCodeErr))
		return EmailConfirmationCodeErr
	}

	err = a.linkMailer.SendNewEmailConfirmationLink(user.NewEmail, user.NewEmailConfirmationCode)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, "linkMailer.SendNewEmailConfirmationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ConfirmNewEmail(newEmail string, confirmationCode string) error {
	a.log.Debug(ConfirmNewEmailFnLogMsg, zap.String("newConfirmationCode", confirmationCode))

	user, err := a.store.getByNewEmailConfirmationCode(confirmationCode)
	if err != nil {
		a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, "store.getByNewEmailConfirmationCode"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	if user.NewEmail != newEmail || user.NewEmailConfirmationCode != confirmationCode {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(NewEmailConfirmationErr))
		return NewEmailConfirmationErr
	}

	if user, err := a.store.getByEmail(newEmail); user != nil || err != nil {
		if err != nil {
			a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
			return err
		} else {
			a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(EmailAlreadyInUseErr))
			return EmailAlreadyInUseErr
		}
	}

	user.Email = newEmail
	user.NewEmail = ""
	user.NewEmailConfirmationCode = ""
	if err = a.store.update(user); err != nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(err))
		return err
	}

	return nil
}

func (a *api) ResetPwd(email string) error {
	a.log.Debug(resetPwdFnLogMsg, zap.String("email", email))

	user, err := a.store.getByEmail(email)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	resetPwdCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	user.ResetPwdCode = resetPwdCode
	if err = a.store.update(user); err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	err = a.linkMailer.SendPwdResetLink(email, resetPwdCode)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, "linkMailer.SendPwdResetLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (string, error) {
	a.log.Debug(setNewPwdFromPwdResetFnLogMsg, zap.String("newPwd", "******"), zap.String("resetPwdCode", "******"))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return "", err
	}

	user, err := a.store.getByResetPwdCode(resetPwdCode)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, "store.getByResetPwdCode"), zap.Error(err))
		return "", err
	}
	if user == nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(NoSuchUserErr))
		return "", NoSuchUserErr
	}

	scryptSalt, err := generateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return "", err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return "", err
	}

	user.ScryptPwd = scryptPwd
	user.ScryptSalt = scryptSalt
	user.ScryptN = a.scryptN
	user.ScryptR = a.scryptR
	user.ScryptP = a.scryptP
	user.ScryptKeyLen = a.scryptKeyLen
	user.ActivationCode = ""
	user.ResetPwdCode = ""
	if err = a.store.update(user); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return "", err
	}

	return user.Id, nil
}

func (a *api) ChangePwd(id, oldPwd, newPwd string) error {
	a.log.Debug(changePwdFnLogMsg, zap.String("id", id), zap.String("oldPwd", "******"), zap.String("newPwd", "******"))

	if err := validateStringParam("password", newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(changePwdFnLogMsg, zap.Error(err))
		return err
	}

	user, err := a.store.getById(id)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
	}
	if user == nil {
		a.log.Info(changePwdFnLogMsg, zap.Error(NoSuchUserErr))
		return NoSuchUserErr
	}

	scryptPwdTry, err := scrypt.Key([]byte(oldPwd), user.ScryptSalt, user.ScryptN, user.ScryptR, user.ScryptP, user.ScryptKeyLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return err
	}

	if !pwdsMatch(user.ScryptPwd, scryptPwdTry) {
		a.log.Info(changePwdFnLogMsg, zap.Error(IncorrectPwdErr))
		return IncorrectPwdErr
	}

	scryptSalt, err := generateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(newPwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return err
	}

	user.ScryptPwd = scryptPwd
	user.ScryptSalt = scryptSalt
	user.ScryptN = a.scryptN
	user.ScryptR = a.scryptR
	user.ScryptP = a.scryptP
	user.ScryptKeyLen = a.scryptKeyLen
	if err = a.store.update(user); err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return err
	}

	return nil
}

func (a *api) GetMe(id string) (*Me, error) {
	a.log.Debug(getMeFnLogMsg, zap.String("id", id))

	user, err := a.store.getById(id)
	if err != nil {
		a.log.Error(getMeFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return nil, err
	}
	if user == nil {
		a.log.Info(getMeFnLogMsg, zap.Error(NoSuchUserErr))
		return nil, NoSuchUserErr
	}

	return &user.Me, nil
}

func (a *api) Delete(id string) error {
	a.log.Debug(deleteFnLogMsg, zap.String("id", id))

	err := a.store.delete(id)
	if err != nil {
		a.log.Error(deleteFnLogMsg, zap.String(subcall, "store.delete"), zap.Error(err))
	}

	return err
}

func (a *api) Get(ids []string) ([]*User, error) {
	a.log.Debug(getFnLogMsg, zap.String("ids", fmt.Sprintf("%v", ids)))

	users, err := a.store.getByIds(ids)
	if err != nil {
		a.log.Error(getFnLogMsg, zap.String(subcall, "store.getByIds"), zap.Error(err))
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

	users, err := a.store.search(search, limit)
	if err != nil {
		a.log.Error(searchFnLogMsg, zap.String(subcall, "store.search"), zap.Error(err))
	}

	return users, err
}

type fullUserInfo struct {
	Me
	RegistrationTime         time.Time  `json:"registrationTime"`
	ActivationCode           string     `json:"activationCode,omitempty"`
	ActivationTime           *time.Time `json:"activationTime,omitempty"`
	NewEmailConfirmationCode string     `json:"newEmailConfirmationCode,omitempty"`
	ResetPwdCode             string     `json:"resetPwdCode,omitempty"`
	ScryptSalt               []byte     `json:"scryptSalt"`
	ScryptPwd                []byte     `json:"scryptPwd"`
	ScryptN                  int        `json:"scryptN"`
	ScryptR                  int        `json:"scryptR"`
	ScryptP                  int        `json:"scryptP"`
	ScryptKeyLen             int        `json:"scryptKeyLen"`
	PersonalStoreId          int        `json:"personalStoreId"`
}

func (u *fullUserInfo) isActivated() bool {
	return len(u.ActivationCode) == 0
}

//helpers

type store interface {
	getByUsername(username string) (*fullUserInfo, error)
	getByEmail(email string) (*fullUserInfo, error)
	getById(id string) (*fullUserInfo, error)
	getByActivationCode(activationCode string) (*fullUserInfo, error)
	getByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error)
	getByResetPwdCode(resetPwdCode string) (*fullUserInfo, error)
	getByIds(ids []string) ([]*User, error)
	search(search string, limit int) ([]*User, error)
	create(user *fullUserInfo) error
	update(user *fullUserInfo) error
	delete(id string) error
}

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

func generateCryptoBytes(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateCryptoString(length int) (string, error) {
	buf := make([]rune, length)
	urlSafeRunesLength := len(urlSafeRunes)
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, big.NewInt(int64(urlSafeRunesLength)))
		if err != nil {
			return "", err
		}
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf), nil
}

func pwdsMatch(a, b []byte) bool {
	return bytes.Compare(a, b) == 0
}
