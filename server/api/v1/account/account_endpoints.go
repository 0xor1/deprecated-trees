package account

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/validate"
	"github.com/0xor1/panic"
	"time"
	"net/http"
)

type setPublicProjectsEnabledArgs struct {
	Shard                 int   `json:"shard"`
	Account               id.Id `json:"account"`
	PublicProjectsEnabled bool  `json:"publicProjectsEnabled"`
}

var setPublicProjectsEnabled = &endpoint.Endpoint{
	Method:          http.MethodPost,
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
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
}

var getPublicProjectsEnabled = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/account/getPublicProjectsEnabled",
	RequiresSession:          true,
	ExampleResponseStructure: false,
	GetArgsStruct: func() interface{} {
		return &getPublicProjectsEnabledArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getPublicProjectsEnabledArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		return db.GetPublicProjectsEnabled(ctx, args.Shard, args.Account)
	},
}

type setMemberRoleArgs struct {
	Shard   int              `json:"shard"`
	Account id.Id            `json:"account"`
	Member  id.Id            `json:"member"`
	Role    cnst.AccountRole `json:"role"`
}

var setMemberRole = &endpoint.Endpoint{
	Method:          http.MethodPost,
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
		panic.IfTrueWith(args.Role == cnst.AccountOwner && *accountRole != cnst.AccountOwner, err.InsufficientPermission)
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
	Method:                   http.MethodGet,
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
	Method:                   http.MethodGet,
	Path:                     "/api/v1/account/getActivities",
	RequiresSession:          true,
	ExampleResponseStructure: []*activity.Activity{{}},
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		panic.IfTrueWith(args.OccurredAfter != nil && args.OccurredBefore != nil, err.InvalidArguments)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		return dbGetActivities(ctx, args.Shard, args.Account, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
}

var getMe = &endpoint.Endpoint{
	Method:                   http.MethodGet,
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
	setPublicProjectsEnabled,
	getPublicProjectsEnabled,
	setMemberRole,
	getMembers,
	getActivities,
	getMe,
}

type Member struct {
	Id          id.Id            `json:"id"`
	Name        string           `json:"name"`
	DisplayName *string          `json:"displayName,omitempty"`
	HasAvatar   bool             `json:"hasAvatar"`
	Role        cnst.AccountRole `json:"role"`
	IsActive    bool             `json:"isActive"`
}
