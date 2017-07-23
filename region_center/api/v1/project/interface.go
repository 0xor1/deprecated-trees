package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
)

type Api interface {
	//must be org owner/admin
	CreateProject(shard int, orgId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel bool) *project
	//must be org owner/admin
	SetName(shard int, orgId, projectId, myId Id, name string)
	//must be org owner/admin
	SetDescription(shard int, orgId, projectId, myId Id, description string)
	//must be org owner/admin and org.publicProjectsEnabled must be true
	SetIsPublic(shard int, orgId, projectId, myId Id, isPublic bool)
	//must be org owner/admin or project admin/writer
	SetIsParallel(shard int, orgId, projectId, myId Id, isParallel bool)
	//must be org owner/admin or project admin
	AddMembers(shard int, orgId, projectId, myId Id, members []*addProjectMember)
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(shard int, orgId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, orgId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, orgId, projectId, myId Id) *projectMember
}

func NewApi(shards map[int]isql.ReplicaSet, maxGetEntityCount int) Api {
	return &api{
		store:             newSqlStore(shards),
		maxGetEntityCount: maxGetEntityCount,
	}
}
