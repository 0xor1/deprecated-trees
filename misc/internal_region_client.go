package misc

type InternalRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, owner Id, ownerName string) int
	DeleteAccount(region string, shard int, account, owner Id)
	AddMembers(region string, shard int, account, admin Id, members []*AddMemberInternal)
	RemoveMembers(region string, shard int, account, admin Id, members []Id)
	MemberIsOnlyAccountOwner(region string, shard int, account, member Id) bool
	RenameMember(region string, shard int, account, member Id, newName string)
	MemberIsAccountOwner(region string, shard int, account, member Id) bool
}
