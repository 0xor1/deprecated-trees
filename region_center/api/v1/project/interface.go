package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
)

type Api interface {
	//must be account owner/admin
	CreateProject(shard int, accountId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*addProjectMember) *project
	//must be account owner/admin
	SetName(shard int, accountId, projectId, myId Id, name string)
	//must be account owner/admin
	SetDescription(shard int, accountId, projectId, myId Id, description string)
	//must be account owner/admin and account.publicProjectsEnabled must be true
	SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool)
	//must be account owner/admin or project admin/writer
	SetIsParallel(shard int, accountId, projectId, myId Id, isParallel bool)
	//must be account owner/admin or project admin
	AddMembers(shard int, accountId, projectId, myId Id, members []*addProjectMember)
	//must be account owner/admin or project admin
	RemoveMembers(shard int, accountId, projectId, myId Id, members []Id)
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
	//for anyone
	GetMe(shard int, accountId, projectId, myId Id) *projectMember
}

func NewApi(shards map[int]isql.ReplicaSet, maxGetEntityCount int) Api {
	return &api{
		store:             newSqlStore(shards),
		maxGetEntityCount: maxGetEntityCount,
	}
}
