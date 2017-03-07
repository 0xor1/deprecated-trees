package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
)

const (
	Owner  = role("owner")
	Admin  = role("admin")
	Writer = role("writer")
	Reader = role("reader")
)

type role string

type InternalApi interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, account, owner Id) (public error, private error)
	AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (public error, private error)
	RemoveMembers(region string, shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(region string, shard int, org, member Id) error
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserCanRenameOrg(region string, shard int, org, user Id) (bool, error)
}