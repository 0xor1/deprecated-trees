package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
	"math/rand"
	"bytes"
)

var (
	invalidShardIdErr = &Error{Code: 23, Msg: "invalid shard id"}
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if shards == nil || len(shards) == 0 {
		NilCriticalParamPanic("shards")
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

var query_registerPersonalAccount = `INSERT INTO orgs (id) VALUES (?);`

func (s *sqlStore) registerPersonalAccount(id Id) (int, error) {
	shardId := rand.Intn(len(s.shards))
	_, err := s.shards[shardId].Exec(query_registerPersonalAccount, []byte(id))
	return shardId, err
}

var query_registerOrgAccount = `CALL registerOrgAccount(?, ?, ?);`

func (s *sqlStore) registerOrgAccount(id Id, ownerId Id, ownerName string) (int, error) {
	shardId := rand.Intn(len(s.shards))
	_, err := s.shards[shardId].Exec(query_registerOrgAccount, []byte(id), []byte(ownerId), ownerName)
	return shardId, err
}

var query_deleteAccount = `CALL deleteAccount(?);`

func (s *sqlStore) deleteAccount(shard int, account Id) error {
	if s.shards[shard] == nil {
		return invalidShardIdErr
	}
	_, err := s.shards[shard].Exec(query_deleteAccount, []byte(account))
	return err
}

var query_getMember = `SELECT org, id, name, totalRemainingTime, totalLoggedTime, isActive, isDeleted, role FROM orgMembers WHERE org=? AND id=?`

func (s *sqlStore) getMember(shard int, org, member Id) (*orgMember, error) {
	if s.shards[shard] == nil {
		return invalidShardIdErr
	}
	row := s.shards[shard].QueryRow(query_getMember, []byte(org), []byte(member))
	res := orgMember{}
	err := row.Scan(&res.Org, &res.Id, &res.Name, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.IsActive, &res.IsDeleted, &res.Role)
	if err == nil {
		return &res, nil
	}
	return nil, err
}

var query_addMembers = `INSERT INTO orgMembers (org, id, name, role) VALUES `

func (s *sqlStore) addMembers(shard int, org Id, members []*AddMemberInternal) error {
	if s.shards[shard] == nil {
		return invalidShardIdErr
	}
	var query bytes.Buffer
	queryArgs := make([]interface{}, 0, 3*len(members))
	query.WriteString(query_addMembers)
	for i, mem := range members {
		if i == 0 {
			query.WriteString(`(?, ?, ?)`)
		} else {
			query.WriteString(`, (?, ?, ?)`)
		}
		queryArgs = append(queryArgs, []byte(org), []byte(mem.Id), mem.Name, mem.Role)
	}
	_, err := s.shards[shard].Exec(query_addMembers, queryArgs...)
	return err
}

var query_updateMembers = `UPDATE orgMembers name=?, role=?, isActive=true WHERE org=? AND id=?`

func (s *sqlStore) updateMembers(shard int, org Id, members []*AddMemberInternal) error {
	if s.shards[shard] == nil {
		return invalidShardIdErr
	}
	for _, mem := range members {
		if _, err := s.shards[shard].Exec(query_updateMembers, mem.Name, mem.Role, []byte(org), []byte(mem.Id)); err != nil {
			return err
		}
	}
	return nil
}

var query_getTotalOrgOwnerCount = `SELECT COUNT(*) WHERE org=? AND isActive=true AND role=0`

func (s *sqlStore) getTotalOrgOwnerCount(shard int, org Id) (int, error) {
	if s.shards[shard] == nil {
		return invalidShardIdErr
	}
	var count int
	err := s.shards[shard].QueryRow(query_getTotalOrgOwnerCount, []byte(org)).Scan(&count)
	return count, err
}

var query_getOwnerCountInSet = `SELECT COUNT(*) WHERE org=? AND isActive=true AND role=0 AND id IN (`

func (s *sqlStore) getOwnerCountInSet(shard int, org Id, members []Id) (int, error) {
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
	err := s.shards[shard].QueryRow(query.String(), queryArgs...).Scan(&count)
	return count, err
}

func (s *sqlStore) setMembersInactive(shard int, org Id, members []Id) error {

}

func (s *sqlStore) memberIsOnlyOwner(shard int, org, member Id) (bool, error) {

}

func (s *sqlStore) setMemberDeleted(shard int, org Id, member Id) error {

}

func (s *sqlStore) renameMember(shard int, org Id, member Id, newName string) error {

}
