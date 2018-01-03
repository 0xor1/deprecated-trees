package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
	"time"
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

func (s *sqlStore) getAccountRole(shard int, accountId, memberId Id) *AccountRole {
	return GetAccountRole(s.shards[shard], accountId, memberId)
}

func (s *sqlStore) getAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	return GetAccountAndProjectRoles(s.shards[shard], accountId, projectId, memberId)
}

func (s *sqlStore) getProjectRole(shard int, accountId, projectId, memberId Id) *ProjectRole {
	return GetProjectRole(s.shards[shard], accountId, projectId, memberId)
}

func (s *sqlStore) getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	return GetAccountAndProjectRolesAndProjectIsPublic(s.shards[shard], accountId, projectId, memberId)
}

func (s *sqlStore) createNode(shard int, accountId, projectId, parentId, myId Id, nextSibling *Id, newNode *node) {
	args := make([]interface{}, 0, 18)
	args = append(args, []byte(accountId), []byte(projectId), []byte(parentId), []byte(myId))
	if nextSibling != nil {
		args = append(args, []byte(*nextSibling))
	} else {
		args = append(args, nil)
	}
	args = append(args, []byte(newNode.Id))
	args = append(args, newNode.IsAbstract)
	args = append(args, newNode.Name)
	args = append(args, newNode.Description)
	args = append(args, newNode.CreatedOn)
	args = append(args, newNode.TotalRemainingTime)
	args = append(args, newNode.IsParallel)
	if newNode.Member != nil {
		args = append(args, []byte(*newNode.Member))
	} else {
		args = append(args, nil)
	}
	MakeChangeHelper(s.shards[shard], `CALL createNode(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
}

func (s *sqlStore) setName(shard int, accountId, projectId, nodeId, myId Id, name string) {
	_, err := s.shards[shard].Exec(`CALL setNodeName(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), name)
	PanicIf(err)
}

func (s *sqlStore) setDescription(shard int, accountId, projectId, nodeId, myId Id, description *string) {
	_, err := s.shards[shard].Exec(`CALL setNodeDescription(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), description)
	PanicIf(err)
}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId, nodeId, myId Id, isParallel bool) {
	MakeChangeHelper(s.shards[shard], `CALL setNodeIsParallel(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), isParallel)
}

func (s *sqlStore) setMember(shard int, accountId, projectId, nodeId, myId Id, memberId *Id) {
	var memArg []byte
	if memberId != nil {
		memArg = []byte(*memberId)
	}
	MakeChangeHelper(s.shards[shard], `CALL setNodeMember(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), memArg)
}

func (s *sqlStore) setTimeRemainingAndOrLogTime(shard int, accountId, projectId, nodeId, myId Id, timeRemaining *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	MakeChangeHelper(s.shards[shard], `CALL setTimeRemainingAndOrLogTime(?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), timeRemaining, loggedOn, duration, note)
}

func (s *sqlStore) moveNode(shard int, accountId, projectId, nodeId, parentId Id, nextSibling *Id) {
	var nextSib []byte
	if nextSibling != nil {
		nextSib = []byte(*nextSibling)
	}
	MakeChangeHelper(s.shards[shard], `CALL moveNode(?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(parentId), nextSib)
}

func (s *sqlStore) deleteNode(shard int, accountId, projectId, nodeId Id) {
	MakeChangeHelper(s.shards[shard], `CALL deleteNode(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId))
}

func (s *sqlStore) getNode(shard int, accountId, projectId, nodeId Id) *node {
	row := s.shards[shard].QueryRow(`SELECT ... shit`, []byte(accountId), []byte(projectId), []byte(nodeId))
	n := node{}
	PanicIf(row.Scan(&n.Id, &n.IsAbstract, &n.Name, &n.Description, &n.CreatedOn, &n.TotalRemainingTime, &n.TotalLoggedTime, &n.MinimumRemainingTime, &n.LinkedFileCount, &n.ChatCount, &n.ChildCount, &n.DescendantCount, &n.IsParallel, &n.Member))
	nilOutPropertiesThatAreNotNilInTheDb(&n)
	return &n
}

func (s *sqlStore) getNodes(shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) []*node {
	var fromSib []byte
	if fromSibling != nil {
		fromSib = []byte(*fromSibling)
	}
	rows, err := s.shards[shard].Query(`CALL getNodes(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(parentId), fromSib, limit)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*node, 0, limit)
	for rows.Next() {
		n := node{}
		PanicIf(rows.Scan(&n.Id, &n.IsAbstract, &n.Name, &n.Description, &n.CreatedOn, &n.TotalRemainingTime, &n.TotalLoggedTime, &n.MinimumRemainingTime, &n.LinkedFileCount, &n.ChatCount, &n.ChildCount, &n.DescendantCount, &n.IsParallel, &n.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&n)
		res = append(res, &n)
	}
	return res
}

func nilOutPropertiesThatAreNotNilInTheDb(n *node) {
	if !n.IsAbstract {
		n.MinimumRemainingTime = nil
		n.ChildCount = nil
		n.DescendantCount = nil
		n.IsParallel = nil
	}
}
