package account

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
	"time"
)

type Client interface {
	//must be account owner
	SetPublicProjectsEnabled(css *clientsession.Store, shard int, account id.Id, publicProjectsEnabled bool) error
	//must be account owner/admin
	GetPublicProjectsEnabled(css *clientsession.Store, shard int, account id.Id) (bool, error)
	//must be account owner/admin
	SetMemberRole(css *clientsession.Store, shard int, account, member id.Id, role cnst.AccountRole) error
	//pointers are optional filters
	GetMembers(css *clientsession.Store, shard int, account id.Id, role *cnst.AccountRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, shard int, account id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
	//for anyone
	GetMe(css *clientsession.Store, shard int, account id.Id) (*member, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) SetPublicProjectsEnabled(css *clientsession.Store, shard int, account id.Id, publicProjectsEnabled bool) error {
	_, e := setPublicProjectsEnabled.DoRequest(css, c.host, &setPublicProjectsEnabledArgs{
		Shard:                 shard,
		Account:               account,
		PublicProjectsEnabled: publicProjectsEnabled,
	}, nil, nil)
	return e
}

func (c *client) GetPublicProjectsEnabled(css *clientsession.Store, shard int, account id.Id) (bool, error) {
	respVal := true
	val, e := getPublicProjectsEnabled.DoRequest(css, c.host, &getPublicProjectsEnabledArgs{
		Shard:   shard,
		Account: account,
	}, nil, &respVal)
	return *val.(*bool), e
}

func (c *client) SetMemberRole(css *clientsession.Store, shard int, account, member id.Id, role cnst.AccountRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, &setMemberRoleArgs{
		Shard:   shard,
		Account: account,
		Member:  member,
		Role:    role,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, shard int, account id.Id, role *cnst.AccountRole, nameContains *string, after *id.Id, limit int) (*getMembersResp, error) {
	val, e := getMembers.DoRequest(css, c.host, &getMembersArgs{
		Shard:        shard,
		Account:      account,
		Role:         role,
		NameContains: nameContains,
		After:        after,
		Limit:        limit,
	}, nil, &getMembersResp{})
	return val.(*getMembersResp), e
}

func (c *client) GetActivities(css *clientsession.Store, shard int, account id.Id, itemId *id.Id, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, &getActivitiesArgs{
		Shard:          shard,
		Account:        account,
		Item:           itemId,
		Member:         member,
		OccurredAfter:  occurredAfter,
		OccurredBefore: occurredBefore,
		Limit:          limit,
	}, nil, &[]*activity.Activity{})
	return *val.(*[]*activity.Activity), e
}

func (c *client) GetMe(css *clientsession.Store, shard int, account id.Id) (*member, error) {
	val, e := getMe.DoRequest(css, c.host, &getMeArgs{
		Shard:   shard,
		Account: account,
	}, nil, &member{})
	return val.(*member), e
}
