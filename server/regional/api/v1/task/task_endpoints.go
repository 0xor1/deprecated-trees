package task

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"time"
)

type createTaskArgs struct {
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

var createTask = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/createTask",
	GetArgsStruct: func() interface{} {
		return &createTaskArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		if (args.IsAbstract && (args.IsParallel == nil || args.Member != nil || args.TotalRemainingTime != nil)) || (!args.IsAbstract && (args.IsParallel != nil || args.TotalRemainingTime == nil)) {
			panic(err.InvalidArguments)
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
		newTask := &task{
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
	Method: cnst.POST,
	Path:   "/api/v1/project/setName",
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
	Method: cnst.POST,
	Path:   "/api/v1/project/setDescription",
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
	Method: cnst.POST,
	Path:   "/api/v1/project/setIsParallel",
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
	Method: cnst.POST,
	Path:   "/api/v1/project/setMember",
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
	Method: cnst.POST,
	Path:   "/api/v1/project/setRemainingTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.RemainingTime, nil, nil)
	},
}

type logTimeArgs struct {
	Shard    int     `json:"shard"`
	Account  id.Id   `json:"account"`
	Project  id.Id   `json:"project"`
	Task     id.Id   `json:"task"`
	Duration uint64  `json:"duration"`
	Note     *string `json:"note,omitempty"`
}

var logTime = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/logTime",
	GetArgsStruct: func() interface{} {
		return &logTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*logTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, nil, &args.Duration, args.Note)
	},
}

type setRemainingTimeAndLogTimeArgs struct {
	Shard         int     `json:"shard"`
	Account       id.Id   `json:"account"`
	Project       id.Id   `json:"project"`
	Task          id.Id   `json:"task"`
	RemainingTime uint64  `json:"remainingTime"`
	Duration      uint64  `json:"duration"`
	Note          *string `json:"note,omitempty"`
}

var setRemainingTimeAndLogTime = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setRemainingTimeAndLogTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeAndLogTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeAndLogTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.RemainingTime, &args.Duration, args.Note)
	},
}

func setRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, accountId, projectId, taskId id.Id, remainingTime *uint64, duration *uint64, note *string) *timeLog {
	if duration != nil {
		validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, shard, accountId, projectId, ctx.Me()))
	} else if remainingTime != nil {
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, shard, accountId, projectId, ctx.Me()))
	} else {
		panic(err.InvalidArguments)
	}

	loggedOn := t.Now()
	dbSetRemainingTimeAndOrLogTime(ctx, shard, accountId, projectId, taskId, remainingTime, &loggedOn, duration, note)
	if duration != nil {
		return &timeLog{
			Project:  projectId,
			Task:     taskId,
			Member:   ctx.Me(),
			LoggedOn: loggedOn,
			Duration: *duration,
			Note:     note,
		}
	}
	return nil
}

type moveTaskArgs struct {
	Shard              int    `json:"shard"`
	Account            id.Id  `json:"account"`
	Project            id.Id  `json:"project"`
	Task               id.Id  `json:"task"`
	NewParent          id.Id  `json:"newParent"`
	NewPreviousSibling *id.Id `json:"newPreviousSibling"`
}

var moveTask = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/moveTask",
	GetArgsStruct: func() interface{} {
		return &moveTaskArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*moveTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbMoveTask(ctx, args.Shard, args.Account, args.Project, args.Task, args.NewParent, args.NewPreviousSibling)
		return nil
	},
}

type deleteTaskArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	Task    id.Id `json:"task"`
}

var deleteTask = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/deleteTask",
	GetArgsStruct: func() interface{} {
		return &deleteTaskArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		dbDeleteTask(ctx, args.Shard, args.Account, args.Project, args.Task)
		return nil
	},
}

type getTasksArgs struct {
	Shard   int     `json:"shard"`
	Account id.Id   `json:"account"`
	Project id.Id   `json:"project"`
	Tasks   []id.Id `json:"tasks"`
}

var getTasks = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getTasks",
	GetArgsStruct: func() interface{} {
		return &getTasksArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getTasksArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		validate.EntityCount(len(args.Tasks), ctx.MaxProcessEntityCount())
		return dbGetTasks(ctx, args.Shard, args.Account, args.Project, args.Tasks)
	},
}

type getChildTasksArgs struct {
	Shard       int    `json:"shard"`
	Account     id.Id  `json:"account"`
	Project     id.Id  `json:"project"`
	Parent      id.Id  `json:"parent"`
	FromSibling *id.Id `json:"fromSibling,omitempty"`
	Limit       int    `json:"limit"`
}

var getChildTasks = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getChildTasks",
	GetArgsStruct: func() interface{} {
		return &getChildTasksArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getChildTasksArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetChildTasks(ctx, args.Shard, args.Account, args.Project, args.Parent, args.FromSibling, args.Limit)
	},
}

var Endpoints = []*endpoint.Endpoint{
	createTask,
	setName,
	setDescription,
	setIsParallel,
	setMember,
	setRemainingTime,
	logTime,
	setRemainingTimeAndLogTime,
	moveTask,
	deleteTask,
	getTasks,
	getChildTasks,
}

type task struct {
	Id                   id.Id     `json:"id"`
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

type timeLog struct {
	Project  id.Id     `json:"project"`
	Task     id.Id     `json:"task"`
	Member   id.Id     `json:"member"`
	LoggedOn time.Time `json:"loggedOn"`
	TaskName string    `json:"taskName"`
	Duration uint64    `json:"duration"`
	Note     *string   `json:"note"`
}
