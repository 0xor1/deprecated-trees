package centralaccount

import (
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bytes"
	"github.com/0xor1/isql"
	"strings"
)

func dbAccountWithCiNameExists(ctx ctx.Ctx, name string) bool {
	row := ctx.AccountQueryRow(`SELECT COUNT(*) FROM accounts WHERE name = ?`, name)
	count := 0
	err.PanicIf(row.Scan(&count))
	return count != 0
}

func dbGetAccountByCiName(ctx ctx.Ctx, name string) *account {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name = ?`, name)
	acc := account{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal)) {
		return nil
	}
	return &acc
}

func dbCreatePersonalAccount(ctx ctx.Ctx, account *fullPersonalAccountInfo, pwdInfo *pwdInfo) {
	_, e := ctx.AccountExec(`CALL createPersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.Email, account.Language, account.Theme, account.NewEmail, account.activationCode, account.activatedOn, account.newEmailConfirmationCode, account.resetPwdCode)
	err.PanicIf(e)
	_, e = ctx.PwdExec(`INSERT INTO pwds (id, salt, pwd, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)`, account.Id, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen)
	err.PanicIf(e)
}

func dbGetPersonalAccountByEmail(ctx ctx.Ctx, email string) *fullPersonalAccountInfo {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE email = ?`, email)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPersonalAccountById(ctx ctx.Ctx, id id.Id) *fullPersonalAccountInfo {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, email, language, theme, newEmail, activationCode, activatedOn, newEmailConfirmationCode, resetPwdCode FROM personalAccounts WHERE id = ?`, id)
	account := fullPersonalAccountInfo{}
	account.IsPersonal = true
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&account.Id, &account.Name, &account.DisplayName, &account.CreatedOn, &account.Region, &account.NewRegion, &account.Shard, &account.HasAvatar, &account.Email, &account.Language, &account.Theme, &account.NewEmail, &account.activationCode, &account.activatedOn, &account.newEmailConfirmationCode, &account.resetPwdCode)) {
		return nil
	}
	return &account
}

func dbGetPwdInfo(ctx ctx.Ctx, id id.Id) *pwdInfo {
	row := ctx.PwdQueryRow(`SELECT salt, pwd, n, r, p, keyLen FROM pwds WHERE id = ?`, id)
	pwd := pwdInfo{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&pwd.salt, &pwd.pwd, &pwd.n, &pwd.r, &pwd.p, &pwd.keyLen)) {
		return nil
	}
	return &pwd
}

func dbUpdatePersonalAccount(ctx ctx.Ctx, personalAccountInfo *fullPersonalAccountInfo) {
	_, e := ctx.AccountExec(`CALL updatePersonalAccount(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, personalAccountInfo.Id, personalAccountInfo.Name, personalAccountInfo.DisplayName, personalAccountInfo.CreatedOn, personalAccountInfo.Region, personalAccountInfo.NewRegion, personalAccountInfo.Shard, personalAccountInfo.HasAvatar, personalAccountInfo.Email, personalAccountInfo.Language, personalAccountInfo.Theme, personalAccountInfo.NewEmail, personalAccountInfo.activationCode, personalAccountInfo.activatedOn, personalAccountInfo.newEmailConfirmationCode, personalAccountInfo.resetPwdCode)
	err.PanicIf(e)
}

func dbUpdateAccount(ctx ctx.Ctx, account *account) {
	_, e := ctx.AccountExec(`CALL updateAccountInfo(?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, account.IsPersonal)
	err.PanicIf(e)
}

func dbUpdatePwdInfo(ctx ctx.Ctx, id id.Id, pwdInfo *pwdInfo) {
	_, e := ctx.PwdExec(`UPDATE pwds SET salt=?, pwd=?, n=?, r=?, p=?, keyLen=? WHERE id = ?`, pwdInfo.salt, pwdInfo.pwd, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen, id)
	err.PanicIf(e)
}

func dbDeleteAccountAndAllAssociatedMemberships(ctx ctx.Ctx, id id.Id) {
	_, e := ctx.AccountExec(`CALL deleteAccountAndAllAssociatedMemberships(?)`, id)
	err.PanicIf(e)
	_, e = ctx.PwdExec(`DELETE FROM pwds WHERE id = ?`, id)
	err.PanicIf(e)
}

func dbGetAccount(ctx ctx.Ctx, id id.Id) *account {
	row := ctx.AccountQueryRow(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id = ?`, id)
	a := account{}
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal)) {
		return nil
	}
	return &a
}

func dbGetAccounts(ctx ctx.Ctx, ids []id.Id) []*account {
	args := make([]interface{}, 0, len(ids))
	args = append(args, ids[0])
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (?`)
	for _, i := range ids[1:] {
		query.WriteString(`,?`)
		args = append(args, i)
	}
	query.WriteString(`)`)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		a := account{}
		err.PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	return res
}

func dbSearchAccounts(ctx ctx.Ctx, nameOrDisplayNameStartsWith string) []*account {
	searchTerm := nameOrDisplayNameStartsWith + "%"
	//rows, err := ctx.AccountQuery(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar, a.isPersonal FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	rows, e := ctx.AccountQuery(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE name LIKE ? OR displayName LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, 0, 100)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar, &acc.IsPersonal))
		res = append(res, &acc)
	}
	return res
}

func dbSearchPersonalAccounts(ctx ctx.Ctx, nameOrDisplayNameOrEmailStartsWith string) []*account {
	//rows, err := ctx.AccountQuery(`SELECT DISTINCT a.id, a.name, a.displayName, a.createdOn, a.region, a.newRegion, a.shard, a.hasAvatar FROM ((SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE displayName LIKE ? ORDER BY name ASC LIMIT ?, ?) UNION (SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?)) AS a ORDER BY name ASC LIMIT ?, ?`, searchTerm, 0, 100, searchTerm, 0, 100, searchTerm, 0, 100, 0, 100)
	//TODO need to profile these queries to check for best performance
	var rows isql.Rows
	var e error
	if strings.Contains(nameOrDisplayNameOrEmailStartsWith, "@") {
		rows, e = ctx.AccountQuery(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE email LIKE ? ORDER BY name ASC LIMIT ?, ?`, nameOrDisplayNameOrEmailStartsWith, 0, 100)
	} else {
		searchTerm := nameOrDisplayNameOrEmailStartsWith + "%"
		rows, e = ctx.AccountQuery(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE name LIKE ? OR displayName LIKE ? ORDER BY name ASC LIMIT ?, ?`, searchTerm, searchTerm, 0, 100)
	}
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*account, 0, 100)
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbGetPersonalAccounts(ctx ctx.Ctx, ids []id.Id) []*account {
	args := make([]interface{}, 0, len(ids))
	args = append(args, ids[0])
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar FROM personalAccounts WHERE activatedOn IS NOT NULL AND id IN (?`)
	for _, i := range ids[1:] {
		query.WriteString(`,?`)
		args = append(args, i)
	}
	query.WriteString(`)`)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, len(ids))
	for rows.Next() {
		acc := account{}
		acc.IsPersonal = true
		err.PanicIf(rows.Scan(&acc.Id, &acc.Name, &acc.DisplayName, &acc.CreatedOn, &acc.Region, &acc.NewRegion, &acc.Shard, &acc.HasAvatar))
		res = append(res, &acc)
	}
	return res
}

func dbCreateGroupAccountAndMembership(ctx ctx.Ctx, account *account, member id.Id) {
	_, e := ctx.AccountExec(`CALL  createGroupAccountAndMembership(?, ?, ?, ?, ?, ?, ?, ?, ?)`, account.Id, account.Name, account.DisplayName, account.CreatedOn, account.Region, account.NewRegion, account.Shard, account.HasAvatar, member)
	err.PanicIf(e)
}

func dbGetGroupAccounts(ctx ctx.Ctx, member id.Id, after *id.Id, limit int) ([]*account, bool) {
	args := make([]interface{}, 0, 3)
	query := bytes.NewBufferString(`SELECT id, name, displayName, createdOn, region, newRegion, shard, hasAvatar, isPersonal FROM accounts WHERE id IN (SELECT account FROM memberships WHERE member = ?)`)
	args = append(args, member)
	if after != nil {
		query.WriteString(` AND name > (SELECT name FROM accounts WHERE id = ?)`)
		args = append(args, *after)
	}
	query.WriteString(` ORDER BY name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, e := ctx.AccountQuery(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*account, 0, limit+1)
	for rows.Next() {
		a := account{}
		err.PanicIf(rows.Scan(&a.Id, &a.Name, &a.DisplayName, &a.CreatedOn, &a.Region, &a.NewRegion, &a.Shard, &a.HasAvatar, &a.IsPersonal))
		res = append(res, &a)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func dbCreateMemberships(ctx ctx.Ctx, account id.Id, members []id.Id) {
	args := make([]interface{}, 0, len(members)*2)
	args = append(args, account, members[0])
	query := bytes.NewBufferString(`INSERT INTO memberships (account, member) VALUES (?,?)`)
	for _, member := range members[1:] {
		query.WriteString(`,(?,?)`)
		args = append(args, account, member)
	}
	_, e := ctx.AccountExec(query.String(), args...)
	err.PanicIf(e)
}

func dbDeleteMemberships(ctx ctx.Ctx, account id.Id, members []id.Id) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, account, members[0])
	query := bytes.NewBufferString(`DELETE FROM memberships WHERE account=? AND member IN (?`)
	for _, member := range members[1:] {
		query.WriteString(`,?`)
		args = append(args, member)
	}
	query.WriteString(`)`)
	_, e := ctx.AccountExec(query.String(), args...)
	err.PanicIf(e)
}
