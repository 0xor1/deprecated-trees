package project

import (
	. "bitbucket.org/0xor1/task/server/util"
	"time"
)

var (
	publicProjectsDisabledErr = &AppError{Code: "r_v1_p_ppd", Message: "public projects disabled", Public: true}
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateProject(shard int, accountId, myId Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) *project {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	if isPublic && !a.store.getPublicProjectsEnabled(shard, accountId) {
		publicProjectsDisabledErr.Panic()
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
	a.store.createProject(shard, accountId, myId, project)
	if accountId.Equal(myId) {
		addMem := &AddProjectMember{}
		addMem.Id = myId
		addMem.Role = ProjectAdmin
		a.store.addMemberOrSetActive(shard, accountId, project.Id, myId, addMem)
	}

	if len(members) > 0 {
		a.AddMembers(shard, accountId, project.Id, myId, members)
	}

	return project
}

func (a *api) SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	if isPublic && !a.store.getPublicProjectsEnabled(shard, accountId) {
		publicProjectsDisabledErr.Panic()
	}

	a.store.setIsPublic(shard, accountId, projectId, myId, isPublic)
}

func (a *api) GetProject(shard int, accountId, projectId, myId Id) *project {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))

	return a.store.getProject(shard, accountId, projectId)
}

func (a *api) GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	myAccountRole := a.store.getAccountRole(shard, accountId, myId)
	limit = ValidateLimit(limit, a.maxProcessEntityCount)
	if myAccountRole == nil {
		return a.store.getPublicProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
	}
	if *myAccountRole != AccountOwner && *myAccountRole != AccountAdmin {
		return a.store.getPublicAndSpecificAccessProjects(shard, accountId, myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
	}
	return a.store.getAllProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func (a *api) ArchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setProjectIsArchived(shard, accountId, projectId, myId, true)
}

func (a *api) UnarchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setProjectIsArchived(shard, accountId, projectId, myId, false)
}

func (a *api) DeleteProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.deleteProject(shard, accountId, projectId, myId)
	//TODO delete s3 data, uploaded files etc
}

func (a *api) AddMembers(shard int, accountId, projectId, myId Id, members []*AddProjectMember) {
	ValidateEntityCount(len(members), a.maxProcessEntityCount)
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}

	ValidateMemberHasProjectAdminAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	ValidateExists(a.store.getProjectExists(shard, accountId, projectId))

	for _, mem := range members {
		mem.Role.Validate()
		accRole := a.store.getAccountRole(shard, accountId, mem.Id)
		if accRole == nil {
			InvalidArgumentsErr.Panic()
		}
		if *accRole == AccountOwner || *accRole == AccountAdmin {
			mem.Role = ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
		}
		a.store.addMemberOrSetActive(shard, accountId, projectId, myId, mem)
	}
}

func (a *api) SetMemberRole(shard int, accountId, projectId, myId Id, member Id, role ProjectRole) {
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}
	role.Validate()

	ValidateMemberHasProjectAdminAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	accRole, projectRole := a.store.getAccountAndProjectRoles(shard, accountId, projectId, member)
	if projectRole == nil {
		InvalidOperationErr.Panic()
	}
	if *projectRole != role {
		if role != ProjectAdmin && (*accRole == AccountOwner || *accRole == AccountAdmin) {
			InvalidArgumentsErr.Panic() // account owners and admins can only be project admins
		}
		a.store.setMemberRole(shard, accountId, projectId, myId, member, role)
	}
}

func (a *api) RemoveMembers(shard int, accountId, projectId, myId Id, members []Id) {
	ValidateEntityCount(len(members), a.maxProcessEntityCount)
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}
	ValidateMemberHasProjectAdminAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	for _, mem := range members {
		a.store.setMemberInactive(shard, accountId, projectId, myId, mem)
	}
}

func (a *api) GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, after *Id, limit int) ([]*member, bool) {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getMembers(shard, accountId, projectId, role, nameContains, after, ValidateLimit(limit, a.maxProcessEntityCount))
}

func (a *api) GetMe(shard int, accountId, projectId, myId Id) *member {
	return a.store.getMember(shard, accountId, projectId, myId)
}

func (a *api) GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
	if occurredAfter != nil && occurredBefore != nil {
		InvalidArgumentsErr.Panic()
	}
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getActivities(shard, accountId, projectId, item, member, occurredAfter, occurredBefore, ValidateLimit(limit, a.maxProcessEntityCount))
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	getProjectExists(shard int, accountId, projectId Id) bool
	getPublicProjectsEnabled(shard int, accountId Id) bool
	createProject(shard int, accountId, myId Id, project *project)
	setIsPublic(shard int, accountId, projectId, myId Id, isPublic bool)
	getProject(shard int, accountId, projectId Id) *project
	getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	setProjectIsArchived(shard int, accountId, projectId, myId Id, isArchived bool)
	deleteProject(shard int, accountId, projectId, myId Id)
	addMemberOrSetActive(shard int, accountId, projectId, myId Id, member *AddProjectMember)
	setMemberRole(shard int, accountId, projectId, myId, member Id, role ProjectRole)
	setMemberInactive(shard int, accountId, projectId, myId Id, member Id)
	getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, after *Id, limit int) ([]*member, bool)
	getMember(shard int, accountId, projectId, member Id) *member
	getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity
}

type member struct {
	Id                 Id          `json:"id"`
	TotalRemainingTime uint64      `json:"totalRemainingTime"`
	TotalLoggedTime    uint64      `json:"totalLoggedTime"`
	IsActive           bool        `json:"isActive"`
	Role               ProjectRole `json:"role"`
}

type project struct {
	Id                   Id         `json:"id"`
	IsArchived           bool       `json:"isArchived"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description"`
	CreatedOn            time.Time  `json:"createdOn"`
	StartOn              *time.Time `json:"startOn,omitempty"`
	DueOn                *time.Time `json:"dueOn,omitempty"`
	TotalRemainingTime   uint64     `json:"totalRemainingTime"`
	TotalLoggedTime      uint64     `json:"totalLoggedTime"`
	MinimumRemainingTime uint64     `json:"minimumRemainingTime"`
	FileCount            uint64     `json:"fileCount"`
	FileSize             uint64     `json:"fileSize"`
	LinkedFileCount      uint64     `json:"linkedFileCount"`
	ChatCount            uint64     `json:"chatCount"`
	ChildCount           uint64     `json:"childCount"`
	DescendantCount      uint64     `json:"descendantCount"`
	IsParallel           bool       `json:"isParallel,omitempty"`
	IsPublic             bool       `json:"isPublic"`
}
