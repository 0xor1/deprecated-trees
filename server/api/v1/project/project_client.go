package project

import (
	"github.com/0xor1/trees/server/util/activity"
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
	"time"
)

type Client interface {
	//must be account owner/admin
	Create(css *clientsession.Store, region cnst.Region, shard int, account id.Id, name string, description *string, hoursPerDay, daysPerWeek uint8, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*Project, error)
	//see individual fields for permissions
	Edit(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, fields Fields) error
	//check project access permission per user
	Get(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id) (*Project, error)
	//check project access permission per user
	GetSet(css *clientsession.Store, region cnst.Region, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) (*GetSetResult, error)
	//must be account owner/admin
	Delete(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id) error
	//must be account owner/admin or project admin
	AddMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, members []*AddProjectMember) error
	//must be account owner/admin or project admin
	SetMemberRole(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, member id.Id, role cnst.ProjectRole) error
	//must be account owner/admin or project admin
	RemoveMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, members []id.Id) error
	//pointers are optional filters, anyone who can see a project can see all the member info for that project
	GetMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) (*GetMembersResult, error)
	//used when typing a chat message after entering @ symbol
	GetAtMentions(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, nameOrDisplayNamePrefix string) ([]*Member, error)
	//for anyone
	GetMe(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id) (*Member, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Create(css *clientsession.Store, region cnst.Region, shard int, account id.Id, name string, description *string, hoursPerDay, daysPerWeek uint8, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*AddProjectMember) (*Project, error) {
	val, e := create.DoRequest(css, c.host, region, &createArgs{
		Shard:       shard,
		Account:     account,
		Name:        name,
		Description: description,
		HoursPerDay: hoursPerDay,
		DaysPerWeek: daysPerWeek,
		StartOn:     startOn,
		DueOn:       dueOn,
		IsParallel:  isParallel,
		IsPublic:    isPublic,
		Members:     members,
	}, nil, &Project{})
	if val != nil {
		return val.(*Project), e
	}
	return nil, e
}

func (c *client) Edit(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, fields Fields) error {
	_, e := edit.DoRequest(css, c.host, region, &editArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Fields:  fields,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, region cnst.Region, shard int, account, proj id.Id) (*Project, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: account,
		Project: proj,
	}, nil, &Project{})
	if val != nil {
		return val.(*Project), e
	}
	return nil, e
}

func (c *client) GetSet(css *clientsession.Store, region cnst.Region, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) (*GetSetResult, error) {
	val, e := getSet.DoRequest(css, c.host, region, &getSetArgs{
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
		SortAsc:         sortAsc,
		After:           after,
		Limit:           limit,
	}, nil, &GetSetResult{})
	if val != nil {
		return val.(*GetSetResult), e
	}
	return nil, e
}

func (c *client) Delete(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id) error {
	_, e := delete.DoRequest(css, c.host, region, &deleteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, members []*AddProjectMember) error {
	_, e := addMembers.DoRequest(css, c.host, region, &addMembersArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Members: members,
	}, nil, nil)
	return e
}

func (c *client) SetMemberRole(css *clientsession.Store, region cnst.Region, shard int, account, project, member id.Id, role cnst.ProjectRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, region, &setMemberRoleArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Member:  member,
		Role:    role,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, members []id.Id) error {
	_, e := removeMembers.DoRequest(css, c.host, region, &removeMembersArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Members: members,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) (*GetMembersResult, error) {
	val, e := getMembers.DoRequest(css, c.host, region, &getMembersArgs{
		Shard:                     shard,
		Account:                   account,
		Project:                   project,
		Role:                      role,
		NameOrDisplayNameContains: nameOrDisplayNameContains,
		After:                     after,
		Limit:                     limit,
	}, nil, &GetMembersResult{})
	if val != nil {
		return val.(*GetMembersResult), e
	}
	return nil, e
}

func (c *client) GetAtMentions(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, nameOrDisplayNamePrefix string) ([]*Member, error) {
	val, e := getMembers.DoRequest(css, c.host, region, &getAtMentionsArgs{
		Shard:                   shard,
		Account:                 account,
		Project:                 project,
		NameOrDisplayNamePrefix: nameOrDisplayNamePrefix,
	}, nil, &[]*Member{})
	if val != nil {
		return *val.(*[]*Member), e
	}
	return nil, e
}

func (c *client) GetMe(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id) (*Member, error) {
	val, e := getMe.DoRequest(css, c.host, region, &getMeArgs{
		Shard:   shard,
		Account: account,
		Project: project,
	}, nil, &Member{})
	if val != nil {
		return val.(*Member), e
	}
	return nil, e
}

func (c *client) GetActivities(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, region, &getActivitiesArgs{
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
