package misc

type PrivateRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, myId Id, myName string, myDisplayName *string) int
	DeleteAccount(region string, shard int, account, myId Id)
	AddMembers(region string, shard int, account, myId Id, members []*AddMemberPrivate)
	RemoveMembers(region string, shard int, account, myId Id, members []Id)
	MemberIsOnlyAccountOwner(region string, shard int, account, myId Id) bool
	RenameMember(region string, shard int, account, myId Id, newName string)
	MemberIsAccountOwner(region string, shard int, account, myId Id) bool
}
