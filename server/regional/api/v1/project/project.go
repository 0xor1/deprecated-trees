package project

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
	"fmt"
	"strings"
	"time"
)

var (
	publicProjectsDisabledErr = &err.Err{Code: "r_v1_p_ppd", Message: "public projects disabled"}
)

type createProjectArgs struct {
	Shard       int                 `json:"shard"`
	AccountId   id.Id               `json:"accountId"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	StartOn     *time.Time          `json:"startOn"`
	DueOn       *time.Time          `json:"dueOn"`
	IsParallel  bool                `json:"isParallel"`
	IsPublic    bool                `json:"isPublic"`
	Members     []*AddProjectMember `json:"members"`
}

var createProject = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/createProject",
	GetArgsStruct: func() interface{} {
		return &createProjectArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createProjectArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		if args.IsPublic && !db.GetPublicProjectsEnabled(ctx, args.Shard, args.AccountId) {
			panic(publicProjectsDisabledErr)
		}

		project := &project{}
		project.Id = id.New()
		project.Name = args.Name
		project.Description = args.Description
		project.CreatedOn = t.Now()
		project.StartOn = args.StartOn
		project.DueOn = args.DueOn
		project.IsParallel = args.IsParallel
		project.IsPublic = args.IsPublic
		dbCreateProject(ctx, args.Shard, args.AccountId, ctx.Me(), project)
		if args.AccountId.Equal(ctx.Me()) {
			addMem := &AddProjectMember{}
			addMem.Id = ctx.Me()
			addMem.Role = cnst.ProjectAdmin
			dbAddMemberOrSetActive(ctx, args.Shard, args.AccountId, project.Id, ctx.Me(), addMem)
		}

		if len(args.Members) > 0 {
			addMembers.CtxHandler(ctx, &addMembersArgs{
				Shard:     args.Shard,
				AccountId: args.AccountId,
				ProjectId: project.Id,
				Members:   args.Members,
			})
		}

		return project
	},
}

type setIsPublicArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	ProjectId id.Id `json:"projectId"`
	IsPublic  bool  `json:"isPublic"`
}

var setIsPublic = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setIsPublic",
	GetArgsStruct: func() interface{} {
		return &setIsPublicArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setIsPublicArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))

		if args.IsPublic && !db.GetPublicProjectsEnabled(ctx, args.Shard, args.AccountId) {
			panic(publicProjectsDisabledErr)
		}

		dbSetIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me(), args.IsPublic)

		return nil
	},
}

type setIsArchivedArgs struct {
	Shard      int   `json:"shard"`
	AccountId  id.Id `json:"accountId"`
	ProjectId  id.Id `json:"projectId"`
	IsArchived bool  `json:"isArchived"`
}

var setIsArchived = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setIsArchived",
	GetArgsStruct: func() interface{} {
		return &setIsArchivedArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setIsArchivedArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		dbSetProjectIsArchived(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me(), args.IsArchived)
		return nil
	},
}

type getProjectArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	ProjectId id.Id `json:"projectId"`
}

var getProject = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getProject",
	GetArgsStruct: func() interface{} {
		return &getProjectArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getProjectArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		return dbGetProject(ctx, args.Shard, args.AccountId, args.ProjectId)
	},
}

type getProjectsArgs struct {
	Shard           int          `json:"shard"`
	AccountId       id.Id        `json:"accountId"`
	NameContains    *string      `json:"nameContains"`
	CreatedOnAfter  *time.Time   `json:"createdOnAfter"`
	CreatedOnBefore *time.Time   `json:"createdOnBefore"`
	StartOnAfter    *time.Time   `json:"startOnAfter"`
	StartOnBefore   *time.Time   `json:"startOnBefore"`
	DueOnAfter      *time.Time   `json:"dueOnAfter"`
	DueOnBefore     *time.Time   `json:"dueOnBefore"`
	IsArchived      bool         `json:"isArchived"`
	SortBy          cnst.SortBy  `json:"sortBy"`
	SortDir         cnst.SortDir `json:"sortDir"`
	After           *id.Id       `json:"after"`
	Limit           int          `json:"limit"`
}

type getProjectsResp struct {
	Projects []*project `json:"projects"`
	More     bool       `json:"more"`
}

var getProjects = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getProjects",
	GetArgsStruct: func() interface{} {
		return &getProjectsArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getProjectsArgs)
		myAccountRole := db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me())
		args.Limit = validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		if myAccountRole == nil {
			return dbGetPublicProjects(ctx, args.Shard, args.AccountId, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		if *myAccountRole != cnst.AccountOwner && *myAccountRole != cnst.AccountAdmin {
			return dbGetPublicAndSpecificAccessProjects(ctx, args.Shard, args.AccountId, ctx.Me(), args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		return dbGetAllProjects(ctx, args.Shard, args.AccountId, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
	},
}

type deleteProjectArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	ProjectId id.Id `json:"projectId"`
}

var deleteProject = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/deleteProject",
	GetArgsStruct: func() interface{} {
		return &deleteProjectArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteProjectArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		dbDeleteProject(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me())
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard     int                 `json:"shard"`
	AccountId id.Id               `json:"accountId"`
	ProjectId id.Id               `json:"projectId"`
	Members   []*AddProjectMember `json:"members"`
}

var addMembers = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/addMembers",
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		validate.Exists(dbGetProjectExists(ctx, args.Shard, args.AccountId, args.ProjectId))

		for _, mem := range args.Members {
			mem.Role.Validate()
			accRole := db.GetAccountRole(ctx, args.Shard, args.AccountId, mem.Id)
			if accRole == nil {
				panic(err.InvalidArguments)
			}
			if *accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin {
				mem.Role = cnst.ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
			}
			dbAddMemberOrSetActive(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me(), mem)
		}
		return nil
	},
}

type setMemberRoleArgs struct {
	Shard     int              `json:"shard"`
	AccountId id.Id            `json:"accountId"`
	ProjectId id.Id            `json:"projectId"`
	Member    id.Id            `json:"member"`
	Role      cnst.ProjectRole `json:"role"`
}

var setMemberRole = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setMemberRole",
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		if args.AccountId.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		args.Role.Validate()

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		accRole, projectRole := db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, args.Member)
		if projectRole == nil {
			panic(err.InvalidOperation)
		}
		if *projectRole != args.Role {
			if args.Role != cnst.ProjectAdmin && (*accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin) {
				panic(err.InvalidArguments) // account owners and admins can only be project admins
			}
			dbSetMemberRole(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me(), args.Member, args.Role)
		}
		return nil
	},
}

type removeMembersArgs struct {
	Shard     int     `json:"shard"`
	AccountId id.Id   `json:"accountId"`
	ProjectId id.Id   `json:"projectId"`
	Members   []id.Id `json:"members"`
}

var removeMembers = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/removeMembers",
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		for _, mem := range args.Members {
			dbSetMemberInactive(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me(), mem)
		}
		return nil
	},
}

type getMembersArgs struct {
	Shard        int               `json:"shard"`
	AccountId    id.Id             `json:"accountId"`
	ProjectId    id.Id             `json:"projectId"`
	Role         *cnst.ProjectRole `json:"role,omitempty"`
	NameContains *string           `json:"nameContains,omitempty"`
	After        *id.Id            `json:"after,omitempty"`
	Limit        int               `json:"limit"`
}

type getMembersResp struct {
	Members []*member `json:"members"`
	More    bool      `json:"more"`
}

var getMembers = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getMembers",
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		return dbGetMembers(ctx, args.Shard, args.AccountId, args.ProjectId, args.Role, args.NameContains, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	ProjectId id.Id `json:"projectId"`
}

var getMe = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getMe",
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me())
	},
}

type getActivitiesArgs struct {
	Shard          int        `json:"shard"`
	AccountId      id.Id      `json:"accountId"`
	ProjectId      id.Id      `json:"projectId"`
	Item           *id.Id     `json:"item,omitempty"`
	Member         *id.Id     `json:"member,omitempty"`
	OccurredAfter  *time.Time `json:"occurredAfter,omitempty"`
	OccurredBefore *time.Time `json:"occurredBefore,omitempty"`
	Limit          int        `json:"limit"`
}

var getActivities = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getActivities",
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		if args.OccurredAfter != nil && args.OccurredBefore != nil {
			panic(err.InvalidArguments)
		}
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		return dbGetActivities(ctx, args.Shard, args.AccountId, args.ProjectId, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

var Endpoints = []*endpoint.Endpoint{
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
	CreateProject(css *clientsession.Store, shard int, accountId id.Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error)
	//must be account owner/admin and account.publicProjectsEnabled must be true
	SetIsPublic(css *clientsession.Store, shard int, accountId, projectId id.Id, isPublic bool) error
	//must be account owner/admin
	SetIsArchived(css *clientsession.Store, shard int, accountId, projectId id.Id, isArchived bool) error
	//check project access permission per user
	GetProject(css *clientsession.Store, shard int, accountId, projectId id.Id) (*project, error)
	//check project access permission per user
	GetProjects(css *clientsession.Store, shard int, accountId id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) (*getProjectsResp, error)
	//must be account owner/admin
	DeleteProject(css *clientsession.Store, shard int, accountId, projectId id.Id) error
	//must be account owner/admin or project admin
	AddMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, members []*AddProjectMember) error
	//must be account owner/admin or project admin
	SetMemberRole(css *clientsession.Store, shard int, accountId, projectId id.Id, member id.Id, role cnst.ProjectRole) error
	//must be account owner/admin or project admin
	RemoveMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, members []id.Id) error
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, role *cnst.ProjectRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error)
	//for anyone
	GetMe(css *clientsession.Store, shard int, accountId, projectId id.Id) (*member, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, shard int, accountId, projectId id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateProject(css *clientsession.Store, shard int, accountId id.Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error) {
	val, e := createProject.DoRequest(css, c.host, &createProjectArgs{
		Shard:       shard,
		AccountId:   accountId,
		Name:        name,
		Description: description,
		StartOn:     startOn,
		DueOn:       dueOn,
		IsParallel:  isParallel,
		IsPublic:    isPublic,
		Members:     members,
	}, nil, &project{})
	if val != nil {
		return val.(*project), e
	}
	return nil, e
}

func (c *client) SetIsPublic(css *clientsession.Store, shard int, accountId, projectId id.Id, isPublic bool) error {
	_, e := setIsPublic.DoRequest(css, c.host, &setIsPublicArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		IsPublic:  isPublic,
	}, nil, nil)
	return e
}

func (c *client) SetIsArchived(css *clientsession.Store, shard int, accountId, projectId id.Id, isArchived bool) error {
	_, e := setIsArchived.DoRequest(css, c.host, &setIsArchivedArgs{
		Shard:      shard,
		AccountId:  accountId,
		ProjectId:  projectId,
		IsArchived: isArchived,
	}, nil, nil)
	return e
}

func (c *client) GetProject(css *clientsession.Store, shard int, accountId, projectId id.Id) (*project, error) {
	val, e := getProject.DoRequest(css, c.host, &getProjectArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, &project{})
	if val != nil {
		return val.(*project), e
	}
	return nil, e
}

func (c *client) GetProjects(css *clientsession.Store, shard int, accountId id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) (*getProjectsResp, error) {
	val, e := getProjects.DoRequest(css, c.host, &getProjectsArgs{
		Shard:           shard,
		AccountId:       accountId,
		NameContains:    nameContains,
		CreatedOnAfter:  createdOnAfter,
		CreatedOnBefore: createdOnBefore,
		StartOnAfter:    startOnAfter,
		StartOnBefore:   startOnBefore,
		DueOnAfter:      dueOnAfter,
		DueOnBefore:     dueOnBefore,
		IsArchived:      isArchived,
		SortBy:          sortBy,
		SortDir:         sortDir,
		After:           after,
		Limit:           limit,
	}, nil, &getProjectsResp{})
	if val != nil {
		return val.(*getProjectsResp), e
	}
	return nil, e
}

func (c *client) DeleteProject(css *clientsession.Store, shard int, accountId, projectId id.Id) error {
	_, e := deleteProject.DoRequest(css, c.host, &deleteProjectArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, members []*AddProjectMember) error {
	_, e := addMembers.DoRequest(css, c.host, &addMembersArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		Members:   members,
	}, nil, nil)
	return e
}

func (c *client) SetMemberRole(css *clientsession.Store, shard int, accountId, projectId, member id.Id, role cnst.ProjectRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		Member:    member,
		Role:      role,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, members []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, &removeMembersArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		Members:   members,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, shard int, accountId, projectId id.Id, role *cnst.ProjectRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error) {
	val, e := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard:        shard,
		AccountId:    accountId,
		ProjectId:    projectId,
		Role:         role,
		NameContains: nameContains,
		After:        after,
		Limit:        limit,
	}, nil, &getMembersResp{})
	if val != nil {
		return val.(*getMembersResp), e
	}
	return nil, e
}

func (c *client) GetMe(css *clientsession.Store, shard int, accountId, projectId id.Id) (*member, error) {
	val, e := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
	}, nil, &member{})
	if val != nil {
		return val.(*member), e
	}
	return nil, e
}

func (c *client) GetActivities(css *clientsession.Store, shard int, accountId, projectId id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard:          shard,
		AccountId:      accountId,
		ProjectId:      projectId,
		Item:           item,
		Member:         member,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		Limit:          limit,
	}, nil, &[]*activity.Activity{})
	if val != nil {
		return *val.(*[]*activity.Activity), e
	}
	return nil, e
}

func dbGetProjectExists(ctx ctx.Ctx, shard int, accountId, projectId id.Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, accountId, projectId)
	exists := false
	err.PanicIf(row.Scan(&exists))
	return exists
}

func dbCreateProject(ctx ctx.Ctx, shard int, accountId, myId id.Id, project *project) {
	_, e := ctx.TreeExec(shard, `CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, accountId, project.Id, myId, project.Name, project.Description, project.CreatedOn, project.StartOn, project.DueOn, project.IsParallel, project.IsPublic)
	err.PanicIf(e)
}

func dbSetIsPublic(ctx ctx.Ctx, shard int, accountId, projectId, myId id.Id, isPublic bool) {
	_, e := ctx.TreeExec(shard, `CALL setProjectIsPublic(?, ?, ?, ?)`, accountId, projectId, myId, isPublic)
	err.PanicIf(e)
}

func dbGetProject(ctx ctx.Ctx, shard int, accountId, projectId id.Id) *project {
	row := ctx.TreeQueryRow(shard, `SELECT p.id, p.isArchived, p.name, p.createdOn, p.startOn, p.dueOn, p.fileCount, p.fileSize, p.isPublic, t.description, t.totalRemainingTime, t.totalLoggedTime, t.minimumRemainingTime, t.linkedFileCount, t.chatCount, t.childCount, t.descendantCount, t.isParallel FROM projects p, tasks t WHERE p.account=? AND p.id=? AND t.account=? AND t.project=? AND t.id=?`, accountId, projectId, accountId, projectId, projectId)
	result := project{}
	err.PanicIf(row.Scan(&result.Id, &result.IsArchived, &result.Name, &result.CreatedOn, &result.StartOn, &result.DueOn, &result.FileCount, &result.FileSize, &result.IsPublic, &result.Description, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.LinkedFileCount, &result.ChatCount, &result.ChildCount, &result.DescendantCount, &result.IsParallel))
	return &result
}

func dbGetPublicProjects(ctx ctx.Ctx, shard int, accountId id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND isPublic=true`, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetPublicAndSpecificAccessProjects(ctx ctx.Ctx, shard int, accountId, myId id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, accountId, &myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetAllProjects(ctx ctx.Ctx, shard int, accountId id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, ``, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbSetProjectIsArchived(ctx ctx.Ctx, shard int, accountId, projectId, myId id.Id, isArchived bool) {
	_, e := ctx.TreeExec(shard, `CALL setProjectIsArchived(?, ?, ?, ?)`, accountId, projectId, myId, isArchived)
	err.PanicIf(e)
}

func dbDeleteProject(ctx ctx.Ctx, shard int, accountId, projectId, myId id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteProject(?, ?, ?)`, accountId, projectId, myId)
	err.PanicIf(e)
}

func dbAddMemberOrSetActive(ctx ctx.Ctx, shard int, accountId, projectId, myId id.Id, member *AddProjectMember) {
	db.MakeChangeHelper(ctx, shard, `CALL addProjectMemberOrSetActive(?, ?, ?, ?, ?)`, accountId, projectId, myId, member.Id, member.Role)
}

func dbSetMemberRole(ctx ctx.Ctx, shard int, accountId, projectId, myId, member id.Id, role cnst.ProjectRole) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberRole(?, ?, ?, ?, ?)`, accountId, projectId, myId, member, role)
}

func dbSetMemberInactive(ctx ctx.Ctx, shard int, accountId, projectId, myId id.Id, member id.Id) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberInactive(?, ?, ?, ?)`, accountId, projectId, myId, member)
}

func dbGetMembers(ctx ctx.Ctx, shard int, accountId, projectId id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) *getMembersResp {
	query := bytes.NewBufferString(`SELECT p1.id, p1.isActive, p1.totalRemainingTime, p1.totalLoggedTime, p1.role FROM projectMembers p1`)
	args := make([]interface{}, 0, 9)
	if after != nil {
		query.WriteString(`, projectMembers p2`)
	}
	query.WriteString(` WHERE p1.account=? AND p1.project=? AND p1.isActive=true`)
	args = append(args, accountId, projectId)
	if after != nil {
		query.WriteString(` AND p2.account=? AND p2.project=? p2.id=? AND ((p1.name>p2.name AND p1.role=p2.role) OR p1.role>p2.role)`)
		args = append(args, accountId, projectId, *after)
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
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		err.PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return &getMembersResp{Members: res[:limit], More: true}
	}
	return &getMembersResp{Members: res, More: false}
}

func dbGetMember(ctx ctx.Ctx, shard int, accountId, projectId, memberId id.Id) *member {
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, accountId, projectId, memberId)
	res := member{}
	err.PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
	return &res
}

func dbGetActivities(ctx ctx.Ctx, shard int, accountId, projectId id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) []*activity.Activity {
	if occurredAfter != nil && occurredBefore != nil {
		panic(err.InvalidArguments)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, extraInfo FROM projectActivities WHERE account=? AND project=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, accountId, projectId)
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, *item)
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, *member)
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
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*activity.Activity, 0, limit)
	for rows.Next() {
		act := activity.Activity{}
		err.PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	return res
}

func dbGetProjects(ctx ctx.Ctx, shard int, specificSqlFilterTxt string, accountId id.Id, myId *id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	query := bytes.NewBufferString(`SELECT id, isArchived, name, createdOn, startOn, dueOn, fileCount, fileSize, isPublic FROM projects WHERE account=? AND isArchived=? %s`)
	args := make([]interface{}, 0, 14)
	args = append(args, accountId, isArchived)
	if myId != nil {
		args = append(args, accountId, *myId)
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
		args = append(args, accountId, *after, *after)
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s, id LIMIT ?`, sortBy, sortDir))
	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, fmt.Sprintf(query.String(), specificSqlFilterTxt), args...)
	err.PanicIf(e)
	res := make([]*project, 0, limit+1)
	idx := 0
	resIdx := map[string]int{}
	for rows.Next() {
		proj := project{}
		err.PanicIf(rows.Scan(&proj.Id, &proj.IsArchived, &proj.Name, &proj.CreatedOn, &proj.StartOn, &proj.DueOn, &proj.FileCount, &proj.FileSize, &proj.IsPublic))
		res = append(res, &proj)
		resIdx[proj.Id.String()] = idx
		idx++
	}
	if len(res) > 0 { //populate task properties
		var i id.Id
		var description *string
		var totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount uint64
		var isParallel bool
		query.Reset()
		args = make([]interface{}, 0, len(res)+1)
		args = append(args, accountId, res[0].Id)
		query.WriteString(`SELECT id, description, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel FROM tasks WHERE account=? AND project=id AND project IN (?`)
		for _, proj := range res[1:] {
			query.WriteString(`,?`)
			args = append(args, proj.Id)
		}
		query.WriteString(fmt.Sprintf(`) LIMIT %d`, len(res)))
		rows, e := ctx.TreeQuery(shard, query.String(), args...)
		err.PanicIf(e)
		for rows.Next() {
			rows.Scan(&i, &description, &totalRemainingTime, &totalLoggedTime, &minimumRemainingTime, &linkedFileCount, &chatCount, &childCount, &descendantCount, &isParallel)
			proj := res[resIdx[i.String()]]
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
	Id                 id.Id            `json:"id"`
	TotalRemainingTime uint64           `json:"totalRemainingTime"`
	TotalLoggedTime    uint64           `json:"totalLoggedTime"`
	IsActive           bool             `json:"isActive"`
	Role               cnst.ProjectRole `json:"role"`
}

type project struct {
	Id                   id.Id      `json:"id"`
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

type AddProjectMember struct {
	Id   id.Id            `json:"id"`
	Role cnst.ProjectRole `json:"role"`
}
