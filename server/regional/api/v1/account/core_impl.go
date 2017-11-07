package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"fmt"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) SetPublicProjectsEnabled(shard int, accountId, myId Id, publicProjectsEnabled bool) {
	ValidateMemberHasAccountOwnerAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setPublicProjectsEnabled(shard, accountId, publicProjectsEnabled)
	a.store.logActivity(shard, accountId, myId, accountId, "account", "setPublicProjectsEnabled", fmt.Sprintf("%t", publicProjectsEnabled))
}

func (a *api) GetPublicProjectsEnabled(shard int, accountId, myId Id) bool {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	return a.store.getPublicProjectsEnabled(shard, accountId)
}

func (a *api) SetMemberRole(shard int, accountId, myId, memberId Id, role AccountRole) {
	accountRole := a.store.getAccountRole(shard, accountId, myId)
	ValidateMemberHasAccountAdminAccess(accountRole)
	role.Validate()
	if role == AccountOwner && *accountRole != AccountOwner {
		InsufficientPermissionErr.Panic()
	}
	a.store.setMemberRole(shard, accountId, memberId, role)
	a.store.logActivity(shard, accountId, myId, memberId, "member", "setRole", role.String())
}

func (a *api) GetMembers(shard int, accountId, myId Id, role *AccountRole, nameContains *string, after *Id, limit int) ([]*member, bool) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	return a.store.getMembers(shard, accountId, role, nameContains, after, ValidateLimitParam(limit, a.maxProcessEntityCount))
}

func (a *api) GetActivities(shard int, accountId, myId Id, itemId *Id, memberId *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		InvalidArgumentsErr.Panic()
	}
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	return a.store.getActivities(shard, accountId, itemId, memberId, occurredAfterUnixMillis, occurredBeforeUnixMillis, ValidateLimitParam(limit, a.maxProcessEntityCount))
}

func (a *api) GetMe(shard int, accountId Id, myId Id) *member {
	return a.store.getMember(shard, accountId, myId)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	setPublicProjectsEnabled(shard int, accountId Id, publicProjectsEnabled bool)
	getPublicProjectsEnabled(shard int, accountId Id) bool
	setMemberRole(shard int, accountId, memberId Id, role AccountRole)
	getMember(shard int, accountId, memberId Id) *member
	getMembers(shard int, accountId Id, role *AccountRole, nameContains *string, after *Id, limit int) ([]*member, bool)
	getActivities(shard int, accountId Id, item *Id, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity
	logActivity(shard int, accountId Id, member, item Id, itemType, action string, newValue string)
}

type member struct {
	Id       Id          `json:"id"`
	Role     AccountRole `json:"role"`
	IsActive bool        `json:"isActive"`
}
