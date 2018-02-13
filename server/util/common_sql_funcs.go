package util

import (
	"github.com/0xor1/isql"
)

var (
	noChangeMadeErr = &AppError{Code: "r_v1_p_nc", Message: "no change made", Public: true}
)

func GetAccountRole(shard isql.ReplicaSet, accountId, memberId Id) *AccountRole {
	row := shard.QueryRow(`SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func GetAccountAndProjectRoles(shard isql.ReplicaSet, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	accountIdBytes := []byte(accountId)
	memberIdBytes := []byte(memberId)
	row := shard.QueryRow(`SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountIdBytes, []byte(projectId), memberIdBytes, accountIdBytes, memberIdBytes)
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&accRole, &projRole)) {
		return nil, nil
	}
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(shard isql.ReplicaSet, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	accountIdBytes := []byte(accountId)
	projectIdBytes := []byte(projectId)
	memberIdBytes := []byte(memberId)
	row := shard.QueryRow(`SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, accountIdBytes, memberIdBytes, accountIdBytes, projectIdBytes, memberIdBytes, accountIdBytes, projectIdBytes)
	isPublic := false
	var accRole *AccountRole
	var projRole *ProjectRole
	if IsSqlErrNoRowsAndPanicIf(row.Scan(&isPublic, &accRole, &projRole)) {
		return nil, nil, nil
	}
	return accRole, projRole, &isPublic
}

func GetPublicProjectsEnabled(shard isql.ReplicaSet, accountId Id) bool {
	row := shard.QueryRow(`SELECT publicProjectsEnabled FROM accounts WHERE id=?`, []byte(accountId))
	res := false
	PanicIf(row.Scan(&res))
	return res
}

func MakeChangeHelper(shard isql.ReplicaSet, sql string, args ...interface{}) {
	row := shard.QueryRow(sql, args...)
	changeMade := false
	PanicIf(row.Scan(&changeMade))
	if !changeMade {
		noChangeMadeErr.Panic()
	}
}
