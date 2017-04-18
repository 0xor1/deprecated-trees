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

var query_createAbstractTask = `CALL createAbstractTask(?, ?, ?, ?)`

func (s *sqlStore) createAbstractTask(at *abstractTask) (int, error) {
	shardId := rand.Intn(len(s.shards))
	_, err := s.shards[shardId].Exec(query_createAbstractTask, at.Org, nil, at.Org, at.Id, at.Name, at.User, at.TotalRemainingTime, at.TotalLoggedTime, at.ChatCount, at.FileCount, at.FileSize, at.CreatedOn, true, at.MinimumRemainingTime, at.IsParallel, at.ChildCount, at.DescendantCount, at.LeafCount, at.SubFileCount, at.SubFileSize, at.ArchivedChildCount, at.ArchivedDescendantCount, at.ArchivedLeafCount, at.ArchivedSubFileCount, at.ArchivedSubFileSize)
	return shardId, err
}

func (s *sqlStore) createMember(shard int, org Id, member *member) error {

}

func (s *sqlStore) deleteAccount(shard int, account Id) error {

}

func (s *sqlStore) getMember(shard int, org, member Id) (*member, error) {

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
