package project

import (
	"bitbucket.org/0xor1/trees/server/util/activity"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	t "bitbucket.org/0xor1/trees/server/util/time"
	"bitbucket.org/0xor1/trees/server/util/validate"
	"github.com/0xor1/panic"
	"net/http"
	"time"
	"bitbucket.org/0xor1/trees/server/util/field"
)

var (
	publicProjectsDisabledErr = &err.Err{Code: "r_v1_p_ppd", Message: "public projects disabled"}
)

type createArgs struct {
	Shard       int                 `json:"shard"`
	Account     id.Id               `json:"account"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	HoursPerDay uint8             `json:"hoursPerDay"`
	DaysPerWeek uint8             `json:"daysPerWeek"`
	StartOn     *time.Time          `json:"startOn"`
	DueOn       *time.Time          `json:"dueOn"`
	IsParallel  bool                `json:"isParallel"`
	IsPublic    bool                `json:"isPublic"`
	Members     []*AddProjectMember `json:"members"`
}

var create = &endpoint.Endpoint{
	Method:                   http.MethodPost,
	Path:                     "/api/v1/project/create",
	RequiresSession:          true,
	ExampleResponseStructure: &Project{},
	GetArgsStruct: func() interface{} {
		return &createArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		panic.IfTrue(args.IsPublic && !db.GetAccount(ctx, args.Shard, args.Account).PublicProjectsEnabled, publicProjectsDisabledErr)

		validate.HoursPerDay(args.HoursPerDay)
		validate.DaysPerWeek(args.DaysPerWeek)

		project := &Project{}
		project.Id = id.New()
		project.Name = args.Name
		project.HoursPerDay = args.HoursPerDay
		project.DaysPerWeek = args.DaysPerWeek
		project.Description = args.Description
		project.CreatedOn = t.Now()
		project.StartOn = args.StartOn
		project.DueOn = args.DueOn
		project.IsParallel = args.IsParallel
		project.IsPublic = args.IsPublic
		dbCreateProject(ctx, args.Shard, args.Account, project)
		if args.Account.Equal(ctx.Me()) {
			addMem := &AddProjectMember{}
			addMem.Id = ctx.Me()
			addMem.Role = cnst.ProjectAdmin
			dbAddMemberOrSetActive(ctx, args.Shard, args.Account, project.Id, addMem)
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

type editArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Project id.Id  `json:"project"`
	Fields  Fields `json:"fields"`
}

var edit = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/project/edit",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &editArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*editArgs)
		accRole, projRole := db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me())
		if args.Fields.IsPublic != nil || args.Fields.IsArchived != nil {
			//must be account owner/admin to set these fields
			validate.MemberHasAccountAdminAccess(accRole)
		} else {
			//other fields only require project admin access
			validate.MemberHasProjectAdminAccess(accRole, projRole)
		}
		if args.Fields.HoursPerDay != nil {
			validate.HoursPerDay(args.Fields.HoursPerDay.Val)
		}
		if args.Fields.DaysPerWeek != nil {
			validate.DaysPerWeek(args.Fields.DaysPerWeek.Val)
		}
		panic.IfTrue(args.Fields.IsPublic != nil && args.Fields.IsPublic.Val && !db.GetAccount(ctx, args.Shard, args.Account).PublicProjectsEnabled, publicProjectsDisabledErr)
		dbEdit(ctx, args.Shard, args.Account, args.Project, args.Fields)
		return nil
	},
}

type getArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var get = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/project/get",
	RequiresSession:          false,
	ExampleResponseStructure: &Project{},
	GetArgsStruct: func() interface{} {
		return &getArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))

		return dbGetProject(ctx, args.Shard, args.Account, args.Project)
	},
}

type getSetArgs struct {
	Shard           int         `json:"shard"`
	Account         id.Id       `json:"account"`
	NameContains    *string     `json:"nameContains"`
	CreatedOnAfter  *time.Time  `json:"createdOnAfter"`
	CreatedOnBefore *time.Time  `json:"createdOnBefore"`
	StartOnAfter    *time.Time  `json:"startOnAfter"`
	StartOnBefore   *time.Time  `json:"startOnBefore"`
	DueOnAfter      *time.Time  `json:"dueOnAfter"`
	DueOnBefore     *time.Time  `json:"dueOnBefore"`
	IsArchived      bool        `json:"isArchived"`
	SortBy          cnst.SortBy `json:"sortBy"`
	SortAsc         bool        `json:"sortAsc"`
	After           *id.Id      `json:"after"`
	Limit           int         `json:"limit"`
}

type GetSetResult struct {
	Projects []*Project `json:"projects"`
	More     bool       `json:"more"`
}

var getSet = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/project/getSet",
	RequiresSession:          false,
	ExampleResponseStructure: &GetSetResult{Projects: []*Project{{}}},
	GetArgsStruct: func() interface{} {
		return &getSetArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getSetArgs)
		var myAccountRole *cnst.AccountRole
		if ctx.TryMe() != nil {
			myAccountRole = db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me())
		}
		args.Limit = validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		if myAccountRole == nil {
			return dbGetPublicProjects(ctx, args.Shard, args.Account, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortAsc, args.After, args.Limit)
		}
		if *myAccountRole != cnst.AccountOwner && *myAccountRole != cnst.AccountAdmin {
			return dbGetPublicAndSpecificAccessProjects(ctx, args.Shard, args.Account, ctx.Me(), args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortAsc, args.After, args.Limit)
		}
		return dbGetAllProjects(ctx, args.Shard, args.Account, args.NameContains, args.CreatedOnAfter, args.CreatedOnBefore, args.StartOnAfter, args.StartOnBefore, args.DueOnAfter, args.DueOnBefore, args.IsArchived, args.SortBy, args.SortAsc, args.After, args.Limit)
	},
}

type deleteArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var delete = &endpoint.Endpoint{
	Method:          http.MethodPost,
	Path:            "/api/v1/project/delete",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &deleteArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteArgs)
		validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.Account, ctx.Me()))
		dbDeleteProject(ctx, args.Shard, args.Account, args.Project)
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
	Method:          http.MethodPost,
	Path:            "/api/v1/project/addMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		panic.IfTrue(args.Account.Equal(ctx.Me()), err.InvalidOperation)

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))
		validate.Exists(dbGetProjectExists(ctx, args.Shard, args.Account, args.Project))

		for _, mem := range args.Members {
			mem.Role.Validate()
			accRole := db.GetAccountRole(ctx, args.Shard, args.Account, mem.Id)
			panic.IfTrue(accRole == nil, err.InvalidArguments)
			if *accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin {
				mem.Role = cnst.ProjectAdmin // account owners and admins cant be added to projects with privelages less than project admin
			}
			dbAddMemberOrSetActive(ctx, args.Shard, args.Account, args.Project, mem)
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
	Method:          http.MethodPost,
	Path:            "/api/v1/project/setMemberRole",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &setMemberRoleArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberRoleArgs)
		panic.IfTrue(args.Account.Equal(ctx.Me()), err.InvalidOperation)
		args.Role.Validate()

		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		accRole, projectRole := db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, args.Member)
		panic.IfTrue(projectRole == nil, err.InvalidOperation)
		if *projectRole != args.Role {
			panic.IfTrue(args.Role != cnst.ProjectAdmin && (*accRole == cnst.AccountOwner || *accRole == cnst.AccountAdmin), err.InvalidArguments) // account owners and admins can only be project admins
			dbSetMemberRole(ctx, args.Shard, args.Account, args.Project, args.Member, args.Role)
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
	Method:          http.MethodPost,
	Path:            "/api/v1/project/removeMembers",
	RequiresSession: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		panic.IfTrue(args.Account.Equal(ctx.Me()), err.InvalidOperation)
		validate.MemberHasProjectAdminAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.Account, args.Project, ctx.Me()))

		for _, mem := range args.Members {
			dbSetMemberInactive(ctx, args.Shard, args.Account, args.Project, mem)
		}
		return nil
	},
}

type getMembersArgs struct {
	Shard                     int               `json:"shard"`
	Account                   id.Id             `json:"account"`
	Project                   id.Id             `json:"project"`
	Role                      *cnst.ProjectRole `json:"role,omitempty"`
	NameOrDisplayNameContains *string           `json:"nameorDisplayNameContains,omitempty"`
	After                     *id.Id            `json:"after,omitempty"`
	Limit                     int               `json:"limit"`
}

type GetMembersResult struct {
	Members []*member `json:"members"`
	More    bool      `json:"more"`
}

var getMembers = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/project/getMembers",
	RequiresSession:          false,
	ExampleResponseStructure: &GetMembersResult{Members: []*member{{}}},
	GetArgsStruct: func() interface{} {
		return &getMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getMembersArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		return dbGetMembers(ctx, args.Shard, args.Account, args.Project, args.Role, args.NameOrDisplayNameContains, args.After, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

type getMeArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Project id.Id `json:"project"`
}

var getMe = &endpoint.Endpoint{
	Method:                   http.MethodGet,
	Path:                     "/api/v1/project/getMe",
	RequiresSession:          true,
	ExampleResponseStructure: &member{},
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
	Method:                   http.MethodGet,
	Path:                     "/api/v1/project/getActivities",
	RequiresSession:          false,
	ExampleResponseStructure: []*activity.Activity{{}},
	GetArgsStruct: func() interface{} {
		return &getActivitiesArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*getActivitiesArgs)
		panic.IfTrue(args.OccurredAfter != nil && args.OccurredBefore != nil, err.InvalidArguments)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.Account, args.Project, ctx.TryMe()))
		return dbGetActivities(ctx, args.Shard, args.Account, args.Project, args.Item, args.Member, args.OccurredAfter, args.OccurredBefore, validate.Limit(args.Limit, ctx.MaxProcessEntityCount()))
	},
}

var Endpoints = []*endpoint.Endpoint{
	create,
	edit,
	get,
	getSet,
	delete,
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

type Project struct {
	Id                   id.Id      `json:"id"`
	IsArchived           bool       `json:"isArchived"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description"`
	HoursPerDay          uint8   `json:"hoursPerDay"`
	DaysPerWeek          uint8    `json:"daysPerWeek"`
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
	IsParallel           bool       `json:"isParallel"`
	IsPublic             bool       `json:"isPublic"`
}

type Fields struct {
	// account owner/admin
	IsPublic             *field.Bool    `json:"isPublic,omitempty"`
	// account owner/admin
	IsArchived           *field.Bool    `json:"isArchived,omitempty"`
	// account owner/admin or project admin
	HoursPerDay			 *field.UInt8	`json:"hoursPerDay,omitempty"`
	// account owner/admin or project admin
	DaysPerWeek			 *field.UInt8	`json:"daysPerWeek,omitempty"`
	// account owner/admin or project admin
	StartOn              *field.TimePtr `json:"startOn,omitempty"`
	// account owner/admin or project admin
	DueOn                *field.TimePtr `json:"dueOn,omitempty"`
}

type AddProjectMember struct {
	Id   id.Id            `json:"id"`
	Role cnst.ProjectRole `json:"role"`
}
