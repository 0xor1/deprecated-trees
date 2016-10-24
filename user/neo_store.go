package user

import (
	"bytes"
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
	ScryptSalt               []byte `json:"scryptSalt"`
	ScryptPwd                []byte `json:"scryptPwd"`
	ScryptN                  int    `json:"scryptN"`
	ScryptR                  int    `json:"scryptR"`
	ScryptP                  int    `json:"scryptP"`
	ScryptKeyLen             int    `json:"scryptKeyLen"`
}

func newNeoFullUserInfo(u *FullUserInfo) *neoFullUserInfo {
	return &neoFullUserInfo{
		Me:                       u.Me,
		RegistrationTime:         u.RegistrationTime.Unix(),
		ActivationCode:           u.ActivationCode,
		ActivationTime:           u.ActivationTime.Unix(),
		NewEmailConfirmationCode: u.NewEmailConfirmationCode(),
		ResetPwdCode:             u.ResetPwdCode,
		ScryptSalt:               u.ScryptSalt,
		ScryptPwd:                u.ScryptPwd,
		ScryptN:                  u.ScryptN,
		ScryptR:                  u.ScryptR,
		ScryptP:                  u.ScryptP,
		ScryptKeyLen:             u.ScryptKeyLen,
	}
}

func (nu *neoFullUserInfo) toFullUserInfo() *FullUserInfo {
	return &FullUserInfo{
		Me:                       nu.Me,
		RegistrationTime:         time.Unix(nu.RegistrationTime, 0),
		ActivationCode:           nu.ActivationCode,
		ActivationTime:           time.Unix(nu.ActivationTime, 0),
		NewEmailConfirmationCode: nu.NewEmailConfirmationCode(),
		ResetPwdCode:             nu.ResetPwdCode,
		ScryptSalt:               nu.ScryptSalt,
		ScryptPwd:                nu.ScryptPwd,
		ScryptN:                  nu.ScryptN,
		ScryptR:                  nu.ScryptR,
		ScryptP:                  nu.ScryptP,
		ScryptKeyLen:             nu.ScryptKeyLen,
	}
}

func NewNeoStore(db *neo.Database, log zap.Logger) (*Store, error) {
	log.Debug("user.NewNeoApi initializing database indexes and constraints")

	err := db.Cypher(&neo.CypherQuery{
		Statement: `
CREATE CONSTRAINT ON (u:USER) ASSERT u.id IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT exists(u.id)
CREATE CONSTRAINT ON (u:USER) ASSERT u.email IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT exists(u.email)
CREATE CONSTRAINT ON (u:USER) ASSERT u.activationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.newEmailConfirmationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.activationCode IS UNIQUE
CREATE CONSTRAINT ON (u:USER) ASSERT u.resetPwdCode IS UNIQUE
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

func (s *neoStore) getByUniqueStringProperty(propName, propValue string) (*FullUserInfo, error) {
	res := []*FullUserInfo{}
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
	return res[0], nil
}

func (s *neoStore) GetByEmail(email string) (*FullUserInfo, error) {
	return s.getByUniqueStringProperty("email", email)
}

func (s *neoStore) GetById(id string) (*FullUserInfo, error) {
	return s.getByUniqueStringProperty("id", id)
}

func (s *neoStore) GetByActivationCode(activationCode string) (*FullUserInfo, error) {
	return s.getByUniqueStringProperty("activationCode", activationCode)
}

func (s *neoStore) GetByNewEmailConfirmationCode(newEmailconfirmationCode string) (*FullUserInfo, error) {
	return s.getByUniqueStringProperty("newEmailconfirmationCode", newEmailconfirmationCode)
}

func (s *neoStore) GetByResetPwdCode(resetPwdCode string) (*FullUserInfo, error) {
	return s.getByUniqueStringProperty("resetPwdCode", resetPwdCode)
}

func (s *neoStore) save(nu *neoFullUserInfo) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("MERGE (:USER {id:%q,firstName:%q,lastName:%q,email:%q,registrationTime:%d", nu.Id, nu.FirstName, nu.LastName, nu.Email, nu.RegistrationTime))
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
	buf.WriteString(fmt.Sprintf(",scryptSalt:{scryptSalt},scryptPwd:{scryptPwd},scryptN:%d,scryptR:%d,scryptP:%d,scryptKeyLen:%d})", nu.ScryptN, nu.ScryptR, nu.ScryptP, nu.ScryptKeyLen))
	props := neo.Props{
		"scryptSalt": nu.ScryptSalt,
		"scryptPwd":  nu.ScryptPwd,
	}
	return s.db.Cypher(&neo.CypherQuery{
		Statement:  buf.String(),
		Parameters: props,
	})
}

func (s *neoStore) Create(user *FullUserInfo) error {
	return s.save(newNeoFullUserInfo(user))
}

func (s *neoStore) Update(user *FullUserInfo) error {
	return s.save(newNeoFullUserInfo(user))
}

func (s *neoStore) Delete(id string) error {
	return s.db.Cypher(&neo.CypherQuery{
		Statement: fmt.Sprintf("MATCH (u:USER {id:%q}) DETACH DELETE u", id),
	})
}
