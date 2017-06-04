package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
)

type InternalApiClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) (int, error)
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(region string, shard int, account, owner Id) (public error, private error)
	AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal) (public error, private error)
	RemoveMembers(region string, shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(region string, shard int, org, member Id) error
	MemberIsOnlyOwner(region string, shard int, org, member Id) (bool, error)
	RenameMember(region string, shard int, org, member Id, newName string) error
	UserIsOrgOwner(region string, shard int, org, user Id) (bool, error)
}

type InternalApi interface {
	CreatePersonalTaskCenter(user Id) (int, error)
	CreateOrgTaskCenter(org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(shard int, account, owner Id) (public error, private error)
	AddMembers(shard int, org, admin Id, members []*AddMemberInternal) (public error, private error)
	RemoveMembers(shard int, org, admin Id, members []Id) (public error, private error)
	SetMemberDeleted(shard int, org, member Id) error
	MemberIsOnlyOwner(shard int, org, member Id) (bool, error)
	RenameMember(shard int, org, member Id, newName string) error
	UserIsOrgOwner(shard int, org, user Id) (bool, error)
}