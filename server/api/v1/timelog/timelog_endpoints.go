package timelog

import (
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/timelog"
	"bitbucket.org/0xor1/trees/server/util/validate"
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
	Path:            "/api/v1/timeLog/setDuration",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setDurationArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setDurationArgs)
		ctx.ReturnBadRequestNowIf(args.Duration == 0, "duration must be > 0")
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
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Task    *id.Id `json:"task,omitempty"`
	Member  *id.Id `json:"member,omitempty"`
	TimeLog *id.Id `json:"timeLog,omitempty"`
	SortAsc bool   `json:"sortAsc"`
	After   *id.Id `json:"after,omitempty"`
	Limit   int    `json:"limit"`
}

type getResp struct {
	TimeLogs []*timelog.TimeLog `json:"timeLogs"`
	More     bool               `json:"more"`
}

var get = &endpoint.Endpoint{
	Path:                     "/api/v1/timeLog/get",
	RequiresSession:          false,
	ExampleResponseStructure: &getResp{TimeLogs: []*timelog.TimeLog{{}}},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		return dbGetTimeLogs(ctx, args.Shard, args.Account, args.Project, args.Task, args.Member, args.TimeLog, args.SortAsc, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
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
