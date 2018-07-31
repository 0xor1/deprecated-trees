package chat

import (
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/id"
)

type Client interface {
	Create(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, content string) (*Entry, error)
	Edit(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, content string) error
	Delete(css *clientsession.Store, region cnst.Region, shard int, account, project, task, entry id.Id) error
	Get(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, after *id.Id, limit int) (*getResp, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Create(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, content string) (*Entry, error) {
	val, e := create.DoRequest(css, c.host, region, &createArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Content: content,
	}, nil, &Entry{})
	if val != nil {
		return val.(*Entry), e
	}
	return nil, e
}

func (c *client) Edit(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, content string) error {
	_, e := edit.DoRequest(css, c.host, region, &editArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Content: content,
	}, nil, nil)
	return e
}

func (c *client) Delete(css *clientsession.Store, region cnst.Region, shard int, account, project, task, entry id.Id) error {
	_, e := delete.DoRequest(css, c.host, region, &deleteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Entry:   entry,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, after *id.Id, limit int) (*getResp, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		After:   after,
		Limit:   limit,
	}, nil, &getResp{})
	if val != nil {
		return val.(*getResp), e
	}
	return nil, e
}

func (c *client) GetAtMentions(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, nameOrDisplayNamePrefix string) ([]*Member, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: account,
		Project: project,
	}, nil, &getResp{})
	if val != nil {
		return val.(*getResp), e
	}
	return nil, e
}
