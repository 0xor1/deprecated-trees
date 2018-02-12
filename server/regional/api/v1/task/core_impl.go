package task

import (
	. "bitbucket.org/0xor1/task/server/util"
	"time"
)

const (
	itemType = "task"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateTask(shard int, accountId, projectId, parentId, myId Id, previousSibling *Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *Id, totalRemainingTime *uint64) *task {
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
	if memberId != nil { //if a member is being assigned to the task then we need to check they have project write access
		ValidateMemberIsAProjectMemberWithWriteAccess(a.store.getProjectRole(shard, accountId, projectId, *memberId))
	}
	newTask := &task{
		Id:                   NewId(),
		IsAbstract:           isAbstract,
		Name:                 name,
		Description:          description,
		CreatedOn:            Now(),
		TotalRemainingTime:   *totalRemainingTime,
		MinimumRemainingTime: zeroPtr,
		ChildCount:           zeroPtr,
		DescendantCount:      zeroPtr,
		IsParallel:           isParallel,
		Member:               memberId,
	}
	a.store.createTask(shard, accountId, projectId, parentId, myId, previousSibling, newTask)
	return newTask
}

func (a *api) SetName(shard int, accountId, projectId, taskId, myId Id, name string) {
	if projectId.Equal(taskId) {
		ValidateMemberHasAccountAdminAccess(a.store.getAccountRole(shard, accountId, myId))
	} else {
		ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	}

	a.store.setName(shard, accountId, projectId, taskId, myId, name)
}

func (a *api) SetDescription(shard int, accountId, projectId, taskId, myId Id, description *string) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setDescription(shard, accountId, projectId, taskId, myId, description)
}

func (a *api) SetIsParallel(shard int, accountId, projectId, taskId, myId Id, isParallel bool) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.setIsParallel(shard, accountId, projectId, taskId, myId, isParallel)
}

func (a *api) SetMember(shard int, accountId, projectId, taskId, myId Id, memberId *Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	if memberId != nil {
		ValidateMemberIsAProjectMemberWithWriteAccess(a.store.getProjectRole(shard, accountId, projectId, *memberId))
	}
	a.store.setMember(shard, accountId, projectId, taskId, myId, memberId)
}

func (a *api) SetRemainingTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining uint64) {
	a.setRemainingTimeAndOrLogTime(shard, accountId, projectId, taskId, myId, &timeRemaining, nil, nil)
}

func (a *api) LogTime(shard int, accountId, projectId, taskId Id, myId Id, duration uint64, note *string) *timeLog {
	return a.setRemainingTimeAndOrLogTime(shard, accountId, projectId, taskId, myId, nil, &duration, note)
}

func (a *api) SetRemainingTimeAndLogTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining uint64, duration uint64, note *string) *timeLog {
	return a.setRemainingTimeAndOrLogTime(shard, accountId, projectId, taskId, myId, &timeRemaining, &duration, note)
}

func (a *api) setRemainingTimeAndOrLogTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining *uint64, duration *uint64, note *string) *timeLog {
	if duration != nil {
		ValidateMemberIsAProjectMemberWithWriteAccess(a.store.getProjectRole(shard, accountId, projectId, myId))
	} else if timeRemaining != nil {
		ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))
	} else {
		InvalidArgumentsErr.Panic()
	}

	loggedOn := Now()
	a.store.setRemainingTimeAndOrLogTime(shard, accountId, projectId, taskId, myId, timeRemaining, &loggedOn, duration, note)
	if duration != nil {
		return &timeLog{
			Project:  projectId,
			Task:     taskId,
			Member:   myId,
			LoggedOn: loggedOn,
			Duration: *duration,
			Note:     note,
		}
	}
	return nil
}

func (a *api) MoveTask(shard int, accountId, projectId, taskId, myId, parentId Id, nextSibling *Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.moveTask(shard, accountId, projectId, taskId, parentId, myId, nextSibling)
}

func (a *api) DeleteTask(shard int, accountId, projectId, taskId, myId Id) {
	ValidateMemberHasProjectWriteAccess(a.store.getAccountAndProjectRoles(shard, accountId, projectId, myId))

	a.store.deleteTask(shard, accountId, projectId, taskId, myId)
}

func (a *api) GetTasks(shard int, accountId, projectId, myId Id, taskIds []Id) []*task {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	ValidateEntityCount(len(taskIds), a.maxProcessEntityCount)
	return a.store.getTasks(shard, accountId, projectId, taskIds)
}

func (a *api) GetChildTasks(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*task {
	ValidateMemberHasProjectReadAccess(a.store.getAccountAndProjectRolesAndProjectIsPublic(shard, accountId, projectId, myId))
	ValidateLimitParam(limit, a.maxProcessEntityCount)
	return a.store.getChildTasks(shard, accountId, projectId, parentId, fromSibling, limit)
}

type store interface {
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	getProjectRole(shard int, accountId, projectId, memberId Id) *ProjectRole
	getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	createTask(shard int, accountId, projectId, parentId, myId Id, previousSibling *Id, newTask *task)
	setName(shard int, accountId, projectId, taskId, myId Id, name string)
	setDescription(shard int, accountId, projectId, taskId, myId Id, description *string)
	setIsParallel(shard int, accountId, projectId, taskId, myId Id, isParallel bool)
	setMember(shard int, accountId, projectId, taskId, myId Id, memberId *Id)
	setRemainingTimeAndOrLogTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining *uint64, loggedOn *time.Time, duration *uint64, note *string)
	moveTask(shard int, accountId, projectId, taskId, parentId, myId Id, nextSibling *Id)
	deleteTask(shard int, accountId, projectId, taskId, myId Id)
	getTasks(shard int, accountId, projectId Id, taskIds []Id) []*task
	getChildTasks(shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) []*task
}

type task struct {
	Id                   Id        `json:"id"`
	IsAbstract           bool      `json:"isAbstract"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"` //only abstract tasks
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	ChildCount           *uint64   `json:"childCount,omitempty"`      //only abstract tasks
	DescendantCount      *uint64   `json:"descendantCount,omitempty"` //only abstract tasks
	IsParallel           *bool     `json:"isParallel,omitempty"`      //only abstract tasks
	Member               *Id       `json:"member,omitempty"`          //only task tasks
}

type timeLog struct {
	Project  Id        `json:"project"`
	Task     Id        `json:"task"`
	Member   Id        `json:"member"`
	LoggedOn time.Time `json:"loggedOn"`
	TaskName string    `json:"taskName"`
	Duration uint64    `json:"duration"`
	Note     *string   `json:"note"`
}
