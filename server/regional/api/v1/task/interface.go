package task

import (
	. "bitbucket.org/0xor1/task/server/util"
	"github.com/0xor1/isql"
)

type Api interface {
	CreateTask(shard int, accountId, projectId, parentId, myId Id, previousSibling *Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *Id, timeRemaining *uint64) *task
	SetName(shard int, accountId, projectId, taskId, myId Id, name string)
	SetDescription(shard int, accountId, projectId, taskId, myId Id, description *string)
	SetIsParallel(shard int, accountId, projectId, taskId, myId Id, isParallel bool)                                                           //only applys to abstract tasks
	SetMember(shard int, accountId, projectId, taskId, myId Id, memberId *Id)                                                                  //only applys to task tasks
	SetRemainingTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining uint64)                                                   //only applys to task tasks
	LogTime(shard int, accountId, projectId, taskId Id, myId Id, duration uint64, note *string) *timeLog                                       //only applys to task tasks
	SetRemainingTimeAndLogTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining uint64, duration uint64, note *string) *timeLog //only applys to task tasks
	MoveTask(shard int, accountId, projectId, taskId, myId, parentId Id, nextSibling *Id)
	DeleteTask(shard int, accountId, projectId, taskId, myId Id)
	GetTasks(shard int, accountId, projectId, myId Id, taskIds []Id) []*task
	GetChildTasks(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*task
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}
