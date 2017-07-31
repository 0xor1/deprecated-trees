package misc

import (
	"github.com/0xor1/isql"
)

func LogAccountActivity(shard isql.ReplicaSet, accountId, member, item Id, itemType, action string, newValue *string) {
	if _, err := shard.Exec(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue); err != nil {
		panic(err)
	}
}

func LogProjectActivity(shard isql.ReplicaSet, accountId, projectId, member, item Id, itemType, action string, newValue *string) {
	if _, err := shard.Exec(`INSERT INTO projectActivities (account, project, occurredOn, member, item, itemType, itemName, action, newValue) VALUES (? , ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), Now().UnixNano()/1000000, []byte(member), []byte(item), itemType, "", action, newValue); err != nil {
		panic(err)
	}
}
