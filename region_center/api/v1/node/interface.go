package node

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
)

type Api interface {
	CreateAbstractNode(shard int, accountId, projectId, parentId, firstChildId, myId Id, name, description string, isParallel bool) *abstractNode
	CreateNode(shard int, accountId, projectId, parentId, firstChildId, myId Id, name, description string, memberId Id) *node
	SetName(shard int, accountId, projectId, nodeId, myId Id, name string)
	SetDescription(shard int, accountId, projectId, nodeId, myId Id, description string)
	SetIsParallel(shard int, accountId, projectId, nodeId, myId Id, isParallel bool)
	SetMember(shard int, accountId, projectId, nodeId, myId Id, MemberId *Id)
	SetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, timeRemaining uint64)
	MoveNode(shard int, accountId, projectId, nodeId, myId, parentId Id, nextSibling *Id)
	LogTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string)
	DeleteNode(shard int, accountId, projectId, nodeId, myId Id)
	GetNode(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) *node
	GetChildren(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int)
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}