package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"time"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateNode(shard int, accountId, projectId, parentId, myId Id, name, description string, isAbstract bool, isParallel *bool, memberId *Id, timeRemaining *uint64) *node {

}

func (a *api) SetName(shard int, accountId, projectId, nodeId, myId Id, name string) {

}

func (a *api) SetDescription(shard int, accountId, projectId, nodeId, myId Id, description string) {

}

func (a *api) SetIsParallel(shard int, accountId, projectId, nodeId, myId Id, isParallel bool) {

}

func (a *api) SetMember(shard int, accountId, projectId, nodeId, myId Id, memberId *Id) {

}

func (a *api) SetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, timeRemaining uint64) {

}

func (a *api) LogTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) {

}

func (a *api) MoveNode(shard int, accountId, projectId, nodeId, myId, parentId Id, nextSibling *Id) {

}

func (a *api) DeleteNode(shard int, accountId, projectId, nodeId, myId Id) {

}

func (a *api) GetNodes(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*node {

}


type store interface {
}

type node struct {
	Id                   Id        `json:"id"`
	Name                 string    `json:"name"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	IsAbstract           bool      `json:"isAbstract"`
	Description          string    `json:"description"`
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"`
	IsParallel           *bool     `json:"isParallel,omitempty"`
	Member               *Id       `json:"member,omitempty"`
}
