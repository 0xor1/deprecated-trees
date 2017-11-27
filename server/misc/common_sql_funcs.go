package misc

import (
	"bytes"
	"github.com/0xor1/isql"
)

func GetProjectExists(shard isql.ReplicaSet, accountId, projectId Id) bool {
	row := shard.QueryRow(`SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, []byte(accountId), []byte(projectId))
	exists := false
	PanicIf(row.Scan(&exists))
	return exists
}

func GetNodeExists(shard isql.ReplicaSet, accountId, projectId Id, nodeId Id) bool {
	row := shard.QueryRow(`SELECT COUNT() = 1 FROM nodes WHERE account=? AND project=? AND id=?`, []byte(accountId), []byte(projectId), []byte(nodeId))
	exists := false
	PanicIf(row.Scan(&exists))
	return exists
}

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

func LogAccountActivity(shard isql.ReplicaSet, accountId, member, item Id, itemType, action string, newValue *string) {
	_, err := shard.Exec(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue)
	PanicIf(err)
}

func LogProjectActivity(shard isql.ReplicaSet, accountId, projectId, member, item Id, itemType, action string, newValue *string) {
	_, err := shard.Exec(`INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue)
	PanicIf(err)
}

func LogAccountBatchAddOrRemoveMembersActivity(shard isql.ReplicaSet, accountId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	args = append(args, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(members[0]), "member", "", action, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(memId), "member", "", action, nil)
	}
	_, err := shard.Exec(query.String(), args...)
	PanicIf(err)
}

func LogProjectBatchAddOrRemoveMembersActivity(shard isql.ReplicaSet, accountId, projectId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (?,?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*9)
	args = append(args, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(members[0]), "member", "", action, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?,?)`)
		args = append(args, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(memId), "member", "", action, nil)
	}
	_, err := shard.Exec(query.String(), args...)
	PanicIf(err)
}
