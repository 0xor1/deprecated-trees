package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

type Api interface {
	CreateNode(shard int, accountId, projectId, parentId, myId Id, name, description string, isAbstract bool, isParallel *bool, memberId *Id, timeRemaining *uint64) *node
	SetName(shard int, accountId, projectId, nodeId, myId Id, name string)
	SetDescription(shard int, accountId, projectId, nodeId, myId Id, description string)
	SetIsParallel(shard int, accountId, projectId, nodeId, myId Id, isParallel bool) //only applys to abstract nodes
	SetMember(shard int, accountId, projectId, nodeId, myId Id, memberId *Id) //only applys to task nodes
	SetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, timeRemaining uint64) //only applys to task nodes
	LogTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) //only applys to task nodes
	MoveNode(shard int, accountId, projectId, nodeId, myId, parentId Id, nextSibling *Id)
	DeleteNode(shard int, accountId, projectId, nodeId, myId Id)
	GetNodes(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*node
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}