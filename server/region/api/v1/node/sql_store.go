package node

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		panic(InvalidArgumentsErr)
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}
