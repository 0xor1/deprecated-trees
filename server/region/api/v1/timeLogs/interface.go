package timeLogs

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

type Api interface {
	LogTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) *timeLog
	GetTimeLogs(shard int, accountId, projectId, nodeId, myId Id) []*timeLog
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}
