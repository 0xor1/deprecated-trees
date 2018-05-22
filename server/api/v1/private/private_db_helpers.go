package private

import (
	"bitbucket.org/0xor1/trees/server/util/cachekey"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/private"
	"bitbucket.org/0xor1/trees/server/util/time"
	"bytes"
	"github.com/0xor1/panic"
	"math/rand"
)

func dbCreateAccount(ctx ctx.Ctx, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) int {
	shard := rand.Intn(ctx.TreeShardCount())
	_, e := ctx.TreeExec(shard, `CALL registerAccount(?, ?, ?, ?, ?)`, account, me, myName, myDisplayName, hasAvatar)
	panic.If(e)
	return shard
}

func dbDeleteAccount(ctx ctx.Ctx, shard int, account id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteAccount(?)`, account)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMaster(account))
}

func dbGetAllInactiveMembersFromInputSet(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) []id.Id {
	res := make([]id.Id, 0, len(members))
	cacheKey := cachekey.NewGet("private.dbGetAllInactiveMembersFromInputSet", shard, account, members).AccountMembers(account, members)
	if ctx.GetCacheValue(&res, cacheKey) {
		return res
	}
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, account, members[0])
	query := bytes.NewBufferString(`SELECT id FROM accountMembers WHERE account=? AND isActive=false AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	rows, e := ctx.TreeQuery(shard, query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	for rows.Next() {
		i := make([]byte, 0, 16)
		rows.Scan(&i)
		res = append(res, id.Id(i))
	}
	ctx.SetCacheValue(res, cacheKey)
	return res
}

func dbAddMembers(ctx ctx.Ctx, shard int, account id.Id, members []*private.AddMember) {
	queryArgs := make([]interface{}, 0, 3*len(members))
	queryArgs = append(queryArgs, account, members[0].Id, members[0].Name, members[0].DisplayName, members[0].HasAvatar, members[0].Role)
	query := bytes.NewBufferString(`INSERT INTO accountMembers (account, id, name, displayName, hasAvatar, role) VALUES (?,?,?,?,?,?)`)
	for _, mem := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?)`)
		queryArgs = append(queryArgs, account, mem.Id, mem.Name, mem.DisplayName, mem.HasAvatar, mem.Role)
	}
	_, e := ctx.TreeExec(shard, query.String(), queryArgs...)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMembersSet(account))
}

func dbUpdateMembersAndSetActive(ctx ctx.Ctx, shard int, account id.Id, members []*private.AddMember) {
	memberIds := make([]id.Id, 0, len(members))
	for _, mem := range members {
		memberIds = append(memberIds, mem.Id)
		_, e := ctx.TreeExec(shard, `CALL updateMembersAndSetActive(?, ?, ?, ?, ?, ?)`, account, mem.Id, mem.Name, mem.DisplayName, mem.HasAvatar, mem.Role)
		panic.If(e)
	}
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMembers(account, memberIds))
}

func dbGetTotalOwnerCount(ctx ctx.Ctx, shard int, account id.Id) int {
	count := 0
	cacheKey := cachekey.NewGet("private.dbGetTotalOwnerCount", shard, account).AccountMembersSet(account)
	if ctx.GetCacheValue(&count, cacheKey) {
		return count
	}
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, `SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, account).Scan(&count))
	ctx.SetCacheValue(count, cacheKey)
	return count
}

func dbGetOwnerCountInSet(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) int {
	count := 0
	cacheKey := cachekey.NewGet("private.dbGetOwnerCountInSet", shard, account, members).AccountMembers(account, members)
	if ctx.GetCacheValue(&count, cacheKey) {
		return count
	}
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, account, members[0])
	query := bytes.NewBufferString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, query.String(), queryArgs...).Scan(&count))
	ctx.SetCacheValue(count, cacheKey)
	return count
}

func dbSetMembersInactive(ctx ctx.Ctx, shard int, account id.Id, members []id.Id) {
	cacheKey := cachekey.NewSetDlms().AccountMembers(account, members)
	for _, mem := range members {
		rows, e := ctx.TreeQuery(shard, `CALL setAccountMemberInactive(?, ?)`, account, mem)
		panic.If(e)
		for rows.Next() {
			var project id.Id
			var task id.Id
			panic.If(rows.Scan(&project, task))
			cacheKey.Task(account, project, task).ProjectMembersSet(account, project).ProjectMember(account, project, mem)
		}
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetMemberName(ctx ctx.Ctx, shard int, account id.Id, member id.Id, newName string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberName(?, ?, ?)`, account, member, newName)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMember(account, member))
}

func dbSetMemberDisplayName(ctx ctx.Ctx, shard int, account, member id.Id, newDisplayName *string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberDisplayName(?, ?, ?)`, account, member, newDisplayName)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMember(account, member))
}

func dbSetMemberHasAvatar(ctx ctx.Ctx, shard int, account, member id.Id, hasAvatar bool) {
	_, e := ctx.TreeExec(shard, `UPDATE accountMembers SET hasAvatar=? WHERE account=? AND id=?`, hasAvatar, account, member)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMember(account, member))
}

func dbLogAccountBatchAddOrRemoveMembersActivity(ctx ctx.Ctx, shard int, account, member id.Id, members []id.Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	now := time.Now()
	args = append(args, account, now, member, members[0], "member", action, nil, nil)
	for _, mem := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, account, now, member, mem, "member", action, nil, nil)
	}
	_, e := ctx.TreeExec(shard, query.String(), args...)
	panic.If(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountActivities(account))
}
