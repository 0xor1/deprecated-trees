package user

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	neo "github.com/jmcvetta/neoism"
	"github.com/uber-go/zap"
	"time"
)

type neoFullUserInfo struct {
	Me
	RegistrationTime         int64  `json:"registrationTime"`
	ActivationCode           string `json:"activationCode,omitempty"`
	ActivationTime           int64  `json:"activationTime,omitempty"`
	NewEmailConfirmationCode string `json:"newEmailConfirmationCode,omitempty"`
	ResetPwdCode             string `json:"resetPwdCode,omitempty"`
	ScryptSalt               string `json:"scryptSalt"`
	ScryptPwd                string `json:"scryptPwd"`
	ScryptN                  int    `json:"scryptN"`
	ScryptR                  int    `json:"scryptR"`
	ScryptP                  int    `json:"scryptP"`
	ScryptKeyLen             int    `json:"scryptKeyLen"`
	PersonalStoreId          int    `json:"personalStoreId"`
}

func newNeoFullUserInfo(u *fullUserInfo) *neoFullUserInfo {
	return &neoFullUserInfo{
		Me:                       u.Me,
		RegistrationTime:         u.RegistrationTime.Unix(),
		ActivationCode:           u.ActivationCode,
		ActivationTime:           u.ActivationTime.Unix(),
		NewEmailConfirmationCode: u.NewEmailConfirmationCode,
		ResetPwdCode:             u.ResetPwdCode,
		ScryptSalt:               hex.EncodeToString(u.ScryptSalt),
		ScryptPwd:                hex.EncodeToString(u.ScryptPwd),
		ScryptN:                  u.ScryptN,
		ScryptR:                  u.ScryptR,
		ScryptP:                  u.ScryptP,
		ScryptKeyLen:             u.ScryptKeyLen,
		PersonalStoreId:          u.PersonalStoreId,
	}
}

func (nu *neoFullUserInfo) toFullUserInfo() (*fullUserInfo, error) {
	saltBytes, err := hex.DecodeString(nu.ScryptSalt)
	if err != nil {
		return nil, errors.New("failed to decode salt hex string to byte slice")
	}
	pwdBytes, err := hex.DecodeString(nu.ScryptPwd)
	if err != nil {
		return nil, errors.New("failed to decode password hex string to byte slice")
	}
	var activationTime *time.Time
	if nu.ActivationTime > 0 {
		tmpTime := time.Unix(nu.ActivationTime, 0).UTC()
		activationTime = &tmpTime
	}
	return &fullUserInfo{
		Me:                       nu.Me,
		RegistrationTime:         time.Unix(nu.RegistrationTime, 0).UTC(),
		ActivationCode:           nu.ActivationCode,
		ActivationTime:           activationTime,
		NewEmailConfirmationCode: nu.NewEmailConfirmationCode,
		ResetPwdCode:             nu.ResetPwdCode,
		ScryptSalt:               saltBytes,
		ScryptPwd:                pwdBytes,
		ScryptN:                  nu.ScryptN,
		ScryptR:                  nu.ScryptR,
		ScryptP:                  nu.ScryptP,
		ScryptKeyLen:             nu.ScryptKeyLen,
	}, nil
}

func NewNeoApi(db *neo.Database, linkMailer LinkMailer, usernameRegexMatchers, pwdRegexMatchers []string, maxSearchLimitResults, usernameMinRuneCount, usernameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, log zap.Logger) (Api, error) {
	if db == nil {
		return nil, NilNeoDbErr
	}

	store, err := newNeoStore(db, log)
	if err != nil {
		return nil, err
	}

	return newApi(store, linkMailer, usernameRegexMatchers, pwdRegexMatchers, maxSearchLimitResults, usernameMinRuneCount, usernameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, cryptoCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen, log)
}

func newNeoStore(db *neo.Database, log zap.Logger) (store, error) {
	log.Debug("user.NewNeoApi initializing database indexes and constraints")

	err := db.Cypher(&neo.CypherQuery{
		Statement: `
CREATE CONSTRAINT ON (u:USER) ASSERT u.id IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT exists(u.id)
CREATE CONSTRAINT ON (u:USER) ASSERT u.username IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT exists(u.username)
CREATE CONSTRAINT ON (u:USER) ASSERT u.email IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT exists(u.email)
CREATE CONSTRAINT ON (u:USER) ASSERT u.activationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.newEmailConfirmationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.activationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.resetPwdCode IS UNIQUE
CREATE CONSTRAINT ON (o:ORG_REF) ASSERT o.id IS UNIQUE
`,
	})
	if err != nil {
		log.Fatal("user.NewNeoStore failed to initialise database indexes and constraints", zap.Error(err))
	}

	return &neoStore{
		db:  db,
		log: log,
	}, nil
}

type neoStore struct {
	db  *neo.Database
	log zap.Logger
}

func (s *neoStore) getByUniqueStringProperty(propName, propValue string) (*fullUserInfo, error) {
	res := []*neoFullUserInfo{}
	if err := s.db.Cypher(&neo.CypherQuery{
		Statement:  fmt.Sprintf("MATCH (u:USER {%s:{%s}}) RETURN u", propName, propName),
		Parameters: neo.Props{propName: propValue},
		Result:     &res,
	}); len(res) != 1 || err != nil {
		if err == nil {
			if len(res) == 0 {
				err = errors.New(fmt.Sprintf("user with %s: %s not found", propName, propValue))
			} else {
				err = errors.New(fmt.Sprintf("Critical data error more than one user exists with %s: %s", propName, propValue))
			}
		}
		return nil, err
	}
	return res[0].toFullUserInfo()
}

func (s *neoStore) getByUsername(username string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("username", username)
}

func (s *neoStore) getByEmail(email string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("email", email)
}

func (s *neoStore) getById(id string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("id", id)
}

func (s *neoStore) getByActivationCode(activationCode string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("activationCode", activationCode)
}

func (s *neoStore) getByNewEmailConfirmationCode(newEmailconfirmationCode string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("newEmailconfirmationCode", newEmailconfirmationCode)
}

func (s *neoStore) getByResetPwdCode(resetPwdCode string) (*fullUserInfo, error) {
	return s.getByUniqueStringProperty("resetPwdCode", resetPwdCode)
}

func (s *neoStore) getByIds(ids []string) ([]*User, error) {
	return nil, errors.New("not implemented")
}

func (s *neoStore) search(search string, limit int) ([]*User, error) {
	return nil, errors.New("not implemented")
}

func (s *neoStore) save(nu *neoFullUserInfo) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("MERGE (:USER {id:%q,username:%q,email:%q,registrationTime:%d", nu.Id, nu.Username, nu.Email, nu.RegistrationTime))
	if len(nu.NewEmail) > 0 {
		buf.WriteString(fmt.Sprintf(",newEmail:%q", nu.NewEmail))
	} else {
		buf.WriteString(",newEmail:null")
	}
	if len(nu.ActivationCode) > 0 {
		buf.WriteString(fmt.Sprintf(",activationCode:%q", nu.ActivationCode))
	} else {
		buf.WriteString(",activationCode:null")
	}
	if nu.ActivationTime > 0 {
		buf.WriteString(fmt.Sprintf(",activationTime:%d", nu.ActivationTime))
	} else {
		buf.WriteString(",activationTime:null")
	}
	if len(nu.ResetPwdCode) > 0 {
		buf.WriteString(fmt.Sprintf(",resetPwdCode:%q", nu.ResetPwdCode))
	} else {
		buf.WriteString(",resetPwdCode:null")
	}
	if len(nu.ResetPwdCode) > 0 {
		buf.WriteString(fmt.Sprintf(",resetPwdCode:%q", nu.ResetPwdCode))
	} else {
		buf.WriteString(",resetPwdCode:null")
	}
	buf.WriteString(fmt.Sprintf(",scryptSalt:%q,scryptPwd:%q,scryptN:%d,scryptR:%d,scryptP:%d,scryptKeyLen:%d})", nu.ScryptSalt, nu.ScryptPwd, nu.ScryptN, nu.ScryptR, nu.ScryptP, nu.ScryptKeyLen))
	return s.db.Cypher(&neo.CypherQuery{
		Statement: buf.String(),
	})
}

func (s *neoStore) create(user *fullUserInfo) error {
	return s.save(newNeoFullUserInfo(user))
}

func (s *neoStore) update(user *fullUserInfo) error {
	return s.save(newNeoFullUserInfo(user))
}

func (s *neoStore) delete(id string) error {
	return s.db.Cypher(&neo.CypherQuery{
		Statement:  "MATCH (u:USER {id:{id}}) DETACH DELETE u",
		Parameters: neo.Props{"id": id},
	})
}
