package task

import (
	"bitbucket.org/0xor1/trees/server/util/cachekey"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/id"
	"github.com/0xor1/panic"
)

func dbCreateTask(ctx ctx.Ctx, shard int, account, project, parent id.Id, nextSibling *id.Id, newTask *Task) {
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
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectActivities(account, project).TaskChildrenSet(account, project, parent).CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)))
}

func dbSetName(ctx ctx.Ctx, shard int, account, project, task id.Id, name string) {
	rows, e := ctx.TreeQuery(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), name)
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	var timeLog *id.Id //on the first pass this is the parent (if any)
	var member *id.Id
	firstRow := true
	cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project).Task(account, project, task)
	for rows.Next() {
		rows.Scan(&timeLog, &member)
		if firstRow && timeLog != nil { //timeLog is the parent id on the first row
			cacheKey.TaskChildrenSet(account, project, *timeLog)
		} else if timeLog != nil && member != nil {
			cacheKey.TimeLog(account, project, *timeLog, &task, member)
		}
		firstRow = false
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetDescription(ctx ctx.Ctx, shard int, account, project, task id.Id, description *string) {
	row := ctx.TreeQueryRow(shard, `CALL setTaskDescription(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), description)
	var parent *id.Id
	panic.IfNotNil(row.Scan(&parent))
	cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project).Task(account, project, task)
	if parent != nil {
		cacheKey.TaskChildrenSet(account, project, *parent)
	}
	ctx.TouchDlms(cacheKey)
}

func dbSetIsParallel(ctx ctx.Ctx, shard int, account, project id.Id, task id.Id, isParallel bool) {
	cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project)
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
	panic.IfNotNil(ctx.TreeQueryRow(shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), memArg).Scan(&changeMade, &parent, &existingMember))
	ctx.ReturnBadRequestNowIf(!changeMade, "no change made")
	cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project).TaskChildrenSet(account, project, parent).Task(account, project, task)
	if member != nil {
		cacheKey.ProjectMember(account, project, *member)
	}
	if existingMember != nil {
		cacheKey.ProjectMember(account, project, *existingMember)
	}
	ctx.TouchDlms(cacheKey)
}

func dbMoveTask(ctx ctx.Ctx, shard int, account, project, task, newParent id.Id, newPreviousSibling *id.Id) {
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectActivities(account, project).CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL moveTask(?, ?, ?, ?, ?, ?)`, account, project, task, newParent, ctx.Me(), newPreviousSibling)))
}

func dbDeleteTask(ctx ctx.Ctx, shard int, account, project, task id.Id) {
	rows, e := ctx.TreeQuery(shard, `CALL deleteTask(?, ?, ?, ?)`, account, project, task, ctx.Me())
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	affectedTasks := make([]id.Id, 0, 100)
	updatedProjectMembers := make([]id.Id, 0, 100)
	cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project)
	for rows.Next() {
		var i id.Id
		var j id.Id
		var k id.Id
		key := ""
		rows.Scan(&i, &j, &k, &key)
		switch key {
		case "t":
			affectedTasks = append(affectedTasks, i)
		case "m":
			updatedProjectMembers = append(updatedProjectMembers, i)
		case "tl":
			cacheKey.TimeLog(account, project, i, &j, &k)
		default:
			panic.If(true, "unknown key value in delete task rows")
		}
	}
	ctx.TouchDlms(cacheKey.CombinedTaskAndTaskChildrenSets(account, project, affectedTasks).ProjectMembers(account, project, updatedProjectMembers))
}

func dbGetTask(ctx ctx.Ctx, shard int, account, project id.Id, task id.Id) *Task {
	cacheKey := cachekey.NewGet("project.dbGetTask", shard, account, project, task).Task(account, project, task)
	res := Task{}
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	panic.IfNotNil(ctx.TreeQueryRow(shard, `SELECT id, parent, firstChild, nextSibling, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member FROM tasks WHERE account = ? AND project = ? AND id = ?`, account, project, task).Scan(&res.Id, &res.Parent, &res.FirstChild, &res.NextSibling, &res.IsAbstract, &res.Name, &res.Description, &res.CreatedOn, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.MinimumRemainingTime, &res.LinkedFileCount, &res.ChatCount, &res.ChildCount, &res.DescendantCount, &res.IsParallel, &res.Member))
	ctx.SetCacheValue(res, cacheKey)
	return &res
}

func dbGetChildTasks(ctx ctx.Ctx, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) *getChildrenResp {
	res := getChildrenResp{}
	cacheKey := cachekey.NewGet("project.dbGetChildTasks", shard, account, project, parent, fromSibling, limit).TaskChildrenSet(account, project, parent)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	rows, e := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, account, project, parent, fromSibling, limit+1)
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	childSet := make([]*Task, 0, limit+1)
	for rows.Next() {
		ta := Task{}
		panic.IfNotNil(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		childSet = append(childSet, &ta)
	}
	if len(childSet) == limit+1 {
		res.Children = childSet[:limit]
		res.More = true
	} else {
		res.Children = childSet
		res.More = false
	}
	ctx.SetCacheValue(res, cacheKey)
	return &res
}

func dbGetAncestorTasks(ctx ctx.Ctx, shard int, account, project, child id.Id, limit int) *getAncestorsResp {
	res := getAncestorsResp{}
	cacheKey := cachekey.NewGet("project.dbGetAncestorTasks", shard, account, project, child, limit).Task(account, project, child)
	innerCacheKey := cachekey.NewGet("project.dbGetAncestorTasks-inner", shard, account, project, child, limit)
	innerRes := true
	if ctx.GetCacheValue(&res, cacheKey) {
		for _, a := range res.Ancestors { //we have to check each task dlm here to ensure the upwards ancestor tree structure is unchanged since the result was cached
			innerCacheKey.Task(account, project, a.Id)
		}
		if ctx.GetCacheValue(&innerRes, innerCacheKey) {
			return &res
		}
		innerCacheKey.DlmKeys = map[string]bool{}
	}
	rows, e := ctx.TreeQuery(shard, `CALL getAncestorTasks(?, ?, ?, ?)`, account, project, child, limit+1)
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	ancestorSet := make([]*Ancestor, 0, limit+1)
	for rows.Next() {
		an := Ancestor{}
		panic.IfNotNil(rows.Scan(&an.Id, &an.Name))
		innerCacheKey.Task(account, project, an.Id)
		ancestorSet = append(ancestorSet, &an)
	}
	if len(ancestorSet) == limit+1 {
		res.Ancestors = ancestorSet[:limit]
		res.More = true
	} else {
		res.Ancestors = ancestorSet
		res.More = false
	}
	ctx.SetCacheValue(res, cacheKey)
	ctx.SetCacheValue(true, innerCacheKey)
	return &res
}

func nilOutPropertiesThatAreNotNilInTheDb(ta *Task) {
	if !ta.IsAbstract {
		ta.MinimumRemainingTime = nil
		ta.ChildCount = nil
		ta.DescendantCount = nil
		ta.IsParallel = nil
	}
}
