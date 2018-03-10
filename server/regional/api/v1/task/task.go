package task

import (
	. "bitbucket.org/0xor1/task/server/util"
	"time"
	"bytes"
	"encoding/hex"
)


type createTaskArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	ParentId Id `json:"parentId"`
	PreviousSibling *Id `json:"previousSibling,omitempty"`
	Name string `json:"name"`
	Description *string `json:"description,omitempty"`
	IsAbstract bool `json:"isAbstract"`
	IsParallel *bool `json:"isParallel,omitempty"`
	MemberId *Id `json:"memberId,omitempty"`
	TotalRemainingTime *uint64 `json:"totalRemainingTime,omitempty"`
}

var createTask = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/createTask",
	GetArgsStruct: func() interface{} {
		return &createTaskArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*createTaskArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		if (args.IsAbstract && (args.IsParallel == nil || args.MemberId != nil || args.TotalRemainingTime != nil)) || (!args.IsAbstract && (args.IsParallel != nil || args.TotalRemainingTime == nil)) {
			InvalidArgumentsErr.Panic()
		}
		zeroVal := uint64(0)
		zeroPtr := &zeroVal
		if !args.IsAbstract {
			zeroPtr = nil
		} else {
			args.TotalRemainingTime = zeroPtr
		}
		if args.MemberId != nil { //if a member is being assigned to the task then we need to check they have project write access
			ValidateMemberIsAProjectMemberWithWriteAccess(GetProjectRole(ctx, args.Shard, args.AccountId, args.ProjectId, *args.MemberId))
		}
		newTask := &task{
			Id:                   NewId(),
			IsAbstract:           args.IsAbstract,
			Name:                 args.Name,
			Description:          args.Description,
			CreatedOn:            Now(),
			TotalRemainingTime:   *args.TotalRemainingTime,
			MinimumRemainingTime: zeroPtr,
			ChildCount:           zeroPtr,
			DescendantCount:      zeroPtr,
			IsParallel:           args.IsParallel,
			Member:               args.MemberId,
		}
		dbCreateTask(ctx, args.Shard, args.AccountId, args.ProjectId, args.ParentId, args.PreviousSibling, newTask)
		return newTask
	},
}

type setNameArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	TaskId Id `json:"taskId"`
	Name string `json:"name"`
}

var setName = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setName",
	GetArgsStruct: func() interface{} {
		return &setNameArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setNameArgs)
		if args.ProjectId.Equal(args.TaskId) {
			ValidateMemberHasAccountAdminAccess(GetAccountRole(ctx, args.Shard, args.AccountId, ctx.MyId()))
		} else {
			ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		}

		dbSetName(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.Name)
		return nil
	},
}

type setDescriptionArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	TaskId Id `json:"taskId"`
	Description *string `json:"description,omitempty"`
}

var setDescription = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setDescription",
	GetArgsStruct: func() interface{} {
		return &setDescriptionArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setDescriptionArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		dbSetDescription(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.Description)
		return nil
	},
}

type setIsParallelArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	TaskId Id `json:"taskId"`
	IsParallel bool `json:"isParallel"`
}

var setIsParallel = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setIsParallel",
	GetArgsStruct: func() interface{} {
		return &setIsParallelArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setIsParallelArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		dbSetIsParallel(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.IsParallel)
		return nil
	},
}

type setMemberArgs struct {
	Shard int `json:"shard"`
	AccountId Id `json:"accountId"`
	ProjectId Id `json:"projectId"`
	TaskId Id `json:"taskId"`
	MemberId *Id `json:"memberId,omitempty"`
}

var setMember = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setMember",
	GetArgsStruct: func() interface{} {
		return &setMemberArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setMemberArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		if args.MemberId != nil {
			ValidateMemberIsAProjectMemberWithWriteAccess(GetProjectRole(ctx, args.Shard, args.AccountId, args.ProjectId, *args.MemberId))
		}
		dbSetMember(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.MemberId)
		return nil
	},
}

type setRemainingTimeArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskId        Id     `json:"taskId"`
	RemainingTime uint64 `json:"remainingTime"`
}

var setRemainingTime = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setRemainingTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, &args.RemainingTime, nil, nil)
	},
}

type logTimeArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskId        Id     `json:"taskId"`
	Duration uint64 `json:"duration"`
	Note *string `json:"note,omitempty"`
}

var logTime = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/logTime",
	GetArgsStruct: func() interface{} {
		return &logTimeArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*logTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, nil, &args.Duration, args.Note)
	},
}

type setRemainingTimeAndLogTimeArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskId        Id     `json:"taskId"`
	RemainingTime uint64 `json:"remainingTime"`
	Duration uint64 `json:"duration"`
	Note *string `json:"note,omitempty"`
}

var setRemainingTimeAndLogTime = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/setRemainingTimeAndLogTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeAndLogTimeArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeAndLogTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, &args.RemainingTime, &args.Duration, args.Note)
	},
}

func setRemainingTimeAndOrLogTime(ctx *Ctx, shard int, accountId, projectId, taskId Id, remainingTime *uint64, duration *uint64, note *string) *timeLog {
	if duration != nil {
		ValidateMemberIsAProjectMemberWithWriteAccess(GetProjectRole(ctx, shard, accountId, projectId, ctx.MyId()))
	} else if remainingTime != nil {
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, shard, accountId, projectId, ctx.MyId()))
	} else {
		InvalidArgumentsErr.Panic()
	}

	loggedOn := Now()
	dbSetRemainingTimeAndOrLogTime(ctx, shard, accountId, projectId, taskId, remainingTime, &loggedOn, duration, note)
	if duration != nil {
		return &timeLog{
			Project:  projectId,
			Task:     taskId,
			Member:   ctx.MyId(),
			LoggedOn: loggedOn,
			Duration: *duration,
			Note:     note,
		}
	}
	return nil
}

type moveTaskArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskId        Id     `json:"taskId"`
	NewParentId     Id     `json:"newParentId"`
	NewPreviousSibling *Id     `json:"newPreviousSibling"`
}

var moveTask = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/moveTask",
	GetArgsStruct: func() interface{} {
		return &moveTaskArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*moveTaskArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		dbMoveTask(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.NewParentId, args.NewPreviousSibling)
		return nil
	},
}

type deleteTaskArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskId        Id     `json:"taskId"`
}

var deleteTask = &Endpoint{
	Method: POST,
	Path:   "/api/v1/project/deleteTask",
	GetArgsStruct: func() interface{} {
		return &deleteTaskArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*deleteTaskArgs)
		ValidateMemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))

		dbDeleteTask(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId)
		return nil
	},
}

type getTasksArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	TaskIds        []Id     `json:"taskId"`
}

var getTasks = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getTasks",
	GetArgsStruct: func() interface{} {
		return &getTasksArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getTasksArgs)
		ValidateMemberHasProjectReadAccess(GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		ValidateEntityCount(len(args.TaskIds), ctx.MaxProcessEntityCount())
		return dbGetTasks(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskIds)
	},
}

type getChildTasksArgs struct {
	Shard         int    `json:"shard"`
	AccountId     Id     `json:"accountId"`
	ProjectId     Id     `json:"projectId"`
	ParentId      Id     `json:"parentId"`
	FromSibling *Id `json:"fromSibling,omitempty"`
	Limit int `json:"limit"`
}

var getChildTasks = &Endpoint{
	Method: GET,
	Path:   "/api/v1/project/getChildTasks",
	GetArgsStruct: func() interface{} {
		return &getChildTasksArgs{}
	},
	CtxHandler: func(ctx *Ctx, a interface{}) interface{} {
		args := a.(*getChildTasksArgs)
		ValidateMemberHasProjectReadAccess(GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.MyId()))
		ValidateLimit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetChildTasks(ctx, args.Shard, args.AccountId, args.ProjectId, args.ParentId, args.FromSibling, args.Limit)
	},
}

var Endpoints = []*Endpoint{
	createTask,
	setName,
	setDescription,
	setIsParallel,
	setMember,
	setRemainingTime,
	logTime,
	setRemainingTimeAndLogTime,
	moveTask,
	deleteTask,
	getTasks,
	getChildTasks,
}

type Client interface {
	CreateTask(css *ClientSessionStore, shard int, accountId, projectId, parentId Id, previousSibling *Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *Id, remainingTime *uint64) (*task, error)
	SetName(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, name string) error
	SetDescription(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, description *string) error
	SetIsParallel(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, isParallel bool) error                                                              //only applys to abstract tasks
	SetMember(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, memberId *Id) error                                                                     //only applys to task tasks
	SetRemainingTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, remainingTime uint64) error                                                      //only applys to task tasks
	LogTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, duration uint64, note *string) (*timeLog, error)                                          //only applys to task tasks
	SetRemainingTimeAndLogTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) //only applys to task tasks
	MoveTask(css *ClientSessionStore, shard int, accountId, projectId, taskId, parentId Id, nextSibling *Id) error
	DeleteTask(css *ClientSessionStore, shard int, accountId, projectId, taskId Id) error
	GetTasks(css *ClientSessionStore, shard int, accountId, projectId Id, taskIds []Id) ([]*task, error)
	GetChildTasks(css *ClientSessionStore, shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) ([]*task, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateTask(css *ClientSessionStore, shard int, accountId, projectId, parentId Id, previousSibling *Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *Id, totalRemainingTime *uint64) (*task, error) {
	val, err := createTask.DoRequest(css, c.host, &createTaskArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		ParentId: parentId,
		PreviousSibling: previousSibling,
		Name: name,
		Description: description,
		IsAbstract: isAbstract,
		IsParallel: isParallel,
		MemberId: memberId,
		TotalRemainingTime: totalRemainingTime,
	}, nil, &task{})
	if val != nil {
		return val.(*task), err
	}
	return nil, err
}

func (c *client) SetName(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, name string) error {
	_, err := setName.DoRequest(css, c.host, &setNameArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		Name: name,
	}, nil, nil)
	return err
}

func (c *client) SetDescription(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, description *string) error {
	_, err := setDescription.DoRequest(css, c.host, &setDescriptionArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		Description: description,
	}, nil, nil)
	return err
}

func (c *client) SetIsParallel(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, isParallel bool) error {
	_, err := setIsParallel.DoRequest(css, c.host, &setIsParallelArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		IsParallel: isParallel,
	}, nil, nil)
	return err
}

func (c *client) SetMember(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, memberId *Id) error {
	_, err := setMember.DoRequest(css, c.host, &setMemberArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		MemberId: memberId,
	}, nil, nil)
	return err
}

func (c *client) SetRemainingTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, remainingTime uint64) error {
	_, err := setRemainingTime.DoRequest(css, c.host, &setRemainingTimeArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		RemainingTime: remainingTime,
	}, nil, nil)
	return err
}

func (c *client) LogTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, duration uint64, note *string) (*timeLog, error) {
	val, err := logTime.DoRequest(css, c.host, &logTimeArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		Duration: duration,
		Note: note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), err
	}
	return nil, err
}

func (c *client) SetRemainingTimeAndLogTime(css *ClientSessionStore, shard int, accountId, projectId, taskId Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) {
	val, err := setRemainingTimeAndLogTime.DoRequest(css, c.host, &setRemainingTimeAndLogTimeArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId: taskId,
		RemainingTime: remainingTime,
		Duration: duration,
		Note: note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), err
	}
	return nil, err
}

func (c *client) MoveTask(css *ClientSessionStore, shard int, accountId, projectId, taskId, newParentId Id, newPreviousSibling *Id) error {
	_, err := moveTask.DoRequest(css, c.host, &moveTaskArgs{
		Shard:              shard,
		AccountId:          accountId,
		ProjectId:          projectId,
		TaskId:             taskId,
		NewParentId:        newParentId,
		NewPreviousSibling: newPreviousSibling,
	}, nil, nil)
	return err
}

func (c *client) DeleteTask(css *ClientSessionStore, shard int, accountId, projectId, taskId Id) error {
	_, err := deleteTask.DoRequest(css, c.host, &deleteTaskArgs{
		Shard:              shard,
		AccountId:          accountId,
		ProjectId:          projectId,
		TaskId:             taskId,
	}, nil, nil)
	return err
}

func (c *client) GetTasks(css *ClientSessionStore, shard int, accountId, projectId Id, taskIds []Id) ([]*task, error) {
	val, err := getTasks.DoRequest(css, c.host, &getTasksArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskIds: taskIds,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), err
	}
	return nil, err
}

func (c *client) GetChildTasks(css *ClientSessionStore, shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) ([]*task, error) {
	val, err := getChildTasks.DoRequest(css, c.host, &getChildTasksArgs{
		Shard: shard,
		AccountId: accountId,
		ProjectId: projectId,
		ParentId: parentId,
		FromSibling: fromSibling,
		Limit: limit,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), err
	}
	return nil, err
}

func dbCreateTask(ctx *Ctx, shard int, accountId, projectId, parentId Id, nextSibling *Id, newTask *task) {
	args := make([]interface{}, 0, 18)
	args = append(args, accountId, projectId, parentId, ctx.MyId())
	if nextSibling != nil {
		args = append(args, *nextSibling)
	} else {
		args = append(args, nil)
	}
	args = append(args, newTask.Id)
	args = append(args, newTask.IsAbstract)
	args = append(args, newTask.Name)
	args = append(args, newTask.Description)
	args = append(args, newTask.CreatedOn)
	args = append(args, newTask.TotalRemainingTime)
	args = append(args, newTask.IsParallel)
	if newTask.Member != nil {
		args = append(args, *newTask.Member)
	} else {
		args = append(args, nil)
	}
	TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
}

func dbSetName(ctx *Ctx, shard int, accountId, projectId, taskId Id, name string) {
	_, err := ctx.TreeExec(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId(), name)
	PanicIf(err)
}

func dbSetDescription(ctx *Ctx, shard int, accountId, projectId, taskId Id, description *string) {
	_, err := ctx.TreeExec(shard, `CALL setTaskDescription(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId(), description)
	PanicIf(err)
}

func dbSetIsParallel(ctx *Ctx, shard int, accountId, projectId, taskId Id, isParallel bool) {
	TreeChangeHelper(ctx, shard, `CALL setTaskIsParallel(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId(), isParallel)
}

func dbSetMember(ctx *Ctx, shard int, accountId, projectId, taskId Id, memberId *Id) {
	var memArg []byte
	if memberId != nil {
		memArg = *memberId
	}
	MakeChangeHelper(ctx, shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId(), memArg)
}

func dbSetRemainingTimeAndOrLogTime(ctx *Ctx, shard int, accountId, projectId, taskId Id, remainingTime *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	TreeChangeHelper(ctx, shard, `CALL setRemainingTimeAndOrLogTime(?, ?, ?, ?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId(), remainingTime, loggedOn, duration, note)
}

func dbMoveTask(ctx *Ctx, shard int, accountId, projectId, taskId, newParentId Id, newPreviousSibling *Id) {
	var prevSib []byte
	if newPreviousSibling != nil {
		prevSib = *newPreviousSibling
	}
	TreeChangeHelper(ctx, shard, `CALL moveTask(?, ?, ?, ?, ?, ?)`, accountId, projectId, taskId, newParentId, ctx.MyId(), prevSib)
}

func dbDeleteTask(ctx *Ctx, shard int, accountId, projectId, taskId Id) {
	TreeChangeHelper(ctx, shard, `CALL deleteTask(?, ?, ?, ?)`, accountId, projectId, taskId, ctx.MyId())
}

func dbGetTasks(ctx *Ctx, shard int, accountId, projectId Id, taskIds []Id) []*task {
	idsStr := bytes.NewBufferString(``)
	for _, id := range taskIds {
		idsStr.WriteString(hex.EncodeToString(id))
	}
	rows, err := ctx.TreeQuery(shard, `CALL getTasks(?, ?, ?)`, accountId, projectId, idsStr.String())
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*task, 0, len(taskIds))
	for rows.Next() {
		t := task{}
		PanicIf(rows.Scan(&t.Id, &t.IsAbstract, &t.Name, &t.Description, &t.CreatedOn, &t.TotalRemainingTime, &t.TotalLoggedTime, &t.MinimumRemainingTime, &t.LinkedFileCount, &t.ChatCount, &t.ChildCount, &t.DescendantCount, &t.IsParallel, &t.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&t)
		res = append(res, &t)
	}
	return res
}

func dbGetChildTasks(ctx *Ctx, shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) []*task {
	var fromSib []byte
	if fromSibling != nil {
		fromSib = *fromSibling
	}
	rows, err := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, accountId, projectId, parentId, fromSib, limit)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*task, 0, limit)
	for rows.Next() {
		t := task{}
		PanicIf(rows.Scan(&t.Id, &t.IsAbstract, &t.Name, &t.Description, &t.CreatedOn, &t.TotalRemainingTime, &t.TotalLoggedTime, &t.MinimumRemainingTime, &t.LinkedFileCount, &t.ChatCount, &t.ChildCount, &t.DescendantCount, &t.IsParallel, &t.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&t)
		res = append(res, &t)
	}
	return res
}

func nilOutPropertiesThatAreNotNilInTheDb(t *task) {
	if !t.IsAbstract {
		t.MinimumRemainingTime = nil
		t.ChildCount = nil
		t.DescendantCount = nil
		t.IsParallel = nil
	}
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
