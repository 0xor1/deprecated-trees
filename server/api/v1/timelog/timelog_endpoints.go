package timelog

import (
	"github.com/0xor1/trees/server/util/ctx"
	"github.com/0xor1/trees/server/util/db"
	"github.com/0xor1/trees/server/util/endpoint"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/timelog"
	"github.com/0xor1/trees/server/util/validate"
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

type editArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	TimeLog id.Id  `json:"timeLog"`
	Fields  Fields `json:"fields"`
}

var edit = &endpoint.Endpoint{
	Path:            "/api/v1/timeLog/edit",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &editArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*editArgs)
		ctx.ReturnBadRequestNowIf(args.Fields.Duration != nil && args.Fields.Duration.Val == 0, "duration must be > 0")
		tl := dbGetTimeLog(ctx, args.Shard, args.Account, args.Project, args.TimeLog)
		if tl.Member.Equal(ctx.Me()) {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		} else {
			validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		}
		if args.Fields.Duration != nil && args.Fields.Duration.Val != tl.Duration {
			dbSetDuration(ctx, args.Shard, args.Account, args.Project, tl.Task, ctx.Me(), tl.Id, args.Fields.Duration.Val)
		}
		if args.Fields.Note != nil && ((args.Fields.Note.Val == nil && tl.Note != nil) || (args.Fields.Note.Val != nil && tl.Note == nil) || (tl.Note != nil && args.Fields.Note.Val != nil && *tl.Note != *args.Fields.Note.Val)) {
			dbSetNote(ctx, args.Shard, args.Account, args.Project, tl.Task, ctx.Me(), tl.Id, args.Fields.Note.Val)
		}
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

type GetResp struct {
	TimeLogs []*timelog.TimeLog `json:"timeLogs"`
	More     bool               `json:"more"`
}

var get = &endpoint.Endpoint{
	Path:                     "/api/v1/timeLog/get",
	RequiresSession:          false,
	ExampleResponseStructure: &GetResp{TimeLogs: []*timelog.TimeLog{{}}},
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
	edit,
	delete,
	get,
}

type Fields struct {
	Duration *field.UInt64    `json:"duration,omitempty"`
	Note     *field.StringPtr `json:"note,omitempty"`
}
