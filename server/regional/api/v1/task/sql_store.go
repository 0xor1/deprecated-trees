package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
	"time"
	"bytes"
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

func (s *sqlStore) createTask(shard int, accountId, projectId, parentId, myId Id, nextSibling *Id, newTask *task) {
	args := make([]interface{}, 0, 18)
	args = append(args, []byte(accountId), []byte(projectId), []byte(parentId), []byte(myId))
	if nextSibling != nil {
		args = append(args, []byte(*nextSibling))
	} else {
		args = append(args, nil)
	}
	args = append(args, []byte(newTask.Id))
	args = append(args, newTask.IsAbstract)
	args = append(args, newTask.Name)
	args = append(args, newTask.Description)
	args = append(args, newTask.CreatedOn)
	args = append(args, newTask.TotalRemainingTime)
	args = append(args, newTask.IsParallel)
	if newTask.Member != nil {
		args = append(args, []byte(*newTask.Member))
	} else {
		args = append(args, nil)
	}
	MakeChangeHelper(s.shards[shard], `CALL createTask(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
}

func (s *sqlStore) setName(shard int, accountId, projectId, taskId, myId Id, name string) {
	_, err := s.shards[shard].Exec(`CALL setTaskName(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(myId), name)
	PanicIf(err)
}

func (s *sqlStore) setDescription(shard int, accountId, projectId, taskId, myId Id, description *string) {
	_, err := s.shards[shard].Exec(`CALL setTaskDescription(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(myId), description)
	PanicIf(err)
}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId, taskId, myId Id, isParallel bool) {
	MakeChangeHelper(s.shards[shard], `CALL setTaskIsParallel(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(myId), isParallel)
}

func (s *sqlStore) setMember(shard int, accountId, projectId, taskId, myId Id, memberId *Id) {
	var memArg []byte
	if memberId != nil {
		memArg = []byte(*memberId)
	}
	MakeChangeHelper(s.shards[shard], `CALL setTaskMember(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(myId), memArg)
}

func (s *sqlStore) setRemainingTimeAndOrLogTime(shard int, accountId, projectId, taskId, myId Id, timeRemaining *uint64, loggedOn *time.Time, duration *uint64, note *string) {
	MakeChangeHelper(s.shards[shard], `CALL setRemainingTimeAndOrLogTime(?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(myId), timeRemaining, loggedOn, duration, note)
}

func (s *sqlStore) moveTask(shard int, accountId, projectId, taskId, parentId, myId Id, newPreviousSibling *Id) {
	var prevSib []byte
	if newPreviousSibling != nil {
		prevSib = []byte(*newPreviousSibling)
	}
	MakeChangeHelper(s.shards[shard], `CALL moveTask(?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId), []byte(parentId), []byte(myId), prevSib)
}

func (s *sqlStore) deleteTask(shard int, accountId, projectId, taskId Id) {
	MakeChangeHelper(s.shards[shard], `CALL deleteTask(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(taskId))
}

func (s *sqlStore) getTasks(shard int, accountId, projectId Id, taskIds []Id) []*task {
	idsStr := bytes.NewBufferString(``)
	for _, id := range taskIds {
		idsStr.WriteString(id.String())
	}
	rows, err := s.shards[shard].Query(`CALL getTasks(?, ?, ?)`, []byte(accountId), []byte(projectId), idsStr.String())
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*task, 0, len(taskIds))
	for rows.Next() {
		t := task{}
		PanicIf(rows.Scan(&t.Id, &t.IsAbstract, &t.Name, &t.Description, &t.CreatedOn, &t.TotalRemainingTime, &t.TotalLoggedTime, &t.MinimumRemainingTime, &t.LinkedFileCount, &t.ChatCount, &t.ChildCount, &t.DescendantCount, &t.IsParallel, &t.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&t)
		res = append(res, &t)
	}
	return res
}

func (s *sqlStore) getChildTasks(shard int, accountId, projectId, parentId Id, fromSibling *Id, limit int) []*task {
	var fromSib []byte
	if fromSibling != nil {
		fromSib = []byte(*fromSibling)
	}
	rows, err := s.shards[shard].Query(`CALL getChildTasks(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(parentId), fromSib, limit)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*task, 0, limit)
	for rows.Next() {
		t := task{}
		PanicIf(rows.Scan(&t.Id, &t.IsAbstract, &t.Name, &t.Description, &t.CreatedOn, &t.TotalRemainingTime, &t.TotalLoggedTime, &t.MinimumRemainingTime, &t.LinkedFileCount, &t.ChatCount, &t.ChildCount, &t.DescendantCount, &t.IsParallel, &t.Member))
		nilOutPropertiesThatAreNotNilInTheDb(&t)
		res = append(res, &t)
	}
	return res
}

func nilOutPropertiesThatAreNotNilInTheDb(t *task) {
	if !t.IsAbstract {
		t.MinimumRemainingTime = nil
		t.ChildCount = nil
		t.DescendantCount = nil
		t.IsParallel = nil
	}
}
