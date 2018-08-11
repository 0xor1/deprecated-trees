package private

import (
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
)

type V1Client interface {
	CreateAccount(region cnst.Region, account, me id.Id, myName string, myDisplayName *string, hasAvatar bool) (int, error)
	DeleteAccount(region cnst.Region, shard int, account, me id.Id) error
	AddMembers(region cnst.Region, shard int, account, me id.Id, members []*AddMember) error
	RemoveMembers(region cnst.Region, shard int, account, me id.Id, members []id.Id) error
	MemberIsOnlyAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error)
	SetMemberName(region cnst.Region, shard int, account, me id.Id, newName string) error
	SetMemberDisplayName(region cnst.Region, shard int, account, me id.Id, newDisplayName *string) error
	SetMemberHasAvatar(region cnst.Region, shard int, account, me id.Id, hasAvatar bool) error
	MemberIsAccountOwner(region cnst.Region, shard int, account, me id.Id) (bool, error)
}

type AddMember struct {
	Id          id.Id            `json:"id"`
	Name        string           `json:"name"`
	DisplayName *string          `json:"displayName"`
	HasAvatar   bool             `json:"hasAvatar"`
	Role        cnst.AccountRole `json:"role"`
}
