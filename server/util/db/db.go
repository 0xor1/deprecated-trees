package db

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"fmt"
)

var (
	noChangeMadeErr = &err.Err{Code: "r_v1_p_nc", Message: "no change made"}
)

func GetProjectExists(ctx ctx.Ctx, shard int, accountId, projectId id.Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, accountId, projectId)
	exists := false
	err.PanicIf(row.Scan(&exists))
	return exists
}

func GetAccountRole(ctx ctx.Ctx, shard int, accountId, memberId id.Id) *cnst.AccountRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountId, memberId)
	res := cnst.AccountRole(3)
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func GetProjectRole(ctx ctx.Ctx, shard int, accountId, projectId, memberId id.Id) *cnst.ProjectRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?`, accountId, projectId, memberId)
	var projRole *cnst.ProjectRole
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&projRole)) {
		return nil
	}
	return projRole
}

func GetAccountAndProjectRoles(ctx ctx.Ctx, shard int, accountId, projectId, memberId id.Id) (*cnst.AccountRole, *cnst.ProjectRole) {
	row := ctx.TreeQueryRow(shard, `SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountId, projectId, memberId, accountId, memberId)
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole)) {
		return nil, nil
	}
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(ctx ctx.Ctx, shard int, accountId, projectId, memberId id.Id) (*cnst.AccountRole, *cnst.ProjectRole, *bool) {
	row := ctx.TreeQueryRow(shard, `SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, accountId, memberId, accountId, projectId, memberId, accountId, projectId)
	isPublic := false
	var accRole *cnst.AccountRole
	var projRole *cnst.ProjectRole
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole)) {
		return nil, nil, nil
	}
	return accRole, projRole, &isPublic
}

func GetPublicProjectsEnabled(ctx ctx.Ctx, shard int, accountId id.Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT publicProjectsEnabled FROM accounts WHERE id=?`, accountId)
	res := false
	err.PanicIf(row.Scan(&res))
	return res
}

func MakeChangeHelper(ctx ctx.Ctx, shard int, sql string, args ...interface{}) {
	row := ctx.TreeQueryRow(shard, sql, args...)
	changeMade := false
	err.PanicIf(row.Scan(&changeMade))
	if !changeMade {
		panic(noChangeMadeErr)
	}
}

func TreeChangeHelper(ctx ctx.Ctx, shard int, sql string, args ...interface{}) {
	rows, e := ctx.TreeQuery(shard, sql, args...)
	err.PanicIf(e)
	res := make([]id.Id, 0, 100)
	for rows.Next() {
		var i id.Id
		rows.Scan(&i)
		res = append(res, i)
	}
	//TODO break cache for all returned items
	fmt.Println("YOLO", res)
}
