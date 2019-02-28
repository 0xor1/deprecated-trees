package api

import (
	"github.com/0xor1/trees/server/api/v1/account"
	"github.com/0xor1/trees/server/api/v1/central"
	"github.com/0xor1/trees/server/api/v1/project"
	"github.com/0xor1/trees/server/api/v1/task"
	"github.com/0xor1/trees/server/api/v1/timelog"
	utilaccount "github.com/0xor1/trees/server/util/account"
	"github.com/0xor1/trees/server/util/activity"
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
	tlog "github.com/0xor1/trees/server/util/timelog"
	"io"
	"time"
)

// API is a helper struct to simplify making calls to the trees backend.
// It enables developers to create an api instance once for a single user and not
// have to manually pass around the clientsessionstore into every call.
type API struct {
	Me *central.Me
	V1 *V1
}

type V1 struct {
	Central *centralClient
	Account *accountClient
	Project *projectClient
	Task    *taskClient
	TimeLog *timeLogClient
}

type centralClient struct {
	css    *clientsession.Store
	client central.Client
}

func (c *centralClient) GetMe() (*central.Me, error) {
	return c.client.GetMe(c.css)
}

func (c *centralClient) SetMyPwd(oldPwd, newPwd string) error {
	return c.client.SetMyPwd(c.css, oldPwd, newPwd)
}

func (c *centralClient) SetMyEmail(newEmail string) error {
	return c.client.SetMyEmail(c.css, newEmail)
}

func (c *centralClient) ResendMyNewEmailConfirmationEmail() error {
	return c.client.ResendMyNewEmailConfirmationEmail(c.css)
}

func (c *centralClient) SetAccountName(account id.Id, newName string) error {
	return c.client.SetAccountName(c.css, account, newName)
}

func (c *centralClient) SetAccountDisplayName(account id.Id, newDisplayName *string) error {
	return c.client.SetAccountDisplayName(c.css, account, newDisplayName)
}

func (c *centralClient) SetAccountAvatar(account id.Id, avatar io.ReadCloser) error {
	return c.client.SetAccountAvatar(c.css, account, avatar)
}

func (c *centralClient) MigrateAccount(account id.Id, newRegion cnst.Region) error {
	return c.client.MigrateAccount(c.css, account, newRegion)
}

func (c *centralClient) CreateAccount(region cnst.Region, name string, displayName *string) (*central.Account, error) {
	return c.client.CreateAccount(c.css, region, name, displayName)
}

func (c *centralClient) GetMyAccounts(after *id.Id, limit int) (*central.GetMyAccountsResult, error) {
	return c.client.GetMyAccounts(c.css, after, limit)
}

func (c *centralClient) DeleteAccount(account id.Id) error {
	return c.client.DeleteAccount(c.css, account)
}

func (c *centralClient) AddMembers(account id.Id, newMembers []*central.AddMember) error {
	return c.client.AddMembers(c.css, account, newMembers)
}

func (c *centralClient) RemoveMembers(account id.Id, existingMembers []id.Id) error {
	return c.client.RemoveMembers(c.css, account, existingMembers)
}

type accountClient struct {
	css    *clientsession.Store
	client account.Client
}

func (c *accountClient) Edit(region cnst.Region, shard int, account id.Id, fields account.Fields) error {
	return c.client.Edit(c.css, region, shard, account, fields)
}

func (c *accountClient) Get(region cnst.Region, shard int, account id.Id) (*utilaccount.Account, error) {
	return c.client.Get(c.css, region, shard, account)
}

func (c *accountClient) SetMemberRole(region cnst.Region, shard int, account, member id.Id, role cnst.AccountRole) error {
	return c.client.SetMemberRole(c.css, region, shard, account, member, role)
}

func (c *accountClient) GetMembers(region cnst.Region, shard int, account id.Id, role *cnst.AccountRole, nameOrDisplayNamePrefix *string, after *id.Id, limit int) (*account.GetMembersResp, error) {
	return c.client.GetMembers(c.css, region, shard, account, role, nameOrDisplayNamePrefix, after, limit)
}

func (c *accountClient) GetActivities(region cnst.Region, shard int, account id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	return c.client.GetActivities(c.css, region, shard, account, item, member, occurredAfter, occurredBefore, limit)
}

func (c *accountClient) GetMe(region cnst.Region, shard int, account id.Id) (*account.Member, error) {
	return c.client.GetMe(c.css, region, shard, account)
}

type projectClient struct {
	css    *clientsession.Store
	client project.Client
}

func (c *projectClient) Create(region cnst.Region, shard int, account id.Id, name string, description *string, hoursPerDay, daysPerWeek uint8, startOn, dueOn *time.Time, isParallel, isPublic bool, members []*project.AddProjectMember) (*project.Project, error) {
	return c.client.Create(c.css, region, shard, account, name, description, hoursPerDay, daysPerWeek, startOn, dueOn, isParallel, isPublic, members)
}

func (c *projectClient) Edit(region cnst.Region, shard int, account, project id.Id, fields project.Fields) error {
	return c.client.Edit(c.css, region, shard, account, project, fields)
}

func (c *projectClient) Get(region cnst.Region, shard int, account, project id.Id) (*project.Project, error) {
	return c.client.Get(c.css, region, shard, account, project)
}

func (c *projectClient) GetSet(region cnst.Region, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) (*project.GetSetResult, error) {
	return c.client.GetSet(c.css, region, shard, account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit)
}

func (c *projectClient) Delete(region cnst.Region, shard int, account, project id.Id) error {
	return c.client.Delete(c.css, region, shard, account, project)
}

func (c *projectClient) AddMembers(region cnst.Region, shard int, account, project id.Id, members []*project.AddProjectMember) error {
	return c.client.AddMembers(c.css, region, shard, account, project, members)
}

func (c *projectClient) SetMemberRole(region cnst.Region, shard int, account, project id.Id, member id.Id, role cnst.ProjectRole) error {
	return c.client.SetMemberRole(c.css, region, shard, account, project, member, role)
}

func (c *projectClient) RemoveMembers(region cnst.Region, shard int, account, project id.Id, members []id.Id) error {
	return c.client.RemoveMembers(c.css, region, shard, account, project, members)
}

func (c *projectClient) GetMembers(region cnst.Region, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) (*project.GetMembersResult, error) {
	return c.client.GetMembers(c.css, region, shard, account, project, role, nameOrDisplayNameContains, after, limit)
}

func (c *projectClient) GetAtMentions(region cnst.Region, shard int, account, project id.Id, nameOrDisplayNamePrefix string) ([]*project.Member, error) {
	return c.client.GetAtMentions(c.css, region, shard, account, project, nameOrDisplayNamePrefix)
}

func (c *projectClient) GetMe(region cnst.Region, shard int, account, project id.Id) (*project.Member, error) {
	return c.client.GetMe(c.css, region, shard, account, project)
}

func (c *projectClient) GetActivities(region cnst.Region, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	return c.client.GetActivities(c.css, region, shard, account, project, item, member, occurredAfter, occurredBefore, limit)
}

type taskClient struct {
	css    *clientsession.Store
	client task.Client
}

func (c *taskClient) Create(region cnst.Region, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, remainingTime *uint64) (*task.Task, error) {
	return c.client.Create(c.css, region, shard, account, project, parent, previousSibling, name, description, isAbstract, isParallel, member, remainingTime)
}

func (c *taskClient) Edit(region cnst.Region, shard int, account, project, task id.Id, fields task.Fields) error {
	return c.client.Edit(c.css, region, shard, account, project, task, fields)
}

func (c *taskClient) Move(region cnst.Region, shard int, account, project, task, parent id.Id, nextSibling *id.Id) error {
	return c.client.Move(c.css, region, shard, account, project, task, parent, nextSibling)
}

func (c *taskClient) Delete(region cnst.Region, shard int, account, project, task id.Id) error {
	return c.client.Delete(c.css, region, shard, account, project, task)
}

func (c *taskClient) Get(region cnst.Region, shard int, account, project id.Id, task id.Id) (*task.Task, error) {
	return c.client.Get(c.css, region, shard, account, project, task)
}

func (c *taskClient) GetChildren(region cnst.Region, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) (*task.GetChildrenResp, error) {
	return c.client.GetChildren(c.css, region, shard, account, project, parent, fromSibling, limit)
}

func (c *taskClient) GetAncestors(region cnst.Region, shard int, account, project, child id.Id, limit int) (*task.GetAncestorsResp, error) {
	return c.client.GetAncestors(c.css, region, shard, account, project, child, limit)
}

type timeLogClient struct {
	css    *clientsession.Store
	client timelog.Client
}

func (c *timeLogClient) Create(region cnst.Region, shard int, account, project, task id.Id, duration uint64, note *string) (*tlog.TimeLog, error) {
	return c.client.Create(c.css, region, shard, account, project, task, duration, note)
}

func (c *timeLogClient) CreateAndSetRemainingTime(region cnst.Region, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*tlog.TimeLog, error) {
	return c.client.CreateAndSetRemainingTime(c.css, region, shard, account, project, task, remainingTime, duration, note)
}

func (c *timeLogClient) Edit(region cnst.Region, shard int, account, project, timeLog id.Id, fields timelog.Fields) error {
	return c.client.Edit(c.css, region, shard, account, project, timeLog, fields)
}

func (c *timeLogClient) Delete(region cnst.Region, shard int, account, project, timeLog id.Id) error {
	return c.client.Delete(c.css, region, shard, account, project, timeLog)
}

func (c *timeLogClient) Get(region cnst.Region, shard int, account, project id.Id, task, member, timeLog *id.Id, sortAsc bool, after *id.Id, limit int) (*timelog.GetResp, error) {
	return c.client.Get(c.css, region, shard, account, project, task, member, timeLog, sortAsc, after, limit)
}

// New returns a new API configured for
func New(host, email, pwd string) (*API, error) {
	css := clientsession.New()
	central := central.NewClient(host)
	account := account.NewClient(host)
	project := project.NewClient(host)
	task := task.NewClient(host)
	timeLog := timelog.NewClient(host)

	authResp, err := central.Authenticate(css, email, pwd)
	if err != nil {
		return nil, err
	}

	return &API{
		Me: authResp.Me,
		V1: &V1{
			Central: &centralClient{
				css:    css,
				client: central,
			},
			Account: &accountClient{
				css:    css,
				client: account,
			},
			Project: &projectClient{
				css:    css,
				client: project,
			},
			Task: &taskClient{
				css:    css,
				client: task,
			},
			TimeLog: &timeLogClient{
				css:    css,
				client: timeLog,
			},
		},
	}, nil
}
