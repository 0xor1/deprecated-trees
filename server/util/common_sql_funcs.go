package util

var (
	noChangeMadeErr = &AppError{Code: "r_v1_p_nc", Message: "no change made", Public: true}
)

func GetProjectExists(ctx *Ctx, shard int, accountId, projectId Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, []byte(accountId), []byte(projectId))
	exists := false
	PanicIf(row.Scan(&exists))
	return exists
}

func GetAccountRole(ctx *Ctx, shard int, accountId, memberId Id) *AccountRole {
	row := ctx.TreeQueryRow(shard ,`SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func GetProjectRole(ctx *Ctx, shard int, accountId, projectId, memberId Id) *ProjectRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?`, []byte(accountId), []byte(projectId), []byte(memberId))
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&projRole)) {
		return nil
	}
	return projRole
}

func GetAccountAndProjectRoles(ctx *Ctx, shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	accountIdBytes := []byte(accountId)
	memberIdBytes := []byte(memberId)
	row := ctx.TreeQueryRow(shard, `SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountIdBytes, []byte(projectId), memberIdBytes, accountIdBytes, memberIdBytes)
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole)) {
		return nil, nil
	}
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(ctx *Ctx, shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	accountIdBytes := []byte(accountId)
	projectIdBytes := []byte(projectId)
	memberIdBytes := []byte(memberId)
	row := ctx.TreeQueryRow(shard, `SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, accountIdBytes, memberIdBytes, accountIdBytes, projectIdBytes, memberIdBytes, accountIdBytes, projectIdBytes)
	isPublic := false
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole)) {
		return nil, nil, nil
	}
	return accRole, projRole, &isPublic
}

func GetPublicProjectsEnabled(ctx *Ctx, shard int, accountId Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT publicProjectsEnabled FROM accounts WHERE id=?`, []byte(accountId))
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
