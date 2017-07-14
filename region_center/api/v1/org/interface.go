package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

type Api interface {
	//must be org owner
	SetPublicProjectsEnabled(shard int, orgId, myId Id, publicProjectsEnabled bool)
	//must be org owner/admin
	GetPublicProjectsEnabled(shard int, orgId, myId Id) bool
	//pointers are optional filters
	GetMembers(shard int, orgId, myId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, orgId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, orgId, myId Id) *Member
}

func NewApi(store store, maxGetEntityCount int) Api {
	if store == nil {
		panic(NilOrInvalidCriticalParamErr)
	}
	return &api{
		store:             store,
		maxGetEntityCount: maxGetEntityCount,
	}
}
