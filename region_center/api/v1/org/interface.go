package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

type Api interface {
	//must be org owner/admin
	//
	GetMembers(shard int, orgId Id, myId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, orgId Id, myId Id, item *Id, member *Id, OccurredAfter *time.Time, OccurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, orgId Id, myId Id) *Member
}
