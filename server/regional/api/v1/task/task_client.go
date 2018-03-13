package task

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/id"
)

type Client interface {
	CreateTask(css *clientsession.Store, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, remainingTime *uint64) (*task, error)
	SetName(css *clientsession.Store, shard int, account, project, task id.Id, name string) error
	SetDescription(css *clientsession.Store, shard int, account, project, task id.Id, description *string) error
	SetIsParallel(css *clientsession.Store, shard int, account, project, task id.Id, isParallel bool) error                                                              //only applys to abstract tasks
	SetMember(css *clientsession.Store, shard int, account, project, task id.Id, member *id.Id) error                                                                  //only applys to task tasks
	SetRemainingTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64) error                                                      //only applys to task tasks
	LogTime(css *clientsession.Store, shard int, account, project, task id.Id, duration uint64, note *string) (*timeLog, error)                                          //only applys to task tasks
	SetRemainingTimeAndLogTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) //only applys to task tasks
	MoveTask(css *clientsession.Store, shard int, account, project, task, parent id.Id, nextSibling *id.Id) error
	DeleteTask(css *clientsession.Store, shard int, account, project, task id.Id) error
	GetTasks(css *clientsession.Store, shard int, account, project id.Id, tasks []id.Id) ([]*task, error)
	GetChildTasks(css *clientsession.Store, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) ([]*task, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateTask(css *clientsession.Store, shard int, account, project, parent id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, member *id.Id, totalRemainingTime *uint64) (*task, error) {
	val, e := createTask.DoRequest(css, c.host, &createTaskArgs{
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
	}, nil, &task{})
	if val != nil {
		return val.(*task), e
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

func (c *client) LogTime(css *clientsession.Store, shard int, account, project, task id.Id, duration uint64, note *string) (*timeLog, error) {
	val, e := logTime.DoRequest(css, c.host, &logTimeArgs{
		Shard:    shard,
		Account:  account,
		Project:  project,
		Task:     task,
		Duration: duration,
		Note:     note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), e
	}
	return nil, e
}

func (c *client) SetRemainingTimeAndLogTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) {
	val, e := setRemainingTimeAndLogTime.DoRequest(css, c.host, &setRemainingTimeAndLogTimeArgs{
		Shard:         shard,
		Account:       account,
		Project:       project,
		Task:          task,
		RemainingTime: remainingTime,
		Duration:      duration,
		Note:          note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), e
	}
	return nil, e
}

func (c *client) MoveTask(css *clientsession.Store, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) error {
	_, e := moveTask.DoRequest(css, c.host, &moveTaskArgs{
		Shard:              shard,
		Account:            account,
		Project:            project,
		Task:               task,
		NewParent:          newParent,
		NewPreviousSibling: newPreviousSibling,
	}, nil, nil)
	return e
}

func (c *client) DeleteTask(css *clientsession.Store, shard int, account, project, task id.Id) error {
	_, e := deleteTask.DoRequest(css, c.host, &deleteTaskArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
	}, nil, nil)
	return e
}

func (c *client) GetTasks(css *clientsession.Store, shard int, account, project id.Id, tasks []id.Id) ([]*task, error) {
	val, e := getTasks.DoRequest(css, c.host, &getTasksArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Tasks:   tasks,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), e
	}
	return nil, e
}

func (c *client) GetChildTasks(css *clientsession.Store, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) ([]*task, error) {
	val, e := getChildTasks.DoRequest(css, c.host, &getChildTasksArgs{
		Shard:       shard,
		Account:     account,
		Project:     project,
		Parent:      parent,
		FromSibling: fromSibling,
		Limit:       limit,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), e
	}
	return nil, e
}