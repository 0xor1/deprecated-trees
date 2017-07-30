package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

type api struct {
	store             store
	maxGetEntityCount int
}

func (a *api) CreateProject(shard int, accountId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*addProjectMember) *project {
	me := a.store.getAccountMember(shard, accountId, myId)
	if me == nil || (me.Role != AccountOwner && me.Role != AccountAdmin) {
		panic(InsufficientPermissionErr)
	}

	project := &project{}
	project.Id = NewId()
	project.Name = name
	project.Description = description
	project.CreatedOn = Now()
	project.StartOn = startOn
	project.DueOn = dueOn
	project.IsParallel = isParallel
	project.IsPublic = isPublic
	a.store.createProject(shard, accountId, project)

	a.store.logAccountActivity(shard, accountId, Now(), myId, project.Id, "project", "created", nil)
	a.store.logProjectActivity(shard, accountId, project.Id, Now(), myId, project.Id, "project", "created", nil)

	a.AddMembers(shard, accountId, project.Id, myId, members)

	return project
}

func (a *api) SetName(shard int, accountId, projectId, myId Id, name string) {
	me := a.store.getAccountMember(shard, accountId, myId)
	if me == nil || (me.Role != AccountOwner && me.Role != AccountAdmin) {
		panic(InsufficientPermissionErr)
	}

	
}

func (a *api) SetDescription(shard int, accountId, projectId, myId Id, description string) {

}

func (a *api) SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {

}

func (a *api) SetIsParallel(shard int, accountId, projectId, myId Id, isParallel bool) {

}

func (a *api) GetProject(shard int, accountId, projectId, myId Id) ([]*project, int) {

}

func (a *api) GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isPublic *bool, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {

}

func (a *api) ArchiveProject(shard int, accountId, projectId, myId Id) {

}

func (a *api) UnarchiveProject(shard int, accountId, projectId, myId Id) {

}

func (a *api) DeleteProject(shard int, accountId, projectId, myId Id) {

}

func (a *api) AddMembers(shard int, accountId, projectId, myId Id, members []*addProjectMember) {

}

func (a *api) RemoveMembers(shard int, accountId, projectId, myId Id, members []Id) {

}

func (a *api) GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int) {

}

func (a *api) GetMe(shard int, accountId, projectId, myId Id) *projectMember {

}

func (a *api) GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {

}


type store interface {
	getAccountMember(shard int, accountId, memberId Id) *AccountMember
	createProject(shard int, accountId Id, project *project) //remember to create projectLocks db row
	logAccountActivity(shard int, accountId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string)
	logProjectActivity(shard int, accountId, projectId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string)
}

type addProjectMember struct{
	Entity
	Role ProjectRole `json:"role"`
}

type projectMember struct{
	Entity
	CommonTimeProps
	Role ProjectRole `json:"role"`
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
