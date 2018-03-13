package account

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
	"fmt"
	"strings"
	"time"
)

type setPublicProjectsEnabledArgs struct {
	Shard                 int   `json:"shard"`
	Account               id.Id `json:"account"`
	PublicProjectsEnabled bool  `json:"publicProjectsEnabled"`
}

var setPublicProjectsEnabled = &endpoint.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/account/setPublicProjectsEnabled",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setPublicProjectsEnabledArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setPublicProjectsEnabledArgs)
		validate.MemberHasAccountOwnerAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		dbSetPublicProjectsEnabled(ctx, args.Shard, args.Account, args.PublicProjectsEnabled)
		return nil
	},
}

type getPublicProjectsEnabledArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
}

var getPublicProjectsEnabled = &endpoint.Endpoint{
	Method:          cnst.GET,
	Path:            "/api/v1/account/getPublicProjectsEnabled",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getPublicProjectsEnabledArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getPublicProjectsEnabledArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		return dbGetPublicProjectsEnabled(ctx, args.Shard, args.AccountId)
	},
}

type setMemberRoleArgs struct {
	Shard     int              `json:"shard"`
	AccountId id.Id            `json:"accountId"`
	MemberId  id.Id            `json:"memberId"`
	Role      cnst.AccountRole `json:"role"`
}

var setMemberRole = &endpoint.Endpoint{
	Method:          cnst.POST,
	Path:            "/api/v1/account/setMemberRole",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		accountRole := db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me())
		validate.MemberHasAccountAdminAccess(accountRole)
		args.Role.Validate()
		if args.Role == cnst.AccountOwner && *accountRole != cnst.AccountOwner {
			panic(err.InsufficientPermission)
		}
		dbSetMemberRole(ctx, args.Shard, args.AccountId, args.MemberId, args.Role)
		return nil
	},
}

type getMembersArgs struct {
	Shard        int               `json:"shard"`
	AccountId    id.Id             `json:"accountId"`
	Role         *cnst.AccountRole `json:"role,omitempty"`
	NameContains *string           `json:"nameContains,omitempty"`
	After        *id.Id            `json:"after,omitempty"`
	Limit        int               `json:"limit"`
}

type getMembersResp struct {
	Members []*member `json:"members"`
	More    bool      `json:"more"`
}

var getMembers = &endpoint.Endpoint{
	Method:          cnst.GET,
	Path:            "/api/v1/account/getMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		return dbGetMembers(ctx, args.Shard, args.AccountId, args.Role, args.NameContains, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getActivitiesArgs struct {
	Shard          int        `json:"shard"`
	AccountId      id.Id      `json:"accountId"`
	Item           *id.Id     `json:"item,omitempty"`
	Member         *id.Id     `json:"member,omitempty"`
	OccurredAfter  *time.Time `json:"occurredAfter,omitempty"`
	OccurredBefore *time.Time `json:"occurredBefore,omitempty"`
	Limit          int        `json:"limit"`
}

var getActivities = &endpoint.Endpoint{
	Method:          cnst.GET,
	Path:            "/api/v1/account/getActivities",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		if args.OccurredAfter != nil && args.OccurredBefore != nil {
			panic(err.InvalidArguments)
		}
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		return dbGetActivities(ctx, args.Shard, args.AccountId, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
}

var getMe = &endpoint.Endpoint{
	Method:          cnst.GET,
	Path:            "/api/v1/account/getMe",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.AccountId, ctx.Me())
	},
}

var Endpoints = []*endpoint.Endpoint{
	setPublicProjectsEnabled,
	getPublicProjectsEnabled,
	setMemberRole,
	getMembers,
	getActivities,
	getMe,
}

type Client interface {
	//must be account owner
	SetPublicProjectsEnabled(css *clientsession.Store, shard int, accountId id.Id, publicProjectsEnabled bool) error
	//must be account owner/admin
	GetPublicProjectsEnabled(css *clientsession.Store, shard int, accountId id.Id) (bool, error)
	//must be account owner/admin
	SetMemberRole(css *clientsession.Store, shard int, accountId, memberId id.Id, role cnst.AccountRole) error
	//pointers are optional filters
	GetMembers(css *clientsession.Store, shard int, accountId id.Id, role *cnst.AccountRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, shard int, accountId id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
	//for anyone
	GetMe(css *clientsession.Store, shard int, accountId id.Id) (*member, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) SetPublicProjectsEnabled(css *clientsession.Store, shard int, accountId id.Id, publicProjectsEnabled bool) error {
	_, e := setPublicProjectsEnabled.DoRequest(css, c.host, &setPublicProjectsEnabledArgs{
		Shard:                 shard,
		Account:               accountId,
		PublicProjectsEnabled: publicProjectsEnabled,
	}, nil, nil)
	return e
}

func (c *client) GetPublicProjectsEnabled(css *clientsession.Store, shard int, accountId id.Id) (bool, error) {
	respVal := true
	val, e := getPublicProjectsEnabled.DoRequest(css, c.host, &getPublicProjectsEnabledArgs{
		Shard:     shard,
		AccountId: accountId,
	}, nil, &respVal)
	return *val.(*bool), e
}

func (c *client) SetMemberRole(css *clientsession.Store, shard int, accountId, memberId id.Id, role cnst.AccountRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard:     shard,
		AccountId: accountId,
		MemberId:  memberId,
		Role:      role,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, shard int, accountId id.Id, role *cnst.AccountRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error) {
	val, e := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard:        shard,
		AccountId:    accountId,
		Role:         role,
		NameContains: nameContains,
		After:        after,
		Limit:        limit,
	}, nil, &getMembersResp{})
	return val.(*getMembersResp), e
}

func (c *client) GetActivities(css *clientsession.Store, shard int, accountId id.Id, itemId *id.Id, memberId *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard:          shard,
		AccountId:      accountId,
		Item:           itemId,
		Member:         memberId,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		Limit:          limit,
	}, nil, &[]*activity.Activity{})
	return *val.(*[]*activity.Activity), e
}

func (c *client) GetMe(css *clientsession.Store, shard int, accountId id.Id) (*member, error) {
	val, e := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard:     shard,
		AccountId: accountId,
	}, nil, &member{})
	return val.(*member), e
}

func dbSetPublicProjectsEnabled(ctx ctx.Ctx, shard int, accountId id.Id, publicProjectsEnabled bool) {
	_, e := ctx.TreeExec(shard, `CALL setPublicProjectsEnabled(?, ?, ?)`, accountId, ctx.Me(), publicProjectsEnabled)
	err.PanicIf(e)
}

func dbGetPublicProjectsEnabled(ctx ctx.Ctx, shard int, accountId id.Id) bool {
	return db.GetPublicProjectsEnabled(ctx, shard, accountId)
}

func dbSetMemberRole(ctx ctx.Ctx, shard int, accountId, memberId id.Id, role cnst.AccountRole) {
	db.MakeChangeHelper(ctx, shard, `CALL setAccountMemberRole(?, ?, ?, ?)`, accountId, ctx.Me(), memberId, role)
}

func dbGetMember(ctx ctx.Ctx, shard int, accountId, memberId id.Id) *member {
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM accountMembers WHERE account=? AND id=?`, accountId, memberId)
	res := member{}
	err.PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
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

func dbGetMembers(ctx ctx.Ctx, shard int, accountId id.Id, role *cnst.AccountRole, nameOrDisplayNameContains *string, after *id.Id, limit int) *getMembersResp {
	query := bytes.NewBufferString(`SELECT a1.id, a1.isActive, a1.role FROM accountMembers a1`)
	args := make([]interface{}, 0, 7)
	if after != nil {
		query.WriteString(`, accountMembers a2`)
	}
	query.WriteString(` WHERE a1.account=? AND a1.isActive=true`)
	args = append(args, accountId)
	if after != nil {
		query.WriteString(` AND a2.account=? AND a2.id=? AND ((a1.name>a2.name AND a1.role=a2.role) OR a1.role>a2.role)`)
		args = append(args, accountId, *after)
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
	err.PanicIf(e)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		err.PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return &getMembersResp{Members: res[:limit], More: true}
	}
	return &getMembersResp{Members: res, More: false}
}

func dbGetActivities(ctx ctx.Ctx, shard int, accountId id.Id, item *id.Id, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) []*activity.Activity {
	if occurredAfter != nil && occurredBefore != nil {
		panic(err.InvalidArguments)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, extraInfo FROM accountActivities WHERE account=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, accountId)
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
	err.PanicIf(e)

	res := make([]*activity.Activity, 0, limit)
	for rows.Next() {
		act := activity.Activity{}
		err.PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	return res
}

type member struct {
	Id       id.Id            `json:"id"`
	Role     cnst.AccountRole `json:"role"`
	IsActive bool             `json:"isActive"`
}
