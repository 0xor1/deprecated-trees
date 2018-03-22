package task

import (
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bytes"
	"encoding/hex"
	"time"
)

func dbCreateTask(ctx ctx.Ctx, shard int, account, project, parent id.Id, nextSibling *id.Id, newTask *task) {
	args := make([]interface{}, 0, 18)
	args = append(args, account, project, parent, ctx.Me())
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
	ctx.UpdateDlms(cachekey.NewSet().ProjectActivities(account, project).Tasks(account, project, db.TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)))
}

func dbSetName(ctx ctx.Ctx, shard int, account, project, task id.Id, name string) {
	_, e := ctx.TreeExec(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), name)
	err.PanicIf(e)
	ctx.UpdateDlms(cachekey.NewSet().ProjectActivities(account, project).Task(account, project, task))
}

func dbSetDescription(ctx ctx.Ctx, shard int, account, project, task id.Id, description *string) {
	_, e := ctx.TreeExec(shard, `CALL setTaskDescription(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), description)
	err.PanicIf(e)
	ctx.UpdateDlms(cachekey.NewSet().ProjectActivities(account, project).Task(account, project, task))
}

func dbSetIsParallel(ctx ctx.Ctx, shard int, account, project, task id.Id, isParallel bool) {
	ctx.UpdateDlms(cachekey.NewSet().ProjectActivities(account, project).Tasks(account, project, db.TreeChangeHelper(ctx, shard, `CALL setTaskIsParallel(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), isParallel)))
}

func dbSetMember(ctx ctx.Ctx, shard int, account, project, task id.Id, member *id.Id) {
	var memArg []byte
	if member != nil {
		memArg = *member
	}
	changeMade := false
	var existingMember *id.Id
	err.PanicIf(ctx.TreeQueryRow(shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), memArg).Scan(&changeMade, &existingMember))
	if changeMade {
		cacheKey := cachekey.NewSet().ProjectActivities(account, project).Task(account, project, task)
		if member != nil {
			cacheKey.ProjectMember(account, project, *member)
		}
		if existingMember != nil {
			cacheKey.ProjectMember(account, project, *existingMember)
		}
		ctx.UpdateDlms(cacheKey)
	} else {
		panic(db.ErrNoChangeMade)
	}
}

func dbSetRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	rows, e := ctx.TreeQuery(shard, `CALL setRemainingTimeAndOrLogTime(?, ?, ?, ?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), remainingTime, loggedOn, duration, note)
	err.PanicIf(e)
	res := make([]id.Id, 0, 100)
	var existingMember *id.Id
	for rows.Next() {
		var i id.Id
		rows.Scan(&i, existingMember)
		res = append(res, i)
	}
	cacheKey := cachekey.NewSet().ProjectActivities(account, project).Tasks(account, project, res)
	if existingMember != nil {
		cacheKey.ProjectMember(account, project, *existingMember)
	}
	ctx.UpdateDlms(cacheKey)
}

func dbMoveTask(ctx ctx.Ctx, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) {
	var prevSib []byte
	if newPreviousSibling != nil {
		prevSib = *newPreviousSibling
	}
	ctx.UpdateDlms(cachekey.NewSet().ProjectActivities(account, project).Tasks(account, project, db.TreeChangeHelper(ctx, shard, `CALL moveTask(?, ?, ?, ?, ?, ?)`, account, project, task, newParent, ctx.Me(), prevSib)))
}

func dbDeleteTask(ctx ctx.Ctx, shard int, account, project, task id.Id) {
	db.TreeChangeHelper(ctx, shard, `CALL deleteTask(?, ?, ?, ?)`, account, project, task, ctx.Me())
}

func dbGetTasks(ctx ctx.Ctx, shard int, account, project id.Id, tasks []id.Id) []*task {
	idsStr := bytes.NewBufferString(``)
	for _, i := range tasks {
		idsStr.WriteString(hex.EncodeToString(i))
	}
	rows, e := ctx.TreeQuery(shard, `CALL getTasks(?, ?, ?)`, account, project, idsStr.String())
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*task, 0, len(tasks))
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	return res
}

func dbGetChildTasks(ctx ctx.Ctx, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) []*task {
	var fromSib []byte
	if fromSibling != nil {
		fromSib = *fromSibling
	}
	rows, e := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, account, project, parent, fromSib, limit)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*task, 0, limit)
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	return res
}

func dbGetAncestorTasks(ctx ctx.Ctx, shard int, account, project, child id.Id, limit int) []*ancestor {
	rows, e := ctx.TreeQuery(shard, `CALL getAncestorTasks(?, ?, ?, ?)`, account, project, child, limit)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*ancestor, 0, limit)
	for rows.Next() {
		an := ancestor{}
		err.PanicIf(rows.Scan(&an.Id, &an.Name))
		res = append(res, &an)
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
