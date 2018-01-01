package private

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

type Api interface {
	CreateAccount(accountId, myId Id, myName string, myDisplayName *string) int
	DeleteAccount(shard int, accountId, myId Id)
	AddMembers(shard int, accountId, myId Id, members []*AddMemberPrivate)
	RemoveMembers(shard int, accountId, myId Id, members []Id)
	MemberIsOnlyAccountOwner(shard int, accountId, memberId Id) bool
	SetMemberName(shard int, accountId, memberId Id, newName string)
	SetMemberDisplayName(shard int, accountId, memberId Id, newDisplayName *string)
	MemberIsAccountOwner(shard int, accountId, memberId Id) bool
}

func New(shards map[int]isql.ReplicaSet, maxProcessEntityCount int) Api {
	return &api{
		store: newSqlStore(shards),
		maxProcessEntityCount: maxProcessEntityCount,
	}
}

func NewClient(regions map[string]Api) PrivateRegionClient {
	if regions == nil {
		InvalidArgumentsErr.Panic()
	}
	return &client{
		regions: regions,
	}
}
