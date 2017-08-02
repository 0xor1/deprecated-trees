package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"fmt"
	"time"
)

var (
	publicProjectsDisabledErr = &Error{Code: "rc_v1_p_ppd", Msg: "public projects disabled", IsPublic: true}
)

type api struct {
	store             store
	maxGetEntityCount int
}

func (a *api) CreateProject(shard int, accountId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*addProjectMember) *project {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

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
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setName(shard, accountId, projectId, name)
	a.store.logProjectActivity(shard, accountId, projectId, Now(), myId, projectId, "project", "setName", &name)
}

func (a *api) SetDescription(shard int, accountId, projectId, myId Id, description string) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setDescription(shard, accountId, projectId, description)
	a.store.logProjectActivity(shard, accountId, projectId, Now(), myId, projectId, "project", "setDescription", &description)
}

func (a *api) SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	if a.store.getPublicProjectsEnabled(shard, accountId) {
		panic(publicProjectsDisabledErr)
	}

	a.store.setIsPublic(shard, accountId, projectId, isPublic)
	action := fmt.Sprintf("%t", isPublic)
	a.store.logAccountActivity(shard, accountId, Now(), myId, projectId, "project", "setIsPublic", &action)
	a.store.logProjectActivity(shard, accountId, projectId, Now(), myId, projectId, "project", "setIsPublic", &action)
}

func (a *api) SetIsParallel(shard int, accountId, projectId, myId Id, isParallel bool) {
	myAccountRole, myProjectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectWriteAccess(myAccountRole, myProjectRole)

	a.store.setIsParallel(shard, accountId, projectId, isParallel)
	action := fmt.Sprintf("%t", isParallel)
	a.store.logProjectActivity(shard, accountId, projectId, Now(), myId, projectId, "project", "setIsParallel", &action)
}

func (a *api) GetProject(shard int, accountId, projectId, myId Id) *project {
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)

	return a.store.getProject(shard, accountId, projectId)
}

func (a *api) GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	myAccountRole := a.store.getAccountRole(shard, accountId, myId)
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxGetEntityCount)
	if myAccountRole == nil {
		return a.store.getPublicProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
	}
	if *myAccountRole != AccountOwner && *myAccountRole != AccountAdmin {
		return a.store.getPublicAndSpecificAccessProjects(shard, accountId, myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
	}
	return a.store.getAllProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (a *api) ArchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	now := Now()
	a.store.setProjectArchivedOn(shard, accountId, projectId, &now)
	a.store.logAccountActivity(shard, accountId, now, myId, projectId, "project", "archived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, now, myId, projectId, "project", "archived", nil)
}

func (a *api) UnarchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setProjectArchivedOn(shard, accountId, projectId, nil)
	a.store.logAccountActivity(shard, accountId, Now(), myId, projectId, "project", "unarchived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, Now(), myId, projectId, "project", "unarchived", nil)
}

func (a *api) DeleteProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.deleteProject(shard, accountId, projectId)
	a.store.logAccountActivity(shard, accountId, Now(), myId, projectId, "project", "deleted", nil)
}

func (a *api) AddMembers(shard int, accountId, projectId, myId Id, members []*addProjectMember) {
	myAccountRole, myProjectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectAdminAccess(myAccountRole, myProjectRole)
	a.store.addMembers(shard, accountId, projectId, members)
	allIds := make([]Id, 0, len(members))
	for _, mem := range members {
		allIds = append(allIds, mem.Id)
	}
	a.store.logProjectBatchAddOrRemoveMembersActivity(shard, accountId, projectId, myId, allIds, "added")
}

func (a *api) RemoveMembers(shard int, accountId, projectId, myId Id, members []Id) {
	myAccountRole, myProjectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectAdminAccess(myAccountRole, myProjectRole)
	a.store.removeMembers(shard, accountId, projectId, members)
	a.store.logProjectBatchAddOrRemoveMembersActivity(shard, accountId, projectId, myId, members, "added")
}

func (a *api) GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int) {
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxGetEntityCount)
	return a.store.getMembers(shard, accountId, projectId, role, nameContains, offset, limit)
}

func (a *api) GetMe(shard int, accountId, projectId, myId Id) *projectMember {
	return a.store.getMember(shard, accountId, projectId, myId)
}

func (a *api) GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)
	_, limit = ValidateOffsetAndLimitParams(0, limit, a.maxGetEntityCount)
	return a.store.getActivities(shard, accountId, projectId, item, member, occurredAfter, occurredBefore, limit)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, bool)
	getPublicProjectsEnabled(shard int, accountId Id) bool
	//remember to create projectLocks db row
	createProject(shard int, accountId Id, project *project)
	setName(shard int, accountId, projectId Id, name string)
	setDescription(shard int, accountId, projectId Id, description string)
	setIsPublic(shard int, accountId, projectId Id, isPublic bool)
	setIsParallel(shard int, accountId, projectId Id, isParallel bool)
	getProject(shard int, accountId, projectId Id) *project
	getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int)
	getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int)
	getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int)
	setProjectArchivedOn(shard int, accountId, projectId Id, now *time.Time)
	deleteProject(shard int, accountId, projectId Id)
	addMembers(shard int, accountId, projectId Id, members []*addProjectMember)
	removeMembers(shard int, accountId, projectId Id, members []Id)
	getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int)
	getMember(shard int, accountId, projectId, members Id) *projectMember
	logAccountActivity(shard int, accountId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string)
	logProjectActivity(shard int, accountId, projectId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string)
	logProjectBatchAddOrRemoveMembersActivity(shard int, accountId, projectId, member Id, members []Id, action string)
	getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
}

type addProjectMember struct {
	Entity
	Role ProjectRole `json:"role"`
}

type projectMember struct {
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
