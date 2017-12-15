package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
)

var (
	noChangeMadeErr = &AppError{Code: "r_v1_p_nc", Message: "no change made", Public: true}
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

func (s *sqlStore) createNode(shard int, accountId, projectId, parentId Id, nextSibling *Id, newNode *node) {
	args := make([]interface{}, 0, 18)
	args = append(args, []byte(accountId), []byte(projectId), []byte(parentId))
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
	args = append(args, newNode.TotalLoggedTime)
	args = append(args, newNode.MinimumRemainingTime)
	args = append(args, newNode.LinkedFileCount)
	args = append(args, newNode.ChatCount)
	args = append(args, newNode.ChildCount)
	args = append(args, newNode.DescendantCount)
	args = append(args, newNode.IsParallel)
	if newNode.Member != nil {
		args = append(args, []byte(*newNode.Member))
	} else {
		args = append(args, nil)
	}
	s.makeChangeHelper(shard, `CALL createNode(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
}

func (s *sqlStore) setName(shard int, accountId, projectId, nodeId Id, name string) {
	_, err := s.shards[shard].Exec(`CALL setNodeName(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), name)
	PanicIf(err)
}

func (s *sqlStore) setDescription(shard int, accountId, projectId, nodeId Id, description *string) {
	_, err := s.shards[shard].Exec(`UPDATE nodes SET description=? WHERE account=? AND project=? AND id=?`, description, []byte(accountId), []byte(projectId), []byte(nodeId))
	PanicIf(err)
}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId, nodeId Id, isParallel bool) {
	s.makeChangeHelper(shard, `CALL setNodeIsParallel(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), isParallel)
}

func (s *sqlStore) setMember(shard int, accountId, projectId, nodeId Id, memberId *Id) {
	var memArg []byte
	if memberId != nil {
		memArg = []byte(*memberId)
	}
	s.makeChangeHelper(shard, `CALL setNodeMember(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), memArg)
}

func (s *sqlStore) setTimeRemaining(shard int, accountId, projectId, nodeId Id, timeRemaining uint64) {
	s.makeChangeHelper(shard, `CALL setTimeRemaining(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), timeRemaining)
}

func (s *sqlStore) logTimeAndSetTimeRemaining(shard int, accountId, projectId, nodeId, myId Id, duration uint64, timeRemaining uint64, note *string) {
	s.makeChangeHelper(shard, `CALL logTimeAndSetTimeRemaining(?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(myId), timeRemaining, note)
}

func (s *sqlStore) moveNode(shard int, accountId, projectId, nodeId, parentId Id, nextSibling *Id) {
	var nextSib []byte
	if nextSibling != nil {
		nextSib = []byte(*nextSibling)
	}
	s.makeChangeHelper(shard, `CALL moveNode(?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId), []byte(parentId), nextSib)
}

func (s *sqlStore) deleteNode(shard int, accountId, projectId, nodeId Id) {
	s.makeChangeHelper(shard, `CALL deleteNode(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(nodeId))
}

func (s *sqlStore) getNode(shard int, accountId, projectId, nodeId Id) *node {
	row := s.shards[shard].QueryRow(`SELECT ... shit`, []byte(accountId), []byte(projectId), []byte(nodeId))
	n := node{}
	PanicIf(row.Scan(&n.Id, &n.IsAbstract, &n.Name, &n.Description, &n.CreatedOn, &n.TotalRemainingTime, &n.TotalLoggedTime, &n.MinimumRemainingTime, &n.LinkedFileCount, &n.ChatCount, &n.ChildCount, &n.DescendantCount, &n.IsParallel, &n.Member))
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
		res = append(res, &n)
	}
	return res
}

func (s *sqlStore) logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string) {
	LogProjectActivity(s.shards[shard], accountId, projectId, member, item, itemType, action, newValue)
}

func (s *sqlStore) makeChangeHelper(shard int, sql string, args ...interface{}) {
	row := s.shards[shard].QueryRow(sql, args...)
	changeMade := false
	PanicIf(row.Scan(&changeMade))
	if !changeMade {
		noChangeMadeErr.Panic()
	}
}
