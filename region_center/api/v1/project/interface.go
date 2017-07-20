package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
)

type Api interface {
	//must be org owner and org.publicProjectsEnabled must be true
	SetIsPublic(shard int, orgId, projId, myId Id, isPublic bool)
	//pointers are optional filters
	GetMembers(shard int, orgId, myId Id, role *OrgRole, nameContains *string, offset, limit int) ([]*Member, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, orgId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, orgId, projId, myId Id) *Member
}

func NewApi(shards map[int]isql.ReplicaSet, maxGetEntityCount int) Api {
	return &api{
		store:             newSqlStore(shards),
		maxGetEntityCount: maxGetEntityCount,
	}
}
