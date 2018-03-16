package private

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
)

type V1Client interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error)
	DeleteAccount(region string, shard int, account, me id.Id) error
	AddMembers(region string, shard int, account, me id.Id, members []*AddMember) error
	RemoveMembers(region string, shard int, account, me id.Id, members []id.Id) error
	MemberIsOnlyAccountOwner(region string, shard int, account, me id.Id) (bool, error)
	SetMemberName(region string, shard int, account, me id.Id, newName string) error
	SetMemberDisplayName(region string, shard int, account, me id.Id, newDisplayName *string) error
	SetMemberHasAvatar(region string, shard int, account, me id.Id, hasAvatar bool) error
	MemberIsAccountOwner(region string, shard int, account, me id.Id) (bool, error)
}

type AddMember struct {
	Id          id.Id            `json:"id"`
	Name        string           `json:"name"`
	DisplayName *string          `json:"displayName"`
	HasAvatar 	bool          	 `json:"hasAvatar"`
	Role        cnst.AccountRole `json:"role"`
}
