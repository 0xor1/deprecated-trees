package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bytes"
	"github.com/0xor1/isql"
)

func newSqlStore(accounts, pwds isql.ReplicaSet) store {
	if accounts == nil || pwds == nil {
		InvalidArgumentsErr.Panic()
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
	row := s.accounts.QueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?`, name)
	count := 0
	PanicIf(row.Scan(&count))
	return count != 0
}

func (s *sqlStore) getAccountByCiName(name string) *account {
	row := s.accounts.QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?`, name)
	acc := account{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)) {
		return nil
	}
	return &acc
}

func (s *sqlStore) createPersonalAccount(account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	id := []byte(account.Id)
	_, err := s.accounts.Exec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode)
	PanicIf(err)
	_, err = s.pwds.Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	PanicIf(err)
}

func (s *sqlStore) getPersonalAccountByEmail(email string) *fullPersonalAccountInfo {
	row := s.accounts.QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func (s *sqlStore) getPersonalAccountById(id Id) *fullPersonalAccountInfo {
	row := s.accounts.QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, []byte(id))
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func (s *sqlStore) getPwdInfo(id Id) *pwdInfo {
	row := s.pwds.QueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?`, []byte(id))
	pwd := pwdInfo{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)) {
		return nil
	}
	return &pwd
}

func (s *sqlStore) updatePersonalAccount(personalAccountInfo *fullPersonalAccountInfo) {
	_, err := s.accounts.Exec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(personalAccountInfo.Id), personalAccountInfo.Name, personalAccountInfo.DisplayName, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode)
	PanicIf(err)
}

func (s *sqlStore) updateAccount(account *account) {
	_, err := s.accounts.Exec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal)
	PanicIf(err)
}

func (s *sqlStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) {
	_, err := s.pwds.Exec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, []byte(id))
	PanicIf(err)
}

func (s *sqlStore) deleteAccountAndAllAssociatedMemberships(id Id) {
	castId := []byte(id)
	_, err := s.accounts.Exec(`CALL deleteAccountAndAllAssociatedMemberships(?)`, castId)
	PanicIf(err)
	_, err = s.pwds.Exec(`DELETE FROM pwds WHERE id = ?`, castId)
	PanicIf(err)
}

func (s *sqlStore) getAccount(id Id) *account {
	row := s.accounts.QueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?`, []byte(id))
	a := account{}
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)) {
		return nil
	}
	return &a
}

func (s *sqlStore) getAccounts(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	castedIds = append(castedIds, []byte(ids[0]))
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (?`)
	for _, id := range ids[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`)`)
	rows, err := s.accounts.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	return res
}

func (s *sqlStore) searchAccounts(nameOrDisplayNameStartsWith string) []*account {
	searchTerm := nameOrDisplayNameStartsWith + "%"
	rows, err := s.accounts.Query(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isPersonal FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal))
		res = append(res, &acc)
	}
	return res
}

func (s *sqlStore) searchPersonalAccounts(nameOrDisplayNameOrEmailStartsWith string) []*account {
	searchTerm := nameOrDisplayNameOrEmailStartsWith + "%"
	rows, err := s.accounts.Query(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func (s *sqlStore) getPersonalAccounts(ids []Id) []*account {
	castedIds := make([]interface{}, 0, len(ids))
	castedIds = append(castedIds, []byte(ids[0]))
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE activatedOn IS NOT NULL AND id IN (?`)
	for _, id := range ids[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(id))
	}
	query.WriteString(`)`)
	rows, err := s.accounts.Query(query.String(), castedIds...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func (s *sqlStore) createGroupAccountAndMembership(account *account, memberId Id) {
	_, err := s.accounts.Exec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(account.Id), account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, []byte(memberId))
	PanicIf(err)
}

func (s *sqlStore) getGroupAccounts(memberId Id, after *Id, limit int) ([]*account, bool) {
	args := make([]interface{}, 0, 3)
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (SELECT account FROM memberships WHERE member = ?)`)
	args = append(args, []byte(memberId))
	if after != nil {
		query.WriteString(` AND name > (SELECT name FROM accounts WHERE id = ?)`)
		args = append(args, []byte(*after))
	}
	query.WriteString(` ORDER BY name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, err := s.accounts.Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*account, 0, limit+1)
	for rows.Next() {
		a := account{}
		PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func (s *sqlStore) createMemberships(accountId Id, members []Id) {
	args := make([]interface{}, 0, len(members)*2)
	args = append(args, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`INSERT INTO memberships (account, member) VALUES (?,?)`)
	for _, member := range members[1:] {
		query.WriteString(`,(?,?)`)
		args = append(args, []byte(accountId), []byte(member))
	}
	_, err := s.accounts.Exec(query.String(), args...)
	PanicIf(err)
}

func (s *sqlStore) deleteMemberships(accountId Id, members []Id) {
	castedIds := make([]interface{}, 0, len(members)+1)
	castedIds = append(castedIds, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`DELETE FROM memberships WHERE account=? AND member IN (?`)
	for _, member := range members[1:] {
		query.WriteString(`,?`)
		castedIds = append(castedIds, []byte(member))
	}
	query.WriteString(`)`)
	_, err := s.accounts.Exec(query.String(), castedIds...)
	PanicIf(err)
}
