package util

import "fmt"

var (
	noChangeMadeErr = &AppError{Code: "r_v1_p_nc", Message: "no change made", Public: true}
)

func GetProjectExists(ctx *Ctx, shard int, accountId, projectId Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, accountId, projectId)
	exists := false
	PanicIf(row.Scan(&exists))
	return exists
}

func GetAccountRole(ctx *Ctx, shard int, accountId, memberId Id) *AccountRole {
	row := ctx.TreeQueryRow(shard ,`SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountId, memberId)
	res := AccountRole(3)
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func GetProjectRole(ctx *Ctx, shard int, accountId, projectId, memberId Id) *ProjectRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?`, accountId, projectId, memberId)
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&projRole)) {
		return nil
	}
	return projRole
}

func GetAccountAndProjectRoles(ctx *Ctx, shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	row := ctx.TreeQueryRow(shard, `SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountId, projectId, memberId, accountId, memberId)
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole)) {
		return nil, nil
	}
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(ctx *Ctx, shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	row := ctx.TreeQueryRow(shard, `SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, accountId, memberId, accountId, projectId, memberId, accountId, projectId)
	isPublic := false
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole)) {
		return nil, nil, nil
	}
	return accRole, projRole, &isPublic
}

func GetPublicProjectsEnabled(ctx *Ctx, shard int, accountId Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT publicProjectsEnabled FROM accounts WHERE id=?`, accountId)
	res := false
	PanicIf(row.Scan(&res))
	return res
}

func MakeChangeHelper(ctx *Ctx, shard int, sql string, args ...interface{}) {
	row := ctx.TreeQueryRow(shard, sql, args...)
	changeMade := false
	PanicIf(row.Scan(&changeMade))
	if !changeMade {
		noChangeMadeErr.Panic()
	}
}

func TreeChangeHelper(ctx *Ctx, shard int, sql string, args ...interface{}) {
	rows, err := ctx.TreeQuery(shard, sql, args...)
	PanicIf(err)
	res := make([]Id, 0, 100)
	for rows.Next() {
		var id Id
		rows.Scan(&id)
		res = append(res, id)
	}
	//TODO break cache for all returned items
	fmt.Println("YOLO", res)
}
