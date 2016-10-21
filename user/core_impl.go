package user

import (
	"github.com/robsix/golog"
	"golang.org/x/crypto/scrypt"
	"io"
	"crypto/rand"
	"errors"
)

func NewStore(guai GetUserAuthenticationInfo, sui SaveUserInfo, sp SavePwd, se SendEmail, g Get, l golog.Log) UserStore {
	return &userStore{
		getUserAuthenticationInfo: guai,
		saveUserInfo:              sui,
		savePwd:                   sp,
		sendEmail:		   se,
		get:                       g,
		log:                       l,
	}
}

type userStore struct {
	getUserAuthenticationInfo GetUserAuthenticationInfo
	saveUserInfo              SaveUserInfo
	savePwd                   SavePwd
	sendEmail		  SendEmail
	get                       Get
	log                       golog.Log
}

func (us *userStore) Authenticate(email, pwd string) (string, error) {
	us.log.Info("userStore.Authenticate(%q, ********)", email)

	userId, scryptPwdActual, salt, N, r, p, keyLen, err := us.getUserAuthenticationInfo(email)
	if err != nil {
		us.log.Error("userStore.Authenticate us.getUserAuthenticationInfo error: %v", err)
		return "", err
	}

	scryptPwdTryBytes, err := scrypt.Key([]byte(pwd), []byte(salt), N, r, p, keyLen)
	if err != nil {
		us.log.Error("userStore.Authenticate scrypt.Key error: %v", err)
		return "", err
	}
	scryptPwdTry := string(scryptPwdTryBytes)

	if scryptPwdActual != scryptPwdTry {
		err = errors.New("pwds do not match")
		us.log.Error("userStore.Authenticate error: %v", err)
		return "", err
	}

	return userId, nil
}

func (us *userStore) Register(email, firstName, lastName, pwd string) error {
	return nil
}

func (us *userStore) ForgotPwd(email string) error {
	return nil
}

func (us *userStore) ChangePwd(id, oldPwd, newPwd string) error {
	return nil
}

func (us *userStore) Get(ids []string) ([]*User, error) {
	return nil, nil
}

//helpers
func generateSalt(length int) ([]byte, error) {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}
