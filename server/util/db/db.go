package db

import (
	"bitbucket.org/0xor1/task/server/util/cachekey"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	t "bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/timeLog"
	"bitbucket.org/0xor1/task/server/util/validate"
	"time"
)

var (
	ErrNoChangeMade = &err.Err{Code: "u_d_ncm", Message: "no change made"}
)

func GetAccountRole(ctx ctx.Ctx, shard int, account, member id.Id) *cnst.AccountRole {
	var accRole *cnst.AccountRole
	cacheKey := cachekey.NewGet().Key("db.GetAccountRole").AccountMember(account, member)
	if ctx.GetCacheValue(&accRole, cacheKey, shard, account, member) {
		return accRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, account, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole))
	ctx.SetCacheValue(accRole, cacheKey, shard, account, member)
	return accRole
}

func GetProjectRole(ctx ctx.Ctx, shard int, account, project, member id.Id) *cnst.ProjectRole {
	var projRole *cnst.ProjectRole
	cacheKey := cachekey.NewGet().Key("db.GetProjectRole").ProjectMember(account, project, member)
	if ctx.GetCacheValue(&projRole, cacheKey, shard, account, project, member) {
		return projRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?`, account, project, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&projRole))
	ctx.SetCacheValue(projRole, cacheKey, shard, account, project, member)
	return projRole
}

func GetAccountAndProjectRoles(ctx ctx.Ctx, shard int, account, project, member id.Id) (*cnst.AccountRole, *cnst.ProjectRole) {
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	cacheKey := cachekey.NewGet().Key("db.GetAccountAndProjectRoles").AccountMember(account, member).ProjectMember(account, project, member)
	if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole}, cacheKey, shard, account, project, member) {
		return accRole, projRole
	}
	row := ctx.TreeQueryRow(shard, `SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, account, project, member, account, member)
	err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole))
	ctx.SetCacheValue([]interface{}{accRole, projRole}, cacheKey, shard, account, project, member)
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(ctx ctx.Ctx, shard int, account, project id.Id, member *id.Id) (*cnst.AccountRole, *cnst.ProjectRole, *bool) {
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	var isPublic *bool
	cacheKey := cachekey.NewGet().Key("db.GetAccountAndProjectRolesAndProjectIsPublic").Project(account, project)
	if member == nil {
		if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole, &isPublic}, cacheKey, []interface{}{shard, account, project}...) {
			return accRole, projRole, isPublic
		}
		row := ctx.TreeQueryRow(shard, `SELECT isPublic FROM projects WHERE account=? AND id=?`, account, project)
		err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic))
		ctx.SetCacheValue([]interface{}{accRole, projRole, isPublic}, cacheKey, shard, account, project)
	} else {
		cacheKey.AccountMember(account, *member).ProjectMember(account, project, *member)
		if ctx.GetCacheValue(&[]interface{}{&accRole, &projRole, &isPublic}, cacheKey, shard, account, project, member) {
			return accRole, projRole, isPublic
		}
		row := ctx.TreeQueryRow(shard, `SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, account, member, account, project, member, account, project)
		err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole))
		ctx.SetCacheValue([]interface{}{accRole, projRole, isPublic}, cacheKey, shard, account, project, member)
	}
	return accRole, projRole, isPublic
}

func GetPublicProjectsEnabled(ctx ctx.Ctx, shard int, account id.Id) bool {
	var enabled bool
	cacheKey := cachekey.NewGet().Key("db.GetPublicProjectsEnabled").Account(account)
	if ctx.GetCacheValue(&enabled, cacheKey, shard, account) {
		return enabled
	}
	row := ctx.TreeQueryRow(shard, `SELECT publicProjectsEnabled FROM accounts WHERE id=?`, account)
	err.PanicIf(row.Scan(&enabled))
	ctx.SetCacheValue(enabled, cacheKey, shard, account)
	return enabled
}

func SetRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, duration *uint64, note *string) *timeLog.TimeLog {
	var timeLogId *id.Id
	if duration != nil {
		validate.MemberIsAProjectMemberWithWriteAccess(GetProjectRole(ctx, shard, account, project, ctx.Me()))
		i := id.New()
		timeLogId = &i
	} else if remainingTime != nil {
		validate.MemberHasProjectWriteAccess(GetAccountAndProjectRoles(ctx, shard, account, project, ctx.Me()))
	} else {
		panic(err.InvalidArguments)
	}

	loggedOn := t.Now()
	setRemainingTimeAndOrLogTime(ctx, shard, account, project, task, remainingTime, timeLogId, &loggedOn, duration, note)
	if duration != nil {
		return &timeLog.TimeLog{
			Id:       *timeLogId,
			Project:  project,
			Task:     task,
			Member:   ctx.Me(),
			LoggedOn: loggedOn,
			Duration: *duration,
			Note:     note,
		}
	}
	return nil
}

func setRemainingTimeAndOrLogTime(ctx ctx.Ctx, shard int, account, project, task id.Id, remainingTime *uint64, timeLogId *id.Id, loggedOn *time.Time, duration *uint64, note *string) {
	rows, e := ctx.TreeQuery(shard, `CALL setRemainingTimeAndOrLogTime( ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account, project, task, ctx.Me(), remainingTime, timeLogId, loggedOn, duration, note)
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
		panic(ErrNoChangeMade)
	}
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
