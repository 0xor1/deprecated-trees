package util

type PrivateRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, myId Id, myName string, myDisplayName *string) int
	DeleteAccount(region string, shard int, account, myId Id)
	AddMembers(region string, shard int, account, myId Id, members []*AddMemberPrivate)
	RemoveMembers(region string, shard int, account, myId Id, members []Id)
	MemberIsOnlyAccountOwner(region string, shard int, account, myId Id) bool
	SetMemberName(region string, shard int, account, myId Id, newName string)
	SetMemberDisplayName(region string, shard int, account, myId Id, newDisplayName *string)
	MemberIsAccountOwner(region string, shard int, account, myId Id) bool
}
