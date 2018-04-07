package timeLog

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/timeLog"
)

type Client interface {
	Create(css *clientsession.Store, shard int, account, project, task id.Id, duration uint64, note *string) (*timeLog.TimeLog, error)                                          //only applys to task tasks
	CreateAndSetRemainingTime(css *clientsession.Store, shard int, account, project, task id.Id, remainingTime uint64, duration uint64, note *string) (*timeLog.TimeLog, error) //only applys to task tasks
	SetDuration(css *clientsession.Store, shard int, account, project, timeLog id.Id, duration uint64)
	SetNote(css *clientsession.Store, shard int, account, project, timeLog id.Id, note *string)
	Delete(css *clientsession.Store, shard int, account, project, timeLog id.Id)
	Get(css *clientsession.Store, shard int, account id.Id, project, task, member, timeLog, after *id.Id, limit int) []*timeLog.TimeLog
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}
