package project

import (
	"bitbucket.org/0xor1/task/server/util/activity"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/ctx"
	"bitbucket.org/0xor1/task/server/util/db"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bytes"
	"fmt"
	"strings"
	"time"
)

func dbGetProjectExists(ctx ctx.Ctx, shard int, account, project id.Id) bool {
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, account, project)
	exists := false
	err.PanicIf(row.Scan(&exists))
	return exists
}

func dbCreateProject(ctx ctx.Ctx, shard int, account id.Id, project *project) {
	_, e := ctx.TreeExec(shard, `CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account, project.Id, ctx.Me(), project.Name, project.Description, project.CreatedOn, project.StartOn, project.DueOn, project.IsParallel, project.IsPublic)
	err.PanicIf(e)
}

func dbSetIsPublic(ctx ctx.Ctx, shard int, account, project id.Id, isPublic bool) {
	_, e := ctx.TreeExec(shard, `CALL setProjectIsPublic(?, ?, ?, ?)`, account, project, ctx.Me(), isPublic)
	err.PanicIf(e)
}

func dbGetProject(ctx ctx.Ctx, shard int, account, proj id.Id) *project {
	row := ctx.TreeQueryRow(shard, `SELECT p.id, p.isArchived, p.name, p.createdOn, p.startOn, p.dueOn, p.fileCount, p.fileSize, p.isPublic, t.description, t.totalRemainingTime, t.totalLoggedTime, t.minimumRemainingTime, t.linkedFileCount, t.chatCount, t.childCount, t.descendantCount, t.isParallel FROM projects p, tasks t WHERE p.account=? AND p.id=? AND t.account=? AND t.project=? AND t.id=?`, account, proj, account, proj, proj)
	result := project{}
	err.PanicIf(row.Scan(&result.Id, &result.IsArchived, &result.Name, &result.CreatedOn, &result.StartOn, &result.DueOn, &result.FileCount, &result.FileSize, &result.IsPublic, &result.Description, &result.TotalRemainingTime, &result.TotalLoggedTime, &result.MinimumRemainingTime, &result.LinkedFileCount, &result.ChatCount, &result.ChildCount, &result.DescendantCount, &result.IsParallel))
	return &result
}

func dbGetPublicProjects(ctx ctx.Ctx, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND isPublic=true`, account, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetPublicAndSpecificAccessProjects(ctx ctx.Ctx, shard int, account, me id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, account, &me, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbGetAllProjects(ctx ctx.Ctx, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	return dbGetProjects(ctx, shard, ``, account, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit)
}

func dbSetProjectIsArchived(ctx ctx.Ctx, shard int, account, project id.Id, isArchived bool) {
	_, e := ctx.TreeExec(shard, `CALL setProjectIsArchived(?, ?, ?, ?)`, account, project, ctx.Me(), isArchived)
	err.PanicIf(e)
}

func dbDeleteProject(ctx ctx.Ctx, shard int, account, project id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteProject(?, ?, ?)`, account, project, ctx.Me())
	err.PanicIf(e)
}

func dbAddMemberOrSetActive(ctx ctx.Ctx, shard int, account, project id.Id, member *AddProjectMember) {
	db.MakeChangeHelper(ctx, shard, `CALL addProjectMemberOrSetActive(?, ?, ?, ?, ?)`, account, project, ctx.Me(), member.Id, member.Role)
}

func dbSetMemberRole(ctx ctx.Ctx, shard int, account, project, member id.Id, role cnst.ProjectRole) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberRole(?, ?, ?, ?, ?)`, account, project, ctx.Me(), member, role)
}

func dbSetMemberInactive(ctx ctx.Ctx, shard int, account, project id.Id, member id.Id) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberInactive(?, ?, ?, ?)`, account, project, ctx.Me(), member)
}

func dbGetMembers(ctx ctx.Ctx, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameContains *string, after *id.Id, limit int) *getMembersResp {
	query := bytes.NewBufferString(`SELECT p1.id, p1.isActive, p1.totalRemainingTime, p1.totalLoggedTime, p1.role FROM projectMembers p1`)
	args := make([]interface{}, 0, 9)
	if after != nil {
		query.WriteString(`, projectMembers p2`)
	}
	query.WriteString(` WHERE p1.account=? AND p1.project=? AND p1.isActive=true`)
	args = append(args, account, project)
	if after != nil {
		query.WriteString(` AND p2.account=? AND p2.project=? p2.id=? AND ((p1.name>p2.name AND p1.role=p2.role) OR p1.role>p2.role)`)
		args = append(args, account, project, *after)
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
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	res := make([]*member, 0, limit+1)
	for rows.Next() {
		mem := member{}
		err.PanicIf(rows.Scan(&mem.Id, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		res = append(res, &mem)
	}
	if len(res) == limit+1 {
		return &getMembersResp{Members: res[:limit], More: true}
	}
	return &getMembersResp{Members: res, More: false}
}

func dbGetMember(ctx ctx.Ctx, shard int, account, project, mem id.Id) *member {
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, account, project, mem)
	res := member{}
	err.PanicIf(row.Scan(&res.Id, &res.IsActive, &res.Role))
	return &res
}

func dbGetActivities(ctx ctx.Ctx, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) []*activity.Activity {
	if occurredAfter != nil && occurredBefore != nil {
		panic(err.InvalidArguments)
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, action, itemName, extraInfo FROM projectActivities WHERE account=? AND project=?`)
	args := make([]interface{}, 0, limit)
	args = append(args, account, project)
	if item != nil {
		query.WriteString(` AND item=?`)
		args = append(args, *item)
	}
	if member != nil {
		query.WriteString(` AND member=?`)
		args = append(args, *member)
	}
	if occurredAfter != nil {
		query.WriteString(` AND occurredOn>? ORDER BY occurredOn ASC`)
		args = append(args, occurredAfter)
	}
	if occurredBefore != nil {
		query.WriteString(` AND occurredOn<? ORDER BY occurredOn DESC`)
		args = append(args, occurredBefore)
	}
	if occurredAfter == nil && occurredBefore == nil {
		query.WriteString(` ORDER BY occurredOn DESC`)
	}
	query.WriteString(` LIMIT ?`)
	args = append(args, limit)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)

	res := make([]*activity.Activity, 0, limit)
	for rows.Next() {
		act := activity.Activity{}
		err.PanicIf(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	return res
}

func dbGetProjects(ctx ctx.Ctx, shard int, specificSqlFilterTxt string, account id.Id, me *id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortDir cnst.SortDir, after *id.Id, limit int) *getProjectsResp {
	query := bytes.NewBufferString(`SELECT id, isArchived, name, createdOn, startOn, dueOn, fileCount, fileSize, isPublic FROM projects WHERE account=? AND isArchived=? %s`)
	args := make([]interface{}, 0, 14)
	args = append(args, account, isArchived)
	if me != nil {
		args = append(args, account, *me)
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
		args = append(args, account, *after, *after)
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s, id LIMIT ?`, sortBy, sortDir))
	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, fmt.Sprintf(query.String(), specificSqlFilterTxt), args...)
	err.PanicIf(e)
	res := make([]*project, 0, limit+1)
	idx := 0
	resIdx := map[string]int{}
	for rows.Next() {
		proj := project{}
		err.PanicIf(rows.Scan(&proj.Id, &proj.IsArchived, &proj.Name, &proj.CreatedOn, &proj.StartOn, &proj.DueOn, &proj.FileCount, &proj.FileSize, &proj.IsPublic))
		res = append(res, &proj)
		resIdx[proj.Id.String()] = idx
		idx++
	}
	if len(res) > 0 { //populate task properties
		var i id.Id
		var description *string
		var totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount uint64
		var isParallel bool
		query.Reset()
		args = make([]interface{}, 0, len(res)+1)
		args = append(args, account, res[0].Id)
		query.WriteString(`SELECT id, description, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel FROM tasks WHERE account=? AND project=id AND project IN (?`)
		for _, proj := range res[1:] {
			query.WriteString(`,?`)
			args = append(args, proj.Id)
		}
		query.WriteString(fmt.Sprintf(`) LIMIT %d`, len(res)))
		rows, e := ctx.TreeQuery(shard, query.String(), args...)
		err.PanicIf(e)
		for rows.Next() {
			rows.Scan(&i, &description, &totalRemainingTime, &totalLoggedTime, &minimumRemainingTime, &linkedFileCount, &chatCount, &childCount, &descendantCount, &isParallel)
			proj := res[resIdx[i.String()]]
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
		return &getProjectsResp{Projects: res[:limit], More: true}
	}
	return &getProjectsResp{Projects: res, More: false}
}
