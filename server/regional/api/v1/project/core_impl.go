package project

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"fmt"
	"time"
)

const(
	itemType = "project"
)

var (
	publicProjectsDisabledErr = &AppError{Code: "r_v1_p_ppd", Message: "public projects disabled", Public: true}
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateProject(shard int, accountId, myId Id, name, description string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*addMember) *project {
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
	project.MinimumRemainingTime = 0
	project.IsPublic = isPublic
	a.store.createProject(shard, accountId, project)
	if accountId.Equal(myId) {
		addMem := &addMember{}
		addMem.Id = myId
		addMem.Role = ProjectAdmin
		a.store.addMemberOrSetActive(shard, accountId, project.Id, addMem)
	}

	a.store.logAccountActivity(shard, accountId, myId, project.Id, itemType, "created", nil)
	a.store.logProjectActivity(shard, accountId, project.Id, myId, project.Id, itemType, "created", nil)

	if len(members) > 0 {
		a.AddMembers(shard, accountId, project.Id, myId, members)
	}

	return project
}

func (a *api) SetName(shard int, accountId, projectId, myId Id, name string) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setName(shard, accountId, projectId, name)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "setName", &name)
}

func (a *api) SetDescription(shard int, accountId, projectId, myId Id, description string) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	a.store.setDescription(shard, accountId, projectId, description)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "setDescription", &description)
}

func (a *api) SetIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))

	if isPublic && !a.store.getPublicProjectsEnabled(shard, accountId) {
		publicProjectsDisabledErr.Panic()
	}

	a.store.setIsPublic(shard, accountId, projectId, isPublic)
	action := fmt.Sprintf("%t", isPublic)
	a.store.logAccountActivity(shard, accountId, myId, projectId, itemType, "setIsPublic", &action)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "setIsPublic", &action)
}

func (a *api) SetIsParallel(shard int, accountId, projectId, myId Id, isParallel bool) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setIsParallel(shard, accountId, projectId, isParallel)
	action := fmt.Sprintf("%t", isParallel)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "setIsParallel", &action)
}

func (a *api) GetProject(shard int, accountId, projectId, myId Id) *project {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))

	return a.store.getProject(shard, accountId, projectId)
}

func (a *api) GetProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	myAccountRole := a.store.getAccountRole(shard, accountId, myId)
	limit = ValidateLimitParam(limit, a.maxProcessEntityCount)
	if myAccountRole == nil {
		return a.store.getPublicProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, after, limit)
	}
	if *myAccountRole != AccountOwner && *myAccountRole != AccountAdmin {
		return a.store.getPublicAndSpecificAccessProjects(shard, accountId, myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, after, limit)
	}
	return a.store.getAllProjects(shard, accountId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, after, limit)
}

func (a *api) ArchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	now := Now()
	a.store.setProjectArchivedOn(shard, accountId, projectId, &now)
	a.store.logAccountActivity(shard, accountId, myId, projectId, itemType, "archived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "archived", nil)
}

func (a *api) UnarchiveProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.setProjectArchivedOn(shard, accountId, projectId, nil)
	a.store.logAccountActivity(shard, accountId, myId, projectId, itemType, "unarchived", nil)
	a.store.logProjectActivity(shard, accountId, projectId, myId, projectId, itemType, "unarchived", nil)
}

func (a *api) DeleteProject(shard int, accountId, projectId, myId Id) {
	ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	a.store.deleteProject(shard, accountId, projectId)
	a.store.logAccountActivity(shard, accountId, myId, projectId, itemType, "deleted", nil)
	//TODO delete s3 data, uploaded files etc
}

func (a *api) AddMembers(shard int, accountId, projectId, myId Id, members []*addMember) {
	ValidateEntityCount(len(members), a.maxProcessEntityCount)
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}
	ValidateMemberHasProjectAdminAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	addedMemberIds := make([]Id, 0, len(members))
	for _, mem := range members {
		mem.Role.Validate()
		accRole := a.store.getAccountRole(shard, accountId, mem.Id)
		if accRole == nil {
			InvalidArgumentsErr.Panic()
		}
		if *accRole == AccountOwner || *accRole == AccountAdmin {
			mem.Role = ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
		}
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
		a.store.setMemberRole(shard, accountId, projectId, member, role)
		roleStr := role.String()
		a.store.logProjectActivity(shard, accountId, projectId, myId, member, "member", "setRole", &roleStr)
	}
}

func (a *api) RemoveMembers(shard int, accountId, projectId, myId Id, members []Id) {
	ValidateEntityCount(len(members), a.maxProcessEntityCount)
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}
	ValidateMemberHasProjectAdminAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	inactivatedMembers := make([]Id, 0, len(members))
	for _, mem := range members {
		if a.store.setMemberInactive(shard, accountId, projectId, mem) {
			inactivatedMembers = append(inactivatedMembers)
		}
	}
	a.store.logProjectBatchAddOrRemoveMembersActivity(shard, accountId, projectId, myId, members, "removed")
}

func (a *api) GetMembers(shard int, accountId, projectId, myId Id, role *ProjectRole, nameContains *string, after *Id, limit int) ([]*member, bool) {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getMembers(shard, accountId, projectId, role, nameContains, after, ValidateLimitParam(limit, a.maxProcessEntityCount))
}

func (a *api) GetMe(shard int, accountId, projectId, myId Id) *member {
	return a.store.getMember(shard, accountId, projectId, myId)
}

func (a *api) GetActivities(shard int, accountId, projectId, myId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		InvalidArgumentsErr.Panic()
	}
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getActivities(shard, accountId, projectId, item, member, occurredAfterUnixMillis, occurredBeforeUnixMillis, ValidateLimitParam(limit, a.maxProcessEntityCount))
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
	getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool)
	setProjectArchivedOn(shard int, accountId, projectId Id, now *time.Time)
	deleteProject(shard int, accountId, projectId Id)
	addMemberOrSetActive(shard int, accountId, projectId Id, member *addMember) bool
	setMemberRole(shard int, accountId, projectId Id, member Id, role ProjectRole)
	setMemberInactive(shard int, accountId, projectId Id, member Id) bool
	getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, after *Id, limit int) ([]*member, bool)
	getMember(shard int, accountId, projectId, member Id) *member
	logAccountActivity(shard int, accountId, member, item Id, itemType, action string, newValue *string)
	logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string)
	logProjectBatchAddOrRemoveMembersActivity(shard int, accountId, projectId, member Id, members []Id, action string)
	getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity
}

type addMember struct {
	Id   Id          `json:"id"`
	Role ProjectRole `json:"role"`
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
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	CreatedOn            time.Time  `json:"createdOn"`
	ArchivedOn           *time.Time `json:"archivedOn,omitempty"`
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
