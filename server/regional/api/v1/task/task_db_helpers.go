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
	ctx.TouchDlms(cachekey.NewDlms().ProjectActivities(account, project).TaskChildrenSet(account, project, parent).CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)))
}

func dbSetName(ctx ctx.Ctx, shard int, account, project, task id.Id, name string) {
	row := ctx.TreeQueryRow(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), name)
	var parent *id.Id
	err.PanicIf(row.Scan(&parent))
	cacheKey := cachekey.NewDlms().ProjectActivities(account, project).Task(account, project, task)
	if parent != nil {
		cacheKey.TaskChildrenSet(account, project, *parent)
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetDescription(ctx ctx.Ctx, shard int, account, project, task id.Id, description *string) {
	row := ctx.TreeQueryRow(shard, `CALL setTaskDescription(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), description)
	var parent *id.Id
	err.PanicIf(row.Scan(&parent))
	cacheKey := cachekey.NewDlms().ProjectActivities(account, project).Task(account, project, task)
	if parent != nil {
		cacheKey.TaskChildrenSet(account, project, *parent)
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetIsParallel(ctx ctx.Ctx, shard int, account, project id.Id, task id.Id, isParallel bool) {
	cacheKey := cachekey.NewDlms().ProjectActivities(account, project)
	ctx.TouchDlms(cacheKey.CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL setTaskIsParallel(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), isParallel)))
}

func dbSetMember(ctx ctx.Ctx, shard int, account, project, task id.Id, member *id.Id) {
	var memArg []byte
	if member != nil {
		memArg = *member
	}
	changeMade := false
	var parent id.Id
	var existingMember *id.Id
	err.PanicIf(ctx.TreeQueryRow(shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), memArg).Scan(&changeMade, &parent, &existingMember))
	if changeMade {
		cacheKey := cachekey.NewDlms().ProjectActivities(account, project).TaskChildrenSet(account, project, parent).Task(account, project, task)
		if member != nil {
			cacheKey.ProjectMember(account, project, *member)
		}
		if existingMember != nil {
			cacheKey.ProjectMember(account, project, *existingMember)
		}
		ctx.TouchDlms(cacheKey)
	} else {
		panic(db.ErrNoChangeMade)
	}
}

func dbSetRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	rows, e := ctx.TreeQuery(shard, `CALL setRemainingTimeAndOrLogTime( ?, ?, ?, ?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), remainingTime, loggedOn, duration, note)
	err.PanicIf(e)
	tasks := make([]id.Id, 0, 100)
	var existingMember *id.Id
	for rows.Next() {
		var i id.Id
		rows.Scan(&i, existingMember)
		tasks = append(tasks, i)
	}
	if len(tasks) > 0 {
		cacheKey := cachekey.NewDlms().ProjectActivities(account, project).CombinedTaskAndTaskChildrenSets(account, project, tasks)
		if existingMember != nil {
			cacheKey.ProjectMember(account, project, *existingMember)
		}
		ctx.TouchDlms(cacheKey)
	} else {
		panic(db.ErrNoChangeMade)
	}
}

func dbMoveTask(ctx ctx.Ctx, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) {
	ctx.TouchDlms(cachekey.NewDlms().ProjectActivities(account, project).CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL moveTask(?, ?, ?, ?, ?, ?)`, account, project, task, newParent, ctx.Me(), newPreviousSibling)))
}

func dbDeleteTask(ctx ctx.Ctx, shard int, account, project, task id.Id) {
	rows, e := ctx.TreeQuery(shard, `CALL deleteTask(?, ?, ?, ?)`, account, project, task, ctx.Me())
	err.PanicIf(e)
	affectedTasks := make([]id.Id, 0, 100)
	updatedProjectMembers := make([]id.Id, 0, 100)
	for rows.Next() {
		deletedTasksCount := 0
		var i id.Id
		rows.Scan(&i, &deletedTasksCount)
		if len(affectedTasks) < deletedTasksCount {
			affectedTasks = append(affectedTasks, i)
		} else {
			updatedProjectMembers = append(updatedProjectMembers, i)
		}
	}
	ctx.TouchDlms(cachekey.NewDlms().ProjectActivities(account, project).CombinedTaskAndTaskChildrenSets(account, project, affectedTasks).ProjectMembers(account, project, updatedProjectMembers))
}

func dbGetTasks(ctx ctx.Ctx, shard int, account, project id.Id, tasks []id.Id) []*task {
	ids := bytes.NewBufferString(``)
	for _, i := range tasks {
		ids.WriteString(hex.EncodeToString(i))
	}
	idsStr := ids.String()
	res := make([]*task, 0, len(tasks))
	cacheKey := cachekey.NewGet().Key("project.dbGetTasks").CombinedTaskAndTaskChildrenSets(account, project, tasks)
	if ctx.GetCacheValue(&res, cacheKey, shard, account, project, idsStr) {
		return res
	}
	rows, e := ctx.TreeQuery(shard, `CALL getTasks(?, ?, ?)`, account, project, idsStr)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	ctx.SetCacheValue(res, cacheKey, shard, account, project, idsStr)
	return res
}

func dbGetChildTasks(ctx ctx.Ctx, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) []*task {
	res := make([]*task, 0, limit)
	cacheKey := cachekey.NewGet().Key("project.dbGetChildTasks").TaskChildrenSet(account, project, parent)
	if ctx.GetCacheValue(&res, cacheKey, shard, account, project, parent, fromSibling, limit) {
		return res
	}
	rows, e := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, account, project, parent, fromSibling, limit)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		ta := task{}
		err.PanicIf(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	ctx.SetCacheValue(res, cacheKey, shard, account, project, parent, fromSibling, limit)
	return res
}

func dbGetAncestorTasks(ctx ctx.Ctx, shard int, account, project, child id.Id, limit int) []*ancestor {
	// note to future dan:
	// I believe this is uncachable, the cache system is based on breaking dlms for entities higher up the entity tree,
	// so we can only cache moving down the entity tree, but this operation goes up the tree, and is therefore uncachable
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
