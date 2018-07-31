package chat

import (
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/id"
	t "bitbucket.org/0xor1/trees/server/util/time"
	"bitbucket.org/0xor1/trees/server/util/validate"
	"time"
)

type createArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    id.Id  `json:"task"`
	Content string `json:"content"`
}

var create = &endpoint.Endpoint{
	Path:                     "/api/v1/chat/create",
	RequiresSession:          true,
	ExampleResponseStructure: &Entry{},
	GetArgsStruct: func() interface{} {
		return &createArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createArgs)
		validate.MemberIsAProjectMemberWithWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		validate.StringArg("chat entry", args.Content, 1, 1000, nil)
		newEntry := &Entry{
			Id:        id.New(),
			CreatedOn: t.Now(),
			Member:    ctx.Me(),
		}
		dbCreateChatEntry(ctx, args.Shard, args.Account, args.Project, args.Task, newEntry)
		return newEntry
	},
}

type editArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    id.Id  `json:"task"`
	Entry   id.Id  `json:"entry"`
	Content string `json:"content"`
}

var edit = &endpoint.Endpoint{
	Path:            "/api/v1/task/edit",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &editArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*editArgs)
		if args.Fields.Name != nil && args.Project.Equal(args.Task) {
			validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		} else {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}
		if args.Fields.Name != nil {
			dbSetName(ctx, args.Shard, args.Account, args.Project, args.Task, args.Fields.Name.Val)
		}
		if args.Fields.Description != nil {
			dbSetDescription(ctx, args.Shard, args.Account, args.Project, args.Task, args.Fields.Description.Val)
		}
		if args.Fields.IsAbstract != nil { //must do isAbstract first before any other tree altering operations
			ctx.ReturnBadRequestNowIf(args.Project.Equal(args.Task), "can't toggle isAbstract on project task node")
			dbSetIsAbstract(ctx, args.Shard, args.Account, args.Project, args.Task, args.Fields.IsAbstract.Val)
		}
		if args.Fields.IsParallel != nil {
			dbSetIsParallel(ctx, args.Shard, args.Account, args.Project, args.Task, args.Fields.IsParallel.Val)
		}
		if args.Fields.Member != nil {
			if args.Fields.Member.Val != nil {
				validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, args.Shard, args.Account, args.Project, *args.Fields.Member.Val))
			}
			dbSetMember(ctx, args.Shard, args.Account, args.Project, args.Task, args.Fields.Member.Val)
		}
		if args.Fields.RemainingTime != nil {
			db.SetRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.Fields.RemainingTime.Val, nil, nil)
		}
		return nil
	},
}

type deleteArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	Task    id.Id `json:"task"`
}

var delete = &endpoint.Endpoint{
	Path:            "/api/v1/task/delete",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteArgs)
		ctx.ReturnBadRequestNowIf(args.Project.Equal(args.Task), "use project delete endpoint to delete the project node")
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbDeleteTask(ctx, args.Shard, args.Account, args.Project, args.Task)
		return nil
	},
}

type getChildrenArgs struct {
	Shard       int    `json:"shard"`
	Account     id.Id  `json:"account"`
	Project     id.Id  `json:"project"`
	Parent      id.Id  `json:"parent"`
	FromSibling *id.Id `json:"fromSibling,omitempty"`
	Limit       int    `json:"limit"`
}

type getChildrenResp struct {
	Children []*Task `json:"children"`
	More     bool    `json:"more"`
}

var getChildren = &endpoint.Endpoint{
	Path:                     "/api/v1/task/getChildren",
	RequiresSession:          false,
	ExampleResponseStructure: &getChildrenResp{Children: []*Task{{}}},
	GetArgsStruct: func() interface{} {
		return &getChildrenArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getChildrenArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetChildTasks(ctx, args.Shard, args.Account, args.Project, args.Parent, args.FromSibling, args.Limit)
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	edit,
	delete,
	get,
}

type Entry struct {
	Id        id.Id     `json:"id"`
	Content   string    `json:"content"`
	CreatedOn time.Time `json:"createdOn"`
	Member    id.Id     `json:"user"`
}
