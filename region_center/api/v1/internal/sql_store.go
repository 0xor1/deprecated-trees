package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"database/sql"
	"github.com/0xor1/isql"
	"math/rand"
	"time"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		panic(InvalidArgumentsErr)
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

func (s *sqlStore) createAccount(id Id, ownerId Id, ownerName string) int {
	shardId := rand.Intn(len(s.shards))
	if _, err := s.shards[shardId].Exec(`CALL registerAccount(?, ?, ?);`, []byte(id), []byte(ownerId), ownerName); err != nil {
		panic(err)
	}
	return shardId
}

func (s *sqlStore) deleteAccount(shard int, account Id) {
	if _, err := s.shards[shard].Exec(`CALL deleteAccount(?);`, []byte(account)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) getMember(shard int, accountId, memberId Id) *AccountMember {
	row := s.shards[shard].QueryRow(`SELECT id, name, totalRemainingTime, totalLoggedTime, isActive, role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountMember{}
	if err := row.Scan(&res.Id, &res.Name, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.IsActive, &res.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &res
}

func (s *sqlStore) addMembers(shard int, accountId Id, members []*AddMemberInternal) {
	var query bytes.Buffer
	queryArgs := make([]interface{}, 0, 3*len(members))
	query.WriteString(`INSERT INTO accountMembers (account, id, name, role) VALUES `)
	for i, mem := range members {
		if i == 0 {
			query.WriteString(`(?, ?, ?, ?)`)
		} else {
			query.WriteString(`, (?, ?, ?, ?)`)
		}
		queryArgs = append(queryArgs, []byte(accountId), []byte(mem.Id), mem.Name, mem.Role)
	}
	if _, err := s.shards[shard].Exec(query.String(), queryArgs...); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updateMembersAndSetActive(shard int, accountId Id, members []*AddMemberInternal) {
	for _, mem := range members {
		if _, err := s.shards[shard].Exec(`UPDATE accountMembers SET name=?, role=?, isActive=true WHERE account=? AND id=?`, mem.Name, mem.Role, []byte(accountId), []byte(mem.Id)); err != nil {
			panic(err)
		}
	}
}

func (s *sqlStore) getTotalOwnerCount(shard int, accountId Id) int {
	count := 0
	if err := s.shards[shard].QueryRow(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, []byte(accountId)).Scan(&count); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	return count
}

func (s *sqlStore) getOwnerCountInSet(shard int, accountId Id, members []Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId))
	var query bytes.Buffer
	query.WriteString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (`)
	for i, mem := range members {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`);`)
	count := 0
	if err := s.shards[shard].QueryRow(query.String(), queryArgs...).Scan(&count); err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	return count
}

func (s *sqlStore) setMembersInactive(shard int, accountId Id, members []Id) {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId))
	var query1 bytes.Buffer
	query1.WriteString(`UPDATE accountMembers SET isActive=false, role=3 WHERE account=? AND id IN (`)
	var query2 bytes.Buffer
	query2.WriteString(`DELETE FROM projectMembers WHERE account=? AND member IN (`)
	for i, mem := range members {
		if i == 0 {
			query1.WriteString(`?`)
			query2.WriteString(`?`)
		} else {
			query1.WriteString(`, ?`)
			query2.WriteString(`, ?`)
		}
		queryArgs = append(queryArgs, []byte(mem))
	}
	query1.WriteString(`);`)
	query2.WriteString(`);`)
	if _, err := s.shards[shard].Exec(query1.String(), queryArgs...); err != nil {
		panic(err)
	}
	if _, err := s.shards[shard].Exec(query2.String(), queryArgs...); err != nil {
		panic(err)
	}
}

func (s *sqlStore) renameMember(shard int, accountId Id, member Id, newName string) {
	if _, err := s.shards[shard].Exec(`UPDATE accountMembers SET name=? WHERE account=? AND id=?`, newName, []byte(accountId), []byte(member)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) logActivity(shard int, accountId Id, occurredOn time.Time, item, member Id, itemType, action string) {
	unixMilli := occurredOn.UnixNano()/1000000
	if _, err := s.shards[shard].Exec(`INSERT INTO accountActivities (account, occurredOn, item, member, itemType, itemName, action) VALUES (? , ?, ?, ?, ?, ?, ?)`, []byte(accountId), unixMilli, []byte(item), []byte(member), itemType, "", action); err != nil {
		panic(err)
	}
}
