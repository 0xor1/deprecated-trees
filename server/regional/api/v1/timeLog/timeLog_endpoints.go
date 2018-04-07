package timeLog

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/timeLog"
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
	Method:                   cnst.POST,
	Path:                     "/api/v1/timeLog/create",
	RequiresSession:          true,
	ExampleResponseStructure: &timeLog.TimeLog{},
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
	Method:                   cnst.POST,
	Path:                     "/api/v1/timeLog/createAndSetRemainingTime",
	RequiresSession:          true,
	ExampleResponseStructure: &timeLog.TimeLog{},
	GetArgsStruct: func() interface{} {
		return &createAndSetRemainingTimeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createAndSetRemainingTimeArgs)
		return db.SetRemainingTimeAndOrLogTime(ctx, args.Shard, args.Account, args.Project, args.Task, &args.RemainingTime, &args.Duration, args.Note)
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	createAndSetRemainingTime,
}
