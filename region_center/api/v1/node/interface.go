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
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}
