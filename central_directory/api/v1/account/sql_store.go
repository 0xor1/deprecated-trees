package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
	"bytes"
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

var query_createUser_accounts = `INSERT INTO accounts (id, name, created, region, newRegion, shard, isUser) VALUES (?, ?, ?, ?, ?, ?, ?);`
var query_createUser_users = `INSERT INTO users (id, email, newEmail, activationCode, activated, newEmailConfirmationCode, resetPwdCode) VALUES (?, ?, ?, ?, ?, ?, ?);`
var query_createUser_pwds = `INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?);`
func (s *sqlStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	id := []byte(user.Id)
	if _, err := s.accountsDB.Exec(query_createUser_accounts, id, user.Name, user.Created, user.Region, user.NewRegion, user.Shard, true); err != nil {
		return err
	}
	if _, err := s.accountsDB.Exec(query_createUser_users, id, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode); err != nil {
		return err
	}
	_, err := s.pwdsDB.Exec(query_createUser_pwds, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	return err
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

var query_getUserById = `SELECT a.id, a.name, a.created, a.region, a.newRegion, a.shard, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.id = ?;`
func (s *sqlStore) getUserById(id Id) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserById, []byte(id))
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.Created, &user.Region, &user.NewRegion, &user.Shard, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	return &user, err
}

var query_getPwdInfo = `SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?;`
func (s *sqlStore) getPwdInfo(id Id) (*pwdInfo, error) {
	row := s.pwdsDB.QueryRow(query_getPwdInfo, []byte(id))
	pwd := pwdInfo{}
	err := row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)
	return &pwd, err
}

var query_updateUser_accounts = `UPDATE accounts SET name=?, created=?, region=?, newRegion=?, shard=? WHERE id = ?;`
var query_updateUser_users = `UPDATE users SET email=?, newEmail=?, activationCode=?, activated=?, newEmailConfirmationCode=?, resetPwdCode=? WHERE id = ?;`
func (s *sqlStore) updateUser(user *fullUserInfo) error {
	id := []byte(user.Id)
	if _, err := s.accountsDB.Exec(query_updateUser_accounts, user.Name, user.Created, user.Region, user.NewRegion, user.Shard, id); err != nil {
		return err
	}
	_, err := s.accountsDB.Exec(query_updateUser_users, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode, id)
	return err
}

var query_updatePwdInfo = `UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?;`
func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) error {
	_, err := s.pwdsDB.Exec(query_updatePwdInfo, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id))
	return err
}

var query_deleteUserAndAllAssociatedMemberships_memberships = `DELETE FROM memberships WHERE user = ?;`
var query_deleteUserAndAllAssociatedMemberships_users = `DELETE FROM users WHERE id = ?;`
var query_deleteUserAndAllAssociatedMemberships_accounts = `DELETE FROM accounts WHERE id = ?;`
var query_deleteUserAndAllAssociatedMemberships_pwds = `DELETE FROM pwds WHERE id = ?;`
func (s *sqlStore) deleteUserAndAllAssociatedMemberships(id Id) error {
	castId := []byte(id)
	_, err := s.accountsDB.Exec(query_deleteUserAndAllAssociatedMemberships_memberships, castId)
	if err != nil {
		return err
	}
	_, err = s.accountsDB.Exec(query_deleteUserAndAllAssociatedMemberships_users, castId)
	if err != nil {
		return err
	}
	_, err = s.accountsDB.Exec(query_deleteUserAndAllAssociatedMemberships_accounts, castId)
	if err != nil {
		return err
	}
	_, err = s.pwdsDB.Exec(query_deleteUserAndAllAssociatedMemberships_pwds, castId)
	return err
}

var query_getUsers = `SELECT id, name, created, region, newRegion, shard, isUser FROM accounts WHERE isUser = true AND id IN (`
func (s *sqlStore) getUsers(ids []Id) ([]*user, error) {
	castedIds := make([]interface{}, 0, len(ids))
	var query bytes.Buffer
	query.WriteString(query_getUsers)
	for i, id := range ids {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	rows, err := s.accountsDB.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	res := make([]*user, 0, len(ids))
	for rows.Next() {
		u := user{}
		err := rows.Scan(&u.Id, &u.Name, &u.Created, &u.Region, &u.NewRegion, &u.Shard, &u.IsUser)
		if err != nil {
			return nil, err
		}
		res = append(res, &u)
	}
	return res, nil
}

var query_createOrgAndMembership_org = `INSERT INTO accounts (id, name, created, region, newRegion, shard, isUser) VALUES (?, ?, ?, ?, ?, ?, ?);`
var query_createOrgAndMembership_membership = `INSERT INTO memberships (org, user) VALUES (?, ?);`
func (s *sqlStore) createOrgAndMembership(org *org, user Id) error {
	orgId := []byte(org.Id)
	_, err := s.accountsDB.Exec(query_createOrgAndMembership_org, orgId, org.Name, org.Created, org.Region, org.NewRegion, org.Shard, false)
	if err != nil {
		return err
	}
	_, err = s.accountsDB.Exec(query_createOrgAndMembership_membership, orgId, []byte(user))
	return err
}

var query_getOrgById = `SELECT id, name, created, region, newRegion, shard, isUser FROM accounts WHERE id = ?;`
func (s *sqlStore) getOrgById(id Id) (*org, error) {
	row := s.accountsDB.QueryRow(query_getOrgById, []byte(id))
	o := org{}
	err := row.Scan(&o.Id, &o.Name, &o.Created, &o.Region, &o.NewRegion, &o.Shard, &o.IsUser)
	return &o, err
}

var query_updateOrg = `UPDATE accounts SET name=?, created=?, region=?, newRegion=?, shard=? WHERE id = ?;`
func (s *sqlStore) updateOrg(org *org) error {
	_, err := s.accountsDB.Exec(query_updateOrg, org.Name, org.Created, org.Region, org.NewRegion, org.Shard, []byte(org.Id))
	return err
}

var query_deleteOrgAndAllAssociatedMemberships_memberships = `DELETE FROM memberships WHERE org = ?;`
var query_deleteOrgAndAllAssociatedMemberships_account = `DELETE FROM accounts WHERE id = ?;`
func (s *sqlStore) deleteOrgAndAllAssociatedMemberships(id Id) error {
	castId := []byte(id)
	_, err := s.accountsDB.Exec(query_deleteOrgAndAllAssociatedMemberships_memberships, castId)
	if err != nil {
		return nil
	}
	_, err = s.accountsDB.Exec(query_deleteOrgAndAllAssociatedMemberships_account, castId)
	return err
}

var query_getOrgs = `SELECT id, name, created, region, newRegion, shard, isUser FROM accounts WHERE isUser = false AND id IN (`
func (s *sqlStore) getOrgs(ids []Id) ([]*org, error) {
	castedIds := make([]interface{}, 0, len(ids))
	var query bytes.Buffer
	query.WriteString(query_getOrgs)
	for i, id := range ids {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	rows, err := s.accountsDB.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	res := make([]*org, 0, len(ids))
	for rows.Next() {
		o := org{}
		err := rows.Scan(&o.Id, &o.Name, &o.Created, &o.Region, &o.NewRegion, &o.Shard, &o.IsUser)
		if err != nil {
			return nil, err
		}
		res = append(res, &o)
	}
	return res, nil
}

var query_getUsersOrgs_total = `SELECT COUNT(*) FROM memberships WHERE user = ?;`
var query_getUsersOrgs = `SELECT a.id, a.name, a.created, a.region, a.newRegion, a.shard, a.isUser FROM accounts AS a JOIN memberships AS m ON a.id = m.org WHERE m.user = ? ORDER BY a.name ASC LIMIT ?, ?;`
func (s *sqlStore) getUsersOrgs(user Id, offset, limit int) ([]*org, int, error) {
	row := s.accountsDB.QueryRow(query_getUsersOrgs_total, []byte(user))
	total := 0
	if err := row.Scan(&total); err != nil {
		return nil, total, err
	}
	rows, err := s.accountsDB.Query(query_getUsersOrgs, []byte(user), offset, limit)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, 0, err
	}
	res := make([]*org, 0, limit)
	for rows.Next() {
		o := org{}
		rows.Scan(&o.Id, &o.Name, &o.Created, &o.Region, &o.NewRegion, &o.Shard, &o.IsUser)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, &o)
	}
	return res, total, nil
}

var query_createMemberships = `INSERT INTO memberships (org, user) VALUES `
func (s *sqlStore) createMemberships(org Id, users []Id) error {
	var query bytes.Buffer
	args := make([]interface{}, 0, len(users)*2)
	query.WriteString(query_createMemberships)
	for i, id := range users {
		if i == 0 {
			query.WriteString(`(?, ?)`)
		} else {
			query.WriteString(`, (?, ?)`)
		}
		args = append(args, []byte(org), []byte(id))
	}
	_, err := s.accountsDB.Exec(query.String(), args...)
	return err
}

var query_deleteMemberships = `DELETE FROM memberships WHERE org = ? AND user IN (`
func (s *sqlStore) deleteMemberships(org Id, users []Id) error {
	castedIds := make([]interface{}, 0, len(users) + 1)
	castedIds = append(castedIds, []byte(org))
	var query bytes.Buffer
	query.WriteString(query_deleteMemberships)
	for i, id := range users {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	_, err := s.accountsDB.Exec(query.String(), castedIds...)
	return err
}

