package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"database/sql"
	"github.com/0xor1/isql"
)

func newSqlStore(accountsDB, pwdsDB isql.DB) store {
	if accountsDB == nil || pwdsDB == nil {
		panic(NilOrInvalidCriticalParamErr)
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

func (s *sqlStore) accountWithCiNameExists(name string) bool {
	row := s.accountsDB.QueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?;`, name)
	count := 0
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	return count != 0
}

func (s *sqlStore) getAccountByCiName(name string) *account {
	row := s.accountsDB.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE name = ?;`, name)
	acc := account{}
	if err := row.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsUser); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &acc
}

func (s *sqlStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) {
	id := []byte(user.Id)
	if _, err := s.accountsDB.Exec(`CALL createUser(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, id, user.Name, user.CreatedOn, user.Region, user.NewRegion, user.Shard, user.HasAvatar, user.IsUser, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode); err != nil {
		panic(err)
	}
	if _, err := s.pwdsDB.Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?);`, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getUserByEmail(email string) *fullUserInfo {
	row := s.accountsDB.QueryRow(`SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE u.email = ?;`, email)
	user := fullUserInfo{}
	if err := row.Scan(&user.Id, &user.Name, &user.CreatedOn, &user.Region, &user.NewRegion, &user.Shard, &user.HasAvatar, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &user
}

func (s *sqlStore) getUserById(id Id) *fullUserInfo {
	row := s.accountsDB.QueryRow(`SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser, u.email, u.newEmail, u.activationCode, u.activated, u.newEmailConfirmationCode, u.resetPwdCode FROM accounts AS a JOIN users AS u ON a.id = u.id WHERE a.id = ?;`, []byte(id))
	user := fullUserInfo{}
	if err := row.Scan(&user.Id, &user.Name, &user.CreatedOn, &user.Region, &user.NewRegion, &user.Shard, &user.HasAvatar, &user.IsUser, &user.Email, &user.NewEmail, &user.activationCode, &user.activated, &user.newEmailConfirmationCode, &user.resetPwdCode); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &user
}

func (s *sqlStore) getPwdInfo(id Id) *pwdInfo {
	row := s.pwdsDB.QueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?;`, []byte(id))
	pwd := pwdInfo{}
	if err := row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &pwd
}

func (s *sqlStore) updateUser(user *fullUserInfo) {
	id := []byte(user.Id)
	if _, err := s.accountsDB.Exec(`CALL updateUser(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, id, user.Name, user.CreatedOn, user.Region, user.NewRegion, user.Shard, user.HasAvatar, user.Email, user.NewEmail, user.activationCode, user.activated, user.newEmailConfirmationCode, user.resetPwdCode); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updateAccount(account *account) {
	if _, err := s.accountsDB.Exec(`UPDATE accounts SET name=?, createdOn=?, region=?, newRegion=?, shard=?, hasAvatar=?, isUser=? WHERE id = ?;`, account.Name, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsUser, []byte(account.Id)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) {
	if _, err := s.pwdsDB.Exec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?;`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) deleteAccountAndAllAssociatedMemberships(id Id) {
	castId := []byte(id)
	if _, err := s.accountsDB.Exec(`CALL deleteAccountAndAllAssociatedMemberships(?);`, castId); err != nil {
		panic(err)
	}
	if _, err := s.pwdsDB.Exec(`DELETE FROM pwds WHERE id = ?;`, castId); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getAccount(id Id) *account {
	row := s.accountsDB.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE id = ?;`, []byte(id))
	a := account{}
	if err := row.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &a
}

func (s *sqlStore) getAccounts(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	var query bytes.Buffer
	query.WriteString(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE id IN (`)
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
		panic(err)
	}
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		err := rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser)
		if err != nil {
			panic(err)
		}
		res = append(res, &a)
	}
	return res
}

func (s *sqlStore) getUsers(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	var query bytes.Buffer
	query.WriteString(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isUser FROM accounts WHERE isUser = true AND id IN (`)
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
		panic(err)
	}
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		u := account{}
		err := rows.Scan(&u.Id, &u.Name, &u.CreatedOn, &u.Region, &u.NewRegion, &u.Shard, &u.HasAvatar, &u.IsUser)
		if err != nil {
			panic(err)
		}
		res = append(res, &u)
	}
	return res
}

func (s *sqlStore) createOrgAndMembership(org *account, user Id) {
	if _, err := s.accountsDB.Exec(`CALL  createOrgAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?);`, []byte(org.Id), org.Name, org.CreatedOn, org.Region, org.NewRegion, org.Shard, org.HasAvatar, false, []byte(user)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getUsersOrgs(user Id, offset, limit int) ([]*account, int) {
	row := s.accountsDB.QueryRow(`SELECT COUNT(*) FROM memberships WHERE user = ?;`, []byte(user))
	total := 0
	if err := row.Scan(&total); err != nil {
		panic(err)
	}
	rows, err := s.accountsDB.Query(`SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isUser FROM accounts AS a JOIN memberships AS m ON a.id = m.org WHERE m.user = ? ORDER BY a.name ASC LIMIT ?, ?;`, []byte(user), offset, limit)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*account, 0, limit)
	for rows.Next() {
		a := account{}
		rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsUser)
		if err != nil {
			panic(err)
		}
		res = append(res, &a)
	}
	return res, total
}

func (s *sqlStore) createMemberships(org Id, users []Id) {
	var query bytes.Buffer
	args := make([]interface{}, 0, len(users)*2)
	query.WriteString(`INSERT INTO memberships (org, user) VALUES `)
	for i, id := range users {
		if i == 0 {
			query.WriteString(`(?, ?)`)
		} else {
			query.WriteString(`, (?, ?)`)
		}
		args = append(args, []byte(org), []byte(id))
	}
	if _, err := s.accountsDB.Exec(query.String(), args...); err != nil {
		panic(err)
	}
}

func (s *sqlStore) deleteMemberships(org Id, users []Id) {
	castedIds := make([]interface{}, 0, len(users)+1)
	castedIds = append(castedIds, []byte(org))
	var query bytes.Buffer
	query.WriteString(`DELETE FROM memberships WHERE org = ? AND user IN (`)
	for i, id := range users {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	if _, err := s.accountsDB.Exec(query.String(), castedIds...); err != nil {
		panic(err)
	}
}
