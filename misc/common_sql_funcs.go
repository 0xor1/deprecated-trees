package misc

import (
	"bytes"
	"database/sql"
	"github.com/0xor1/isql"
)

func GetAccountRole(shard isql.ReplicaSet, accountId, memberId Id) *AccountRole {
	row := shard.QueryRow(`SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if err := row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &res
}

func GetAccountAndProjectRoles(shard isql.ReplicaSet, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	accountIdBytes := []byte(accountId)
	memberIdBytes := []byte(memberId)
	row := shard.QueryRow(`SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountIdBytes, []byte(projectId), memberIdBytes, accountIdBytes, memberIdBytes)
	var accRole *AccountRole
	var projRole *ProjectRole
	if err := row.Scan(&accRole, &projRole); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		panic(err)
	}
	return accRole, projRole
}

func GetAccountAndProjectRolesAndProjectIsPublic(shard isql.ReplicaSet, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	accountIdBytes := []byte(accountId)
	projectIdBytes := []byte(projectId)
	memberIdBytes := []byte(memberId)
	row := shard.QueryRow(`SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND project=?`, accountIdBytes, memberIdBytes, accountIdBytes, projectIdBytes, memberIdBytes, accountIdBytes, projectIdBytes)
	isPublic := false
	accRole := AccountRole(3)
	projRole := ProjectRole(2)
	if err := row.Scan(&isPublic, &accRole, &projRole); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		panic(err)
	}
	return &accRole, &projRole, &isPublic
}

func GetPublicProjectsEnabled(shard isql.ReplicaSet, accountId Id) bool {
	row := shard.QueryRow(`SELECT publicProjectsEnabled FROM accounts WHERE id=?`, []byte(accountId))
	res := false
	if err := row.Scan(&res); err != nil {
		panic(err)
	}
	return res
}

func LogAccountActivity(shard isql.ReplicaSet, accountId, member, item Id, itemType, action string, newValue *string) {
	if _, err := shard.Exec(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue); err != nil {
		panic(err)
	}
}

func LogProjectActivity(shard isql.ReplicaSet, accountId, projectId, member, item Id, itemType, action string, newValue *string) {
	if _, err := shard.Exec(`INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue); err != nil {
		panic(err)
	}
}

func LogAccountBatchAddOrRemoveMembersActivity(shard isql.ReplicaSet, accountId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, itemName, action, newValue) VALUES `)
	args := make([]interface{}, 0, len(members)*8)
	for i, memId := range members {
		if i != 0 {
			query.WriteString(`, `)
		}
		query.WriteString(` (? , ?, ?, ?, ?, ?, ?, ?)`)
		args = append(args, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(memId), "member", "", action, nil)
	}
	if _, err := shard.Exec(query.String(), args...); err != nil {
		panic(err)
	}
}

func LogProjectBatchAddOrRemoveMembersActivity(shard isql.ReplicaSet, accountId, projectId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, itemName, action, newValue) VALUES `)
	args := make([]interface{}, 0, len(members)*9)
	for i, memId := range members {
		if i != 0 {
			query.WriteString(`, `)
		}
		query.WriteString(` (? , ?, ?, ?, ?, ?, ?, ?, ?)`)
		args = append(args, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(memId), "member", "", action, nil)
	}
	if _, err := shard.Exec(query.String(), args...); err != nil {
		panic(err)
	}
}
