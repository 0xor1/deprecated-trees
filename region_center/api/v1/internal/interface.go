package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
)

type Api interface {
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, account Id) error
	AddMember(region string, shard int, org, member Id, memberName string) error
	RemoveMember(region string, shard int, org, member Id) error
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserCanRenameOrg(region string, shard int, org, user Id) (bool, error)
	UserCanMigrateOrg(region string, shard int, org, user Id) (bool, error)
	UserCanDeleteOrg(region string, shard int, org, user Id) (bool, error)
	UserCanManageMembers(region string, shard int, org, user Id) (bool, error)
}
