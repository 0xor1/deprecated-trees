package timelog

import (
	"bitbucket.org/0xor1/trees/server/util/cachekey"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/sortdir"
	tlog "bitbucket.org/0xor1/trees/server/util/timelog"
	"bytes"
	"fmt"
	"github.com/0xor1/panic"
)

func dbGetTimeLog(ctx ctx.Ctx, shard int, account, project, timeLog id.Id) *tlog.TimeLog {
	cacheKey := cachekey.NewGet("timelog.dbGetTimeLog", shard, account, project, timeLog).TimeLog(account, project, timeLog, nil, nil)
	tl := tlog.TimeLog{}
	if ctx.GetCacheValue(&tl, cacheKey) {
		return &tl
	}
	panic.If(ctx.TreeQueryRow(shard, `SELECT project, task, id, member, loggedOn, taskHasBeenDeleted, taskName, duration, note FROM timeLogs WHERE account=? AND project=? AND id=?`, account, project, timeLog).Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskHasBeenDeleted, &tl.TaskName, &tl.Duration, &tl.Note))
	ctx.SetCacheValue(tl, cacheKey)
	return &tl
}

func dbSetDuration(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id, duration uint64) {
	ctx.TouchDlms(cachekey.NewSetDlms().CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL setTimeLogDuration(?, ?, ?, ?, ?)`, account, project, timeLog, ctx.Me(), duration)).TimeLog(account, project, timeLog, &task, &member).ProjectActivities(account, project))
}

func dbSetNote(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id, note *string) {
	db.MakeChangeHelper(ctx, shard, `CALL setTimeLogNote(?, ?, ?, ?, ?)`, account, project, timeLog, ctx.Me(), note)
	ctx.TouchDlms(cachekey.NewSetDlms().TimeLog(account, project, timeLog, &task, &member).ProjectActivities(account, project))
}

func dbDelete(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id) {
	ctx.TouchDlms(cachekey.NewSetDlms().CombinedTaskAndTaskChildrenSets(account, project, db.TreeChangeHelper(ctx, shard, `CALL deleteTimeLog(?, ?, ?, ?)`, account, project, timeLog, ctx.Me())).TimeLog(account, project, timeLog, &task, &member).ProjectActivities(account, project))
}

func dbGetTimeLogs(ctx ctx.Ctx, shard int, account, project id.Id, task, member, timeLog *id.Id, sortAsc bool, after *id.Id, limit int) *getResp {
	if timeLog != nil {
		return &getResp{TimeLogs: []*tlog.TimeLog{dbGetTimeLog(ctx, shard, account, project, *timeLog)}}
	}
	cacheKey := cachekey.NewGet("timelog.dbGetTimeLogs", shard, account, project, task, member, timeLog, sortAsc, after, limit)
	if task != nil {
		cacheKey.TaskTimeLogSet(account, project, *task, member)
	}
	if member != nil {
		cacheKey.ProjectMemberTimeLogSet(account, project, *member)
	}
	if task == nil && member == nil {
		cacheKey.ProjectTimeLogSet(account, project)
	}
	res := getResp{}
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	query := bytes.NewBufferString(`SELECT project, task, id, member, loggedOn, taskHasBeenDeleted, taskName, duration, note FROM timeLogs WHERE account=? AND project=?`)
	args := make([]interface{}, 0, 9)
	args = append(args, account, project)
	if task != nil {
		query.WriteString(` AND task=?`)
		args = append(args, *task)
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, *member)
	}
	if after != nil {
		query.WriteString(fmt.Sprintf(` AND loggedOn %s= (SELECT loggedOn FROM timeLogs WHERE account=? AND project=? AND id=?) AND id %s ?`, sortdir.GtLtSymbol(sortAsc), sortdir.GtLtSymbol(sortAsc)))
		args = append(args, account, project, *after, *after)
	}
	query.WriteString(fmt.Sprintf(` ORDER BY loggedOn %s, id %s LIMIT ?`, sortdir.String(sortAsc), sortdir.String(sortAsc)))
	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	panic.If(e)
	timeLogsSet := make([]*tlog.TimeLog, 0, limit+1)
	for rows.Next() {
		tl := tlog.TimeLog{}
		panic.If(rows.Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskHasBeenDeleted, &tl.TaskName, &tl.Duration, &tl.Note))
		timeLogsSet = append(timeLogsSet, &tl)
	}
	if len(timeLogsSet) == limit+1 {
		res.TimeLogs = timeLogsSet[:limit]
		res.More = true
	} else {
		res.TimeLogs = timeLogsSet
		res.More = false
	}
	ctx.SetCacheValue(res, cacheKey)
	return &res
}
