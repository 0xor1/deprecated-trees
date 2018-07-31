package account

import (
	"bitbucket.org/0xor1/trees/server/util/account"
	"bitbucket.org/0xor1/trees/server/util/activity"
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/id"
	"time"
)

type Client interface {
	//must be account owner
	Edit(css *clientsession.Store, region cnst.Region, shard int, account id.Id, fields Fields) error
	//must be account owner/admin
	Get(css *clientsession.Store, region cnst.Region, shard int, account id.Id) (*account.Account, error)
	//must be account owner/admin
	SetMemberRole(css *clientsession.Store, region cnst.Region, shard int, account, member id.Id, role cnst.AccountRole) error
	//pointers are optional filters
	GetMembers(css *clientsession.Store, region cnst.Region, shard int, account id.Id, role *cnst.AccountRole, nameOrDisplayNamePrefix *string, after *id.Id, limit int) (*GetMembersResp, error)
	//either one or both of OccurredAfter/Before must be nil
	GetActivities(css *clientsession.Store, region cnst.Region, shard int, account id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error)
	//for anyone
	GetMe(css *clientsession.Store, region cnst.Region, shard int, account id.Id) (*Member, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Edit(css *clientsession.Store, region cnst.Region, shard int, account id.Id, fields Fields) error {
	_, e := edit.DoRequest(css, c.host, region, &editArgs{
		Shard:   shard,
		Account: account,
		Fields:  fields,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, region cnst.Region, shard int, acc id.Id) (*account.Account, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: acc,
	}, nil, &account.Account{})
	if val != nil {
		return val.(*account.Account), e
	}
	return nil, e
}

func (c *client) SetMemberRole(css *clientsession.Store, region cnst.Region, shard int, account, member id.Id, role cnst.AccountRole) error {
	_, e := setMemberRole.DoRequest(css, c.host, region, &setMemberRoleArgs{
		Shard:   shard,
		Account: account,
		Member:  member,
		Role:    role,
	}, nil, nil)
	return e
}

func (c *client) GetMembers(css *clientsession.Store, region cnst.Region, shard int, account id.Id, role *cnst.AccountRole, nameOrDisplayNamePrefix *string, after *id.Id, limit int) (*GetMembersResp, error) {
	val, e := getMembers.DoRequest(css, c.host, region, &getMembersArgs{
		Shard:   shard,
		Account: account,
		Role:    role,
		NameOrDisplayNamePrefix: nameOrDisplayNamePrefix,
		After: after,
		Limit: limit,
	}, nil, &GetMembersResp{})
	if val != nil {
		return val.(*GetMembersResp), e
	}
	return nil, e
}

func (c *client) GetActivities(css *clientsession.Store, region cnst.Region, shard int, account id.Id, itemId *id.Id, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) ([]*activity.Activity, error) {
	val, e := getActivities.DoRequest(css, c.host, region, &getActivitiesArgs{
		Shard:          shard,
		Account:        account,
		Item:           itemId,
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

func (c *client) GetMe(css *clientsession.Store, region cnst.Region, shard int, account id.Id) (*Member, error) {
	val, e := getMe.DoRequest(css, c.host, region, &getMeArgs{
		Shard:   shard,
		Account: account,
	}, nil, &Member{})
	if val != nil {
		return val.(*Member), e
	}
	return nil, e
}
