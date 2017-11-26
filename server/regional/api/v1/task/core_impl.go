package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"time"
)

const(
	itemType = "node"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateNode(shard int, accountId, projectId, parentId, myId Id, nextSibling *Id, name, description string, isAbstract bool, isParallel *bool, memberId *Id, timeRemaining *uint64) *node {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	if (isAbstract && (isParallel == nil || memberId != nil || timeRemaining != nil)) || (!isAbstract && (isParallel != nil || timeRemaining == nil)) {
		InvalidArgumentsErr.Panic()
	}
	zero := uint64(0)
	totalTimeRemaining := uint64(0)
	if timeRemaining != nil {
		totalTimeRemaining = *timeRemaining
	}
	if memberId != nil { //if a member is being assigned to the node then we need to check they have project write access
		ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, *memberId))
	}
	newNode := &node{
		Id: NewId(),
		IsAbstract: isAbstract,
		Name: name,
		Description: description,
		CreatedOn: Now(),
		TotalRemainingTime: totalTimeRemaining,
		TotalLoggedTime: 0,
		MinimumRemainingTime: &zero,
		LinkedFileCount: 0,
		ChatCount: 0,
		ChildCount: &zero,
		DescendantCount: &zero,
		IsParallel: isParallel,
		Member: memberId,
	}
	a.store.createNode(shard, accountId, projectId, parentId, nextSibling, newNode)
	a.store.logProjectActivity(shard, accountId, projectId, myId, newNode.Id, itemType, "created", nil)
	return newNode
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
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	createNode(shard int, accountId, projectId, parentId Id, nextSibling *Id, newNode *node)
	logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string)
}

type node struct {
	Id                   Id        `json:"id"`
	IsAbstract           bool      `json:"isAbstract"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"` //only abstract nodes
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	ChildCount           *uint64   `json:"childCount,omitempty"` //only abstract nodes
	DescendantCount      *uint64   `json:"descendantCount,omitempty"` //only abstract nodes
	IsParallel           *bool     `json:"isParallel,omitempty"` //only abstract nodes
	Member               *Id       `json:"member,omitempty"` //only task nodes
}
