package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"fmt"
	"time"
)

type api struct {
	store             store
	maxGetEntityCount int
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
	role.Validate()
	if (role == AccountOwner && *accountRole != AccountOwner) || *accountRole != AccountAdmin {
		panic(InsufficientPermissionErr)
	}
	a.store.setMemberRole(shard, accountId, memberId, role)
	a.store.logActivity(shard, accountId, myId, memberId, "member", "setRole", role.String())
}

func (a *api) GetMembers(shard int, accountId, myId Id, role *AccountRole, nameContains *string, offset, limit int) ([]*AccountMember, int) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxGetEntityCount)
	return a.store.getMembers(shard, accountId, role, nameContains, offset, limit)
}

func (a *api) GetActivities(shard int, accountId, myId Id, itemId *Id, memberId *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity {
	if occurredBefore != nil && occurredAfter != nil {
		panic(InvalidArgumentsErr)
	}
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	_, limit = ValidateOffsetAndLimitParams(0, limit, a.maxGetEntityCount)
	return a.store.getActivities(shard, accountId, itemId, memberId, occurredAfter, occurredBefore, limit)
}

func (a *api) GetMe(shard int, accountId Id, myId Id) *AccountMember {
	return a.store.getMember(shard, accountId, myId)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	setPublicProjectsEnabled(shard int, accountId Id, publicProjectsEnabled bool)
	getPublicProjectsEnabled(shard int, accountId Id) bool
	setMemberRole(shard int, accountId, memberId Id, role AccountRole)
	getMember(shard int, accountId, memberId Id) *AccountMember
	getMembers(shard int, accountId Id, role *AccountRole, nameContains *string, offset, limit int) ([]*AccountMember, int)
	getActivities(shard int, accountId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity
	logActivity(shard int, accountId Id, member, item Id, itemType, action string, newValue string)
}
