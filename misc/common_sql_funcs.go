package misc

import (
	"bytes"
	"github.com/0xor1/isql"
)

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
	args := make([]interface{}, 0, len(members)*8)
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
