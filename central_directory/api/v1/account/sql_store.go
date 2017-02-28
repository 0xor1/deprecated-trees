package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
)

func newSqlStore(accountsDB, pwdsDB isql.DB) store {
	if accountsDB == nil {
		NilCriticalParamPanic("accountsDB")
	}
	if pwdsDB == nil {
		NilCriticalParamPanic("pwdsDB")
	}
	return &sqlStore{
		accountsDB: accountsDB,
		pwdsDB: pwdsDB,
	}
}

type sqlStore struct{
	accountsDB isql.DB
	pwdsDB isql.DB
}

var query_accountWithCiNameExists = `SELECT COUNT(*) FROM accounts WHERE name = ?;`
func (s *sqlStore) accountWithCiNameExists(name string) (bool, error) {
	row := s.accountsDB.QueryRow(query_accountWithCiNameExists, name)
	count := 0
	err := row.Scan(&count)
	return count != 0, err
}

var query_getAccountByCiName = `SELECT id, name, created, region, newRegion, shard, isUser FROM accounts WHERE name = ?;`
func (s *sqlStore) getAccountByCiName(name string) (*account, error) {
	row := s.accountsDB.QueryRow(query_getAccountByCiName, name)
	acc := account{}
	err := row.Scan(&acc.Id, &acc.Name, &acc.Created, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.IsUser)
	return &acc, err
}

var query_createUser_accounts = `INSERT INTO accounts (id, name, created, region, newRegion, shard, isUser) VALUES (?, ?, ?, ?, ?, ?); INSERT INTO users (id, email, newEmail, activationCode, activated, newEmailConfirmationCode, resetPwdCode) VALUES (?, ?, ?, ?, ?, ?, ?);`
var query_createUser_pwds = `INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?);`
func (s *sqlStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	if _, err := s.accountsDB.Exec(query_createUser_accounts, user.Id, user.Name, user.Created, user.Region, user.NewRegion, user.Shard, user.IsUser, user.Id, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode); err != nil {
		return err
	}
	return s.pwdsDB.Exec(query_createUser_pwds, user.Id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
}

var query_getUserByCiName = `SELECT a.id, a.name, a.created, a.region, a.newRegion, a.shard, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.name = ?;`
func (s *sqlStore) getUserByCiName(name string) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserByCiName, name)
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.Created, &user.Region, &user.NewRegion, &user.Shard, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	return &user, err
}

var query_getUserByEmail = `SELECT a.id, a.name, a.created, a.region, a.newRegion, a.shard, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE u.email = ?;`
func (s *sqlStore) getUserByEmail(email string) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserByEmail, email)
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.Created, &user.Region, &user.NewRegion, &user.Shard, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	return &user, err
}

var query_getUserByEmail = `SELECT a.id, a.name, a.created, a.region, a.newRegion, a.shard, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.id = ?;`
func (s *sqlStore) getUserById(id Id) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserByEmail, id)
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.Created, &user.Region, &user.NewRegion, &user.Shard, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	return &user, err
}

var query_getPwdInfo = `SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?;`
func (s *sqlStore) getPwdInfo(id Id) (*pwdInfo, error) {
	row := s.pwdsDB.QueryRow(query_getUserByEmail, id)
	pwd := pwdInfo{}
	err := row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)
	return &pwd, err
}

var query_updateUser = `UPDATE accounts SET name=?, created=?, region=?, newRegion=?, shard=?, isUser=? WHERE id = ?; UPDATE users SET email=?, newEmail=?, activationCode=?, activated=?, newEmailConfirmationCode=?, resetPwdCode=? WHERE id = ?;`
func (s *sqlStore) updateUser(user *fullUserInfo) error {
	_, err := s.accountsDB.Exec(query_updateUser, user.Name, user.Created, user.Region, user.NewRegion, user.Shard, user.IsUser, user.Id, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode, user.Id)
	return err
}

var query_updatePwdInfo = `UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?;`
func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) error {
	_, err := s.pwdsDB.Exec(query_updatePwdInfo, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, id)
	return err
}

var query_deleteUserAndAllAssociatedMemberships_accounts = `DELETE FROM Memberships WHERE user = ?; DELETE FROM users WHERE id = ?; DELETE FROM account WHERE id = ?;`
var query_deleteUserAndAllAssociatedMemberships_pwds = `DELETE FROM pwds WHERE id = ?;`
func (s *sqlStore) deleteUserAndAllAssociatedMemberships(id Id) error {
	_, err := s.accountsDB.Exec(query_deleteUserAndAllAssociatedMemberships_accounts, id, id, id)
	if err != nil {
		return err
	}
	_, err = s.pwdsDB.Exec(query_deleteUserAndAllAssociatedMemberships_pwds, id)
	return err
}

func (s *sqlStore) getUsers(ids []Id) ([]*user, error) {

}

func (s *sqlStore) createOrgAndMembership(user Id, org *org) error {

}

func (s *sqlStore) getOrgById(id Id) (*org, error) {

}

func (s *sqlStore) getOrgByName(name string) (*org, error) {

}

func (s *sqlStore) updateOrg(org *org) error {

}

func (s *sqlStore) deleteOrgAndAllAssociatedMemberships(id Id) error {

}

func (s *sqlStore) getOrgs(ids []Id) ([]*org, error) {

}

func (s *sqlStore) getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error) {

}

func (s *sqlStore) createMemberships(org Id, users []Id) error {

}

func (s *sqlStore) deleteMemberships(org Id, users []Id) error {

}

