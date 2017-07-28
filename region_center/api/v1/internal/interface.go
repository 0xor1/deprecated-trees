package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
)

type Api interface {
	CreateAccount(accountId, myId Id, ownerName string) int
	DeleteAccount(shard int, accountId, myId Id)
	AddMembers(shard int, accountId, myId Id, members []*AddMemberInternal)
	RemoveMembers(shard int, accountId, myId Id, members []Id)
	MemberIsOnlyAccountOwner(shard int, accountId, memberId Id) bool
	RenameMember(shard int, accountId, memberId Id, newName string)
	MemberIsAccountOwner(shard int, accountId, memberId Id) bool
}

func NewApi(shards map[int]isql.ReplicaSet) Api {
	return &api{
		store: newSqlStore(shards),
	}
}

func NewClient(regions map[string]Api) InternalRegionClient {
	if regions == nil {
		panic(InvalidArgumentsErr)
	}
	return &client{
		regions: regions,
	}
}
