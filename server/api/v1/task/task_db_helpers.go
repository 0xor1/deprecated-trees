package task

import (
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bytes"
	"github.com/0xor1/panic"
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
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectActivities(account, project).TaskChildrenSet(account, project, parent).CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)))
}

func dbSetName(ctx ctx.Ctx, shard int, account, project, task id.Id, name string) {
	rows, e := ctx.TreeQuery(shard, `CALL setTaskName(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), name)
	panic.If(e)
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
	panic.If(row.Scan(&parent))
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
	panic.If(ctx.TreeQueryRow(shard, `CALL setTaskMember(?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), memArg).Scan(&changeMade, &parent, &existingMember))
	panic.IfTrueWith(!changeMade, db.ErrNoChangeMade)
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
	panic.If(e)
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
			panic.If(err.InvalidOperation)
		}
	}
	ctx.TouchDlms(cacheKey.CombinedTaskAndTaskChildrenSets(account, project, affectedTasks).ProjectMembers(account, project, updatedProjectMembers))
}

func dbGetTasks(ctx ctx.Ctx, shard int, account, project id.Id, tasks []id.Id) []*task {
	cacheKey := cachekey.NewGet("project.dbGetTasks", shard, account, project, tasks)
	res := make([]*task, 0, len(tasks))
	if ctx.GetCacheValue(&res, cacheKey) {
		return res
	}
	query := bytes.NewBufferString(`SELECT id, parent, firstChild, nextSibling, isAbstract, name, description, createdOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel, member FROM tasks WHERE account = ? AND project = ? AND id IN (?`)
	queryArgs := make([]interface{}, 0, 2+len(tasks))
	queryArgs = append(queryArgs, account, project, tasks[0])
	for _, task := range tasks[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, task)
	}
	query.WriteString(`)`)
	rows, e := ctx.TreeQuery(shard, query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	for rows.Next() {
		ta := task{}
		panic.If(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	ctx.SetCacheValue(res, cacheKey)
	return res
}

func dbGetChildTasks(ctx ctx.Ctx, shard int, account, project, parent id.Id, fromSibling *id.Id, limit int) []*task {
	res := make([]*task, 0, limit)
	cacheKey := cachekey.NewGet("project.dbGetChildTasks", shard, account, project, parent, fromSibling, limit).TaskChildrenSet(account, project, parent)
	if ctx.GetCacheValue(&res, cacheKey) {
		return res
	}
	rows, e := ctx.TreeQuery(shard, `CALL getChildTasks(?, ?, ?, ?, ?)`, account, project, parent, fromSibling, limit)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	for rows.Next() {
		ta := task{}
		panic.If(rows.Scan(&ta.Id, &ta.Parent, &ta.FirstChild, &ta.NextSibling, &ta.IsAbstract, &ta.Name, &ta.Description, &ta.CreatedOn, &ta.TotalRemainingTime, &ta.TotalLoggedTime, &ta.MinimumRemainingTime, &ta.LinkedFileCount, &ta.ChatCount, &ta.ChildCount, &ta.DescendantCount, &ta.IsParallel, &ta.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&ta)
		res = append(res, &ta)
	}
	ctx.SetCacheValue(res, cacheKey)
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
	panic.If(e)
	res := make([]*ancestor, 0, limit)
	for rows.Next() {
		an := ancestor{}
		panic.If(rows.Scan(&an.Id, &an.Name))
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
