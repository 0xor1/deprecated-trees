package org

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

type Api interface {
	//must be org owner/admin
	GetMembers(shard int, orgId Id, myId Id, roleFilter *OrgRole, nameSearch *string, offset, limit uint64) ([]*Member, int)
	GetActivities(shard int, orgId Id, myId Id, before) []*Activity
	//for anyone
	GetMe(shard int, orgId Id, myId Id) *Member
}
