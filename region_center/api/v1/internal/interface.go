package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

type Api interface {
	CreatePersonalTaskCenter(user Id) int
	CreateOrgTaskCenter(org, owner Id, ownerName string) int
	DeleteTaskCenter(shard int, account, owner Id)
	AddMembers(shard int, org, admin Id, members []*AddMemberInternal)
	RemoveMembers(shard int, org, admin Id, members []Id)
	MemberIsOnlyOwner(shard int, org, member Id) bool
	RenameMember(shard int, org, member Id, newName string)
	UserIsOrgOwner(shard int, org, user Id) bool
}
