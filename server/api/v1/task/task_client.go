package task

import (
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/id"
)

type Client interface {
	Create(css *clientsession.Store, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, remainingTime *uint64) (*Task, error)
	SetName(css *clientsession.Store, shard int, account, project, task id.Id, name string) error
	SetDescription(css *clientsession.Store, shard int, account, project, task id.Id, description *string) error
	SetIsParallel(css *clientsession.Store, shard int, account, project, task id.Id, isParallel bool) error         //only applies to abstract tasks
	SetMember(css *clientsession.Store, shard int, account, project, task id.Id, member *id.Id) error               //only applies to tasks
	SetRemainingTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64) error //only applies to tasks
	Move(css *clientsession.Store, shard int, account, project, task, parent id.Id, nextSibling *id.Id) error
	Delete(css *clientsession.Store, shard int, account, project, task id.Id) error
	Get(css *clientsession.Store, shard int, account, project id.Id, task id.Id) (*Task, error)
	GetChildren(css *clientsession.Store, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) ([]*Task, error)
	GetAncestors(css *clientsession.Store, shard int, account, project, child id.Id, limit int) ([]*Ancestor, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Create(css *clientsession.Store, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, totalRemainingTime *uint64) (*Task, error) {
	val, e := create.DoRequest(css, c.host, &createArgs{
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

func (c *client) SetName(css *clientsession.Store, shard int, account, project, task id.Id, name string) error {
	_, e := setName.DoRequest(css, c.host, &setNameArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Name:    name,
	}, nil, nil)
	return e
}

func (c *client) SetDescription(css *clientsession.Store, shard int, account, project, task id.Id, description *string) error {
	_, e := setDescription.DoRequest(css, c.host, &setDescriptionArgs{
		Shard:       shard,
		Account:     account,
		Project:     project,
		Task:        task,
		Description: description,
	}, nil, nil)
	return e
}

func (c *client) SetIsParallel(css *clientsession.Store, shard int, account, project, task id.Id, isParallel bool) error {
	_, e := setIsParallel.DoRequest(css, c.host, &setIsParallelArgs{
		Shard:      shard,
		Account:    account,
		Project:    project,
		Task:       task,
		IsParallel: isParallel,
	}, nil, nil)
	return e
}

func (c *client) SetMember(css *clientsession.Store, shard int, account, project, task id.Id, member *id.Id) error {
	_, e := setMember.DoRequest(css, c.host, &setMemberArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Member:  member,
	}, nil, nil)
	return e
}

func (c *client) SetRemainingTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64) error {
	_, e := setRemainingTime.DoRequest(css, c.host, &setRemainingTimeArgs{
		Shard:         shard,
		Account:       account,
		Project:       project,
		Task:          task,
		RemainingTime: remainingTime,
	}, nil, nil)
	return e
}

func (c *client) Move(css *clientsession.Store, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) error {
	_, e := move.DoRequest(css, c.host, &moveArgs{
		Shard:              shard,
		Account:            account,
		Project:            project,
		Task:               task,
		NewParent:          newParent,
		NewPreviousSibling: newPreviousSibling,
	}, nil, nil)
	return e
}

func (c *client) Delete(css *clientsession.Store, shard int, account, project, task id.Id) error {
	_, e := delete.DoRequest(css, c.host, &deleteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, shard int, account, project id.Id, task id.Id) (*Task, error) {
	val, e := get.DoRequest(css, c.host, &getArgs{
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

func (c *client) GetChildren(css *clientsession.Store, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) ([]*Task, error) {
	val, e := getChildren.DoRequest(css, c.host, &getChildrenArgs{
		Shard:       shard,
		Account:     account,
		Project:     project,
		Parent:      parent,
		FromSibling: fromSibling,
		Limit:       limit,
	}, nil, &[]*Task{})
	if val != nil {
		return *val.(*[]*Task), e
	}
	return nil, e
}

func (c *client) GetAncestors(css *clientsession.Store, shard int, account, project, child id.Id, limit int) ([]*Ancestor, error) {
	val, e := getAncestors.DoRequest(css, c.host, &getAncestorTasksArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    child,
		Limit:   limit,
	}, nil, &[]*Ancestor{})
	if val != nil {
		return *val.(*[]*Ancestor), e
	}
	return nil, e
}
