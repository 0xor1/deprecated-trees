package task

import (
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	t "bitbucket.org/0xor1/trees/server/util/time"
	"bitbucket.org/0xor1/trees/server/util/validate"
	"github.com/0xor1/panic"
	"net/http"
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
	Method:                   http.MethodPost,
	Path:                     "/api/v1/task/create",
	RequiresSession:          true,
	ExampleResponseStructure: &Task{},
	GetArgsStruct: func() interface{} {
		return &createArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		panic.IfTrue((args.IsAbstract && (args.IsParallel == nil || args.Member != nil || args.TotalRemainingTime != nil)) || (!args.IsAbstract && (args.IsParallel != nil || args.TotalRemainingTime == nil)), err.InvalidArguments)
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

type setNameArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    id.Id  `json:"task"`
	Name    string `json:"name"`
}

var setName = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/setName",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setNameArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setNameArgs)
		if args.Project.Equal(args.Task) {
			validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		} else {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}

		dbSetName(ctx, args.Shard, args.Account, args.Project, args.Task, args.Name)
		return nil
	},
}

type setDescriptionArgs struct {
	Shard       int     `json:"shard"`
	Account     id.Id   `json:"account"`
	Project     id.Id   `json:"project"`
	Task        id.Id   `json:"task"`
	Description *string `json:"description,omitempty"`
}

var setDescription = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/setDescription",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setDescriptionArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setDescriptionArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbSetDescription(ctx, args.Shard, args.Account, args.Project, args.Task, args.Description)
		return nil
	},
}

type setIsParallelArgs struct {
	Shard      int   `json:"shard"`
	Account    id.Id `json:"account"`
	Project    id.Id `json:"project"`
	Task       id.Id `json:"task"`
	IsParallel bool  `json:"isParallel"`
}

var setIsParallel = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/setIsParallel",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setIsParallelArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setIsParallelArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbSetIsParallel(ctx, args.Shard, args.Account, args.Project, args.Task, args.IsParallel)
		return nil
	},
}

type setMemberArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    id.Id  `json:"task"`
	Member  *id.Id `json:"member,omitempty"`
}

var setMember = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/setMember",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMemberArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		if args.Member != nil {
			validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, args.Shard, args.Account, args.Project, *args.Member))
		}
		dbSetMember(ctx, args.Shard, args.Account, args.Project, args.Task, args.Member)
		return nil
	},
}

type setRemainingTimeArgs struct {
	Shard         int    `json:"shard"`
	Account       id.Id  `json:"account"`
	Project       id.Id  `json:"project"`
	Task          id.Id  `json:"task"`
	RemainingTime uint64 `json:"remainingTime"`
}

var setRemainingTime = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/setRemainingTime",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeArgs)
		return db.SetRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.RemainingTime, nil, nil)
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
	Method:          http.MethodPost,
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

var deleteTask = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/task/delete",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteArgs)
		panic.IfTrue(args.Project.Equal(args.Task), err.InvalidArguments)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbDeleteTask(ctx, args.Shard, args.Account, args.Project, args.Task)
		return nil
	},
}

type getArgs struct {
	Shard   int     `json:"shard"`
	Account id.Id   `json:"account"`
	Project id.Id   `json:"project"`
	Tasks   []id.Id `json:"tasks"`
}

var get = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/task/get",
	RequiresSession:          false,
	ExampleResponseStructure: []*Task{{}},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		validate.EntityCount(len(args.Tasks), ctx.MaxProcessEntityCount())
		return dbGetTasks(ctx, args.Shard, args.Account, args.Project, args.Tasks)
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

var getChildren = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/task/getChildren",
	RequiresSession:          false,
	ExampleResponseStructure: []*Task{{}},
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

type getAncestorTasksArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	Task    id.Id `json:"task"`
	Limit   int   `json:"limit"`
}

var getAncestorTasks = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/task/getAncestorTasks",
	RequiresSession:          false,
	ExampleResponseStructure: []*Ancestor{{}},
	GetArgsStruct: func() interface{} {
		return &getAncestorTasksArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getAncestorTasksArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetAncestorTasks(ctx, args.Shard, args.Account, args.Project, args.Task, args.Limit)
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	setName,
	setDescription,
	setIsParallel,
	setMember,
	setRemainingTime,
	move,
	deleteTask,
	get,
	getChildren,
	getAncestorTasks,
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
