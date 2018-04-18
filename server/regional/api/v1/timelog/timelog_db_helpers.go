package timelog

import (
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	tlog "bitbucket.org/0xor1/task/server/util/timelog"
	"bytes"
	"fmt"
)

func dbGetTimeLog(ctx ctx.Ctx, shard int, account, project, timeLog id.Id) *tlog.TimeLog {
	cacheKey := cachekey.NewGet().Key("timelog.dbGetTimeLog").TimeLog(account, project, timeLog, nil, nil)
	tl := tlog.TimeLog{}
	if ctx.GetCacheValue(&tl, cacheKey, shard, account, project, timeLog) {
		return &tl
	}
	err.PanicIf(ctx.TreeQueryRow(shard, `SELECT project, task, id, member, loggedOn, taskName, duration, note FROM timeLogs WHERE account=? AND project=? AND id=?`, account, project, timeLog).Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskName, &tl.Duration, &tl.Note))
	ctx.SetCacheValue(tl, cacheKey, shard, account, project, timeLog)
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

func dbGetTimeLogs(ctx ctx.Ctx, shard int, account, project id.Id, task, member, timeLog *id.Id, sortDir cnst.SortDir, after *id.Id, limit int) []*tlog.TimeLog {
	if timeLog != nil {
		return []*tlog.TimeLog{dbGetTimeLog(ctx, shard, account, project, *timeLog)}
	}
	cacheKey := cachekey.NewGet().Key("timelog.dbGetTimeLogs")
	if task != nil {
		cacheKey.TaskTimeLogSet(account, project, *task, member)
	} else if member != nil {
		cacheKey.ProjectMemberTimeLogSet(account, project, *member)
	}
	timeLogsSet := make([]*tlog.TimeLog, 0, limit)
	if ctx.GetCacheValue(&timeLogsSet, cacheKey, shard, account, project, task, member, timeLog, sortDir, limit) {
		return timeLogsSet
	}
	query := bytes.NewBufferString(`SELECT project, task, id, member, loggedOn, taskName, duration, note FROM timeLogs WHERE account=? AND project=?`)
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
		query.WriteString(fmt.Sprintf(` AND loggedOn %s= (SELECT loggedOn FROM timeLogs WHERE account=? AND project=? AND id=?) AND id > ?`, sortDir.GtLtSymbol()))
		args = append(args, account, project, *after, *after)
	}
	query.WriteString(fmt.Sprintf(` ORDER BY loggedOn %s, id LIMIT ?`, sortDir.String()))
	args = append(args, limit)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		tl := tlog.TimeLog{}
		err.PanicIf(rows.Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskName, &tl.Duration, &tl.Note))
		timeLogsSet = append(timeLogsSet, &tl)
	}
	ctx.SetCacheValue(timeLogsSet, cacheKey, shard, account, project, task, member, timeLog, sortDir, limit)
	return timeLogsSet
}
