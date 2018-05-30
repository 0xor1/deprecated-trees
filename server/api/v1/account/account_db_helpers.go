package account

import (
	"bitbucket.org/0xor1/trees/server/util/activity"
	"bitbucket.org/0xor1/trees/server/util/cachekey"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bytes"
	"fmt"
	"github.com/0xor1/panic"
	"strings"
	"time"
)

func dbSetPublicProjectsEnabled(ctx ctx.Ctx, shard int, account id.Id, publicProjectsEnabled bool) {
	_, e := ctx.TreeExec(shard, `CALL setPublicProjectsEnabled(?, ?, ?)`, account, ctx.Me(), publicProjectsEnabled)
	panic.If(e)
	cacheKey := cachekey.NewSetDlms().Account(account).AccountActivities(account)
	if !publicProjectsEnabled { //if setting publicProjectsEnabled to false this could have set some projects to not public
		cacheKey.AccountProjectsSet(account)
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetMemberRole(ctx ctx.Ctx, shard int, account, member id.Id, role cnst.AccountRole) {
	db.MakeChangeHelper(ctx, shard, `CALL setAccountMemberRole(?, ?, ?, ?)`, account, ctx.Me(), member, role)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountMember(account, member).AccountActivities(account))
}

func dbGetMember(ctx ctx.Ctx, shard int, account, mem id.Id) *Member {
	res := Member{}
	cacheKey := cachekey.NewGet("account.dbGetMember", shard, account, mem).AccountMember(account, mem)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	row := ctx.TreeQueryRow(shard, `SELECT id, name, displayName, hasAvatar, isActive, role FROM accountMembers WHERE account=? AND id=?`, account, mem)
	panic.If(row.Scan(&res.Id, &res.Name, &res.DisplayName, &res.HasAvatar, &res.IsActive, &res.Role))
	ctx.SetCacheValue(res, cacheKey)
	return &res
}

/***
TODO need to determine which of these is most efficient on the db (this applys to project.GetMembers too):
1)
SELECT id, isActive, role
FROM accountMembers
WHERE account=:acc
AND isActive=true
AND (
        (
            name > (SELECT name FROM accountMembers WHERE account=:acc AND id=:id)
            AND role = (SELECT role FROM accountMembers WHERE account=:acc AND id=:id)
        )
        OR role > (SELECT role FROM accountMembers WHERE account=:acc AND id=:id)
)
ORDER BY role ASC, name ASC LIMIT :lim

2)
SELECT a1.id, a1.name, a1.displayName, a1.hasAvatar, a1.isActive, a1.role
FROM accountMembers a1, accountMembers a2
WHERE a1.account=:acc
AND a1.isActive=true
AND a2.account=:acc
AND a2.id=:id
AND (
        (
            a1.name>a2.name
            AND a1.role=a2.role
        )
        OR a1.role>a2.role
)
ORDER BY a1.role ASC, a1.name ASC LIMIT :lim
***/

func dbGetMembers(ctx ctx.Ctx, shard int, account id.Id, role *cnst.AccountRole, nameOrDisplayNameContains *string, after *id.Id, limit int) *GetMembersResp {
	res := GetMembersResp{}
	cacheKey := cachekey.NewGet("account.dbGetMembers", shard, account, role, nameOrDisplayNameContains, after, limit).AccountMembersSet(account)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	query := bytes.NewBufferString(`SELECT a1.id, a1.name, a1.displayName, a1.hasAvatar, a1.isActive, a1.role FROM accountMembers a1`)
	args := make([]interface{}, 0, 7)
	if after != nil {
		query.WriteString(`, accountMembers a2`)
	}
	query.WriteString(` WHERE a1.account=? AND a1.isActive=true`)
	args = append(args, account)
	if after != nil {
		query.WriteString(` AND a2.account=? AND a2.id=? AND ((a1.name>a2.name AND a1.role=a2.role) OR a1.role>a2.role)`)
		args = append(args, account, *after)
	}
	if role != nil {
		query.WriteString(` AND a1.role=?`)
		args = append(args, role)
	}
	if nameOrDisplayNameContains != nil {
		query.WriteString(` AND (a1.name LIKE ? OR a1.displayName LIKE ?)`)
		strVal := strings.Trim(*nameOrDisplayNameContains, " ")
		strVal = fmt.Sprintf("%%%s%%", strVal)
		args = append(args, strVal, strVal)
	}
	query.WriteString(` ORDER BY a1.role ASC, a1.name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	memSet := make([]*Member, 0, limit+1)
	for rows.Next() {
		mem := Member{}
		panic.If(rows.Scan(&mem.Id, &mem.Name, &mem.DisplayName, &mem.HasAvatar, &mem.IsActive, &mem.Role))
		memSet = append(memSet, &mem)
	}
	if len(memSet) == limit+1 {
		res.Members = memSet[:limit]
		res.More = true
	} else {
		res.Members = memSet
		res.More = false
	}
	ctx.SetCacheValue(&res, cacheKey)
	return &res
}

func dbGetActivities(ctx ctx.Ctx, shard int, account id.Id, item *id.Id, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) []*activity.Activity {
	panic.IfTrue(occurredAfter != nil && occurredBefore != nil, err.InvalidArguments)
	res := make([]*activity.Activity, 0, limit)
	cacheKey := cachekey.NewGet("account.dbGetActivities", shard, account, item, member, occurredAfter, occurredBefore, limit).AccountActivities(account)
	if ctx.GetCacheValue(&res, cacheKey) {
		return res
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, itemHasBeenDeleted, action, itemName, extraInfo FROM accountActivities WHERE account=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, account)
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, *item)
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, *member)
	}
	if occurredAfter != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfter)
	}
	if occurredBefore != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBefore)
	}
	if occurredAfter == nil && occurredBefore == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	for rows.Next() {
		act := activity.Activity{}
		panic.If(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.ItemHasBeenDeleted, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	ctx.SetCacheValue(res, cacheKey)
	return res
}
