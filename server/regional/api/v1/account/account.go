package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bytes"
	"fmt"
	"strings"
	"time"
)

type setPublicProjectsEnabledArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	PublicProjectsEnabled bool `json:"publicProjectsEnabled"`
}

var setPublicProjectsEnabled = &Endpoint{
	Method: POST,
	Path: "/api/v1/account/setPublicProjectsEnabled",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setPublicProjectsEnabledArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setPublicProjectsEnabledArgs)
		ValidateMemberHasAccountOwnerAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		dbSetPublicProjectsEnabled(ctx, args.Shard, args.AccountId, ctx.MyId(), args.PublicProjectsEnabled)
		return nil
	},
}

type getPublicProjectsEnabledArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
}

var getPublicProjectsEnabled = &Endpoint{
	Method: GET,
	Path: "/api/v1/account/getPublicProjectsEnabled",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getPublicProjectsEnabledArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getPublicProjectsEnabledArgs)
		ValidateMemberHasAccountAdminAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		return dbGetPublicProjectsEnabled(ctx, args.Shard, args.AccountId)
	},
}

type setMemberRoleArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	MemberId Id `json:"memberId"`
	Role AccountRole `json:"role"`
}

var setMemberRole = &Endpoint{
	Method: POST,
	Path: "/api/v1/account/setMemberRole",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId())
		ValidateMemberHasAccountAdminAccess(accountRole)
		args.Role.Validate()
		if args.Role == AccountOwner && *accountRole != AccountOwner {
			InsufficientPermissionErr.Panic()
		}
		dbSetMemberRole(ctx, args.Shard, args.AccountId, ctx.MyId(), args.MemberId, args.Role)
		return nil
	},
}

type getMembersArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	Role *AccountRole `json:"role,omitempty"`
	NameContains *string `json:"nameContains,omitempty"`
	After *Id `json:"after,omitempty"`
	Limit int `json:"limit"`
}

type getMembersResp struct {
	Members []*member `json:"members"`
	More     bool       `json:"more"`
}

var getMembers = &Endpoint{
	Method: GET,
	Path: "/api/v1/account/getMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		ValidateMemberHasAccountAdminAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		return dbGetMembers(ctx, args.Shard, args.AccountId, args.Role, args.NameContains, args.After, ValidateLimit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getActivitiesArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	Item *Id `json:"item,omitempty"`
	Member *Id `json:"member,omitempty"`
	OccurredAfter *time.Time `json:"occurredAfter,omitempty"`
	OccurredBefore *time.Time `json:"occurredBefore,omitempty"`
	Limit int `json:"limit"`
}

var getActivities = &Endpoint{
	Method: GET,
	Path: "/api/v1/account/getActivities",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		if args.OccurredAfter != nil && args.OccurredBefore != nil {
			InvalidArgumentsErr.Panic()
		}
		ValidateMemberHasAccountAdminAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		return dbGetActivities(ctx, args.Shard, args.AccountId, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, ValidateLimit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
}

var getMe = &Endpoint{
	Method: GET,
	Path: "/api/v1/account/getMe",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.AccountId, ctx.MyId())
	},
}

var Endpoints = []*Endpoint{
	setPublicProjectsEnabled,
	getPublicProjectsEnabled,
	setMemberRole,
	getMembers,
	getActivities,
	getMe,
}

type Client interface {
	//must be account owner
	SetPublicProjectsEnabled(css *ClientSessionStore, shard int, accountId Id, publicProjectsEnabled bool) error
	//must be account owner/admin
	GetPublicProjectsEnabled(css *ClientSessionStore, shard int, accountId Id) (bool, error)
	//must be account owner/admin
	SetMemberRole(css *ClientSessionStore, shard int, accountId, memberId Id, role AccountRole) error
	//pointers are optional filters
	GetMembers(css *ClientSessionStore, shard int, accountId Id, role *AccountRole, nameContains *string, after *Id, limit int) (*getMembersResp, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *ClientSessionStore, shard int, accountId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*Activity, error)
	//for anyone
	GetMe(css *ClientSessionStore, shard int, accountId Id) (*member, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct{
	host string
}

func (c *client) SetPublicProjectsEnabled(css *ClientSessionStore, shard int, accountId Id, publicProjectsEnabled bool) error {
	_, err := setPublicProjectsEnabled.DoRequest(css, c.host, &setPublicProjectsEnabledArgs{
		Shard: shard,
		AccountId: accountId,
		PublicProjectsEnabled: publicProjectsEnabled,
	}, nil, nil)
	return err
}

func (c *client) GetPublicProjectsEnabled(css *ClientSessionStore, shard int, accountId Id) (bool, error) {
	respVal := true
	val, err := getPublicProjectsEnabled.DoRequest(css, c.host, &getPublicProjectsEnabledArgs{
		Shard: shard,
		AccountId: accountId,
	}, nil, &respVal)
	return *val.(*bool), err
}

func (c *client) SetMemberRole(css *ClientSessionStore, shard int, accountId, memberId Id, role AccountRole) error {
	_, err := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard: shard,
		AccountId: accountId,
		MemberId: memberId,
		Role: role,
	}, nil, nil)
	return err
}

func (c *client) GetMembers(css *ClientSessionStore, shard int, accountId Id, role *AccountRole, nameContains *string, after *Id, limit int) (*getMembersResp, error) {
	val, err := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard: shard,
		AccountId: accountId,
		Role: role,
		NameContains: nameContains,
		After: after,
		Limit: limit,
	}, nil, &getMembersResp{})
	return val.(*getMembersResp), err
}

func (c *client) GetActivities(css *ClientSessionStore, shard int, accountId Id, itemId *Id, memberId *Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*Activity, error) {
	val, err := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard: shard,
		AccountId: accountId,
		Item: itemId,
		Member: memberId,
		OccurredAfter: occurredAfter,
		OccurredBefore: occurredBefore,
		Limit: limit,
	}, nil, &[]*Activity{})
	return *val.(*[]*Activity), err
}

func (c *client) GetMe(css *ClientSessionStore, shard int, accountId Id) (*member, error) {
	val, err := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard: shard,
		AccountId: accountId,
	}, nil, &member{})
	return val.(*member), err
}

func dbGetAccountRole(ctx *Ctx, shard int, accountId, memberId Id) *AccountRole {
	return GetAccountRole(ctx, shard, accountId, memberId)
}

func dbSetPublicProjectsEnabled(ctx *Ctx, shard int, accountId, myId Id, publicProjectsEnabled bool) {
	_, err := ctx.TreeExec(shard, `CALL setPublicProjectsEnabled(?, ?, ?)`, []byte(accountId), []byte(ctx.MyId()), publicProjectsEnabled)
	PanicIf(err)
}

func dbGetPublicProjectsEnabled(ctx *Ctx, shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(ctx, shard, accountId)
}

func dbSetMemberRole(ctx *Ctx, shard int, accountId, myId, memberId Id, role AccountRole) {
	MakeChangeHelper(ctx, shard, `CALL setAccountMemberRole(?, ?, ?, ?)`, []byte(accountId), []byte(ctx.MyId()), []byte(memberId), role)
}

func dbGetMember(ctx *Ctx, shard int, accountId, memberId Id) *member {
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
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

func dbGetMembers(ctx *Ctx, shard int, accountId Id, role *AccountRole, nameOrDisplayNameContains *string, after *Id, limit int) *getMembersResp {
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
	rows, err := ctx.TreeQuery(shard, query.String(), args...)
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
		return &getMembersResp{Members: res[:limit], More: true}
	}
	return &getMembersResp{Members: res, More: false}
}

func dbGetActivities(ctx *Ctx, shard int, accountId Id, item *Id, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
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
	rows, err := ctx.TreeQuery(shard, query.String(), args...)
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
