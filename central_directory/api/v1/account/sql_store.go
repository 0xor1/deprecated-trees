package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"bytes"
	"database/sql"
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
		pwdsDB:     pwdsDB,
	}
}

type sqlStore struct {
	accountsDB isql.DB
	pwdsDB     isql.DB
}

var query_accountWithCiNameExists = `SELECT COUNT(*) FROM accounts WHERE name = ?;`

func (s *sqlStore) accountWithCiNameExists(name string) (bool, error) {
	row := s.accountsDB.QueryRow(query_accountWithCiNameExists, name)
	count := 0
	err := row.Scan(&count)
	return count != 0, err
}

var query_getAccountByCiName = `SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE name = ?;`

func (s *sqlStore) getAccountByCiName(name string) (*account, error) {
	row := s.accountsDB.QueryRow(query_getAccountByCiName, name)
	acc := account{}
	err := row.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsUser)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &acc, err
}

var query_createUser_accounts = `CALL createUser(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
var query_createUser_pwds = `INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?);`

func (s *sqlStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	id := []byte(user.Id)
	if _, err := s.accountsDB.Exec(query_createUser_accounts, id, user.Name, user.CreatedOn, user.Region, user.NewRegion, user.Shard, user.HasAvatar, user.IsUser, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode); err != nil {
		return err
	}
	_, err := s.pwdsDB.Exec(query_createUser_pwds, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	return err
}

var query_getUserByCiName = `SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.name = ?;`

func (s *sqlStore) getUserByCiName(name string) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserByCiName, name)
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.CreatedOn, &user.Region, &user.NewRegion, &user.Shard, &user.HasAvatar, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

var query_getUserByEmail = `SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE u.email = ?;`

func (s *sqlStore) getUserByEmail(email string) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserByEmail, email)
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.CreatedOn, &user.Region, &user.NewRegion, &user.Shard, &user.HasAvatar, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

var query_getUserById = `SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.id = ?;`

func (s *sqlStore) getUserById(id Id) (*fullUserInfo, error) {
	row := s.accountsDB.QueryRow(query_getUserById, []byte(id))
	user := fullUserInfo{}
	err := row.Scan(&user.Id, &user.Name, &user.CreatedOn, &user.Region, &user.NewRegion, &user.Shard, &user.HasAvatar, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

var query_getPwdInfo = `SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?;`

func (s *sqlStore) getPwdInfo(id Id) (*pwdInfo, error) {
	row := s.pwdsDB.QueryRow(query_getPwdInfo, []byte(id))
	pwd := pwdInfo{}
	err := row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pwd, err
}

var query_updateUser = `CALL updateUser(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

func (s *sqlStore) updateUser(user *fullUserInfo) error {
	id := []byte(user.Id)
	_, err := s.accountsDB.Exec(query_updateUser, id, user.Name, user.CreatedOn, user.Region, user.NewRegion, user.Shard, user.HasAvatar, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode)
	return err
}

var query_updateAccount = `UPDATE accounts SET name=?, createdOn=?, region=?, newRegion=?, shard=?, hasAvatar=?, isUser=? WHERE id = ?;`

func (s *sqlStore) updateAccount(account *account) error {
	_, err := s.accountsDB.Exec(query_updateAccount, account.Name, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsUser, []byte(account.Id))
	return err
}

var query_updatePwdInfo = `UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?;`

func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) error {
	_, err := s.pwdsDB.Exec(query_updatePwdInfo, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id))
	return err
}

var query_deleteAccountAndAllAssociatedMemberships_accounts = `CALL deleteAccountAndAllAssociatedMemberships(?);`
var query_deleteAccountAndAllAssociatedMemberships_pwds = `DELETE FROM pwds WHERE id = ?;`

func (s *sqlStore) deleteAccountAndAllAssociatedMemberships(id Id) error {
	castId := []byte(id)
	_, err := s.accountsDB.Exec(query_deleteAccountAndAllAssociatedMemberships_accounts, castId)
	if err != nil {
		return err
	}
	_, err = s.pwdsDB.Exec(query_deleteAccountAndAllAssociatedMemberships_pwds, castId)
	return err
}

var query_getAccount = `SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE id = ?;`

func (s *sqlStore) getAccount(id Id) (*account, error) {
	row := s.accountsDB.QueryRow(query_getAccount, []byte(id))
	a := account{}
	err := row.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

var query_getAccounts = `SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE id IN (`

func (s *sqlStore) getAccounts(ids []Id) ([]*account, error) {
	castedIds := make([]interface{}, 0, len(ids))
	var query bytes.Buffer
	query.WriteString(query_getAccounts)
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
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		err := rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser)
		if err != nil {
			return nil, err
		}
		res = append(res, &a)
	}
	return res, nil
}

var query_getUsers = `SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE isUser = true AND id IN (`

func (s *sqlStore) getUsers(ids []Id) ([]*account, error) {
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
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		u := account{}
		err := rows.Scan(&u.Id, &u.Name, &u.CreatedOn, &u.Region, &u.NewRegion, &u.Shard, &u.HasAvatar, &u.IsUser)
		if err != nil {
			return nil, err
		}
		res = append(res, &u)
	}
	return res, nil
}

var query_createOrgAndMembership = `CALL  createOrgAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?);`

func (s *sqlStore) createOrgAndMembership(org *account, user Id) error {
	_, err := s.accountsDB.Exec(query_createOrgAndMembership, []byte(org.Id), org.Name, org.CreatedOn, org.Region, org.NewRegion, org.Shard, org.HasAvatar, false, []byte(user))
	return err
}

var query_getUsersOrgs_total = `SELECT COUNT(*) FROM memberships WHERE user = ?;`
var query_getUsersOrgs = `SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser FROM accounts AS a JOIN memberships AS m ON a.id = m.org WHERE m.user = ? ORDER BY a.name ASC LIMIT ?, ?;`

func (s *sqlStore) getUsersOrgs(user Id, offset, limit int) ([]*account, int, error) {
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
	res := make([]*account, 0, limit)
	for rows.Next() {
		a := account{}
		rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser)
		if err != nil {
			return nil, 0, err
		}
		res = append(res, &a)
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
	castedIds := make([]interface{}, 0, len(users)+1)
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
