package private

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/time"
	"bytes"
	"math/rand"
)

func dbCreateAccount(ctx ctx.Ctx, account, me id.Id, myName string, myDisplayName *string) int {
	shard := rand.Intn(ctx.TreeShardCount())
	_, e := ctx.TreeExec(shard, `CALL registerAccount(?, ?, ?, ?)`, account, me, myName, myDisplayName)
	err.PanicIf(e)
	return shard
}

func dbDeleteAccount(ctx ctx.Ctx, shard int, account id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteAccount(?)`, account)
	err.PanicIf(e)
}

func dbGetAllInactiveMembersFromInputSet(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) []id.Id {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, account, members[0])
	query := bytes.NewBufferString(`SELECT id FROM accountMembers WHERE account=? AND isActive=false AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	res := make([]id.Id, 0, len(members))
	rows, e := ctx.TreeQuery(shard, query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		i := make([]byte, 0, 16)
		rows.Scan(&i)
		res = append(res, id.Id(i))
	}
	return res
}

func dbGetAccountRole(ctx ctx.Ctx, shard int, account, member id.Id) *cnst.AccountRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM accountMembers WHERE account=? AND id=?`, account, member)
	res := cnst.AccountRole(3)
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func dbAddMembers(ctx ctx.Ctx, shard int, account id.Id, members []*private.AddMember) {
	queryArgs := make([]interface{}, 0, 3*len(members))
	queryArgs = append(queryArgs, account, members[0].Id, members[0].Name, members[0].DisplayName, members[0].Role)
	query := bytes.NewBufferString(`INSERT INTO accountMembers (account, id, name, displayName, role) VALUES (?,?,?,?,?)`)
	for _, mem := range members[1:] {
		query.WriteString(`,(?,?,?,?,?)`)
		queryArgs = append(queryArgs, account, mem.Id, mem.Name, mem.DisplayName, mem.Role)
	}
	_, e := ctx.TreeExec(shard, query.String(), queryArgs...)
	err.PanicIf(e)
}

func dbUpdateMembersAndSetActive(ctx ctx.Ctx, shard int, account id.Id, members []*private.AddMember) {
	for _, mem := range members {
		_, e := ctx.TreeExec(shard, `CALL updateMembersAndSetActive(?, ?, ?, ?, ?)`, account, mem.Id, mem.Name, mem.DisplayName, mem.Role)
		err.PanicIf(e)
	}
}

func dbGetTotalOwnerCount(ctx ctx.Ctx, shard int, account id.Id) int {
	count := 0
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, `SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, account).Scan(&count))
	return count
}

func dbGetOwnerCountInSet(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, account, members[0])
	query := bytes.NewBufferString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	count := 0
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, query.String(), queryArgs...).Scan(&count))
	return count
}

func dbSetMembersInactive(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) {
	for _, mem := range members {
		_, e := ctx.TreeExec(shard, `CALL setAccountMemberInactive(?, ?)`, account, mem)
		err.PanicIf(e)
	}
}

func dbSetMemberName(ctx ctx.Ctx, shard int, account id.Id, member id.Id, newName string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberName(?, ?, ?)`, account, member, newName)
	err.PanicIf(e)
}

func dbSetMemberDisplayName(ctx ctx.Ctx, shard int, account, member id.Id, newDisplayName *string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberDisplayName(?, ?, ?)`, account, member, newDisplayName)
	err.PanicIf(e)
}

func dbLogAccountBatchAddOrRemoveMembersActivity(ctx ctx.Ctx, shard int, account, member id.Id, members []id.Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	now := time.Now()
	args = append(args, account, now, member, members[0], "member", action, nil, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, account, now, member, memId, "member", action, nil, nil)
	}
	_, e := ctx.TreeExec(shard, query.String(), args...)
	err.PanicIf(e)
}
