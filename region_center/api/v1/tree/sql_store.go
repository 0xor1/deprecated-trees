package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/robsix/isql"
	"math/rand"
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

func (s *sqlStore) getMember(shard int, org, member Id) (*orgMember, error) {

}

func (s *sqlStore) addMembers(shard int, org Id, members []*NamedEntity) error {

}

func (s *sqlStore) setMembersActive(shard int, org Id, members []*NamedEntity) error {

}

func (s *sqlStore) getTotalOrgOwnerCount(shard int, org Id) (int, error) {

}

func (s *sqlStore) getOwnerCountInSet(shard int, org Id, members []Id) (int, error) {

}

func (s *sqlStore) setMembersInactive(shard int, org Id, members []Id) error {

}

func (s *sqlStore) memberIsOnlyOwner(shard int, org, member Id) (bool, error) {

}

func (s *sqlStore) setMemberDeleted(shard int, org Id, member Id) error {

}

func (s *sqlStore) renameMember(shard int, org Id, member Id, newName string) error {

}
