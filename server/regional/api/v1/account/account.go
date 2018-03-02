package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Api interface {
	//must be account owner
	SetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId Id, publicProjectsEnabled bool)
	//must be account owner/admin
	GetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId Id) bool
	//must be account owner/admin
	SetMemberRole(ctx RegionalCtx, shard int, accountId, memberId Id, role AccountRole)
	//pointers are optional filters
	GetMembers(ctx RegionalCtx, shard int, accountId Id, role *AccountRole, nameContains *string, after *Id, limit int) ([]*member, bool)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(ctx RegionalCtx, shard int, accountId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(ctx RegionalCtx, shard int, accountId Id) *member
}

func New() Api {
	return &api{}
}

type api struct{}

func (a *api) SetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId Id, publicProjectsEnabled bool) {
	ctx.Validate().MemberHasAccountOwnerAccess(dbGetAccountRole(ctx, shard, accountId, ctx.MyId()))
	dbSetPublicProjectsEnabled(ctx, shard, accountId, ctx.MyId(), publicProjectsEnabled)
}

func (a *api) GetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId Id) bool {
	ctx.Validate().MemberHasAccountAdminAccess(dbGetAccountRole(ctx, shard, accountId, ctx.MyId()))
	return dbGetPublicProjectsEnabled(ctx, shard, accountId)
}

func (a *api) SetMemberRole(ctx RegionalCtx, shard int, accountId, memberId Id, role AccountRole) {
	accountRole := dbGetAccountRole(ctx, shard, accountId, ctx.MyId())
	ctx.Validate().MemberHasAccountAdminAccess(accountRole)
	role.Validate()
	if role == AccountOwner && *accountRole != AccountOwner {
		InsufficientPermissionErr.Panic()
	}
	dbSetMemberRole(ctx, shard, accountId, ctx.MyId(), memberId, role)
}

func (a *api) GetMembers(ctx RegionalCtx, shard int, accountId Id, role *AccountRole, nameContains *string, after *Id, limit int) ([]*member, bool) {
	ctx.Validate().MemberHasAccountAdminAccess(dbGetAccountRole(ctx, shard, accountId, ctx.MyId()))
	return dbGetMembers(ctx, shard, accountId, role, nameContains, after, ctx.Validate().Limit(limit))
}

func (a *api) GetActivities(ctx RegionalCtx, shard int, accountId Id, itemId *Id, memberId *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
	if occurredAfter != nil && occurredBefore != nil {
		InvalidArgumentsErr.Panic()
	}
	ctx.Validate().MemberHasAccountAdminAccess(dbGetAccountRole(ctx, shard, accountId, ctx.MyId()))
	return dbGetActivities(ctx, shard, accountId, itemId, memberId, occurredAfter, occurredBefore, ctx.Validate().Limit(limit))
}

func (a *api) GetMe(ctx RegionalCtx, shard int, accountId Id) *member {
	return dbGetMember(ctx, shard, accountId, ctx.MyId())
}

func dbGetAccountRole(ctx RegionalCtx, shard int, accountId, memberId Id) *AccountRole {
	return ctx.Db().GetAccountRole(shard, accountId, memberId)
}

func dbSetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId, myId Id, publicProjectsEnabled bool) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL setPublicProjectsEnabled(?, ?, ?)`, []byte(accountId), []byte(ctx.MyId()), publicProjectsEnabled)
	PanicIf(err)
}

func dbGetPublicProjectsEnabled(ctx RegionalCtx, shard int, accountId Id) bool {
	return ctx.Db().GetPublicProjectsEnabled(shard, accountId)
}

func dbSetMemberRole(ctx RegionalCtx, shard int, accountId, myId, memberId Id, role AccountRole) {
	ctx.Db().MakeChangeHelper(shard, `CALL setAccountMemberRole(?, ?, ?, ?)`, []byte(accountId), []byte(ctx.MyId()), []byte(memberId), role)
}

func dbGetMember(ctx RegionalCtx, shard int, accountId, memberId Id) *member {
	row := ctx.Db().Tree(shard).QueryRow(`SELECT id, isActive, role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := member{}
	PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
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
SELECT a1.id, a1.isActive, a1.role
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

func dbGetMembers(ctx RegionalCtx, shard int, accountId Id, role *AccountRole, nameOrDisplayNameContains *string, after *Id, limit int) ([]*member, bool) {
	query := bytes.NewBufferString(`SELECT a1.id, a1.isActive, a1.role FROM accountMembers a1`)
	args := make([]interface{}, 0, 7)
	if after != nil {
		query.WriteString(`, accountMembers a2`)
	}
	query.WriteString(` WHERE a1.account=? AND a1.isActive=true`)
	args = append(args, []byte(accountId))
	if after != nil {
		query.WriteString(` AND a2.account=? AND a2.id=? AND ((a1.name>a2.name AND a1.role=a2.role) OR a1.role>a2.role)`)
		args = append(args, []byte(accountId), []byte(*after))
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
	rows, err := ctx.Db().Tree(shard).Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func dbGetActivities(ctx RegionalCtx, shard int, accountId Id, item *Id, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
	if occurredAfter != nil && occurredBefore != nil {
		InvalidArgumentsErr.Panic()
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, extraInfo FROM accountActivities WHERE account=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(accountId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
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
	rows, err := ctx.Db().Tree(shard).Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*Activity, 0, limit)
	for rows.Next() {
		act := Activity{}
		PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	return res
}

type member struct {
	Id       Id          `json:"id"`
	Role     AccountRole `json:"role"`
	IsActive bool        `json:"isActive"`
}
