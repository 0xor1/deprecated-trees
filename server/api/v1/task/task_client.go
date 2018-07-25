package task

import (
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/id"
)

type Client interface {
	Create(css *clientsession.Store, region cnst.Region, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, remainingTime *uint64) (*Task, error)
	Edit(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, fields Fields) error
	Move(css *clientsession.Store, region cnst.Region, shard int, account, project, task, parent id.Id, nextSibling *id.Id) error
	Delete(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id) error
	Get(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, task id.Id) (*Task, error)
	GetChildren(css *clientsession.Store, region cnst.Region, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) (*getChildrenResp, error)
	GetAncestors(css *clientsession.Store, region cnst.Region, shard int, account, project, child id.Id, limit int) (*getAncestorsResp, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Create(css *clientsession.Store, region cnst.Region, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, totalRemainingTime *uint64) (*Task, error) {
	val, e := create.DoRequest(css, c.host, region, &createArgs{
		Shard:              shard,
		Account:            account,
		Project:            project,
		Parent:             parent,
		PreviousSibling:    previousSibling,
		Name:               name,
		Description:        description,
		IsAbstract:         isAbstract,
		IsParallel:         isParallel,
		Member:             member,
		TotalRemainingTime: totalRemainingTime,
	}, nil, &Task{})
	if val != nil {
		return val.(*Task), e
	}
	return nil, e
}

func (c *client) Edit(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id, fields Fields) error {
	_, e := edit.DoRequest(css, c.host, region, &editArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Fields:  fields,
	}, nil, nil)
	return e
}

func (c *client) Move(css *clientsession.Store, region cnst.Region, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) error {
	_, e := move.DoRequest(css, c.host, region, &moveArgs{
		Shard:              shard,
		Account:            account,
		Project:            project,
		Task:               task,
		NewParent:          newParent,
		NewPreviousSibling: newPreviousSibling,
	}, nil, nil)
	return e
}

func (c *client) Delete(css *clientsession.Store, region cnst.Region, shard int, account, project, task id.Id) error {
	_, e := delete.DoRequest(css, c.host, region, &deleteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, region cnst.Region, shard int, account, project id.Id, task id.Id) (*Task, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
	}, nil, &Task{})
	if val != nil {
		return val.(*Task), e
	}
	return nil, e
}

func (c *client) GetChildren(css *clientsession.Store, region cnst.Region, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) (*getChildrenResp, error) {
	val, e := getChildren.DoRequest(css, c.host, region, &getChildrenArgs{
		Shard:       shard,
		Account:     account,
		Project:     project,
		Parent:      parent,
		FromSibling: fromSibling,
		Limit:       limit,
	}, nil, &getChildrenResp{})
	if val != nil {
		return val.(*getChildrenResp), e
	}
	return nil, e
}

func (c *client) GetAncestors(css *clientsession.Store, region cnst.Region, shard int, account, project, child id.Id, limit int) (*getAncestorsResp, error) {
	val, e := getAncestors.DoRequest(css, c.host, region, &getAncestorsArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Child:   child,
		Limit:   limit,
	}, nil, &getAncestorsResp{})
	if val != nil {
		return val.(*getAncestorsResp), e
	}
	return nil, e
}
