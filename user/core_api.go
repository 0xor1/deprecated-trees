package user

import (
	"bitbucket.org/robsix/core"
	"bytes"
	"crypto/rand"
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
	changeEmailFnLogMsg                     = "user.api.ChangeEmail"
	resendNewEmailConfirmationEmailFnLogMsg = "user.api.ResendNewEmailConfirmationEmail"
	ConfirmNewEmailFnLogMsg                 = "user.api.ConfirmNewEmail"
	resetPwdFnLogMsg                        = "user.api.ResetPwd"
	setNewPwdFromPwdResetFnLogMsg           = "user.api.SetNewPwdFromPwdReset"
	changePwdFnLogMsg                       = "user.api.ChangePwd"
	getMeFnLogMsg                           = "user.api.GetMe"
	subcall                                 = "subcall"
)

func NewCoreApi(store Store, linkMailer LinkMailer, pwdRegexMatchers []string, cryptoCodeLen, pwdMinRuneCount, pwdMaxRuneCount, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log zap.Logger) CoreApi {
	return &coreApi{
		store:            store,
		linkMailer:       linkMailer,
		cryptoCodeLen:    cryptoCodeLen,
		pwdMinRuneCount:  pwdMinRuneCount,
		pwdMaxRuneCount:  pwdMaxRuneCount,
		pwdRegexMatchers: append(make([]string, 0, len(pwdRegexMatchers)), pwdRegexMatchers...),
		saltLen:          saltLen,
		scryptN:          scryptN,
		scryptR:          scryptR,
		scryptP:          scryptP,
		scryptKeyLen:     scryptKeyLen,
		log:              log,
	}
}

type coreApi struct {
	store            Store
	linkMailer       LinkMailer
	cryptoCodeLen    int
	pwdMinRuneCount  int
	pwdMaxRuneCount  int
	pwdRegexMatchers []string
	saltLen          int
	scryptN          int
	scryptR          int
	scryptP          int
	scryptKeyLen     int
	log              zap.Logger
}

func (a *coreApi) Register(email, firstName, lastName, pwd string) error {
	a.log.Debug(registerFnLogMsg, zap.String("email", email), zap.String("firstName", firstName), zap.String("lastName", lastName), zap.String("pwd", "******"))

	if err := validateEmail(email); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	if err := validatePwd(pwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	firstName = strings.Trim(firstName, " ")
	if utf8.RuneCountInString(firstName) == 0 {
		a.log.Info(registerFnLogMsg, zap.Error(FirstNameErr))
		return FirstNameErr
	}

	lastName = strings.Trim(lastName, " ")
	if utf8.RuneCountInString(lastName) == 0 {
		a.log.Info(registerFnLogMsg, zap.Error(LastNameErr))
		return LastNameErr
	}

	if user, err := a.store.GetByEmail(email); user != nil || err != nil {
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
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	scryptPwd, err := scrypt.Key([]byte(pwd), scryptSalt, a.scryptN, a.scryptR, a.scryptP, a.scryptKeyLen)
	if err != nil {
		a.log.Error(registerFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return err
	}

	activationCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	id, err := core.NewId()
	if err != nil {
		a.log.Info(registerFnLogMsg, zap.Error(err))
		return err
	}

	err = a.store.Create(&FullUserInfo{
		Me: Me{
			User: User{
				Entity: core.Entity{
					Id: id,
				},
				FirstName: firstName,
				LastName:  lastName,
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

func (a *coreApi) ResendActivationEmail(email string) error {
	a.log.Debug(resendActivationEmailFnLogMsg, zap.String("email", email))

	user, err := a.store.GetByEmail(email)
	if err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
		return err
	}

	if user.IsActivated() {
		a.log.Info(resendActivationEmailFnLogMsg, zap.Error(UserAlreadyActivatedErr))
		return UserAlreadyActivatedErr
	}

	if err = a.linkMailer.SendActivationLink(email, user.ActivationCode); err != nil {
		a.log.Error(resendActivationEmailFnLogMsg, zap.String(subcall, "linkMailer.SendActivationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) Activate(activationCode string) (string, error) {
	a.log.Debug(activateFnLogMsg, zap.String("activationCode", "******"))

	user, err := a.store.GetByActivationCode(activationCode)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, "store.getByActivationCode"), zap.Error(err))
		return "", err
	}

	if user.IsActivated() {
		a.log.Error(activateFnLogMsg, zap.Error(UserAlreadyActivatedErr))
		return "", UserAlreadyActivatedErr
	}

	user.ActivationCode = ""
	err = a.store.Update(user)
	if err != nil {
		a.log.Error(activateFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return "", err
	}

	return user.Id, nil
}

func (a *coreApi) Authenticate(email, pwdTry string) (string, error) {
	a.log.Debug(authenticateFnLogMsg, zap.String("email", email), zap.String("pwdTry", "******"))

	user, err := a.store.GetByEmail(email)
	if err != nil {
		a.log.Error(authenticateFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
		return "", err
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

	if !user.IsActivated() {
		a.log.Info(authenticateFnLogMsg, zap.Error(UserNotActivated))
		return "", UserNotActivated
	}

	userUpdated := false
	//if there was an outstanding password reset on this user, remove it, they have since remembered their password
	if utf8.RuneCountInString(user.ResetPwdCode) > 0 {
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
		if err = a.store.Update(user); err != nil {
			a.log.Error(authenticateFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
			return "", err
		}
	}

	return user.Id, nil
}

func (a *coreApi) ChangeEmail(id, newEmail string) error {
	a.log.Debug(changeEmailFnLogMsg, zap.String("id", id), zap.String("newEmail", newEmail))

	if err := validateEmail(newEmail); err != nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(err))
		return err
	}

	user, err := a.store.GetByEmail(newEmail)
	if err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
		return err
	}

	confirmationCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(changeEmailFnLogMsg, zap.Error(err))
		return err
	}

	user.NewEmail = newEmail
	user.NewEmailConfirmationCode = confirmationCode
	if err = a.store.Update(user); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return err
	}

	if err = a.linkMailer.SendNewEmailConfirmationLink(newEmail, confirmationCode); err != nil {
		a.log.Error(changeEmailFnLogMsg, zap.String(subcall, "linkMailer.SendNewEmailConfirmationLink"), zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) ResendNewEmailConfirmationEmail(id string) error {
	a.log.Debug(resendNewEmailConfirmationEmailFnLogMsg, zap.String("id", id))

	user, err := a.store.GetById(id)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
	}

	// just in case something has gone crazy wrong
	if utf8.RuneCountInString(user.NewEmailConfirmationCode) == 0 {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.Error(EmailConfirmationCodeErr))
		return EmailConfirmationCodeErr
	}

	err = a.linkMailer.SendNewEmailConfirmationLink(id, user.NewEmailConfirmationCode)
	if err != nil {
		a.log.Error(resendNewEmailConfirmationEmailFnLogMsg, zap.String(subcall, "linkMailer.SendEmailConfirmationCode"), zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) ConfirmNewEmail(newEmail string, confirmationCode string) error {
	a.log.Debug(ConfirmNewEmailFnLogMsg, zap.String("newConfirmationCode", confirmationCode))

	user, err := a.store.GetByNewEmailConfirmationCode(confirmationCode)
	if err != nil {
		a.log.Error(ConfirmNewEmailFnLogMsg, zap.String(subcall, "store.getByNewEmailConfirmationCode"), zap.Error(err))
		return err
	}

	if user.NewEmail != newEmail || user.NewEmailConfirmationCode != confirmationCode {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(NewEmailConfirmationErr))
		return NewEmailConfirmationErr
	}

	user.Email = newEmail
	user.NewEmail = ""
	user.NewEmailConfirmationCode = ""
	if err = a.store.Update(user); err != nil {
		a.log.Info(ConfirmNewEmailFnLogMsg, zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) ResetPwd(email string) error {
	a.log.Debug(resetPwdFnLogMsg, zap.String("email", email))

	user, err := a.store.GetByEmail(email)
	if err != nil {
		a.log.Error(resetPwdFnLogMsg, zap.String(subcall, "store.getByEmail"), zap.Error(err))
		return err
	}

	resetPwdCode, err := generateCryptoString(a.cryptoCodeLen)
	if err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	user.ResetPwdCode = resetPwdCode
	if err = a.store.Update(user); err != nil {
		a.log.Info(resetPwdFnLogMsg, zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) SetNewPwdFromPwdReset(newPwd, resetPwdCode string) (string, error) {
	a.log.Debug(setNewPwdFromPwdResetFnLogMsg, zap.String("newPwd", "******"), zap.String("resetPwdCode", "******"))

	if err := validatePwd(newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(setNewPwdFromPwdResetFnLogMsg, zap.Error(err))
		return "", err
	}

	user, err := a.store.GetByResetPwdCode(resetPwdCode)
	if err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, "store.getByResetPwdCode"), zap.Error(err))
		return "", err
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
	if err = a.store.Update(user); err != nil {
		a.log.Error(setNewPwdFromPwdResetFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return "", err
	}

	return user.Id, nil
}

func (a *coreApi) ChangePwd(id, oldPwd, newPwd string) error {
	a.log.Debug(changePwdFnLogMsg, zap.String("id", id), zap.String("oldPwd", "******"), zap.String("newPwd", "******"))

	if err := validatePwd(newPwd, a.pwdMinRuneCount, a.pwdMaxRuneCount, a.pwdRegexMatchers); err != nil {
		a.log.Info(changePwdFnLogMsg, zap.Error(err))
		return err
	}

	user, err := a.store.GetById(id)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return err
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

	user.ScryptSalt, err = generateCryptoBytes(a.saltLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.Error(err))
		return err
	}
	user.ScryptN = a.scryptN
	user.ScryptR = a.scryptR
	user.ScryptP = a.scryptP
	user.ScryptKeyLen = a.scryptKeyLen
	user.ScryptPwd, err = scrypt.Key([]byte(newPwd), user.ScryptSalt, user.ScryptN, user.ScryptR, user.ScryptP, user.ScryptKeyLen)
	if err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "scrypt.Key"), zap.Error(err))
		return err
	}
	if err = a.store.Update(user); err != nil {
		a.log.Error(changePwdFnLogMsg, zap.String(subcall, "store.update"), zap.Error(err))
		return err
	}

	return nil
}

func (a *coreApi) GetMe(id string) (*Me, error) {
	a.log.Debug(getMeFnLogMsg)

	user, err := a.store.GetById(id)
	if err != nil {
		a.log.Error(getMeFnLogMsg, zap.String(subcall, "store.getById"), zap.Error(err))
		return nil, err
	}

	return user.ToMe(), nil
}

//helpers
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

func validateEmail(email string) error {
	matches, err := regexp.MatchString(`.+@.+\..+`, email)
	if err == nil && !matches {
		err = InvalidEmailErr
	}

	return err
}

func validatePwd(pwd string, minRuneCount, maxRuneCount int, regexMatchers []string) error {
	pwdRuneCount := utf8.RuneCountInString(pwd)
	if pwdRuneCount < minRuneCount || pwdRuneCount > maxRuneCount {
		return newInvalidPwdErr(minRuneCount, maxRuneCount, regexMatchers)
	}
	for _, regex := range regexMatchers {
		if matches, err := regexp.MatchString(regex, pwd); !matches || err != nil {
			if err != nil {
				return err
			}
			return newInvalidPwdErr(minRuneCount, maxRuneCount, regexMatchers)
		}
	}
	return nil
}
