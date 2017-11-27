package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

func newSqlStore(shards map[int]isql.ReplicaSet) store {
	if len(shards) == 0 {
		InvalidArgumentsErr.Panic()
	}
	return &sqlStore{
		shards: shards,
	}
}

type sqlStore struct {
	shards map[int]isql.ReplicaSet
}

func (s *sqlStore) getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {

}

func (s *sqlStore) getProjectRole(shard int, accountId, projectId, memberId Id) *ProjectRole {

}

func (s *sqlStore) getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {

}

func (s *sqlStore) getProjectExists(shard int, accountId, projectId Id) bool {

}

func (s *sqlStore) getNodeExists(shard int, accountId, projectId, nodeId Id) bool {

}

func (s *sqlStore) createNode(shard int, accountId, projectId, parentId Id, nextSibling *Id, newNode *node) {

}

func (s *sqlStore) setName(shard int, accountId, projectId, nodeId Id, name string) {

}

func (s *sqlStore) setDescription(shard int, accountId, projectId, nodeId Id, description string) {

}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId, nodeId Id, isParallel bool) {

}

func (s *sqlStore) setMember(shard int, accountId, projectId, nodeId, myId Id, memberId *Id) {

}

func (s *sqlStore) setTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, timeRemaining uint64) {

}

func (s *sqlStore) logTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) {

}

func (s *sqlStore) moveNode(shard int, accountId, projectId, nodeId, myId, parentId Id, nextSibling *Id) {

}

func (s *sqlStore) deleteNode(shard int, accountId, projectId, nodeId, myId Id) {

}

func (s *sqlStore) getNodes(shard int, accountId, projectId, parentId, myId Id, fromSibling *Id, limit int) []*node {

}

func (s *sqlStore) logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string) {

}
