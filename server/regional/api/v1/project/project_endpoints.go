package project

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/endpoint"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"time"
)

var (
	publicProjectsDisabledErr = &err.Err{Code: "r_v1_p_ppd", Message: "public projects disabled"}
)

type createProjectArgs struct {
	Shard       int                 `json:"shard"`
	Account     id.Id               `json:"account"`
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
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		if args.IsPublic && !db.GetPublicProjectsEnabled(ctx, args.Shard, args.Account) {
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
		dbCreateProject(ctx, args.Shard, args.Account, ctx.Me(), project)
		if args.Account.Equal(ctx.Me()) {
			addMem := &AddProjectMember{}
			addMem.Id = ctx.Me()
			addMem.Role = cnst.ProjectAdmin
			dbAddMemberOrSetActive(ctx, args.Shard, args.Account, project.Id, ctx.Me(), addMem)
		}

		if len(args.Members) > 0 {
			addMembers.CtxHandler(ctx, &addMembersArgs{
				Shard:   args.Shard,
				Account: args.Account,
				Project: project.Id,
				Members: args.Members,
			})
		}

		return project
	},
}

type setIsPublicArgs struct {
	Shard    int   `json:"shard"`
	Account  id.Id `json:"account"`
	Project  id.Id `json:"project"`
	IsPublic bool  `json:"isPublic"`
}

var setIsPublic = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setIsPublic",
	GetArgsStruct: func() interface{} {
		return &setIsPublicArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setIsPublicArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))

		if args.IsPublic && !db.GetPublicProjectsEnabled(ctx, args.Shard, args.Account) {
			panic(publicProjectsDisabledErr)
		}

		dbSetIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me(), args.IsPublic)

		return nil
	},
}

type setIsArchivedArgs struct {
	Shard      int   `json:"shard"`
	Account    id.Id `json:"account"`
	Project    id.Id `json:"project"`
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
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		dbSetProjectIsArchived(ctx, args.Shard, args.Account, args.Project, ctx.Me(), args.IsArchived)
		return nil
	},
}

type getProjectArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var getProject = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getProject",
	GetArgsStruct: func() interface{} {
		return &getProjectArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getProjectArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		return dbGetProject(ctx, args.Shard, args.Account, args.Project)
	},
}

type getProjectsArgs struct {
	Shard           int          `json:"shard"`
	Account         id.Id        `json:"account"`
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
		myAccountRole := db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me())
		args.Limit = validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		if myAccountRole == nil {
			return dbGetPublicProjects(ctx, args.Shard, args.Account, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		if *myAccountRole != cnst.AccountOwner && *myAccountRole != cnst.AccountAdmin {
			return dbGetPublicAndSpecificAccessProjects(ctx, args.Shard, args.Account, ctx.Me(), args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
		}
		return dbGetAllProjects(ctx, args.Shard, args.Account, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortDir, args.After, args.Limit)
	},
}

type deleteProjectArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var deleteProject = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/deleteProject",
	GetArgsStruct: func() interface{} {
		return &deleteProjectArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteProjectArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		dbDeleteProject(ctx, args.Shard, args.Account, args.Project, ctx.Me())
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard   int                 `json:"shard"`
	Account id.Id               `json:"account"`
	Project id.Id               `json:"project"`
	Members []*AddProjectMember `json:"members"`
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
		if args.Account.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		validate.Exists(dbGetProjectExists(ctx, args.Shard, args.Account, args.Project))

		for _, mem := range args.Members {
			mem.Role.Validate()
			accRole := db.GetAccountRole(ctx, args.Shard, args.Account, mem.Id)
			if accRole == nil {
				panic(err.InvalidArguments)
			}
			if *accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin {
				mem.Role = cnst.ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
			}
			dbAddMemberOrSetActive(ctx, args.Shard, args.Account, args.Project, ctx.Me(), mem)
		}
		return nil
	},
}

type setMemberRoleArgs struct {
	Shard   int              `json:"shard"`
	Account id.Id            `json:"account"`
	Project id.Id            `json:"project"`
	Member  id.Id            `json:"member"`
	Role    cnst.ProjectRole `json:"role"`
}

var setMemberRole = &endpoint.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setMemberRole",
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		if args.Account.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		args.Role.Validate()

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		accRole, projectRole := db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, args.Member)
		if projectRole == nil {
			panic(err.InvalidOperation)
		}
		if *projectRole != args.Role {
			if args.Role != cnst.ProjectAdmin && (*accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin) {
				panic(err.InvalidArguments) // account owners and admins can only be project admins
			}
			dbSetMemberRole(ctx, args.Shard, args.Account, args.Project, ctx.Me(), args.Member, args.Role)
		}
		return nil
	},
}

type removeMembersArgs struct {
	Shard   int     `json:"shard"`
	Account id.Id   `json:"account"`
	Project id.Id   `json:"project"`
	Members []id.Id `json:"members"`
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
		if args.Account.Equal(ctx.Me()) {
			panic(err.InvalidOperation)
		}
		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		for _, mem := range args.Members {
			dbSetMemberInactive(ctx, args.Shard, args.Account, args.Project, ctx.Me(), mem)
		}
		return nil
	},
}

type getMembersArgs struct {
	Shard        int               `json:"shard"`
	Account      id.Id             `json:"account"`
	Project      id.Id             `json:"project"`
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
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		return dbGetMembers(ctx, args.Shard, args.Account, args.Project, args.Role, args.NameContains, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var getMe = &endpoint.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getMe",
	GetArgsStruct: func() interface{} {
		return &getMeArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMeArgs)
		return dbGetMember(ctx, args.Shard, args.Account, args.Project, ctx.Me())
	},
}

type getActivitiesArgs struct {
	Shard          int        `json:"shard"`
	Account        id.Id      `json:"account"`
	Project        id.Id      `json:"project"`
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
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		return dbGetActivities(ctx, args.Shard, args.Account, args.Project, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
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
