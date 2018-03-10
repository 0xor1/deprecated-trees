package project

import (
	. "bitbucket.org/0xor1/task/server/util"
	"time"
	"fmt"
	"bytes"
	"strings"
)

var (
	publicProjectsDisabledErr = &AppError{Code: "r_v1_p_ppd", Message: "public projects disabled", Public: true}
)

type createProjectArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	Name string `json:"name"`
	Description *string `json:"description"`
	StartOn *time.Time `json:"startOn"`
	DueOn *time.Time `json:"dueOn"`
	IsParallel bool `json:"isParallel"`
	IsPublic bool `json:"isPublic"`
	Members []*AddProjectMember `json:"members"`
}

var createProject = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/createProject",
	GetArgsStruct: func() interface{} {
		return &createProjectArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*createProjectArgs)
		ValidateMemberHasAccountAdminAccess(GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		if args.IsPublic && !dbGetPublicProjectsEnabled(ctx, args.Shard, args.AccountId) {
			publicProjectsDisabledErr.Panic()
		}

		project := &project{}
		project.Id = NewId()
		project.Name = args.Name
		project.Description = args.Description
		project.CreatedOn = Now()
		project.StartOn = args.StartOn
		project.DueOn = args.DueOn
		project.IsParallel = args.IsParallel
		project.IsPublic = args.IsPublic
		dbCreateProject(ctx, args.Shard, args.AccountId, ctx.MyId(), project)
		if args.AccountId.Equal(ctx.MyId()) {
			addMem := &AddProjectMember{}
			addMem.Id = ctx.MyId()
			addMem.Role = ProjectAdmin
			dbAddMemberOrSetActive(ctx, args.Shard, args.AccountId, project.Id, ctx.MyId(), addMem)
		}

		if len(args.Members) > 0 {
			addMembers.CtxHandler(ctx, &addMembersArgs{
				Shard: args.Shard,
				AccountId: args.AccountId,
				ProjectId: project.Id,
				Members: args.Members,
			})
		}

		return project
	},
}

type setIsPublicArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	IsPublic bool `json:"isPublic"`
}

var setIsPublic = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setIsPublic",
	GetArgsStruct: func() interface{} {
		return &setIsPublicArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setIsPublicArgs)
		ValidateMemberHasAccountAdminAccess(GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))

		if args.IsPublic && !dbGetPublicProjectsEnabled(ctx, args.Shard, args.AccountId) {
			publicProjectsDisabledErr.Panic()
		}

		dbSetIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId(), args.IsPublic)

		return nil
	},
}

type setIsArchivedArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	IsArchived bool `json:"isArchived"`
}

var setIsArchived = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setIsArchived",
	GetArgsStruct: func() interface{} {
		return &setIsArchivedArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setIsArchivedArgs)
		ValidateMemberHasAccountAdminAccess(GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		dbSetProjectIsArchived(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId(), args.IsArchived)
		return nil
	},
}

type getProjectArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
}

var getProject = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getProject",
	GetArgsStruct: func() interface{} {
		return &getProjectArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getProjectArgs)
		ValidateMemberHasProjectReadAccess(GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		return dbGetProject(ctx, args.Shard, args.AccountId, args.ProjectId)
	},
}

type getProjectsArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	NameContains *string `json:"nameContains"`
	CreatedOnAfter *time.Time `json:"createdOnAfter"`
	CreatedOnBefore *time.Time `json:"createdOnBefore"`
	StartOnAfter *time.Time `json:"startOnAfter"`
	StartOnBefore *time.Time `json:"startOnBefore"`
	DueOnAfter *time.Time `json:"dueOnAfter"`
	DueOnBefore *time.Time `json:"dueOnBefore"`
	IsArchived bool `json:"isArchived"`
	SortBy SortBy `json:"sortBy"`
	SortDir SortDir `json:"sortDir"`
	After *Id `json:"after"`
	Limit int `json:"limit"`
}

type getProjectsResp struct {
	Projects []*project `json:"projects"`
	More bool `json:"more"`
}

var getProjects = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getProjects",
	GetArgsStruct: func() interface{} {
		return &getProjectsArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getProjectsArgs)
		myAccountRole := GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId())
		args.Limit = ValidateLimit(args.Limit, ctx.MaxProcessEntityCount())
		if myAccountRole == nil {
			return dbGetPublicProjects(ctx, args.Shard, args.AccountId, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		if *myAccountRole != AccountOwner && *myAccountRole != AccountAdmin {
			return dbGetPublicAndSpecificAccessProjects(ctx, args.Shard, args.AccountId, ctx.MyId(), args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		return dbGetAllProjects(ctx, args.Shard, args.AccountId, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
	},
}

type deleteProjectArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
}

var deleteProject = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/deleteProject",
	GetArgsStruct: func() interface{} {
		return &deleteProjectArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*deleteProjectArgs)
		ValidateMemberHasAccountAdminAccess(GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		dbDeleteProject(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId())
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	Members []*AddProjectMember `json:"members"`
}

var addMembers = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/addMembers",
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		ValidateEntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(ctx.MyId()) {
			InvalidOperationErr.Panic()
		}

		ValidateMemberHasProjectAdminAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		ValidateExists(dbGetProjectExists(ctx, args.Shard, args.AccountId, args.ProjectId))

		for _, mem := range args.Members {
			mem.Role.Validate()
			accRole := GetAccountRole(ctx, args.Shard, args.AccountId, mem.Id)
			if accRole == nil {
				InvalidArgumentsErr.Panic()
			}
			if *accRole == AccountOwner || *accRole == AccountAdmin {
				mem.Role = ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
			}
			dbAddMemberOrSetActive(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId(), mem)
		}
		return nil
	},
}

type setMemberRoleArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	Member Id `json:"member"`
	Role ProjectRole `json:"role"`
}

var setMemberRole = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setMemberRole",
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		if args.AccountId.Equal(ctx.MyId()) {
			InvalidOperationErr.Panic()
		}
		args.Role.Validate()

		ValidateMemberHasProjectAdminAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		accRole, projectRole := GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, args.Member)
		if projectRole == nil {
			InvalidOperationErr.Panic()
		}
		if *projectRole != args.Role {
			if args.Role != ProjectAdmin && (*accRole == AccountOwner || *accRole == AccountAdmin) {
				InvalidArgumentsErr.Panic() // account owners and admins can only be project admins
			}
			dbSetMemberRole(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId(), args.Member, args.Role)
		}
		return nil
	},
}

type removeMembersArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	Members []Id `json:"members"`
}

var removeMembers = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/removeMembers",
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		ValidateEntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(ctx.MyId()) {
			InvalidOperationErr.Panic()
		}
		ValidateMemberHasProjectAdminAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		for _, mem := range args.Members {
			dbSetMemberInactive(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId(), mem)
		}
		return nil
	},
}

type getMembersArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	Role *ProjectRole `json:"role,omitempty"`
	NameContains *string `json:"nameContains,omitempty"`
	After *Id `json:"after,omitempty"`
	Limit int `json:"limit"`
}

type getMembersResp struct {
	Members []*member `json:"members"`
	More bool `json:"more"`
}

var getMembers = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getMembers",
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		ValidateMemberHasProjectReadAccess(GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		return dbGetMembers(ctx, args.Shard, args.AccountId, args.ProjectId, args.Role, args.NameContains, args.After, ValidateLimit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
}

var getMe = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getMe",
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId())
	},
}

type getActivitiesArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	Item *Id `json:"item,omitempty"`
	Member *Id `json:"member,omitempty"`
	OccurredAfter *time.Time `json:"occurredAfter,omitempty"`
	OccurredBefore *time.Time `json:"occurredBefore,omitempty"`
	Limit int `json:"limit"`
}

var getActivities = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getActivities",
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		if args.OccurredAfter != nil && args.OccurredBefore != nil {
			InvalidArgumentsErr.Panic()
		}
		ValidateMemberHasProjectReadAccess(GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		return dbGetActivities(ctx, args.Shard, args.AccountId, args.ProjectId, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, ValidateLimit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

var Endpoints = []*Endpoint{
	createProject,
	setIsPublic,
	setIsArchived,
	getProject,
	getProjects,
	deleteProject,
	addMembers,
	setMemberRole,
	removeMembers,
	getMembers,
	getMe,
	getActivities,
}

type Client interface {
	//must be account owner/admin
	CreateProject(css *ClientSessionStore, shard int, accountId Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error)
	//must be account owner/admin and account.publicProjectsEnabled must be true
	SetIsPublic(css *ClientSessionStore, shard int, accountId, projectId Id, isPublic bool) error
	//must be account owner/admin
	SetIsArchived(css *ClientSessionStore, shard int, accountId, projectId Id, isArchived bool) error
	//check project access permission per user
	GetProject(css *ClientSessionStore, shard int, accountId, projectId Id) (*project, error)
	//check project access permission per user
	GetProjects(css *ClientSessionStore, shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) (*getProjectsResp, error)
	//must be account owner/admin
	DeleteProject(css *ClientSessionStore, shard int, accountId, projectId Id) error
	//must be account owner/admin or project admin
	AddMembers(css *ClientSessionStore, shard int, accountId, projectId Id, members []*AddProjectMember) error
	//must be account owner/admin or project admin
	SetMemberRole(css *ClientSessionStore, shard int, accountId, projectId Id, member Id, role ProjectRole) error
	//must be account owner/admin or project admin
	RemoveMembers(css *ClientSessionStore, shard int, accountId, projectId Id, members []Id) error
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(css *ClientSessionStore, shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, after *Id, limit int) (*getMembersResp, error)
	//for anyone
	GetMe(css *ClientSessionStore, shard int, accountId, projectId Id) (*member, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *ClientSessionStore, shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*Activity, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateProject(css *ClientSessionStore, shard int, accountId Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error) {
	val, err := createProject.DoRequest(css, c.host, &createProjectArgs{
		Shard: shard,
		AccountId: accountId,
		Name: name,
		Description: description,
		StartOn: startOn,
		DueOn: dueOn,
		IsParallel: isParallel,
		IsPublic: isPublic,
		Members: members,
	}, nil, &project{})
	if val != nil {
		return val.(*project), err
	}
	return nil, err
}

func (c *client) SetIsPublic(css *ClientSessionStore, shard int, accountId, projectId Id, isPublic bool) error {
	_, err := setIsPublic.DoRequest(css, c.host, &setIsPublicArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		IsPublic: isPublic,
	}, nil, nil)
	return err
}

func (c *client) SetIsArchived(css *ClientSessionStore, shard int, accountId, projectId Id, isArchived bool) error {
	_, err := setIsArchived.DoRequest(css, c.host, &setIsArchivedArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		IsArchived: isArchived,
	}, nil, nil)
	return err
}

func (c *client) GetProject(css *ClientSessionStore, shard int, accountId, projectId Id) (*project, error) {
	val, err := getProject.DoRequest(css, c.host, &getProjectArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, &project{})
	if val != nil {
		return val.(*project), err
	}
	return nil, err
}

func (c *client) GetProjects(css *ClientSessionStore, shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) (*getProjectsResp, error) {
	val, err := getProjects.DoRequest(css, c.host, &getProjectsArgs{
		Shard: shard,
		AccountId: accountId,
		NameContains: nameContains,
		CreatedOnAfter: createdOnAfter,
		CreatedOnBefore: createdOnBefore,
		StartOnAfter: startOnAfter,
		StartOnBefore: startOnBefore,
		DueOnAfter: dueOnAfter,
		DueOnBefore: dueOnBefore,
		IsArchived: isArchived,
		SortBy: sortBy,
		SortDir: sortDir,
		After: after,
		Limit: limit,
	}, nil, &getProjectsResp{})
	if val != nil {
		return val.(*getProjectsResp), err
	}
	return nil, err
}

func (c *client) DeleteProject(css *ClientSessionStore, shard int, accountId, projectId Id) error {
	_, err := deleteProject.DoRequest(css, c.host, &deleteProjectArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, nil)
	return err
}

func (c *client) AddMembers(css *ClientSessionStore, shard int, accountId, projectId Id, members []*AddProjectMember) error {
	_, err := addMembers.DoRequest(css, c.host, &addMembersArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		Members: members,
	}, nil, nil)
	return err
}

func (c *client) SetMemberRole(css *ClientSessionStore, shard int, accountId, projectId, member Id, role ProjectRole) error {
	_, err := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		Member: member,
		Role: role,
	}, nil, nil)
	return err
}

func (c *client) RemoveMembers(css *ClientSessionStore, shard int, accountId, projectId Id, members []Id) error {
	_, err := removeMembers.DoRequest(css, c.host, &removeMembersArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		Members: members,
	}, nil, nil)
	return err
}

func (c *client) GetMembers(css *ClientSessionStore, shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, after *Id, limit int) (*getMembersResp, error) {
	val, err := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		Role: role,
		NameContains: nameContains,
		After: after,
		Limit: limit,
	}, nil, &getMembersResp{})
	if val != nil {
		return val.(*getMembersResp), err
	}
	return nil, err
}

func (c *client) GetMe(css *ClientSessionStore, shard int, accountId, projectId Id) (*member, error) {
	val, err := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, &member{})
	if val != nil {
		return val.(*member), err
	}
	return nil, err
}

func (c *client) GetActivities(css *ClientSessionStore, shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*Activity, error) {
	val, err := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		Item: item,
		Member: member,
		OccurredAfter: occurredAfter,
		OccurredBefore: occurredBefore,
		Limit: limit,
	}, nil, &[]*Activity{})
	if val != nil {
		return *val.(*[]*Activity), err
	}
	return nil, err
}

func dbGetProjectExists(ctx *Ctx, shard int, accountId, projectId Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, []byte(accountId), []byte(projectId))
	exists := false
	PanicIf(row.Scan(&exists))
	return exists
}

func dbGetPublicProjectsEnabled(ctx *Ctx, shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(ctx, shard, accountId)
}

func dbCreateProject(ctx *Ctx, shard int, accountId, myId Id, project *project) {
	_, err := ctx.TreeExec(shard,  `CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(project.Id), []byte(myId), project.Name, project.Description, project.CreatedOn, project.StartOn, project.DueOn, project.IsParallel, project.IsPublic)
	PanicIf(err)
}

func dbSetIsPublic(ctx *Ctx, shard int, accountId, projectId, myId Id, isPublic bool) {
	_, err := ctx.TreeExec(shard,  `CALL setProjectIsPublic(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), isPublic)
	PanicIf(err)
}

func dbGetProject(ctx *Ctx, shard int, accountId, projectId Id) *project {
	row := ctx.TreeQueryRow(shard, `SELECT p.id, p.isArchived, p.name, p.createdOn, p.startOn, p.dueOn, p.fileCount, p.fileSize, p.isPublic, t.description, t.totalRemainingTime, t.totalLoggedTime, t.minimumRemainingTime, t.linkedFileCount, t.chatCount, t.childCount, t.descendantCount, t.isParallel FROM projects p, tasks t WHERE p.account=? AND p.id=? AND t.account=? AND t.project=? AND t.id=?`, []byte(accountId), []byte(projectId), []byte(accountId), []byte(projectId), []byte(projectId))
	result := project{}
	PanicIf(row.Scan(&result.Id, &result.IsArchived, &result.Name, &result.CreatedOn, &result.StartOn, &result.DueOn, &result.FileCount, &result.FileSize, &result.IsPublic, &result.Description, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.LinkedFileCount, &result.ChatCount, &result.ChildCount, &result.DescendantCount, &result.IsParallel))
	return &result
}

func dbGetPublicProjects(ctx *Ctx, shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND isPublic=true`, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetPublicAndSpecificAccessProjects(ctx *Ctx, shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, accountId, &myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetAllProjects(ctx *Ctx, shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, ``, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbSetProjectIsArchived(ctx *Ctx, shard int, accountId, projectId, myId Id, isArchived bool) {
	_, err := ctx.TreeExec(shard,  `CALL setProjectIsArchived(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), isArchived)
	PanicIf(err)
}

func dbDeleteProject(ctx *Ctx, shard int, accountId, projectId, myId Id) {
	_, err := ctx.TreeExec(shard,  `CALL deleteProject(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId))
	PanicIf(err)
}

func dbAddMemberOrSetActive(ctx *Ctx, shard int, accountId, projectId, myId Id, member *AddProjectMember) {
	MakeChangeHelper(ctx, shard, `CALL addProjectMemberOrSetActive(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member.Id), member.Role)
}

func dbSetMemberRole(ctx *Ctx, shard int, accountId, projectId, myId, member Id, role ProjectRole) {
	MakeChangeHelper(ctx, shard, `CALL setProjectMemberRole(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member), role)
}

func dbSetMemberInactive(ctx *Ctx, shard int, accountId, projectId, myId Id, member Id) {
	MakeChangeHelper(ctx, shard, `CALL setProjectMemberInactive(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member))
}

func dbGetMembers(ctx *Ctx, shard int, accountId, projectId Id, role *ProjectRole, nameOrDisplayNameContains *string, after *Id, limit int) *getMembersResp {
	query := bytes.NewBufferString(`SELECT p1.id, p1.isActive, p1.totalRemainingTime, p1.totalLoggedTime, p1.role FROM projectMembers p1`)
	args := make([]interface{}, 0, 9)
	if after != nil {
		query.WriteString(`, projectMembers p2`)
	}
	query.WriteString(` WHERE p1.account=? AND p1.project=? AND p1.isActive=true`)
	args = append(args, []byte(accountId), []byte(projectId))
	if after != nil {
		query.WriteString(` AND p2.account=? AND p2.project=? p2.id=? AND ((p1.name>p2.name AND p1.role=p2.role) OR p1.role>p2.role)`)
		args = append(args, []byte(accountId), []byte(projectId), []byte(*after))
	}
	if role != nil {
		query.WriteString(` AND p1.role=?`)
		args = append(args, role)
	}
	if nameOrDisplayNameContains != nil {
		query.WriteString(` AND (p1.name LIKE ? OR p1.displayName LIKE ?)`)
		strVal := strings.Trim(*nameOrDisplayNameContains, " ")
		strVal = fmt.Sprintf("%%%s%%", strVal)
		args = append(args, strVal, strVal)
	}
	query.WriteString(` ORDER BY p1.role ASC, p1.name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, err := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return &getMembersResp{Members: res[:limit], More: true}
	}
	return &getMembersResp{Members: res, More: false}
}

func dbGetMember(ctx *Ctx, shard int, accountId, projectId, memberId Id) *member {
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, []byte(accountId), []byte(projectId), []byte(memberId))
	res := member{}
	PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
	return &res
}

func dbGetActivities(ctx *Ctx, shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {
	if occurredAfter != nil && occurredBefore != nil {
		InvalidArgumentsErr.Panic()
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, extraInfo FROM projectActivities WHERE account=? AND project=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(accountId), []byte(projectId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
	}
	if occurredAfter != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfter)
	}
	if occurredBefore != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBefore)
	}
	if occurredAfter == nil && occurredBefore == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, err := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*Activity, 0, limit)
	for rows.Next() {
		act := Activity{}
		PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	return res
}

func dbGetProjects(ctx *Ctx, shard int, specificSqlFilterTxt string, accountId Id, myId *Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) *getProjectsResp {
	query := bytes.NewBufferString(`SELECT id, isArchived, name, createdOn, startOn, dueOn, fileCount, fileSize, isPublic FROM projects WHERE account=? AND isArchived=? %s`)
	args := make([]interface{}, 0, 14)
	args = append(args, []byte(accountId), isArchived)
	if myId != nil {
		args = append(args, []byte(accountId), []byte(*myId))
	}
	if nameContains != nil {
		query.WriteString(` AND name LIKE ?`)
		args = append(args, fmt.Sprintf(`%%%s%%`, strings.Trim(*nameContains, " ")))
	}
	if createdOnAfter != nil {
		query.WriteString(` AND createdOn>?`)
		args = append(args, createdOnAfter)
	}
	if createdOnBefore != nil {
		query.WriteString(` AND createdOn<?`)
		args = append(args, createdOnBefore)
	}
	if startOnAfter != nil {
		query.WriteString(` AND startOn>?`)
		args = append(args, startOnAfter)
	}
	if startOnBefore != nil {
		query.WriteString(` AND startOn<?`)
		args = append(args, startOnBefore)
	}
	if dueOnAfter != nil {
		query.WriteString(` AND dueOn>?`)
		args = append(args, dueOnAfter)
	}
	if dueOnBefore != nil {
		query.WriteString(` AND dueOn<?`)
		args = append(args, dueOnBefore)
	}
	if after != nil {
		query.WriteString(fmt.Sprintf(` AND %s %s= (SELECT %s FROM projects WHERE account=? AND id=?) AND id > ?`, sortBy, sortDir.GtLtSymbol(), sortBy))
		args = append(args, []byte(accountId), []byte(*after), []byte(*after))
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s, id LIMIT ?`, sortBy, sortDir))
	args = append(args, limit+1)
	rows, err := ctx.TreeQuery(shard, fmt.Sprintf(query.String(), specificSqlFilterTxt), args...)
	PanicIf(err)
	res := make([]*project, 0, limit+1)
	idx := 0
	resIdx := map[string]int{}
	for rows.Next() {
		proj := project{}
		PanicIf(rows.Scan(&proj.Id, &proj.IsArchived, &proj.Name, &proj.CreatedOn, &proj.StartOn, &proj.DueOn, &proj.FileCount, &proj.FileSize, &proj.IsPublic))
		res = append(res, &proj)
		resIdx[proj.Id.String()] = idx
		idx++
	}
	if len(res) > 0 { //populate task properties
		var id Id
		var description *string
		var totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount uint64
		var isParallel bool
		query.Reset()
		args = make([]interface{}, 0, len(res)+1)
		args = append(args, []byte(accountId), []byte(res[0].Id))
		query.WriteString(`SELECT id, description, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel FROM tasks WHERE account=? AND project=id AND project IN (?`)
		for _, proj := range res[1:] {
			query.WriteString(`,?`)
			args = append(args, []byte(proj.Id))
		}
		query.WriteString(fmt.Sprintf(`) LIMIT %d`, len(res)))
		rows, err := ctx.TreeQuery(shard, query.String(), args...)
		PanicIf(err)
		for rows.Next() {
			rows.Scan(&id, &description, &totalRemainingTime, &totalLoggedTime, &minimumRemainingTime, &linkedFileCount, &chatCount, &childCount, &descendantCount, &isParallel)
			proj := res[resIdx[id.String()]]
			proj.Description = description
			proj.TotalRemainingTime = totalRemainingTime
			proj.TotalLoggedTime = totalLoggedTime
			proj.MinimumRemainingTime = minimumRemainingTime
			proj.LinkedFileCount = linkedFileCount
			proj.ChatCount = chatCount
			proj.ChildCount = childCount
			proj.DescendantCount = descendantCount
			proj.IsParallel = isParallel
		}
	}
	if len(res) == limit+1 {
		return &getProjectsResp{Projects: res[:limit], More: true}
	}
	return &getProjectsResp{Projects: res, More: false}
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
