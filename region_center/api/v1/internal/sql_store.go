package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"database/sql"
	"github.com/0xor1/isql"
	"math/rand"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		panic(NilCriticalParamErr)
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

var query_registerOrgAccount = `CALL registerOrgAccount(?, ?, ?);`

func (s *sqlStore) registerAccount(id Id, ownerId Id, ownerName string) int {
	shardId := rand.Intn(len(s.shards))
	if _, err := s.shards[shardId].Exec(query_registerOrgAccount, []byte(id), []byte(ownerId), ownerName); err != nil {
		panic(err)
	}
	return shardId
}

var query_deleteAccount = `CALL deleteAccount(?);`

func (s *sqlStore) deleteAccount(shard int, account Id) {
	if _, err := s.shards[shard].Exec(query_deleteAccount, []byte(account)); err != nil {
		panic(err)
	}
}

var query_getMember = `SELECT org, id, name, totalRemainingTime, totalLoggedTime, isActive, role FROM orgMembers WHERE org=? AND id=?`

func (s *sqlStore) getMember(shard int, org, member Id) *orgMember {
	row := s.shards[shard].QueryRow(query_getMember, []byte(org), []byte(member))
	res := orgMember{}
	if err := row.Scan(&res.Org, &res.Id, &res.Name, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.IsActive, &res.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		panic(err)
	}
	return &res
}

var query_addMembers = `INSERT INTO orgMembers (org, id, name, role) VALUES `

func (s *sqlStore) addMembers(shard int, org Id, members []*AddMemberInternal) {
	var query bytes.Buffer
	queryArgs := make([]interface{}, 0, 3*len(members))
	query.WriteString(query_addMembers)
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

var query_updateMembersAndSetActive = `UPDATE orgMembers SET name=?, role=?, isActive=true WHERE org=? AND id=?`

func (s *sqlStore) updateMembersAndSetActive(shard int, org Id, members []*AddMemberInternal) {
	for _, mem := range members {
		if _, err := s.shards[shard].Exec(query_updateMembersAndSetActive, mem.Name, mem.Role, []byte(org), []byte(mem.Id)); err != nil {
			panic(err)
		}
	}
}

var query_getTotalOrgOwnerCount = `SELECT COUNT(*) FROM orgMembers WHERE org=? AND isActive=true AND role=0`

func (s *sqlStore) getTotalOrgOwnerCount(shard int, org Id) int {
	var count int
	if err := s.shards[shard].QueryRow(query_getTotalOrgOwnerCount, []byte(org)).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	return count
}

var query_getOwnerCountInSet = `SELECT COUNT(*) FROM orgMembers WHERE org=? AND isActive=true AND role=0 AND id IN (`

func (s *sqlStore) getOwnerCountInSet(shard int, org Id, members []Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(org))
	var query bytes.Buffer
	query.WriteString(query_getOwnerCountInSet)
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

var query_setMembersInactive_1 = `UPDATE orgMembers SET isActive=false, role=3 WHERE org=? AND id IN (`
var query_setMembersInactive_2 = `DELETE FROM projectMembers WHERE org=? AND member IN (`

func (s *sqlStore) setMembersInactive(shard int, org Id, members []Id) {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(org))
	var query1 bytes.Buffer
	query1.WriteString(query_setMembersInactive_1)
	var query2 bytes.Buffer
	query2.WriteString(query_setMembersInactive_2)
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

var query_renameMember = `UPDATE orgMembers SET name=? WHERE org=? AND id=?`

func (s *sqlStore) renameMember(shard int, org Id, member Id, newName string) {
	if _, err := s.shards[shard].Exec(query_renameMember, newName, []byte(org), []byte(member)); err != nil {
		panic(err)
	}
}
