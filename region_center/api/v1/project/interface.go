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
	//check project access permission per user
	GetProject(shard int, accountId, projectId, myId Id) ([]*project, int)
	//check project access permission per user
	GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int)
	//must be account owner/admin
	ArchiveProject(shard int, accountId, projectId, myId Id)
	//must be account owner/admin
	UnarchiveProject(shard int, accountId, projectId, myId Id)
	//must be account owner/admin
	DeleteProject(shard int, accountId, projectId, myId Id)
	//must be account owner/admin or project admin
	AddMembers(shard int, accountId, projectId, myId Id, members []*addProjectMember)
	//must be account owner/admin or project admin
	RemoveMembers(shard int, accountId, projectId, myId Id, members []Id)
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int)
	//for anyone
	GetMe(shard int, accountId, projectId, myId Id) *projectMember
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
}

func NewApi(shards map[int]isql.ReplicaSet, maxGetEntityCount int) Api {
	return &api{
		store:             newSqlStore(shards),
		maxGetEntityCount: maxGetEntityCount,
	}
}
