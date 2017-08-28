package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"database/sql"
	"github.com/0xor1/isql"
)

func newSqlStore(accounts, pwds isql.ReplicaSet) store {
	if accounts == nil || pwds == nil {
		panic(InvalidArgumentsErr)
	}
	return &sqlStore{
		accounts: accounts,
		pwds:     pwds,
	}
}

type sqlStore struct {
	accounts isql.ReplicaSet
	pwds     isql.ReplicaSet
}

func (s *sqlStore) accountWithCiNameExists(name string) bool {
	row := s.accounts.QueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?;`, name)
	count := 0
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	return count != 0
}

func (s *sqlStore) getAccountByCiName(name string) *account {
	row := s.accounts.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?;`, name)
	acc := account{}
	if err := row.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &acc
}

func (s *sqlStore) createPersonalAccount(account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	id := []byte(account.Id)
	if _, err := s.accounts.Exec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, id, account.Name, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode); err != nil {
		panic(err)
	}
	if _, err := s.pwds.Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?);`, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getPersonalAccountByEmail(email string) *fullPersonalAccountInfo {
	row := s.accounts.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err := row.Scan(&account.Id, &account.Name, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &account
}

func (s *sqlStore) getPersonalAccountById(id Id) *fullPersonalAccountInfo {
	row := s.accounts.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, []byte(id))
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err := row.Scan(&account.Id, &account.Name, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &account
}

func (s *sqlStore) getPwdInfo(id Id) *pwdInfo {
	row := s.pwds.QueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?;`, []byte(id))
	pwd := pwdInfo{}
	if err := row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &pwd
}

func (s *sqlStore) updatePersonalAccount(personalAccountInfo *fullPersonalAccountInfo) {
	if _, err := s.accounts.Exec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, []byte(personalAccountInfo.Id), personalAccountInfo.Name, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updateAccount(account *account) {
	if _, err := s.accounts.Exec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) {
	if _, err := s.pwds.Exec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?;`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) deleteAccountAndAllAssociatedMemberships(id Id) {
	castId := []byte(id)
	if _, err := s.accounts.Exec(`CALL deleteAccountAndAllAssociatedMemberships(?);`, castId); err != nil {
		panic(err)
	}
	if _, err := s.pwds.Exec(`DELETE FROM pwds WHERE id = ?;`, castId); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getAccount(id Id) *account {
	row := s.accounts.QueryRow(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?;`, []byte(id))
	a := account{}
	if err := row.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &a
}

func (s *sqlStore) getAccounts(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	query := bytes.NewBufferString(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (`)
	for i, id := range ids {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	rows, err := s.accounts.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		err := rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)
		if err != nil {
			panic(err)
		}
		res = append(res, &a)
	}
	return res
}

func (s *sqlStore) searchAccounts(nameStartsWith string) []*account {
	rows, err := s.accounts.Query(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?`, nameStartsWith+"%", 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		err := rows.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)
		if err != nil {
			panic(err)
		}
		res = append(res, &acc)
	}
	return res
}

func (s *sqlStore) searchPersonalAccounts(nameOrEmailStartsWith string) []*account {
	rows, err := s.accounts.Query(`SELECT a.id, a.name, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar FROM ((SELECT id, name, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, nameOrEmailStartsWith+"%", 0, 100, nameOrEmailStartsWith+"%", 0, 100, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err := rows.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar)
		if err != nil {
			panic(err)
		}
		res = append(res, &acc)
	}
	return res
}


func (s *sqlStore) getPersonalAccounts(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	query := bytes.NewBufferString(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE activatedOn IS NOT NULL AND id IN (`)
	for i, id := range ids {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`);`)
	rows, err := s.accounts.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err := rows.Scan(&acc.Id, &acc.Name, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar)
		if err != nil {
			panic(err)
		}
		res = append(res, &acc)
	}
	return res
}

func (s *sqlStore) createGroupAccountAndMembership(account *account, memberId Id) {
	if _, err := s.accounts.Exec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?);`, []byte(account.Id), account.Name, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, []byte(memberId)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getGroupAccounts(memberId Id, offset, limit int) ([]*account, int) {
	row := s.accounts.QueryRow(`SELECT COUNT(*) FROM memberships WHERE member=?;`, []byte(memberId))
	total := 0
	if err := row.Scan(&total); err != nil {
		panic(err)
	}
	rows, err := s.accounts.Query(`SELECT id, name, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (SELECT account FROM memberships WHERE member = ?) ORDER BY name ASC LIMIT ?, ?;`, []byte(memberId), offset, limit)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		panic(err)
	}
	res := make([]*account, 0, limit)
	for rows.Next() {
		a := account{}
		rows.Scan(&a.Id, &a.Name, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)
		if err != nil {
			panic(err)
		}
		res = append(res, &a)
	}
	return res, total
}

func (s *sqlStore) createMemberships(accountId Id, members []Id) {
	args := make([]interface{}, 0, len(members)*2)
	query := bytes.NewBufferString(`INSERT INTO memberships (account, member) VALUES `)
	for i, member := range members {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`(?, ?)`)
		args = append(args, []byte(accountId), []byte(member))
	}
	if _, err := s.accounts.Exec(query.String(), args...); err != nil {
		panic(err)
	}
}

func (s *sqlStore) deleteMemberships(accountId Id, members []Id) {
	castedIds := make([]interface{}, 0, len(members)+1)
	castedIds = append(castedIds, []byte(accountId))
	query := bytes.NewBufferString(`DELETE FROM memberships WHERE account=? AND member IN (`)
	for i, member := range members {
		if i != 0 {
			query.WriteString(`,`)
		}
		query.WriteString(`?`)
		castedIds = append(castedIds, []byte(member))
	}
	query.WriteString(`);`)
	if _, err := s.accounts.Exec(query.String(), castedIds...); err != nil {
		panic(err)
	}
}
