package misc

type InternalRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreatePersonalTaskCenter(region string, user Id) int
	CreateOrgTaskCenter(region string, org, owner Id, ownerName string) int
	DeleteTaskCenter(region string, shard int, account, owner Id)
	AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal)
	RemoveMembers(region string, shard int, org, admin Id, members []Id)
	SetMemberDeleted(region string, shard int, org, member Id)
	MemberIsOnlyOwner(region string, shard int, org, member Id) bool
	RenameMember(region string, shard int, org, member Id, newName string)
	UserIsOrgOwner(region string, shard int, org, user Id) bool
}
