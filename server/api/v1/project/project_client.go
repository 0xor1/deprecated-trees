package project

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
	"time"
)

type Client interface {
	//must be account owner/admin
	CreateProject(css *clientsession.Store, shard int, account id.Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error)
	//must be account owner/admin and account.publicProjectsEnabled must be true
	SetIsPublic(css *clientsession.Store, shard int, account, project id.Id, isPublic bool) error
	//must be account owner/admin
	SetIsArchived(css *clientsession.Store, shard int, account, project id.Id, isArchived bool) error
	//check project access permission per user
	GetProject(css *clientsession.Store, shard int, account, project id.Id) (*project, error)
	//check project access permission per user
	GetProjects(css *clientsession.Store, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) (*getProjectsResp, error)
	//must be account owner/admin
	DeleteProject(css *clientsession.Store, shard int, account, project id.Id) error
	//must be account owner/admin or project admin
	AddMembers(css *clientsession.Store, shard int, account, project id.Id, members []*AddProjectMember) error
	//must be account owner/admin or project admin
	SetMemberRole(css *clientsession.Store, shard int, account, project id.Id, member id.Id, role cnst.ProjectRole) error
	//must be account owner/admin or project admin
	RemoveMembers(css *clientsession.Store, shard int, account, project id.Id, members []id.Id) error
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(css *clientsession.Store, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) (*getMembersResp, error)
	//for anyone
	GetMe(css *clientsession.Store, shard int, account, project id.Id) (*member, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateProject(css *clientsession.Store, shard int, account id.Id, name string, description *string, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*project, error) {
	val, e := createProject.DoRequest(css, c.host, &createProjectArgs{
		Shard:       shard,
		Account:     account,
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

func (c *client) SetIsPublic(css *clientsession.Store, shard int, account, project id.Id, isPublic bool) error {
	_, e := setIsPublic.DoRequest(css, c.host, &setIsPublicArgs{
		Shard:    shard,
		Account:  account,
		Project:  project,
		IsPublic: isPublic,
	}, nil, nil)
	return e
}

func (c *client) SetIsArchived(css *clientsession.Store, shard int, account, project id.Id, isArchived bool) error {
	_, e := setIsArchived.DoRequest(css, c.host, &setIsArchivedArgs{
		Shard:      shard,
		Account:    account,
		Project:    project,
		IsArchived: isArchived,
	}, nil, nil)
	return e
}

func (c *client) GetProject(css *clientsession.Store, shard int, account, proj id.Id) (*project, error) {
	val, e := getProject.DoRequest(css, c.host, &getProjectArgs{
		Shard:   shard,
		Account: account,
		Project: proj,
	}, nil, &project{})
	if val != nil {
		return val.(*project), e
	}
	return nil, e
}

func (c *client) GetProjects(css *clientsession.Store, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) (*getProjectsResp, error) {
	val, e := getProjects.DoRequest(css, c.host, &getProjectsArgs{
		Shard:           shard,
		Account:         account,
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

func (c *client) DeleteProject(css *clientsession.Store, shard int, account, project id.Id) error {
	_, e := deleteProject.DoRequest(css, c.host, &deleteProjectArgs{
		Shard:   shard,
		Account: account,
		Project: project,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, shard int, account, project id.Id, members []*AddProjectMember) error {
	_, e := addMembers.DoRequest(css, c.host, &addMembersArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Members: members,
	}, nil, nil)
	return e
}

func (c *client) SetMemberRole(css *clientsession.Store, shard int, account, project, member id.Id, role cnst.ProjectRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Member:  member,
		Role:    role,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, shard int, account, project id.Id, members []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, &removeMembersArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Members: members,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) (*getMembersResp, error) {
	val, e := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard:                     shard,
		Account:                   account,
		Project:                   project,
		Role:                      role,
		NameOrDisplayNameContains: nameOrDisplayNameContains,
		After:                     after,
		Limit:                     limit,
	}, nil, &getMembersResp{})
	if val != nil {
		return val.(*getMembersResp), e
	}
	return nil, e
}

func (c *client) GetMe(css *clientsession.Store, shard int, account, project id.Id) (*member, error) {
	val, e := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard:   shard,
		Account: account,
		Project: project,
	}, nil, &member{})
	if val != nil {
		return val.(*member), e
	}
	return nil, e
}

func (c *client) GetActivities(css *clientsession.Store, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard:          shard,
		Account:        account,
		Project:        project,
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
