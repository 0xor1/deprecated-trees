package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	"time"
	"bytes"
	"fmt"
	"strings"
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
	s.shards[shard].Exec(`CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(project.Id), project.Name, project.Description, project.CreatedOn, project.ArchivedOn, project.StartOn, project.DueOn, project.TotalRemainingTime, project.TotalLoggedTime, project.MinimumRemainingTime, project.FileCount, project.FileSize, project.LinkedFileCount, project.ChatCount, project.IsParallel, project.IsPublic)
}

func (s *sqlStore) setName(shard int, accountId, projectId Id, name string) {
	s.shards[shard].Exec(`UPDATE projects SET name=? WHERE account=? AND id=?`, name, []byte(accountId), []byte(projectId))
}

func (s *sqlStore) setDescription(shard int, accountId, projectId Id, description string) {
	s.shards[shard].Exec(`UPDATE projects SET description=? WHERE account=? AND id=?`, description, []byte(accountId), []byte(projectId))
}

func (s *sqlStore) setIsPublic(shard int, accountId, projectId Id, isPublic bool) {
	s.shards[shard].Exec(`UPDATE projects SET isPublic=? WHERE account=? AND id=?`, isPublic, []byte(accountId), []byte(projectId))
}

func (s *sqlStore) setIsParallel(shard int, accountId, projectId Id, isParallel bool) {
	s.shards[shard].Exec(`CALL setProjectIsParallel(?, ?, ?)`, []byte(accountId), []byte(projectId), isParallel)
}

func (s *sqlStore) getProject(shard int, accountId, projectId Id) *project {
	row := s.shards[shard].QueryRow(`SELECT name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic FROM projects WHERE account=? AND id=?`, []byte(accountId), []byte(projectId))
	result := project{}
	if err := row.Scan(&result.Name, &result.Description, &result.CreatedOn, &result.ArchivedOn, &result.StartOn, &result.DueOn, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.FileCount, &result.FileSize, &result.LinkedFileCount, &result.ChatCount, &result.IsParallel, &result.IsPublic); err != nil {
		panic(err)
	}
	return &result
}

func (s *sqlStore) getPublicProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], `AND isPublic=true`, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) getPublicAndSpecificAccessProjects(shard int, accountId, myId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND id=?))`, accountId, &myId, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) getAllProjects(shard int, accountId Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	return getProjects(s.shards[shard], ``, accountId, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, archived, sortBy, sortDir, offset, limit)
}

func (s *sqlStore) setProjectArchivedOn(shard int, accountId, projectId Id, now *time.Time) {
	s.shards[shard].Exec(`UPDATE projects SET archivedOn=? WHERE account=? && project=?`, []byte(accountId), []byte(projectId))
}

func (s *sqlStore) deleteProject(shard int, accountId, projectId Id) {
	s.shards[shard].Exec(`CALL deleteProject(?, ?)`, []byte(accountId), []byte(projectId))
}

func (s *sqlStore) addMembers(shard int, accountId, projectId Id, members []*addProjectMember) {

}

func (s *sqlStore) removeMembers(shard int, accountId, projectId Id, members []Id) {

}

func (s *sqlStore) getMembers(shard int, accountId, projectId Id, role *ProjectRole, nameContains *string, offset, limit int) ([]*projectMember, int) {

}

func (s *sqlStore) getMember(shard int, accountId, projectId, members Id) *projectMember {

}

func (s *sqlStore) logAccountActivity(shard int, accountId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string) {

}

func (s *sqlStore) logProjectActivity(shard int, accountId, projectId Id, occurredOn time.Time, member, item Id, itemType, action string, newValue *string) {

}

func (s *sqlStore) logProjectBatchAddOrRemoveMembersActivity(shard int, accountId, projectId, member Id, members []Id, action string) {

}

func (s *sqlStore) getActivities(shard int, accountId, projectId Id, item, member *Id, occurredAfter, occurredBefore *time.Time, limit int) []*Activity {

}

func getProjects(shard isql.ReplicaSet, specificSqlFilterTxt string, accountId Id, myId *Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, archived bool, sortBy SortBy, sortDir SortDir, offset, limit int) ([]*project, int) {
	columns := ` name, description, createdOn, archivedOn, startOn, dueOn, totalRemainingTime, totalLoggedTime, minimumRemainingTime, fileCount, fileSize, linkedFileCount, chatCount, isParallel, isPublic `
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
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	if count == 0 {
		return nil, count
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s LIMIT ? OFFSET ?`, sortBy, sortDir))
	args = append(args, sortBy, sortDir, limit, offset)
	rows, err := shard.Query(fmt.Sprintf(query.String(), columns, specificSqlFilterTxt), args...)
	if err != nil {
		panic(err)
	}
	result := make([]*project, 0, limit)
	for rows.Next() {
		res := project{}
		if err := rows.Scan(&res.Name, &res.Description, &res.CreatedOn, &res.ArchivedOn, &res.StartOn, &res.DueOn, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.MinimumRemainingTime, &res.FileCount, &res.FileSize, &res.LinkedFileCount, &res.ChatCount, &res.IsParallel, &res.IsPublic); err != nil {
			panic(err)
		}
		result = append(result, &res)
	}
	return result, count
}

