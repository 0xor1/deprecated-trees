package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
)

type Api interface {
	//must be account owner
	SetPublicProjectsEnabled(shard int, accountId, myId Id, publicProjectsEnabled bool)
	//must be account owner/admin
	GetPublicProjectsEnabled(shard int, accountId, myId Id) bool
	//must be account owner/admin
	SetMemberRole(shard int, accountId, myId, memberId Id, role AccountRole)
	//pointers are optional filters
	GetMembers(shard int, accountId, myId Id, role *AccountRole, nameContains *string, offset, limit int) ([]*AccountMember, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, accountId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, accountId, myId Id) *AccountMember
}

func NewApi(shards map[int]isql.ReplicaSet, maxGetEntityCount int) Api {
	return &api{
		store:             newSqlStore(shards),
		maxGetEntityCount: maxGetEntityCount,
	}
}
