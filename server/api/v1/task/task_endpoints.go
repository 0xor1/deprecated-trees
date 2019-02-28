package task

import (
	"github.com/0xor1/trees/server/util/ctx"
	"github.com/0xor1/trees/server/util/db"
	"github.com/0xor1/trees/server/util/endpoint"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/id"
	t "github.com/0xor1/trees/server/util/time"
	"github.com/0xor1/trees/server/util/validate"
	"time"
)

type createArgs struct {
	Shard              int     `json:"shard"`
	Account            id.Id   `json:"account"`
	Project            id.Id   `json:"project"`
	Parent             id.Id   `json:"parent"`
	PreviousSibling    *id.Id  `json:"previousSibling,omitempty"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	IsAbstract         bool    `json:"isAbstract"`
	IsParallel         *bool   `json:"isParallel,omitempty"`
	Member             *id.Id  `json:"member,omitempty"`
	TotalRemainingTime *uint64 `json:"totalRemainingTime,omitempty"`
}

var create = &endpoint.Endpoint{
	Path:                     "/api/v1/task/create",
	RequiresSession:          true,
	ExampleResponseStructure: &Task{},
	GetArgsStruct: func() interface{} {
		return &createArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		if args.IsAbstract {
			ctx.ReturnBadRequestNowIf(args.IsParallel == nil, "abstract tasks must have isParallel set")
			ctx.ReturnBadRequestNowIf(args.Member != nil, "abstract tasks do not accept a member arg")
			ctx.ReturnBadRequestNowIf(args.TotalRemainingTime != nil, "abstract tasks do not accept a totalRemainingTime arg")
		} else {
			ctx.ReturnBadRequestNowIf(args.IsParallel != nil, "concrete tasks do not accept an isParallel arg")
			ctx.ReturnBadRequestNowIf(args.TotalRemainingTime == nil, "concrete tasks must have a totalRemainingTime set")
		}
		zeroVal := uint64(0)
		zeroPtr := &zeroVal
		if !args.IsAbstract {
			zeroPtr = nil
		} else {
			args.TotalRemainingTime = zeroPtr
		}
		if args.Member != nil { //if a member is being assigned to the task then we need to check they have project write access
			validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, args.Shard, args.Account, args.Project, *args.Member))
		}
		newTask := &Task{
			Id:                   id.New(),
			IsAbstract:           args.IsAbstract,
			Name:                 args.Name,
			Description:          args.Description,
			CreatedOn:            t.Now(),
			TotalRemainingTime:   *args.TotalRemainingTime,
			MinimumRemainingTime: zeroPtr,
			ChildCount:           zeroPtr,
			DescendantCount:      zeroPtr,
			IsParallel:           args.IsParallel,
			Member:               args.Member,
		}
		dbCreateTask(ctx, args.Shard, args.Account, args.Project, args.Parent, args.PreviousSibling, newTask)
		return newTask
	},
}

type editArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    id.Id  `json:"task"`
	Fields  Fields `json:"fields"`
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

type moveArgs struct {
	Shard              int    `json:"shard"`
	Account            id.Id  `json:"account"`
	Project            id.Id  `json:"project"`
	Task               id.Id  `json:"task"`
	NewParent          id.Id  `json:"newParent"`
	NewPreviousSibling *id.Id `json:"newPreviousSibling"`
}

var move = &endpoint.Endpoint{
	Path:            "/api/v1/task/move",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &moveArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*moveArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbMoveTask(ctx, args.Shard, args.Account, args.Project, args.Task, args.NewParent, args.NewPreviousSibling)
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

type getArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	Task    id.Id `json:"task"`
}

var get = &endpoint.Endpoint{
	Path:                     "/api/v1/task/get",
	RequiresSession:          false,
	ExampleResponseStructure: &Task{},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		validate.EntityCount(len(args.Task), ctx.MaxProcessEntityCount())
		return dbGetTask(ctx, args.Shard, args.Account, args.Project, args.Task)
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

type GetChildrenResp struct {
	Children []*Task `json:"children"`
	More     bool    `json:"more"`
}

var getChildren = &endpoint.Endpoint{
	Path:                     "/api/v1/task/getChildren",
	RequiresSession:          false,
	ExampleResponseStructure: &GetChildrenResp{Children: []*Task{{}}},
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

type getAncestorsArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	Child   id.Id `json:"child"`
	Limit   int   `json:"limit"`
}

type GetAncestorsResp struct {
	Ancestors []*Ancestor `json:"ancestors"`
	More      bool        `json:"more"`
}

var getAncestors = &endpoint.Endpoint{
	Path:                     "/api/v1/task/getAncestors",
	RequiresSession:          false,
	ExampleResponseStructure: &GetAncestorsResp{Ancestors: []*Ancestor{{}}},
	GetArgsStruct: func() interface{} {
		return &getAncestorsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getAncestorsArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetAncestorTasks(ctx, args.Shard, args.Account, args.Project, args.Child, args.Limit)
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	edit,
	move,
	delete,
	get,
	getChildren,
	getAncestors,
}

type Task struct {
	Id                   id.Id     `json:"id"`
	Parent               *id.Id    `json:"parent,omitempty"`
	FirstChild           *id.Id    `json:"firstChild,omitempty"`
	NextSibling          *id.Id    `json:"nextSibling,omitempty"`
	IsAbstract           bool      `json:"isAbstract"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"` //only abstract tasks
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	ChildCount           *uint64   `json:"childCount,omitempty"`      //only abstract tasks
	DescendantCount      *uint64   `json:"descendantCount,omitempty"` //only abstract tasks
	IsParallel           *bool     `json:"isParallel,omitempty"`      //only abstract tasks
	Member               *id.Id    `json:"member,omitempty"`          //only task tasks
}

type Ancestor struct {
	Id   id.Id  `json:"id"`
	Name string `json:"name"`
	//may want to add on time values here to render progress bars within breadcrumb ui component
}

type Fields struct {
	Name          *field.String    `json:"name,omitempty"`
	Description   *field.StringPtr `json:"description,omitempty"`
	IsAbstract    *field.Bool      `json:"isAbstract,omitempty"`    //limit to only editable on abstract tasks which have no children and concrete tasks which have no timelogs
	IsParallel    *field.Bool      `json:"isParallel,omitempty"`    //only relevant to abstract tasks
	Member        *field.IdPtr     `json:"member,omitempty"`        //only relevant to concrete tasks
	RemainingTime *field.UInt64    `json:"remainingTime,omitempty"` //only relevant to concrete tasks
}
