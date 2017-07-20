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

func (s *sqlStore) registerAccount(id Id, ownerId Id, ownerName string) int {
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

func (s *sqlStore) getMember(shard int, org, member Id) *Member {
	row := s.shards[shard].QueryRow(`SELECT id, name, totalRemainingTime, totalLoggedTime, isActive, role FROM orgMembers WHERE org=? AND id=?`, []byte(org), []byte(member))
	res := Member{}
	if err := row.Scan(&res.Id, &res.Name, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.IsActive, &res.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &res
}

func (s *sqlStore) addMembers(shard int, org Id, members []*AddMemberInternal) {
	var query bytes.Buffer
	queryArgs := make([]interface{}, 0, 3*len(members))
	query.WriteString(`INSERT INTO orgMembers (org, id, name, role) VALUES `)
	for i, mem := range members {
		if i == 0 {
			query.WriteString(`(?, ?, ?, ?)`)
		} else {
			query.WriteString(`, (?, ?, ?, ?)`)
		}
		queryArgs = append(queryArgs, []byte(org), []byte(mem.Id), mem.Name, mem.Role)
	}
	if _, err := s.shards[shard].Exec(query.String(), queryArgs...); err != nil {
		panic(err)
	}
}

func (s *sqlStore) updateMembersAndSetActive(shard int, org Id, members []*AddMemberInternal) {
	for _, mem := range members {
		if _, err := s.shards[shard].Exec(`UPDATE orgMembers SET name=?, role=?, isActive=true WHERE org=? AND id=?`, mem.Name, mem.Role, []byte(org), []byte(mem.Id)); err != nil {
			panic(err)
		}
	}
}

func (s *sqlStore) getTotalOrgOwnerCount(shard int, org Id) int {
	var count int
	if err := s.shards[shard].QueryRow(`SELECT COUNT(*) FROM orgMembers WHERE org=? AND isActive=true AND role=0`, []byte(org)).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	return count
}

func (s *sqlStore) getOwnerCountInSet(shard int, org Id, members []Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(org))
	var query bytes.Buffer
	query.WriteString(`SELECT COUNT(*) FROM orgMembers WHERE org=? AND isActive=true AND role=0 AND id IN (`)
	for i, mem := range members {
		if i == 0 {
			query.WriteString(`?`)
		} else {
			query.WriteString(`, ?`)
		}
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`);`)
	var count int
	if err := s.shards[shard].QueryRow(query.String(), queryArgs...).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	return count
}

func (s *sqlStore) setMembersInactive(shard int, org Id, members []Id) {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(org))
	var query1 bytes.Buffer
	query1.WriteString(`UPDATE orgMembers SET isActive=false, role=3 WHERE org=? AND id IN (`)
	var query2 bytes.Buffer
	query2.WriteString(`DELETE FROM projectMembers WHERE org=? AND member IN (`)
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

func (s *sqlStore) renameMember(shard int, org Id, member Id, newName string) {
	if _, err := s.shards[shard].Exec(`UPDATE orgMembers SET name=? WHERE org=? AND id=?`, newName, []byte(org), []byte(member)); err != nil {
		panic(err)
	}
}

func (s *sqlStore) logActivity(shard int, org Id, occurredOn time.Time, item, member Id, itemType, action string) {
	unixMilli := occurredOn.UnixNano()/1000000
	if _, err := s.shards[shard].Exec(`INSERT INTO orgACtivities (org, occurredOn, item, member, itemType, itemName, action) VALUES (? , ?, ?, ?, ?, ?, ?)`, []byte(org), unixMilli, []byte(item), []byte(member), itemType, "", action); err != nil {
		panic(err)
	}
}
