package timelog

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/timelog"
	"bitbucket.org/0xor1/task/server/util/validate"
	"github.com/0xor1/panic"
	"net/http"
)

type createArgs struct {
	Shard    int     `json:"shard"`
	Account  id.Id   `json:"account"`
	Project  id.Id   `json:"project"`
	Task     id.Id   `json:"task"`
	Duration uint64  `json:"duration"`
	Note     *string `json:"note,omitempty"`
}

var create = &endpoint.Endpoint{
	Method:                   http.MethodPost,
	Path:                     "/api/v1/timeLog/create",
	RequiresSession:          true,
	ExampleResponseStructure: &timelog.TimeLog{},
	GetArgsStruct: func() interface{} {
		return &createArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createArgs)
		return db.SetRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, nil, &args.Duration, args.Note)
	},
}

type createAndSetRemainingTimeArgs struct {
	Shard         int     `json:"shard"`
	Account       id.Id   `json:"account"`
	Project       id.Id   `json:"project"`
	Task          id.Id   `json:"task"`
	RemainingTime uint64  `json:"remainingTime"`
	Duration      uint64  `json:"duration"`
	Note          *string `json:"note,omitempty"`
}

var createAndSetRemainingTime = &endpoint.Endpoint{
	Method:                   http.MethodPost,
	Path:                     "/api/v1/timeLog/createAndSetRemainingTime",
	RequiresSession:          true,
	ExampleResponseStructure: &timelog.TimeLog{},
	GetArgsStruct: func() interface{} {
		return &createAndSetRemainingTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createAndSetRemainingTimeArgs)
		return db.SetRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.RemainingTime, &args.Duration, args.Note)
	},
}

type setDurationArgs struct {
	Shard    int    `json:"shard"`
	Account  id.Id  `json:"account"`
	Project  id.Id  `json:"project"`
	TimeLog  id.Id  `json:"timeLog"`
	Duration uint64 `json:"duration"`
}

var setDuration = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/timeLog/setDuration",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setDurationArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setDurationArgs)
		panic.IfTrueWith(args.Duration == 0, err.InvalidArguments)
		tl := dbGetTimeLog(ctx, args.Shard, args.Account, args.Project, args.TimeLog)
		if args.Duration == tl.Duration {
			return nil
		}
		if tl.Member.Equal(ctx.Me()) {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		} else {
			validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}
		dbSetDuration(ctx, args.Shard, args.Account, args.Project, tl.Task, tl.Member, args.TimeLog, args.Duration)
		return nil
	},
}

type setNoteArgs struct {
	Shard   int     `json:"shard"`
	Account id.Id   `json:"account"`
	Project id.Id   `json:"project"`
	TimeLog id.Id   `json:"timeLog"`
	Note    *string `json:"note"`
}

var setNote = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/timeLog/setNote",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setNoteArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setNoteArgs)
		tl := dbGetTimeLog(ctx, args.Shard, args.Account, args.Project, args.TimeLog)
		if (args.Note == nil && tl.Note == nil) || (*args.Note == *tl.Note) {
			return nil
		}
		if tl.Member.Equal(ctx.Me()) {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		} else {
			validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}
		dbSetNote(ctx, args.Shard, args.Account, args.Project, tl.Task, tl.Member, args.TimeLog, args.Note)
		return nil
	},
}

type deleteArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
	TimeLog id.Id `json:"timeLog"`
}

var delete = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/timeLog/delete",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteArgs)
		tl := dbGetTimeLog(ctx, args.Shard, args.Account, args.Project, args.TimeLog)
		if tl.Member.Equal(ctx.Me()) {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		} else {
			validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}
		dbDelete(ctx, args.Shard, args.Account, args.Project, tl.Task, tl.Member, args.TimeLog)
		return nil
	},
}

type getArgs struct {
	Shard   int          `json:"shard"`
	Account id.Id        `json:"account"`
	Project id.Id        `json:"project"`
	Task    *id.Id       `json:"task,omitempty"`
	Member  *id.Id       `json:"member,omitempty"`
	TimeLog *id.Id       `json:"timeLog,omitempty"`
	SortDir cnst.SortDir `json:"sortDir"`
	After   *id.Id       `json:"after,omitempty"`
	Limit   int          `json:"limit"`
}

var get = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/timeLog/get",
	RequiresSession:          false,
	ExampleResponseStructure: []*timelog.TimeLog{{}},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		return dbGetTimeLogs(ctx, args.Shard, args.Account, args.Project, args.Task, args.Member, args.TimeLog, args.SortDir, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	createAndSetRemainingTime,
	setDuration,
	setNote,
	delete,
	get,
}
