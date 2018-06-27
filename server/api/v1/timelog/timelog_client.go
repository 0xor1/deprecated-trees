package timelog

import (
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/id"
	tlog "bitbucket.org/0xor1/trees/server/util/timelog"
)

type Client interface {
	Create(css *clientsession.Store, region string, shard int, account, project, task id.Id, duration uint64, note *string) (*tlog.TimeLog, error)                                          //only applys to task tasks
	CreateAndSetRemainingTime(css *clientsession.Store, region string, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*tlog.TimeLog, error) //only applys to task tasks
	SetDuration(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id, duration uint64) error
	SetNote(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id, note *string) error
	Delete(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id) error
	Get(css *clientsession.Store, region string, shard int, account, project id.Id, task, member, timeLog *id.Id, sortAsc bool, after *id.Id, limit int) (*getResp, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) Create(css *clientsession.Store, region string, shard int, account, project, task id.Id, duration uint64, note *string) (*tlog.TimeLog, error) {
	val, e := create.DoRequest(css, c.host, region, &createArgs{
		Shard:    shard,
		Account:  account,
		Project:  project,
		Task:     task,
		Duration: duration,
		Note:     note,
	}, nil, &tlog.TimeLog{})
	if val != nil {
		return val.(*tlog.TimeLog), e
	}
	return nil, e
}

func (c *client) CreateAndSetRemainingTime(css *clientsession.Store, region string, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*tlog.TimeLog, error) {
	val, e := createAndSetRemainingTime.DoRequest(css, c.host, region, &createAndSetRemainingTimeArgs{
		Shard:         shard,
		Account:       account,
		Project:       project,
		Task:          task,
		RemainingTime: remainingTime,
		Duration:      duration,
		Note:          note,
	}, nil, &tlog.TimeLog{})
	if val != nil {
		return val.(*tlog.TimeLog), e
	}
	return nil, e
}

func (c *client) SetDuration(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id, duration uint64) error {
	_, e := setDuration.DoRequest(css, c.host, region, &setDurationArgs{
		Shard:    shard,
		Account:  account,
		Project:  project,
		TimeLog:  timeLog,
		Duration: duration,
	}, nil, nil)
	return e
}

func (c *client) SetNote(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id, note *string) error {
	_, e := setNote.DoRequest(css, c.host, region, &setNoteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		TimeLog: timeLog,
		Note:    note,
	}, nil, nil)
	return e
}

func (c *client) Delete(css *clientsession.Store, region string, shard int, account, project, timeLog id.Id) error {
	_, e := delete.DoRequest(css, c.host, region, &deleteArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		TimeLog: timeLog,
	}, nil, nil)
	return e
}

func (c *client) Get(css *clientsession.Store, region string, shard int, account, project id.Id, task, member, timeLog *id.Id, sortAsc bool, after *id.Id, limit int) (*getResp, error) {
	val, e := get.DoRequest(css, c.host, region, &getArgs{
		Shard:   shard,
		Account: account,
		Project: project,
		Task:    task,
		Member:  member,
		TimeLog: timeLog,
		SortAsc: sortAsc,
		After:   after,
		Limit:   limit,
	}, nil, &getResp{})
	if val != nil {
		return val.(*getResp), e
	}
	return nil, e
}
