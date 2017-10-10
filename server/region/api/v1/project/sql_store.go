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
		panic(InvalidArgumentsErr)
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

func (s *sqlStore) getPublicProjectsEnabled(shard int, accountId Id) bool {
	return GetPublicProjectsEnabled(s.shards[shard], accountId)
}

func (s *sqlStore) createProject(shard int, accountId Id, project *project) {
	_, err := s.shards[shard].Exec(`CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(project.Id), project.Name, project.Description, project.CreatedOn, project.ArchivedOn, project.StartOn, project.DueOn, project.TotalRemainingTime, project.TotalLoggedTime, project.MinimumRemainingTime, project.FileCount, project.FileSize, project.LinkedFileCount, project.ChatCount, project.IsParallel, project.IsPublic)
	PanicIf(err)
}

func (s *sqlStore) setName(shard int, accountId, projectId Id, name string) {
	_, err := s.shards[shard].Exec(`UPDATE projects SET name=? WHERE account=? AND id=?`, name, []byte(accountId), []byte(projectId))
	PanicIf(err)
}

func (s *sqlStore) setDescription(shard int, accountId, projectId Id, description string) {
	_, err := s.shards[shard].Exec(`UPDATE projects SET description=? WHERE account=? AND id=?`, description, []byte(accountId), []byte(projectId))
	PanicIf(err)
}

func (s *sqlStore) setIsPublic(shard int, accountId, projectId Id, isPublic bool) {
	_, err := s.shards[shard].Exec(`UPDATE projects SET isPublic=? WHERE account=? AND id=?`, isPublic, []byte(accountId), []byte(projectId))
	PanicIf(err)
}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId Id, isParallel bool) {
	_, err := s.shards[shard].Exec(`CALL setProjectIsParallel(?, ?, ?)`, []byte(accountId), []byte(projectId), isParallel)
	PanicIf(err)
}

func (s *sqlStore) getProject(shard int, accountId, projectId Id) *project {
	row := s.shards[shard].QueryRow(`SELECT id, name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic FROM projects WHERE account=? AND id=?`, []byte(accountId), []byte(projectId))
	result := project{}
	PanicIf(row.Scan(&result.Id, &result.Name, &result.Description, &result.CreatedOn, &result.ArchivedOn, &result.StartOn, &result.DueOn, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.FileCount, &result.FileSize, &result.LinkedFileCount, &result.ChatCount, &result.IsParallel, &result.IsPublic))
	return &result
}

func (s *sqlStore) getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], `AND isPublic=true`, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, accountId, &myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], ``, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) setProjectArchivedOn(shard int, accountId, projectId Id, now *time.Time) {
	_, err := s.shards[shard].Exec(`UPDATE projects SET archivedOn=? WHERE account=? && id=?`, Now(), []byte(accountId), []byte(projectId))
	PanicIf(err)
}

func (s *sqlStore) deleteProject(shard int, accountId, projectId Id) {
	_, err := s.shards[shard].Exec(`CALL deleteProject(?, ?)`, []byte(accountId), []byte(projectId))
	PanicIf(err)
}

func (s *sqlStore) addMemberOrSetActive(shard int, accountId, projectId Id, member *addMember) bool {
	row := s.shards[shard].QueryRow(`CALL addProjectMemberOrSetActive(?, ?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(member.Id), member.Role)
	added := false
	PanicIf(row.Scan(&added))
	return added
}

func (s *sqlStore) setMemberRole(shard int, accountId, projectId Id, member Id, role ProjectRole) {
	_, err := s.shards[shard].Exec(`UPDATE projectMembers SET role=? WHERE account=? AND project=? AND id=? AND isActive=true`, role, []byte(accountId), []byte(projectId), []byte(member))
	PanicIf(err)
}

func (s *sqlStore) setMemberInactive(shard int, accountId, projectId Id, member Id) bool {
	row := s.shards[shard].QueryRow(`CALL setProjectMemberInactive(?, ?, ?)`, []byte(accountId), []byte(projectId), []byte(member))
	setInactive := false
	PanicIf(row.Scan(&setInactive))
	return setInactive
}

func (s *sqlStore) getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*member, int) {
	query := bytes.NewBufferString(`SELECT %s FROM projectMembers WHERE account=? AND project=? AND isActive=true`)
	columns := ` id, name, isActive, totalRemainingTime, totalLoggedTime, role `
	args := make([]interface{}, 0, 6)
	args = append(args, []byte(accountId), []byte(projectId))
	if role != nil {
		query.WriteString(` AND role=?`)
		args = append(args, role)
	}
	if nameContains != nil {
		query.WriteString(` AND name LIKE ?`)
		strVal := strings.Trim(*nameContains, " ")
		args = append(args, fmt.Sprintf("%%%s%%", strVal))
	}
	count := 0
	row := s.shards[shard].QueryRow(fmt.Sprintf(query.String(), ` COUNT(*) `), args...)
	PanicIf(row.Scan(&count))
	if count == 0 {
		return nil, count
	}
	query.WriteString(` ORDER BY role ASC, name ASC LIMIT ? OFFSET ?`)
	args = append(args, limit, offset)
	rows, err := s.shards[shard].Query(fmt.Sprintf(query.String(), columns), args...)
	if rows != nil {
		defer rows.Close()
	}
	PanicIf(err)
	res := make([]*member, 0, limit)
	for rows.Next() {
		mem := member{}
		PanicIf(rows.Scan(&mem.Id, &mem.Name, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		res = append(res, &mem)
	}
	return res, count
}

func (s *sqlStore) getMember(shard int, accountId, projectId, memberId Id) *member {
	row := s.shards[shard].QueryRow(`SELECT id, name, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, []byte(accountId), []byte(projectId), []byte(memberId))
	res := member{}
	PanicIf(row.Scan(&res.Id, &res.Name, &res.IsActive, &res.Role))
	return &res
}

func (s *sqlStore) logAccountActivity(shard int, accountId, member, item Id, itemType, action string, newValue *string) {
	LogAccountActivity(s.shards[shard], accountId, member, item, itemType, action, newValue)
}

func (s *sqlStore) logProjectActivity(shard int, accountId, projectId, member, item Id, itemType, action string, newValue *string) {
	LogProjectActivity(s.shards[shard], accountId, projectId, member, item, itemType, action, newValue)
}

func (s *sqlStore) logProjectBatchAddOrRemoveMembersActivity(shard int, accountId, projectId, member Id, members []Id, action string) {
	LogProjectBatchAddOrRemoveMembersActivity(s.shards[shard], accountId, projectId, member, members, action)
}

func (s *sqlStore) getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfterUnixMillis, occurredBeforeUnixMillis *uint64, limit int) []*Activity {
	if occurredAfterUnixMillis != nil && occurredBeforeUnixMillis != nil {
		panic(InvalidArgumentsErr)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, itemName, action, newValue FROM projectActivities WHERE account=? AND project=?`)
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
		unixMilli := int64(0)
		PanicIf(rows.Scan(&unixMilli, &act.Item, &act.Member, &act.ItemType, &act.ItemName, &act.Action, &act.NewValue))
		act.OccurredOn = time.Unix(unixMilli/1000, (unixMilli%1000)*1000000).UTC()
		res = append(res, &act)
	}
	return res
}

func getProjects(shard isql.ReplicaSet, specificSqlFilterTxt string, accountId Id, myId *Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	columns := ` id, name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic `
	query := bytes.NewBufferString(`SELECT %s FROM projects WHERE account=? %s`)
	args := make([]interface{}, 0, 13)
	args = append(args, []byte(accountId))
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
	if archived {
		query.WriteString(` AND archivedOn IS NOT NULL`)
	} else {
		query.WriteString(` AND archivedOn IS NULL`)
	}
	row := shard.QueryRow(fmt.Sprintf(query.String(), ` COUNT(*) `, specificSqlFilterTxt), args...)
	count := 0
	PanicIf(row.Scan(&count))
	if count == 0 {
		return nil, count
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s LIMIT ? OFFSET ?`, sortBy, sortDir))
	args = append(args, limit, offset)
	rows, err := shard.Query(fmt.Sprintf(query.String(), columns, specificSqlFilterTxt), args...)
	PanicIf(err)
	result := make([]*project, 0, limit)
	for rows.Next() {
		res := project{}
		PanicIf(rows.Scan(&res.Id, &res.Name, &res.Description, &res.CreatedOn, &res.ArchivedOn, &res.StartOn, &res.DueOn, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.MinimumRemainingTime, &res.FileCount, &res.FileSize, &res.LinkedFileCount, &res.ChatCount, &res.IsParallel, &res.IsPublic))
		result = append(result, &res)
	}
	return result, count
}
