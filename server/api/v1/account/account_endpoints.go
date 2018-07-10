package account

import (
	"bitbucket.org/0xor1/trees/server/util/activity"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/validate"
	"time"
	"bitbucket.org/0xor1/trees/server/util/field"
	"bitbucket.org/0xor1/trees/server/util/account"
)

type editArgs struct {
	Shard                 int   `json:"shard"`
	Account               id.Id `json:"account"`
	Fields Fields  `json:"fields"`
}

var edit = &endpoint.Endpoint{
	Path:            "/api/v1/account/edit",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &editArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*editArgs)
		validate.MemberHasAccountOwnerAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		if args.Fields.HoursPerDay != nil {
			validate.HoursPerDay(args.Fields.HoursPerDay.Val)
		}
		if args.Fields.DaysPerWeek != nil {
			validate.DaysPerWeek(args.Fields.DaysPerWeek.Val)
		}
		dbEdit(ctx, args.Shard, args.Account, args.Fields)
		return nil
	},
}

type getArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
}

var get = &endpoint.Endpoint{
	Path:                     "/api/v1/account/get",
	RequiresSession:          true,
	ExampleResponseStructure: &account.Account{},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		return db.GetAccount(ctx, args.Shard, args.Account)
	},
}

type setMemberRoleArgs struct {
	Shard   int              `json:"shard"`
	Account id.Id            `json:"account"`
	Member  id.Id            `json:"member"`
	Role    cnst.AccountRole `json:"role"`
}

var setMemberRole = &endpoint.Endpoint{
	Path:            "/api/v1/account/setMemberRole",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		accountRole := db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me())
		validate.MemberHasAccountAdminAccess(accountRole)
		args.Role.Validate()
		ctx.ReturnUnauthorizedNowIf(args.Role == cnst.AccountOwner && *accountRole != cnst.AccountOwner)
		dbSetMemberRole(ctx, args.Shard, args.Account, args.Member, args.Role)
		return nil
	},
}

type getMembersArgs struct {
	Shard        int               `json:"shard"`
	Account      id.Id             `json:"account"`
	Role         *cnst.AccountRole `json:"role,omitempty"`
	NameContains *string           `json:"nameContains,omitempty"`
	After        *id.Id            `json:"after,omitempty"`
	Limit        int               `json:"limit"`
}

type GetMembersResp struct {
	Members []*Member `json:"members"`
	More    bool      `json:"more"`
}

var getMembers = &endpoint.Endpoint{
	Path:                     "/api/v1/account/getMembers",
	RequiresSession:          true,
	ExampleResponseStructure: &GetMembersResp{Members: []*Member{{}}},
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		return dbGetMembers(ctx, args.Shard, args.Account, args.Role, args.NameContains, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getActivitiesArgs struct {
	Shard          int        `json:"shard"`
	Account        id.Id      `json:"account"`
	Item           *id.Id     `json:"item,omitempty"`
	Member         *id.Id     `json:"member,omitempty"`
	OccurredAfter  *time.Time `json:"occurredAfter,omitempty"`
	OccurredBefore *time.Time `json:"occurredBefore,omitempty"`
	Limit          int        `json:"limit"`
}

var getActivities = &endpoint.Endpoint{
	Path:                     "/api/v1/account/getActivities",
	RequiresSession:          true,
	ExampleResponseStructure: []*activity.Activity{{}},
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		ctx.ReturnUnauthorizedNowIf(args.OccurredAfter != nil && args.OccurredBefore != nil)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		return dbGetActivities(ctx, args.Shard, args.Account, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
}

var getMe = &endpoint.Endpoint{
	Path:                     "/api/v1/account/getMe",
	RequiresSession:          true,
	ExampleResponseStructure: &Member{},
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.Account, ctx.Me())
	},
}

var Endpoints = []*endpoint.Endpoint{
	edit,
	get,
	setMemberRole,
	getMembers,
	getActivities,
	getMe,
}

type Fields struct {
	PublicProjectsEnabled *field.Bool `json:"publicProjectsEnabled,omitempty"`
	HoursPerDay *field.UInt8  `json:"hoursPerDay,omitempty"`
	DaysPerWeek *field.UInt8  `json:"daysPerWeek,omitempty"`
}

type Member struct {
	Id          id.Id            `json:"id"`
	Name        string           `json:"name"`
	DisplayName *string          `json:"displayName,omitempty"`
	HasAvatar   bool             `json:"hasAvatar"`
	Role        cnst.AccountRole `json:"role"`
	IsActive    bool             `json:"isActive"`
}
