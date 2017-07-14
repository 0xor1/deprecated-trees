package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

func newApi(store store, maxGetEntityCount int) Api {
	if store == nil {
		panic(NilOrInvalidCriticalParamErr)
	}
	return &api{
		store: store,
		maxGetEntityCount: maxGetEntityCount,
	}
}

type api struct {
	store store
	maxGetEntityCount int
}

func (a *api) SetPublicProjectsEnabled(shard int, orgId, myId Id, publicProjectsEnabled bool) {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner {
		panic(InsufficientPermissionErr)
	}
	a.store.setPublicProjectsEnabled(shard, orgId, publicProjectsEnabled)
}

func (a *api) GetPublicProjectsEnabled(shard int, orgId Id, myId Id) bool {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	return a.store.getPublicProjectsEnabled(shard, orgId)
}


func (a *api) GetMembers(shard int, orgId Id, myId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int) {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxGetEntityCount)
	return a.store.getMembers(shard, orgId, role, nameContains, offset, limit)
}

func (a *api) GetActivities(shard int, orgId Id, myId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity {
	actor := a.store.getMember(shard, orgId, myId)
	if actor.Role != OrgOwner && actor.Role != OrgAdmin {
		panic(InsufficientPermissionErr)
	}
	_, limit = ValidateOffsetAndLimitParams(0, limit, a.maxGetEntityCount)
	return a.store.getActivities(shard, orgId, item, member, occurredAfter, occurredBefore, limit)
}

func (a *api) GetMe(shard int, orgId Id, myId Id) *Member {
	return a.store.getMember(shard, orgId, myId)
}

type store interface {
	setPublicProjectsEnabled(shard int, orgId Id, publicProjectsEnabled bool)
	getPublicProjectsEnabled(shard int, orgId Id) bool
	getMember(shard int, org, member Id) *Member
	getMembers(shard int, org, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int)
	getActivities(shard int, orgId Id, item *Id, member *Id, occurredAfter *time.Time, occurredBefore *time.Time, limit int) []*Activity
}

type project struct {
	CommonNodeProps
	CommonAbstractNodeProps
	ArchivedOn *time.Time `json:"archivedOn,omitempty"`
	StartOn    *time.Time `json:"startOn,omitempty"`
	DueOn      *time.Time `json:"dueOn,omitempty"`
	FileCount  uint64     `json:"fileCount"`
	FileSize   uint64     `json:"fileSize"`
	IsPublic   bool       `json:"isPublic"`
}
