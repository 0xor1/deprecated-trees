package timelog

import (
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/id"
	tlog "bitbucket.org/0xor1/task/server/util/timelog"
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/err"
)

func dbGetTimeLog(ctx ctx.Ctx, shard int, account, project, timeLog id.Id) *tlog.TimeLog {
	cacheKey := cachekey.NewGet().Key("timelog.dbGetTimeLog").TimeLog(account, project, timeLog, nil, nil)
	tl := tlog.TimeLog{}
	if ctx.GetCacheValue(&tl, cacheKey, shard, account, project, timeLog) {
		return &tl
	}
	ctx.TreeQueryRow(shard, `SELECT project, task, id, member, loggedOn, taskName, duration, note FROM timeLogs WHERE account=? AND project=? AND id=?`, account, project, timeLog).Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskName, &tl.Duration, &tl.Note)
	ctx.SetCacheValue(tl, cacheKey, shard, account, project, timeLog)
	return &tl
}

func dbSetDuration(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id, duration uint64) {
	ctx.TreeExec(shard, `CALL setTimeLogDuration(?, ?, ?, ?, ?, ?)`, account, project, task, timeLog, ctx.Me(), duration)
	ctx.TouchDlms(cachekey.NewSetDlms().TimeLog(account, project, timeLog, &task, &member))
}

func dbSetNote(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id, note *string) {
	ctx.TreeExec(shard, `CALL setTimeLogNote(?, ?, ?, ?, ?, ?)`, account, project, task, timeLog, ctx.Me(), note)
	ctx.TouchDlms(cachekey.NewSetDlms().TimeLog(account, project, timeLog, &task, &member))
}

func dbDelete(ctx ctx.Ctx, shard int, account, project, task, member, timeLog id.Id) {
	ctx.TreeExec(shard, `CALL deleteTimeLogNote(?, ?, ?, ?, ?, ?)`, account, project, task, timeLog, ctx.Me())
	ctx.TouchDlms(cachekey.NewSetDlms().TimeLog(account, project, timeLog, &task, &member))
}

func dbGetTimeLogs(ctx ctx.Ctx, shard int, account, project id.Id, task, member, timeLog *id.Id, limit int) []*tlog.TimeLog {
	if timeLog != nil {
		return []*tlog.TimeLog{dbGetTimeLog(ctx, shard, account, project, *timeLog)}
	}
	cacheKey := cachekey.NewGet().Key("timelog.dbGetTimeLogs")
	if task != nil {
		cacheKey.TaskTimeLogSet(account, project, *task, member)
	} else if member != nil {
		cacheKey.ProjectMemberTimeLogSet(account, project, *member)
	}
	res := make([]*tlog.TimeLog, 0, limit)
	if ctx.GetCacheValue(&res, cacheKey, shard, account, project, task, member, timeLog, limit) {
		return res
	}
	rows, e := ctx.TreeQuery(shard, `CALL getTimeLogs(?, ?, ?, ?, ?, ?)`)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		tl := tlog.TimeLog{}
		err.PanicIf(rows.Scan(&tl.Project, &tl.Task, &tl.Id, &tl.Member, &tl.LoggedOn, &tl.TaskName, &tl.Duration, &tl.Note))
		res = append(res, &tl)
	}
	ctx.SetCacheValue(res, cacheKey, shard, account, project, task, member, timeLog, limit)
	return res
}