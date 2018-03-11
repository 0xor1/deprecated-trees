package task

import (
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/core"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
	"encoding/hex"
	"time"
)

type createTaskArgs struct {
	Shard              int     `json:"shard"`
	AccountId          id.Id   `json:"accountId"`
	ProjectId          id.Id   `json:"projectId"`
	ParentId           id.Id   `json:"parentId"`
	PreviousSibling    *id.Id  `json:"previousSibling,omitempty"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	IsAbstract         bool    `json:"isAbstract"`
	IsParallel         *bool   `json:"isParallel,omitempty"`
	MemberId           *id.Id  `json:"memberId,omitempty"`
	TotalRemainingTime *uint64 `json:"totalRemainingTime,omitempty"`
}

var createTask = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/createTask",
	GetArgsStruct: func() interface{} {
		return &createTaskArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*createTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		if (args.IsAbstract && (args.IsParallel == nil || args.MemberId != nil || args.TotalRemainingTime != nil)) || (!args.IsAbstract && (args.IsParallel != nil || args.TotalRemainingTime == nil)) {
			panic(err.InvalidArguments)
		}
		zeroVal := uint64(0)
		zeroPtr := &zeroVal
		if !args.IsAbstract {
			zeroPtr = nil
		} else {
			args.TotalRemainingTime = zeroPtr
		}
		if args.MemberId != nil { //if a member is being assigned to the task then we need to check they have project write access
			validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, args.Shard, args.AccountId, args.ProjectId, *args.MemberId))
		}
		newTask := &task{
			Id:                   id.New(),
			IsAbstract:           args.IsAbstract,
			Name:                 args.Name,
			Description:          args.Description,
			CreatedOn:            t.Now(),
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
	Shard     int    `json:"shard"`
	AccountId id.Id  `json:"accountId"`
	ProjectId id.Id  `json:"projectId"`
	TaskId    id.Id  `json:"taskId"`
	Name      string `json:"name"`
}

var setName = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setName",
	GetArgsStruct: func() interface{} {
		return &setNameArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setNameArgs)
		if args.ProjectId.Equal(args.TaskId) {
			validate.MemberHasAccountAdminAccess(db.GetAccountRole(ctx, args.Shard, args.AccountId, ctx.Me()))
		} else {
			validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		}

		dbSetName(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.Name)
		return nil
	},
}

type setDescriptionArgs struct {
	Shard       int     `json:"shard"`
	AccountId   id.Id   `json:"accountId"`
	ProjectId   id.Id   `json:"projectId"`
	TaskId      id.Id   `json:"taskId"`
	Description *string `json:"description,omitempty"`
}

var setDescription = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setDescription",
	GetArgsStruct: func() interface{} {
		return &setDescriptionArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setDescriptionArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		dbSetDescription(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.Description)
		return nil
	},
}

type setIsParallelArgs struct {
	Shard      int   `json:"shard"`
	AccountId  id.Id `json:"accountId"`
	ProjectId  id.Id `json:"projectId"`
	TaskId     id.Id `json:"taskId"`
	IsParallel bool  `json:"isParallel"`
}

var setIsParallel = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setIsParallel",
	GetArgsStruct: func() interface{} {
		return &setIsParallelArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setIsParallelArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		dbSetIsParallel(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.IsParallel)
		return nil
	},
}

type setMemberArgs struct {
	Shard     int    `json:"shard"`
	AccountId id.Id  `json:"accountId"`
	ProjectId id.Id  `json:"projectId"`
	TaskId    id.Id  `json:"taskId"`
	MemberId  *id.Id `json:"memberId,omitempty"`
}

var setMember = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setMember",
	GetArgsStruct: func() interface{} {
		return &setMemberArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setMemberArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		if args.MemberId != nil {
			validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, args.Shard, args.AccountId, args.ProjectId, *args.MemberId))
		}
		dbSetMember(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.MemberId)
		return nil
	},
}

type setRemainingTimeArgs struct {
	Shard         int    `json:"shard"`
	AccountId     id.Id  `json:"accountId"`
	ProjectId     id.Id  `json:"projectId"`
	TaskId        id.Id  `json:"taskId"`
	RemainingTime uint64 `json:"remainingTime"`
}

var setRemainingTime = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setRemainingTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, &args.RemainingTime, nil, nil)
	},
}

type logTimeArgs struct {
	Shard     int     `json:"shard"`
	AccountId id.Id   `json:"accountId"`
	ProjectId id.Id   `json:"projectId"`
	TaskId    id.Id   `json:"taskId"`
	Duration  uint64  `json:"duration"`
	Note      *string `json:"note,omitempty"`
}

var logTime = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/logTime",
	GetArgsStruct: func() interface{} {
		return &logTimeArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*logTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, nil, &args.Duration, args.Note)
	},
}

type setRemainingTimeAndLogTimeArgs struct {
	Shard         int     `json:"shard"`
	AccountId     id.Id   `json:"accountId"`
	ProjectId     id.Id   `json:"projectId"`
	TaskId        id.Id   `json:"taskId"`
	RemainingTime uint64  `json:"remainingTime"`
	Duration      uint64  `json:"duration"`
	Note          *string `json:"note,omitempty"`
}

var setRemainingTimeAndLogTime = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/setRemainingTimeAndLogTime",
	GetArgsStruct: func() interface{} {
		return &setRemainingTimeAndLogTimeArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setRemainingTimeAndLogTimeArgs)
		return setRemainingTimeAndOrLogTime(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, &args.RemainingTime, &args.Duration, args.Note)
	},
}

func setRemainingTimeAndOrLogTime(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, remainingTime *uint64, duration *uint64, note *string) *timeLog {
	if duration != nil {
		validate.MemberIsAProjectMemberWithWriteAccess(db.GetProjectRole(ctx, shard, accountId, projectId, ctx.Me()))
	} else if remainingTime != nil {
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, shard, accountId, projectId, ctx.Me()))
	} else {
		panic(err.InvalidArguments)
	}

	loggedOn := t.Now()
	dbSetRemainingTimeAndOrLogTime(ctx, shard, accountId, projectId, taskId, remainingTime, &loggedOn, duration, note)
	if duration != nil {
		return &timeLog{
			Project:  projectId,
			Task:     taskId,
			Member:   ctx.Me(),
			LoggedOn: loggedOn,
			Duration: *duration,
			Note:     note,
		}
	}
	return nil
}

type moveTaskArgs struct {
	Shard              int    `json:"shard"`
	AccountId          id.Id  `json:"accountId"`
	ProjectId          id.Id  `json:"projectId"`
	TaskId             id.Id  `json:"taskId"`
	NewParentId        id.Id  `json:"newParentId"`
	NewPreviousSibling *id.Id `json:"newPreviousSibling"`
}

var moveTask = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/moveTask",
	GetArgsStruct: func() interface{} {
		return &moveTaskArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*moveTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		dbMoveTask(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId, args.NewParentId, args.NewPreviousSibling)
		return nil
	},
}

type deleteTaskArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	ProjectId id.Id `json:"projectId"`
	TaskId    id.Id `json:"taskId"`
}

var deleteTask = &core.Endpoint{
	Method: cnst.POST,
	Path:   "/api/v1/project/deleteTask",
	GetArgsStruct: func() interface{} {
		return &deleteTaskArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*deleteTaskArgs)
		validate.MemberHasProjectWriteAccess(db.GetAccountAndProjectRoles(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))

		dbDeleteTask(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskId)
		return nil
	},
}

type getTasksArgs struct {
	Shard     int     `json:"shard"`
	AccountId id.Id   `json:"accountId"`
	ProjectId id.Id   `json:"projectId"`
	TaskIds   []id.Id `json:"taskId"`
}

var getTasks = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getTasks",
	GetArgsStruct: func() interface{} {
		return &getTasksArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*getTasksArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		validate.EntityCount(len(args.TaskIds), ctx.MaxProcessEntityCount())
		return dbGetTasks(ctx, args.Shard, args.AccountId, args.ProjectId, args.TaskIds)
	},
}

type getChildTasksArgs struct {
	Shard       int    `json:"shard"`
	AccountId   id.Id  `json:"accountId"`
	ProjectId   id.Id  `json:"projectId"`
	ParentId    id.Id  `json:"parentId"`
	FromSibling *id.Id `json:"fromSibling,omitempty"`
	Limit       int    `json:"limit"`
}

var getChildTasks = &core.Endpoint{
	Method: cnst.GET,
	Path:   "/api/v1/project/getChildTasks",
	GetArgsStruct: func() interface{} {
		return &getChildTasksArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*getChildTasksArgs)
		validate.MemberHasProjectReadAccess(db.GetAccountAndProjectRolesAndProjectIsPublic(ctx, args.Shard, args.AccountId, args.ProjectId, ctx.Me()))
		validate.Limit(args.Limit, ctx.MaxProcessEntityCount())
		return dbGetChildTasks(ctx, args.Shard, args.AccountId, args.ProjectId, args.ParentId, args.FromSibling, args.Limit)
	},
}

var Endpoints = []*core.Endpoint{
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
	CreateTask(css *clientsession.Store, shard int, accountId, projectId, parentId id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *id.Id, remainingTime *uint64) (*task, error)
	SetName(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, name string) error
	SetDescription(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, description *string) error
	SetIsParallel(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, isParallel bool) error                                                              //only applys to abstract tasks
	SetMember(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, memberId *id.Id) error                                                                  //only applys to task tasks
	SetRemainingTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, remainingTime uint64) error                                                      //only applys to task tasks
	LogTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, duration uint64, note *string) (*timeLog, error)                                          //only applys to task tasks
	SetRemainingTimeAndLogTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) //only applys to task tasks
	MoveTask(css *clientsession.Store, shard int, accountId, projectId, taskId, parentId id.Id, nextSibling *id.Id) error
	DeleteTask(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id) error
	GetTasks(css *clientsession.Store, shard int, accountId, projectId id.Id, taskIds []id.Id) ([]*task, error)
	GetChildTasks(css *clientsession.Store, shard int, accountId, projectId, parentId id.Id, fromSibling *id.Id, limit int) ([]*task, error)
}

func NewClient(host string) Client {
	return &client{
		host: host,
	}
}

type client struct {
	host string
}

func (c *client) CreateTask(css *clientsession.Store, shard int, accountId, projectId, parentId id.Id, previousSibling *id.Id, name string, description *string, isAbstract bool, isParallel *bool, memberId *id.Id, totalRemainingTime *uint64) (*task, error) {
	val, e := createTask.DoRequest(css, c.host, &createTaskArgs{
		Shard:              shard,
		AccountId:          accountId,
		ProjectId:          projectId,
		ParentId:           parentId,
		PreviousSibling:    previousSibling,
		Name:               name,
		Description:        description,
		IsAbstract:         isAbstract,
		IsParallel:         isParallel,
		MemberId:           memberId,
		TotalRemainingTime: totalRemainingTime,
	}, nil, &task{})
	if val != nil {
		return val.(*task), e
	}
	return nil, e
}

func (c *client) SetName(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, name string) error {
	_, e := setName.DoRequest(css, c.host, &setNameArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId:    taskId,
		Name:      name,
	}, nil, nil)
	return e
}

func (c *client) SetDescription(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, description *string) error {
	_, e := setDescription.DoRequest(css, c.host, &setDescriptionArgs{
		Shard:       shard,
		AccountId:   accountId,
		ProjectId:   projectId,
		TaskId:      taskId,
		Description: description,
	}, nil, nil)
	return e
}

func (c *client) SetIsParallel(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, isParallel bool) error {
	_, e := setIsParallel.DoRequest(css, c.host, &setIsParallelArgs{
		Shard:      shard,
		AccountId:  accountId,
		ProjectId:  projectId,
		TaskId:     taskId,
		IsParallel: isParallel,
	}, nil, nil)
	return e
}

func (c *client) SetMember(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, memberId *id.Id) error {
	_, e := setMember.DoRequest(css, c.host, &setMemberArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId:    taskId,
		MemberId:  memberId,
	}, nil, nil)
	return e
}

func (c *client) SetRemainingTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, remainingTime uint64) error {
	_, e := setRemainingTime.DoRequest(css, c.host, &setRemainingTimeArgs{
		Shard:         shard,
		AccountId:     accountId,
		ProjectId:     projectId,
		TaskId:        taskId,
		RemainingTime: remainingTime,
	}, nil, nil)
	return e
}

func (c *client) LogTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, duration uint64, note *string) (*timeLog, error) {
	val, e := logTime.DoRequest(css, c.host, &logTimeArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId:    taskId,
		Duration:  duration,
		Note:      note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), e
	}
	return nil, e
}

func (c *client) SetRemainingTimeAndLogTime(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id, remainingTime uint64, duration uint64, note *string) (*timeLog, error) {
	val, e := setRemainingTimeAndLogTime.DoRequest(css, c.host, &setRemainingTimeAndLogTimeArgs{
		Shard:         shard,
		AccountId:     accountId,
		ProjectId:     projectId,
		TaskId:        taskId,
		RemainingTime: remainingTime,
		Duration:      duration,
		Note:          note,
	}, nil, &timeLog{})
	if val != nil {
		return val.(*timeLog), e
	}
	return nil, e
}

func (c *client) MoveTask(css *clientsession.Store, shard int, accountId, projectId, taskId, newParentId id.Id, newPreviousSibling *id.Id) error {
	_, e := moveTask.DoRequest(css, c.host, &moveTaskArgs{
		Shard:              shard,
		AccountId:          accountId,
		ProjectId:          projectId,
		TaskId:             taskId,
		NewParentId:        newParentId,
		NewPreviousSibling: newPreviousSibling,
	}, nil, nil)
	return e
}

func (c *client) DeleteTask(css *clientsession.Store, shard int, accountId, projectId, taskId id.Id) error {
	_, e := deleteTask.DoRequest(css, c.host, &deleteTaskArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskId:    taskId,
	}, nil, nil)
	return e
}

func (c *client) GetTasks(css *clientsession.Store, shard int, accountId, projectId id.Id, taskIds []id.Id) ([]*task, error) {
	val, e := getTasks.DoRequest(css, c.host, &getTasksArgs{
		Shard:     shard,
		AccountId: accountId,
		ProjectId: projectId,
		TaskIds:   taskIds,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), e
	}
	return nil, e
}

func (c *client) GetChildTasks(css *clientsession.Store, shard int, accountId, projectId, parentId id.Id, fromSibling *id.Id, limit int) ([]*task, error) {
	val, e := getChildTasks.DoRequest(css, c.host, &getChildTasksArgs{
		Shard:       shard,
		AccountId:   accountId,
		ProjectId:   projectId,
		ParentId:    parentId,
		FromSibling: fromSibling,
		Limit:       limit,
	}, nil, &[]*task{})
	if val != nil {
		return *val.(*[]*task), e
	}
	return nil, e
}

func dbCreateTask(ctx *core.Ctx, shard int, accountId, projectId, parentId id.Id, nextSibling *id.Id, newTask *task) {
	args := make([]interface{}, 0, 18)
	args = append(args, accountId, projectId, parentId, ctx.Me())
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
	db.TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
}

func dbSetName(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, name string) {
	_, e := ctx.TreeExec(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me(), name)
	err.PanicIf(e)
}

func dbSetDescription(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, description *string) {
	_, e := ctx.TreeExec(shard, `CALL setTaskDescription(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me(), description)
	err.PanicIf(e)
}

func dbSetIsParallel(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, isParallel bool) {
	db.TreeChangeHelper(ctx, shard, `CALL setTaskIsParallel(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me(), isParallel)
}

func dbSetMember(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, memberId *id.Id) {
	var memArg []byte
	if memberId != nil {
		memArg = *memberId
	}
	db.MakeChangeHelper(ctx, shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me(), memArg)
}

func dbSetRemainingTimeAndOrLogTime(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id, remainingTime *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	db.TreeChangeHelper(ctx, shard, `CALL setRemainingTimeAndOrLogTime(?, ?, ?, ?, ?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me(), remainingTime, loggedOn, duration, note)
}

func dbMoveTask(ctx *core.Ctx, shard int, accountId, projectId, taskId, newParentId id.Id, newPreviousSibling *id.Id) {
	var prevSib []byte
	if newPreviousSibling != nil {
		prevSib = *newPreviousSibling
	}
	db.TreeChangeHelper(ctx, shard, `CALL moveTask(?, ?, ?, ?, ?, ?)`, accountId, projectId, taskId, newParentId, ctx.Me(), prevSib)
}

func dbDeleteTask(ctx *core.Ctx, shard int, accountId, projectId, taskId id.Id) {
	db.TreeChangeHelper(ctx, shard, `CALL deleteTask(?, ?, ?, ?)`, accountId, projectId, taskId, ctx.Me())
}

func dbGetTasks(ctx *core.Ctx, shard int, accountId, projectId id.Id, taskIds []id.Id) []*task {
	idsStr := bytes.NewBufferString(``)
	for _, i := range taskIds {
		idsStr.WriteString(hex.EncodeToString(i))
	}
	rows, e := ctx.TreeQuery(shard, `CALL getTasks(?, ?, ?)`, accountId, projectId, idsStr.String())
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*task, 0, len(taskIds))
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	return res
}

func dbGetChildTasks(ctx *core.Ctx, shard int, accountId, projectId, parentId id.Id, fromSibling *id.Id, limit int) []*task {
	var fromSib []byte
	if fromSibling != nil {
		fromSib = *fromSibling
	}
	rows, e := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, accountId, projectId, parentId, fromSib, limit)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*task, 0, limit)
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	return res
}

func nilOutPropertiesThatAreNotNilInTheDb(ta *task) {
	if !ta.IsAbstract {
		ta.MinimumRemainingTime = nil
		ta.ChildCount = nil
		ta.DescendantCount = nil
		ta.IsParallel = nil
	}
}

type task struct {
	Id                   id.Id     `json:"id"`
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
	Member               *id.Id    `json:"member,omitempty"`          //only task tasks
}

type timeLog struct {
	Project  id.Id     `json:"project"`
	Task     id.Id     `json:"task"`
	Member   id.Id     `json:"member"`
	LoggedOn time.Time `json:"loggedOn"`
	TaskName string    `json:"taskName"`
	Duration uint64    `json:"duration"`
	Note     *string   `json:"note"`
}
