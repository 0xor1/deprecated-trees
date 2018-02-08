package project

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bytes"
	"fmt"
	"github.com/0xor1/isql"
	"strings"
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

func (s *sqlStore) getAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	return GetAccountAndProjectRolesAndProjectIsPublic(s.shards[shard], accountId, projectId, memberId)
}

func (s *sqlStore) getProjectExists(shard int, accountId, projectId Id) bool {
	return GetProjectExists(s.shards[shard], accountId, projectId)
}

func (s *sqlStore) getPublicProjectsEnabled(shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(s.shards[shard], accountId)
}

func (s *sqlStore) createProject(shard int, accountId, myId Id, project *project) {
	_, err := s.shards[shard].Exec(`CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(project.Id), []byte(myId), project.Name, project.Description, project.CreatedOn, project.StartOn, project.DueOn, project.IsParallel, project.IsPublic)
	PanicIf(err)
}

func (s *sqlStore) setIsPublic(shard int, accountId, projectId, myId Id, isPublic bool) {
	_, err := s.shards[shard].Exec(`CALL setProjectIsPublic(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), isPublic)
	PanicIf(err)
}

func (s *sqlStore) getProject(shard int, accountId, projectId Id) *project {
	row := s.shards[shard].QueryRow(`SELECT p.id, p.isArchived, p.name, p.createdOn, p.startOn, p.dueOn, p.fileCount, p.fileSize, p.isPublic, t.description, t.totalRemainingTime, t.totalLoggedTime, t.minimumRemainingTime, t.linkedFileCount, t.chatCount, t.childCount, t.descendantCount, t.isParallel FROM projects p, tasks t WHERE p.account=? AND p.id=? AND t.account=? AND t.project=? AND t.id=?`, []byte(accountId), []byte(projectId), []byte(accountId), []byte(projectId), []byte(projectId))
	result := project{}
	PanicIf(row.Scan(&result.Id, &result.IsArchived, &result.Name, &result.CreatedOn, &result.StartOn, &result.DueOn, &result.FileCount, &result.FileSize, &result.IsPublic, &result.Description, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.LinkedFileCount, &result.ChatCount, &result.ChildCount, &result.DescendantCount, &result.IsParallel))
	return &result
}

func (s *sqlStore) getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	return getProjects(s.shards[shard], `AND isPublic=true`, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func (s *sqlStore) getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	return getProjects(s.shards[shard], `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, accountId, &myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func (s *sqlStore) getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	return getProjects(s.shards[shard], ``, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func (s *sqlStore) setProjectIsArchived(shard int, accountId, projectId, myId Id, isArchived bool) {
	_, err := s.shards[shard].Exec(`CALL setProjectIsArchived(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), isArchived)
	PanicIf(err)
}

func (s *sqlStore) deleteProject(shard int, accountId, projectId, myId Id) {
	_, err := s.shards[shard].Exec(`CALL deleteProject(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId))
	PanicIf(err)
}

func (s *sqlStore) addMemberOrSetActive(shard int, accountId, projectId, myId Id, member *AddProjectMember) {
	MakeChangeHelper(s.shards[shard], `CALL addProjectMemberOrSetActive(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member.Id), member.Role)
}

func (s *sqlStore) setMemberRole(shard int, accountId, projectId, myId, member Id, role ProjectRole) {
	MakeChangeHelper(s.shards[shard], `CALL setProjectMemberRole(?, ?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member), role)
}

func (s *sqlStore) setMemberInactive(shard int, accountId, projectId, myId Id, member Id) {
	MakeChangeHelper(s.shards[shard],`CALL setProjectMemberInactive(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(myId), []byte(member))
}

func (s *sqlStore) getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameOrDisplayNameContains *string, after *Id, limit int) ([]*member, bool) {
	query := bytes.NewBufferString(`SELECT p1.id, p1.isActive, p1.totalRemainingTime, p1.totalLoggedTime, p1.role FROM projectMembers p1`)
	args := make([]interface{}, 0, 9)
	if after != nil {
		query.WriteString(`, projectMembers p2`)
	}
	query.WriteString(` WHERE p1.account=? AND p1.project=? AND p1.isActive=true`)
	args = append(args, []byte(accountId), []byte(projectId))
	if after != nil {
		query.WriteString(` AND p2.account=? AND p2.project=? p2.id=? AND ((p1.name>p2.name AND p1.role=p2.role) OR p1.role>p2.role)`)
		args = append(args, []byte(accountId), []byte(projectId), []byte(*after))
	}
	if role != nil {
		query.WriteString(` AND p1.role=?`)
		args = append(args, role)
	}
	if nameOrDisplayNameContains != nil {
		query.WriteString(` AND (p1.name LIKE ? OR p1.displayName LIKE ?)`)
		strVal := strings.Trim(*nameOrDisplayNameContains, " ")
		strVal = fmt.Sprintf("%%%s%%", strVal)
		args = append(args, strVal, strVal)
	}
	query.WriteString(` ORDER BY p1.role ASC, p1.name ASC LIMIT ?`)
	args = append(args, limit+1)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}

func (s *sqlStore) getMember(shard int, accountId, projectId, memberId Id) *member {
	row := s.shards[shard].QueryRow(`SELECT id, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, []byte(accountId), []byte(projectId), []byte(memberId))
	res := member{}
	PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
	return &res
}

func (s *sqlStore) getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		InvalidArgumentsErr.Panic()
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, newValue FROM projectActivities WHERE account=? AND project=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, []byte(accountId), []byte(projectId))
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, []byte(*item))
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, []byte(*member))
	}
	if occurredAfterUnixMillis != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfterUnixMillis)
	}
	if occurredBeforeUnixMillis != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBeforeUnixMillis)
	}
	if occurredAfterUnixMillis == nil && occurredBeforeUnixMillis == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, err := s.shards[shard].Query(query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)

	res := make([]*Activity, 0, limit)
	for rows.Next() {
		act := Activity{}
		PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.NewValue))
		res = append(res, &act)
	}
	return res
}

func getProjects(shard isql.ReplicaSet, specificSqlFilterTxt string, accountId Id, myId *Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy SortBy, sortDir SortDir, after *Id, limit int) ([]*project, bool) {
	query := bytes.NewBufferString(`SELECT id, isArchived, name, createdOn, startOn, dueOn, fileCount, fileSize, isPublic FROM projects WHERE account=? AND isArchived=? %s`)
	args := make([]interface{}, 0, 14)
	args = append(args, []byte(accountId), isArchived)
	if myId != nil {
		args = append(args, []byte(accountId), []byte(*myId))
	}
	if nameContains != nil {
		query.WriteString(` AND name LIKE ?`)
		args = append(args, fmt.Sprintf(`%%%s%%`, strings.Trim(*nameContains, " ")))
	}
	if createdOnAfter != nil {
		query.WriteString(` AND createdOn>?`)
		args = append(args, createdOnAfter)
	}
	if createdOnBefore != nil {
		query.WriteString(` AND createdOn<?`)
		args = append(args, createdOnBefore)
	}
	if startOnAfter != nil {
		query.WriteString(` AND startOn>?`)
		args = append(args, startOnAfter)
	}
	if startOnBefore != nil {
		query.WriteString(` AND startOn<?`)
		args = append(args, startOnBefore)
	}
	if dueOnAfter != nil {
		query.WriteString(` AND dueOn>?`)
		args = append(args, dueOnAfter)
	}
	if dueOnBefore != nil {
		query.WriteString(` AND dueOn<?`)
		args = append(args, dueOnBefore)
	}
	if after != nil {
		query.WriteString(fmt.Sprintf(` AND %s %s= (SELECT %s FROM projects WHERE account=? AND id=?) AND id > ?`, sortBy, sortDir.GtLtSymbol(), sortBy))
		args = append(args, []byte(accountId), []byte(*after), []byte(*after))
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s, id LIMIT ?`, sortBy, sortDir))
	args = append(args, limit+1)
	rows, err := shard.Query(fmt.Sprintf(query.String(), specificSqlFilterTxt), args...)
	PanicIf(err)
	res := make([]*project, 0, limit+1)
	idx := 0
	resIdx := map[string]int{}
	for rows.Next() {
		proj := project{}
		PanicIf(rows.Scan(&proj.Id, &proj.IsArchived, &proj.Name, &proj.CreatedOn, &proj.StartOn, &proj.DueOn, &proj.FileCount, &proj.FileSize, &proj.IsPublic))
		res = append(res, &proj)
		resIdx[proj.Id.String()] = idx
		idx++
	}
	if len(res) > 0 { //populate task properties
		var id Id
		var description *string
		var totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount uint64
		var isParallel bool
		query.Reset()
		args = make([]interface{}, 0, len(res)+1)
		args = append(args, []byte(accountId), []byte(res[0].Id))
		query.WriteString(`SELECT id, description, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel FROM tasks WHERE account=? AND project=id AND project IN (?`)
		for _, proj := range res[1:] {
			query.WriteString(`,?`)
			args = append(args, []byte(proj.Id))
		}
		query.WriteString(fmt.Sprintf(`) LIMIT %d`, len(res)))
		rows, err := shard.Query(query.String(), args...)
		PanicIf(err)
		for rows.Next() {
			rows.Scan(&id, &description, &totalRemainingTime, &totalLoggedTime, &minimumRemainingTime, &linkedFileCount, &chatCount, &childCount, &descendantCount, &isParallel)
			proj := res[resIdx[id.String()]]
			proj.Description = description
			proj.TotalRemainingTime = totalRemainingTime
			proj.TotalLoggedTime = totalLoggedTime
			proj.MinimumRemainingTime = minimumRemainingTime
			proj.LinkedFileCount = linkedFileCount
			proj.ChatCount = chatCount
			proj.ChildCount = childCount
			proj.DescendantCount = descendantCount
			proj.IsParallel = isParallel
		}
	}
	if len(res) == limit+1 {
		return res[:limit], true
	}
	return res, false
}
