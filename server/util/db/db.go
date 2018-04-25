package db

import (
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/timelog"
	"bitbucket.org/0xor1/task/server/util/validate"
	"time"
)

var (
	ErrNoChangeMade = &err.Err{Code: "u_d_ncm", Message: "no change made"}
)

func GetAccountRole(ctx ctx.Ctx, shard int, account, member id.Id) *cnst.AccountRole {
	var accRole *cnst.AccountRole
	cacheKey := cachekey.NewGet("db.GetAccountRole", shard, account, member).AccountMember(account, member)
	if ctx.GetCacheValue(&accRole, cacheKey) {
		return accRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, account, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole))
	ctx.SetCacheValue(accRole, cacheKey)
	return accRole
}

func GetProjectRole(ctx ctx.Ctx, shard int, account, project, member id.Id) *cnst.ProjectRole {
	var projRole *cnst.ProjectRole
	cacheKey := cachekey.NewGet("db.GetProjectRole", shard, account, project, member).ProjectMember(account, project, member)
	if ctx.GetCacheValue(&projRole, cacheKey) {
		return projRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?`, account, project, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&projRole))
	ctx.SetCacheValue(projRole, cacheKey)
	return projRole
}

func GetAccountAndProjectRoles(ctx ctx.Ctx, shard int, account, project, member id.Id) (*cnst.AccountRole, *cnst.ProjectRole) {
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	cacheKey := cachekey.NewGet("db.GetAccountAndProjectRoles", shard, account, project, member).AccountMember(account, member).ProjectMember(account, project, member)
	if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole}, cacheKey) {
		return accRole, projRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, account, project, member, account, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole))
	ctx.SetCacheValue([]interface{}{accRole, projRole}, cacheKey)
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(ctx ctx.Ctx, shard int, account, project id.Id, member *id.Id) (*cnst.AccountRole, *cnst.ProjectRole, *bool) {
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	var isPublic *bool
	cacheKey := cachekey.NewGet("db.GetAccountAndProjectRolesAndProjectIsPublic", shard, account, project, member).Project(account, project)
	if member == nil {
		if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole, &isPublic}, cacheKey) {
			return accRole, projRole, isPublic
		}
		row := ctx.TreeQueryRow(shard, `SELECT isPublic FROM projects WHERE account=? AND id=?`, account, project)
		err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic))
		ctx.SetCacheValue([]interface{}{accRole, projRole, isPublic}, cacheKey)
	} else {
		cacheKey.AccountMember(account, *member).ProjectMember(account, project, *member)
		if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole, &isPublic}, cacheKey) {
			return accRole, projRole, isPublic
		}
		row := ctx.TreeQueryRow(shard, `SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, account, member, account, project, member, account, project)
		err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole))
		ctx.SetCacheValue([]interface{}{accRole, projRole, isPublic}, cacheKey)
	}
	return accRole, projRole, isPublic
}

func GetPublicProjectsEnabled(ctx ctx.Ctx, shard int, account id.Id) bool {
	var enabled bool
	cacheKey := cachekey.NewGet("db.GetPublicProjectsEnabled", shard, account).Account(account)
	if ctx.GetCacheValue(&enabled, cacheKey) {
		return enabled
	}
	row := ctx.TreeQueryRow(shard, `SELECT publicProjectsEnabled FROM accounts WHERE id=?`, account)
	err.PanicIf(row.Scan(&enabled))
	ctx.SetCacheValue(enabled, cacheKey)
	return enabled
}

func SetRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, duration *uint64, note *string) *timelog.TimeLog {
	var timeLog *id.Id
	if duration != nil {
		if *duration == 0 {
			panic(err.InvalidArguments)
		}
		validate.MemberIsAProjectMemberWithWriteAccess(GetProjectRole(ctx, shard, account, project, ctx.Me()))
		i := id.New()
		timeLog = &i
	} else if remainingTime != nil {
		validate.MemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, shard, account, project, ctx.Me()))
	} else {
		panic(err.InvalidArguments)
	}

	loggedOn := t.Now()
	return setRemainingTimeAndOrLogTime(ctx, shard, account, project, task, remainingTime, timeLog, &loggedOn, duration, note)
}

func setRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, timeLog *id.Id, loggedOn *time.Time, duration *uint64, note *string) *timelog.TimeLog {
	rows, e := ctx.TreeQuery(shard, `CALL setRemainingTimeAndOrLogTime( ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), remainingTime, timeLog, loggedOn, duration, note)
	err.PanicIf(e)
	tasks := make([]id.Id, 0, 100)
	var existingMember *id.Id
	var taskName string
	for rows.Next() {
		var i id.Id
		rows.Scan(&i, existingMember, &taskName)
		tasks = append(tasks, i)
	}
	if len(tasks) > 0 {
		cacheKey := cachekey.NewSetDlms().ProjectActivities(account, project).CombinedTaskAndTaskChildrenSets(account, project, tasks)
		if existingMember != nil {
			cacheKey.ProjectMember(account, project, *existingMember)
		}
		if timeLog != nil {
			cacheKey.TimeLog(account, project, *timeLog, &task, ctx.TryMe())
		}
		ctx.TouchDlms(cacheKey)
	} else {
		panic(ErrNoChangeMade)
	}
	if duration != nil {
		return &timelog.TimeLog{
			Id:                 *timeLog,
			Project:            project,
			Task:               task,
			Member:             ctx.Me(),
			LoggedOn:           *loggedOn,
			TaskHasBeenDeleted: false,
			TaskName:           taskName,
			Duration:           *duration,
			Note:               note,
		}
	}
	return nil
}

func MakeChangeHelper(ctx ctx.Ctx, shard int, sql string, args ...interface{}) {
	row := ctx.TreeQueryRow(shard, sql, args...)
	changeMade := false
	err.PanicIf(row.Scan(&changeMade))
	if !changeMade {
		panic(ErrNoChangeMade)
	}
}

func TreeChangeHelper(ctx ctx.Ctx, shard int, sql string, args ...interface{}) []id.Id {
	rows, e := ctx.TreeQuery(shard, sql, args...)
	err.PanicIf(e)
	res := make([]id.Id, 0, 100)
	for rows.Next() {
		var i id.Id
		rows.Scan(&i)
		res = append(res, i)
	}
	if len(res) == 0 {
		panic(ErrNoChangeMade)
	}
	return res
}
