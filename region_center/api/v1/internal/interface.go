package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
)

type Api interface {
	CreateTaskCenter(org, owner Id, ownerName string) int
	DeleteTaskCenter(shard int, account, owner Id)
	AddMembers(shard int, org, admin Id, members []*AddMemberInternal)
	RemoveMembers(shard int, org, admin Id, members []Id)
	MemberIsOnlyOwner(shard int, org, member Id) bool
	RenameMember(shard int, org, member Id, newName string)
	UserIsOrgOwner(shard int, org, user Id) bool
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
