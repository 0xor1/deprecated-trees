package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
)

type Api interface {
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, owner, account Id) (public error, private error)
	AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (public error, private error)
	RemoveMembers(region string, shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(region string, shard int, org, member Id) error
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserCanRenameOrg(region string, shard int, org, user Id) (bool, error)
}
