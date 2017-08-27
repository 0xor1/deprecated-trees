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
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateProject(shard int, accountId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*addMember) *project {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	if isPublic && !a.store.getPublicProjectsEnabled(shard, accountId) {
		panic(publicProjectsDisabledErr)
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
	if accountId.Equal(myId) {
		addMem := &addMember{}
		addMem.Id = myId
		addMem.Role = ProjectAdmin
		a.store.addMemberOrSetActive(shard, accountId, project.Id, addMem)
	}

	a.store.logAccountActivity(shard, accountId, myId, project.Id, "project", "created", nil)
	a.store.logProjectActivity(shard, accountId, project.Id, myId, project.Id, "project", "created", nil)

	a.AddMembers(shard, accountId, project.Id, myId, members)

	return project
}

func (a *api) SetName(shard int, accountId, projectId, myId Id, name string) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setName(shard, accountId, projectId, name)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "setName", &name)
}

func (a *api) SetDescription(shard int, accountId, projectId, myId Id, description string) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setDescription(shard, accountId, projectId, description)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "setDescription", &description)
}

func (a *api) SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	if isPublic && !a.store.getPublicProjectsEnabled(shard, accountId) {
		panic(publicProjectsDisabledErr)
	}

	a.store.setIsPublic(shard, accountId, projectId, isPublic)
	action := fmt.Sprintf("%t", isPublic)
	a.store.logAccountActivity(shard, accountId, myId, projectId, "project", "setIsPublic", &action)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "setIsPublic", &action)
}

func (a *api) SetIsParallel(shard int, accountId, projectId, myId Id, isParallel bool) {
	myAccountRole, myProjectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectWriteAccess(myAccountRole, myProjectRole)

	a.store.setIsParallel(shard, accountId, projectId, isParallel)
	action := fmt.Sprintf("%t", isParallel)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "setIsParallel", &action)
}

func (a *api) GetProject(shard int, accountId, projectId, myId Id) *project {
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)

	return a.store.getProject(shard, accountId, projectId)
}

func (a *api) GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	myAccountRole := a.store.getAccountRole(shard, accountId, myId)
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxProcessEntityCount)
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
	a.store.logAccountActivity(shard, accountId, myId, projectId, "project", "archived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "archived", nil)
}

func (a *api) UnarchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setProjectArchivedOn(shard, accountId, projectId, nil)
	a.store.logAccountActivity(shard, accountId, myId, projectId, "project", "unarchived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, "project", "unarchived", nil)
}

func (a *api) DeleteProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.deleteProject(shard, accountId, projectId)
	a.store.logAccountActivity(shard, accountId, myId, projectId, "project", "deleted", nil)
	//TODO delete s3 data, uploaded files etc
}

func (a *api) AddMembers(shard int, accountId, projectId, myId Id, members []*addMember) {
	if len(members) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	accountRole, projectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectAdminAccess(accountRole, projectRole)

	addedMemberIds := make([]Id, 0, len(members))
	for _, mem := range members {
		mem.Role.Validate()
		if a.store.addMemberOrSetActive(shard, accountId, projectId, mem) {
			addedMemberIds = append(addedMemberIds, mem.Id)
		}
	}

	if len(addedMemberIds) > 0 {
		a.store.logProjectBatchAddOrRemoveMembersActivity(shard, accountId, projectId, myId, addedMemberIds, "added")
	}
}

func (a *api) SetMemberRole(shard int, accountId, projectId, myId Id, member Id, role ProjectRole) {
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	role.Validate()

	accountRole, projectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectAdminAccess(accountRole, projectRole)

	_, projectRole = a.store.getAccountAndProjectRoles(shard, accountId, projectId, member)
	if projectRole == nil {
		panic(InvalidOperationErr)
	}
	if *projectRole != role {
		a.store.setMemberRole(shard, accountId, projectId, member, role)
	}
}

func (a *api) RemoveMembers(shard int, accountId, projectId, myId Id, members []Id) {
	if len(members) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	myAccountRole, myProjectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId)
	ValidateMemberHasProjectAdminAccess(myAccountRole, myProjectRole)

	inactivatedMembers := make([]Id, 0, len(members))
	for _, mem := range members {
		if a.store.setMemberInactive(shard, accountId, projectId, mem) {
			inactivatedMembers = append(inactivatedMembers)
		}
	}
	a.store.logProjectBatchAddOrRemoveMembersActivity(shard, accountId, projectId, myId, members, "removed")
}

func (a *api) GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*member, int) {
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)
	offset, limit = ValidateOffsetAndLimitParams(offset, limit, a.maxProcessEntityCount)
	return a.store.getMembers(shard, accountId, projectId, role, nameContains, offset, limit)
}

func (a *api) GetMe(shard int, accountId, projectId, myId Id) *member {
	return a.store.getMember(shard, accountId, projectId, myId)
}

func (a *api) GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		panic(InvalidArgumentsErr)
	}
	myAccountRole, myProjectRole, projectIsPublic := a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId)
	ValidateMemberHasProjectReadAccess(myAccountRole, myProjectRole, projectIsPublic)
	_, limit = ValidateOffsetAndLimitParams(0, limit, a.maxProcessEntityCount)
	return a.store.getActivities(shard, accountId, projectId, item, member, occurredAfterUnixMillis, occurredBeforeUnixMillis, limit)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	getPublicProjectsEnabled(shard int, accountId Id) bool
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
	addMemberOrSetActive(shard int, accountId, projectId Id, member *addMember) bool
	setMemberRole(shard int, accountId, projectId Id, member Id, role ProjectRole)
	setMemberInactive(shard int, accountId, projectId Id, member Id) bool
	getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*member, int)
	getMember(shard int, accountId, projectId, member Id) *member
	logAccountActivity(shard int, accountId, member, item Id, itemType, action string, newValue *string)
	logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string)
	logProjectBatchAddOrRemoveMembersActivity(shard int, accountId, projectId, member Id, members []Id, action string)
	getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity
}

type addMember struct {
	Entity
	Role ProjectRole `json:"role"`
}

type member struct {
	NamedEntity
	CommonTimeProps
	IsActive bool        `json:"isActive"`
	Role     ProjectRole `json:"role"`
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
