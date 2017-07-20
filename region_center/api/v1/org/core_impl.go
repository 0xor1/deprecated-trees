package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

type api struct {
	store             store
	maxGetEntityCount int
}

func (a *api) SetPublicProjectsEnabled(shard int, orgId, myId Id, publicProjectsEnabled bool) {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner {
		panic(InsufficientPermissionErr)
	}
	a.store.setPublicProjectsEnabled(shard, orgId, publicProjectsEnabled)
}

func (a *api) GetPublicProjectsEnabled(shard int, orgId, myId Id) bool {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	return a.store.getPublicProjectsEnabled(shard, orgId)
}

func (a *api) SetUserRole(shard int, orgId, myId, userId Id, role OrgRole) {
	actor := a.store.getMember(shard, orgId, myId)
	role.Validate()
	if (role == OrgOwner && actor.Role != OrgOwner) || actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	a.store.setUserRole(shard, orgId, userId, role)
}

func (a *api) GetMembers(shard int, orgId, myId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*OrgMember, int) {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxGetEntityCount)
	return a.store.getMembers(shard, orgId, role, nameContains, offset, limit)
}

func (a *api) GetActivities(shard int, orgId, myId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity {
	if occurredBefore != nil && occurredAfter != nil {
		panic(InvalidArgumentsErr)
	}
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	_, limit = ValidateOffsetAndLimitParams(0, limit, a.maxGetEntityCount)
	return a.store.getActivities(shard, orgId, item, member, occurredAfter, occurredBefore, limit)
}

func (a *api) GetMe(shard int, orgId Id, myId Id) *OrgMember {
	return a.store.getMember(shard, orgId, myId)
}

type store interface {
	setPublicProjectsEnabled(shard int, orgId Id, publicProjectsEnabled bool)
	getPublicProjectsEnabled(shard int, orgId Id) bool
	setUserRole(shard int, orgId, userId Id, role OrgRole)
	getMember(shard int, org, member Id) *OrgMember
	getMembers(shard int, orgId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*OrgMember, int)
	getActivities(shard int, orgId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity
}
