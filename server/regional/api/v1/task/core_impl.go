package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"fmt"
	"time"
)

const (
	itemType = "node"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateNode(shard int, accountId, projectId, parentId, myId Id, previousSibling *Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *Id, totalRemainingTime *uint64) *node {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	if (isAbstract && (isParallel == nil || memberId != nil || totalRemainingTime != nil)) || (!isAbstract && (isParallel != nil || totalRemainingTime == nil)) {
		InvalidArgumentsErr.Panic()
	}
	zeroVal := uint64(0)
	zeroPtr := &zeroVal
	if !isAbstract {
		zeroPtr = nil
	} else {
		totalRemainingTime = zeroPtr
	}
	if memberId != nil { //if a member is being assigned to the node then we need to check they have project write access
		ValidateMemberIsAProjectMemberWithWriteAccess(a.store.getProjectRole(shard, accountId, projectId, *memberId))
	}
	newNode := &node{
		Id:                   NewId(),
		IsAbstract:           isAbstract,
		Name:                 name,
		Description:          description,
		CreatedOn:            Now(),
		TotalRemainingTime:   *totalRemainingTime,
		TotalLoggedTime:      0,
		MinimumRemainingTime: zeroPtr,
		LinkedFileCount:      0,
		ChatCount:            0,
		ChildCount:           zeroPtr,
		DescendantCount:      zeroPtr,
		IsParallel:           isParallel,
		Member:               memberId,
	}
	a.store.createNode(shard, accountId, projectId, parentId, previousSibling, newNode)
	a.store.logProjectActivity(shard, accountId, projectId, myId, newNode.Id, itemType, "created", nil)
	return newNode
}

func (a *api) SetName(shard int, accountId, projectId, nodeId, myId Id, name string) {
	if projectId.Equal(nodeId) {
		ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	} else {
		ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	}

	a.store.setName(shard, accountId, projectId, nodeId, name)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "setName", &name)
}

func (a *api) SetDescription(shard int, accountId, projectId, nodeId, myId Id, description *string) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setDescription(shard, accountId, projectId, nodeId, description)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "setDescription", description)
}

func (a *api) SetIsParallel(shard int, accountId, projectId, nodeId, myId Id, isParallel bool) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setIsParallel(shard, accountId, projectId, nodeId, isParallel)
	action := fmt.Sprintf("%t", isParallel)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "setIsParallel", &action)
}

func (a *api) SetMember(shard int, accountId, projectId, nodeId, myId Id, memberId *Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	if memberId != nil {
		ValidateMemberIsAProjectMemberWithWriteAccess(a.store.getProjectRole(shard, accountId, projectId, *memberId))
	}
	a.store.setMember(shard, accountId, projectId, nodeId, memberId)
	var newValue *string
	if memberId != nil {
		str := memberId.String()
		newValue = &str
	}
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "setMember", newValue)
}

func (a *api) SetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, timeRemaining uint64) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setTimeRemaining(shard, accountId, projectId, nodeId, timeRemaining)
	newValue := fmt.Sprintf("%d", timeRemaining)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "setTimeRemaining", &newValue)
}

func (a *api) LogTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) {
	a.SetTimeRemaining(shard, accountId, projectId, nodeId, myId, timeRemaining)

	a.store.logTime(shard, accountId, projectId, nodeId, myId, duration, note)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "loggedTime", note)
}

func (a *api) MoveNode(shard int, accountId, projectId, nodeId, myId, parentId Id, nextSibling *Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.moveNode(shard, accountId, projectId, nodeId, parentId, nextSibling)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "moveNode", nil)
}

func (a *api) DeleteNode(shard int, accountId, projectId, nodeId, myId Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.deleteNode(shard, accountId, projectId, nodeId)
	a.store.logProjectActivity(shard, accountId, projectId, myId, nodeId, itemType, "moveNode", nil)
}

func (a *api) GetNode(shard int, accountId, projectId, nodeId, myId Id) *node {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getNode(shard, accountId, projectId, nodeId)
}

func (a *api) GetNodes(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*node {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	return a.store.getNodes(shard, accountId, projectId, parentId, fromSibling, limit)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getProjectRole(shard int, accountId, projectId, memberId Id) *ProjectRole
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	createNode(shard int, accountId, projectId, parentId Id, previousSibling *Id, newNode *node)
	setName(shard int, accountId, projectId, nodeId Id, name string)
	setDescription(shard int, accountId, projectId, nodeId Id, description *string)
	setIsParallel(shard int, accountId, projectId, nodeId Id, isParallel bool)
	setMember(shard int, accountId, projectId, nodeId Id, memberId *Id)
	setTimeRemaining(shard int, accountId, projectId, nodeId Id, timeRemaining uint64)
	logTime(shard int, accountId, projectId, nodeId, myId Id, duration uint64, note *string)
	moveNode(shard int, accountId, projectId, nodeId, parentId Id, nextSibling *Id)
	deleteNode(shard int, accountId, projectId, nodeId Id)
	getNode(shard int, accountId, projectId, nodeId Id) *node
	getNodes(shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) []*node
	logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string)
}

type node struct {
	Id                   Id        `json:"id"`
	IsAbstract           bool      `json:"isAbstract"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"` //only abstract nodes
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	ChildCount           *uint64   `json:"childCount,omitempty"`      //only abstract nodes
	DescendantCount      *uint64   `json:"descendantCount,omitempty"` //only abstract nodes
	IsParallel           *bool     `json:"isParallel,omitempty"`      //only abstract nodes
	Member               *Id       `json:"member,omitempty"`          //only task nodes
}
