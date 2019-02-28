package project

import (
	"bytes"
	"fmt"
	"github.com/0xor1/panic"
	"github.com/0xor1/trees/server/util/activity"
	"github.com/0xor1/trees/server/util/cachekey"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/ctx"
	"github.com/0xor1/trees/server/util/db"
	"github.com/0xor1/trees/server/util/err"
	"github.com/0xor1/trees/server/util/field"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/sortdir"
	"github.com/0xor1/trees/server/util/validate"
	"strings"
	"time"
)

func dbGetProjectExists(ctx ctx.Ctx, shard int, account, project id.Id) bool {
	var exists bool
	cacheKey := cachekey.NewGet("db.GetProjectExists", shard, account, project).Project(account, project)
	if ctx.GetCacheValue(&exists, cacheKey) {
		return exists
	}
	row := ctx.TreeQueryRow(shard, `SELECT COUNT(*) = 1 FROM projects WHERE account=? AND id=?`, account, project)
	panic.IfNotNil(row.Scan(&exists))
	ctx.SetCacheValue(exists, cacheKey)
	return exists
}

func dbCreateProject(ctx ctx.Ctx, shard int, account id.Id, project *Project) {
	_, e := ctx.TreeExec(shard, `CALL createProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account, project.Id, ctx.Me(), project.Name, project.Description, project.HoursPerDay, project.DaysPerWeek, project.CreatedOn, project.StartOn, project.DueOn, project.IsParallel, project.IsPublic)
	panic.IfNotNil(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountActivities(account).AccountProjectsSet(account))
}

func dbEdit(ctx ctx.Ctx, shard int, account, project id.Id, fields Fields) {
	setIsPublic := fields.IsPublic != nil
	if !setIsPublic {
		fields.IsPublic = &field.Bool{}
	}
	setIsArchived := fields.IsArchived != nil
	if !setIsArchived {
		fields.IsArchived = &field.Bool{}
	}
	setHoursPerDay := fields.HoursPerDay != nil
	if !setHoursPerDay {
		fields.HoursPerDay = &field.UInt8{}
	}
	setDaysPerWeek := fields.DaysPerWeek != nil
	if !setDaysPerWeek {
		fields.DaysPerWeek = &field.UInt8{}
	}
	setStartOn := fields.StartOn != nil
	if !setStartOn {
		fields.StartOn = &field.TimePtr{}
	}
	setDueOn := fields.DueOn != nil
	if !setDueOn {
		fields.DueOn = &field.TimePtr{}
	}
	if !setIsPublic && !setIsArchived && !setHoursPerDay && !setDaysPerWeek && !setStartOn && !setDueOn {
		return
	}
	_, e := ctx.TreeExec(shard, `CALL editProject(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, account, project, ctx.Me(), setIsPublic, fields.IsPublic.Val, setIsArchived, fields.IsArchived.Val, setHoursPerDay, fields.HoursPerDay.Val, setDaysPerWeek, fields.DaysPerWeek.Val, setStartOn, fields.StartOn.Val, setDueOn, fields.DueOn.Val)
	panic.IfNotNil(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountActivities(account).Project(account, project).ProjectActivities(account, project))
}

func dbGetProject(ctx ctx.Ctx, shard int, account, proj id.Id) *Project {
	res := Project{}
	cacheKey := cachekey.NewGet("project.dbGetProject", shard, account, proj).Project(account, proj)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	row := ctx.TreeQueryRow(shard, `SELECT p.id, p.isArchived, p.name, p.hoursPerDay, p.daysPerWeek, p.createdOn, p.startOn, p.dueOn, p.fileCount, p.fileSize, p.isPublic, t.description, t.totalRemainingTime, t.totalLoggedTime, t.minimumRemainingTime, t.linkedFileCount, t.chatCount, t.childCount, t.descendantCount, t.isParallel FROM projects p, tasks t WHERE p.account=? AND p.id=? AND t.account=? AND t.project=? AND t.id=?`, account, proj, account, proj, proj)
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&res.Id, &res.IsArchived, &res.Name, &res.HoursPerDay, &res.DaysPerWeek, &res.CreatedOn, &res.StartOn, &res.DueOn, &res.FileCount, &res.FileSize, &res.IsPublic, &res.Description, &res.TotalRemainingTime, &res.TotalLoggedTime, &res.MinimumRemainingTime, &res.LinkedFileCount, &res.ChatCount, &res.ChildCount, &res.DescendantCount, &res.IsParallel)) {
		return nil
	}
	ctx.SetCacheValue(res, cacheKey)
	return &res
}

func dbGetPublicProjects(ctx ctx.Ctx, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) *GetSetResult {
	return dbGetProjects(ctx, shard, `AND isPublic=true`, account, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit)
}

func dbGetPublicAndSpecificAccessProjects(ctx ctx.Ctx, shard int, account, me id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) *GetSetResult {
	return dbGetProjects(ctx, shard, `AND (isPublic=true OR id IN (SELECT project FROM projectMembers WHERE account=? AND isActive=true AND id=?))`, account, &me, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit)
}

func dbGetAllProjects(ctx ctx.Ctx, shard int, account id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) *GetSetResult {
	return dbGetProjects(ctx, shard, ``, account, nil, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit)
}

func dbDeleteProject(ctx ctx.Ctx, shard int, account, project id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteProject(?, ?, ?)`, account, project, ctx.Me())
	panic.IfNotNil(e)
	ctx.TouchDlms(cachekey.NewSetDlms().AccountProjectsSet(account).ProjectMaster(account, project))
}

func dbAddMemberOrSetActive(ctx ctx.Ctx, shard int, account, project id.Id, member *AddProjectMember) {
	db.MakeChangeHelper(ctx, shard, `CALL addProjectMemberOrSetActive(?, ?, ?, ?, ?)`, account, project, ctx.Me(), member.Id, member.Role)
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectMember(account, project, member.Id).ProjectActivities(account, project))
}

func dbSetMemberRole(ctx ctx.Ctx, shard int, account, project, member id.Id, role cnst.ProjectRole) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberRole(?, ?, ?, ?, ?)`, account, project, ctx.Me(), member, role)
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectMember(account, project, member).ProjectActivities(account, project))
}

func dbSetMemberInactive(ctx ctx.Ctx, shard int, account, project id.Id, member id.Id) {
	db.MakeChangeHelper(ctx, shard, `CALL setProjectMemberInactive(?, ?, ?, ?)`, account, project, ctx.Me(), member)
	ctx.TouchDlms(cachekey.NewSetDlms().ProjectMember(account, project, member).ProjectActivities(account, project))
}

func dbGetMembers(ctx ctx.Ctx, shard int, account, project id.Id, role *cnst.ProjectRole, nameOrDisplayNameFilter *string, nameOrDisplayNameFilterIsPrefix bool, after *id.Id, limit int) *GetMembersResult {
	if nameOrDisplayNameFilter != nil {
		validate.StringArg("nameOrDisplayName", *nameOrDisplayNameFilter, ctx.DisplayNameMinRuneCount(), ctx.DisplayNameMaxRuneCount(), ctx.DisplayNameRegexMatchers())
	}
	res := GetMembersResult{}
	cacheKey := cachekey.NewGet("project.dbGetMembers", shard, account, project, role, nameOrDisplayNameFilter, nameOrDisplayNameFilterIsPrefix, after, limit).ProjectMembersSet(account, project)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
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
	if nameOrDisplayNameFilter != nil {
		query.WriteString(` AND (p1.name LIKE ? OR p1.displayName LIKE ?)`)
		strVal := strings.Trim(*nameOrDisplayNameFilter, " ")
		if nameOrDisplayNameFilterIsPrefix {
			strVal = fmt.Sprintf("%s%%", strVal)
		} else {
			strVal = fmt.Sprintf("%%%s%%", strVal)
		}
		args = append(args, strVal, strVal)
	}
	if nameOrDisplayNameFilter == nil || !nameOrDisplayNameFilterIsPrefix {
		query.WriteString(` ORDER BY p1.role ASC, p1.name ASC LIMIT ?`)
	} else {
		query.WriteString(` ORDER BY p1.name ASC LIMIT ?`)
	}

	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, query.String(), args...)
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	memSet := make([]*Member, 0, limit+1)
	for rows.Next() {
		mem := Member{}
		panic.IfNotNil(rows.Scan(&mem.Id, &mem.IsActive, &mem.TotalRemainingTime, &mem.TotalLoggedTime, &mem.Role))
		memSet = append(memSet, &mem)
	}
	if len(memSet) == limit+1 {
		res.Members = memSet[:limit]
		res.More = true
	} else {
		res.Members = memSet
		res.More = false
	}
	ctx.SetCacheValue(&res, cacheKey)
	return &res
}

func dbGetMember(ctx ctx.Ctx, shard int, account, project, mem id.Id) *Member {
	res := Member{}
	cacheKey := cachekey.NewGet("project.dbGetMember", shard, account, project, mem).ProjectMember(account, project, mem)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	row := ctx.TreeQueryRow(shard, `SELECT id, isActive, role FROM projectMembers WHERE account=? AND project=? AND id=?`, account, project, mem)
	panic.IfNotNil(row.Scan(&res.Id, &res.IsActive, &res.Role))
	ctx.SetCacheValue(&res, cacheKey)
	return &res
}

func dbGetActivities(ctx ctx.Ctx, shard int, account, project id.Id, item, member *id.Id, occurredAfter, occurredBefore *time.Time, limit int) []*activity.Activity {
	ctx.ReturnBadRequestNowIf(occurredAfter != nil && occurredBefore != nil, "only one of occurredAfter or occurredBefore can be set")
	res := make([]*activity.Activity, 0, limit)
	cacheKey := cachekey.NewGet("project.dbGetActivities", shard, account, project, item, member, occurredAfter, occurredBefore, limit).ProjectActivities(account, project)
	if ctx.GetCacheValue(&res, cacheKey) {
		return res
	}
	query := bytes.NewBufferString(`SELECT occurredOn, item, member, itemType, itemHasBeenDeleted, action, itemName, extraInfo FROM projectActivities WHERE account=? AND project=?`)
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
	panic.IfNotNil(e)
	for rows.Next() {
		act := activity.Activity{}
		panic.IfNotNil(rows.Scan(&act.OccurredOn, &act.Item, &act.Member, &act.ItemType, &act.ItemHasBeenDeleted, &act.Action, &act.ItemName, &act.ExtraInfo))
		res = append(res, &act)
	}
	ctx.SetCacheValue(res, cacheKey)
	return res
}

func dbGetProjects(ctx ctx.Ctx, shard int, specificSqlFilterTxt string, account id.Id, me *id.Id, nameContains *string, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore *time.Time, isArchived bool, sortBy cnst.SortBy, sortAsc bool, after *id.Id, limit int) *GetSetResult {
	res := GetSetResult{}
	cacheKey := cachekey.NewGet("project.dbGetProjects", shard, specificSqlFilterTxt, account, me, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit).AccountProjectsSet(account)
	if ctx.GetCacheValue(&res, cacheKey) {
		return &res
	}
	query := bytes.NewBufferString(`SELECT id, isArchived, name, hoursPerDay, daysPerWeek, createdOn, startOn, dueOn, fileCount, fileSize, isPublic FROM projects WHERE account=? AND isArchived=? %s`)
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
		query.WriteString(fmt.Sprintf(` AND %s %s= (SELECT %s FROM projects WHERE account=? AND id=?) AND id > ?`, sortBy, sortdir.GtLtSymbol(sortAsc), sortBy))
		args = append(args, account, *after, *after)
	}
	query.WriteString(fmt.Sprintf(` ORDER BY %s %s, id LIMIT ?`, sortBy, sortdir.String(sortAsc)))
	args = append(args, limit+1)
	rows, e := ctx.TreeQuery(shard, fmt.Sprintf(query.String(), specificSqlFilterTxt), args...)
	if rows != nil {
		defer rows.Close()
	}
	panic.IfNotNil(e)
	projSet := make([]*Project, 0, limit+1)
	idx := 0
	resIdx := map[string]int{}
	for rows.Next() {
		proj := Project{}
		panic.IfNotNil(rows.Scan(&proj.Id, &proj.IsArchived, &proj.Name, &proj.HoursPerDay, &proj.DaysPerWeek, &proj.CreatedOn, &proj.StartOn, &proj.DueOn, &proj.FileCount, &proj.FileSize, &proj.IsPublic))
		projSet = append(projSet, &proj)
		resIdx[proj.Id.String()] = idx
		idx++
	}
	if len(projSet) > 0 { //populate task properties
		var i id.Id
		var description *string
		var totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount uint64
		var isParallel bool
		query.Reset()
		args = make([]interface{}, 0, len(projSet)+1)
		args = append(args, account, projSet[0].Id)
		query.WriteString(`SELECT id, description, totalRemainingTime, totalLoggedTime, minimumRemainingTime, linkedFileCount, chatCount, childCount, descendantCount, isParallel FROM tasks WHERE account=? AND project=id AND project IN (?`)
		for _, proj := range projSet[1:] {
			query.WriteString(`,?`)
			args = append(args, proj.Id)
		}
		query.WriteString(fmt.Sprintf(`) LIMIT %d`, len(projSet)))
		rows, e := ctx.TreeQuery(shard, query.String(), args...)
		if rows != nil {
			defer rows.Close()
		}
		panic.IfNotNil(e)
		for rows.Next() {
			rows.Scan(&i, &description, &totalRemainingTime, &totalLoggedTime, &minimumRemainingTime, &linkedFileCount, &chatCount, &childCount, &descendantCount, &isParallel)
			proj := projSet[resIdx[i.String()]]
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
	if len(projSet) == limit+1 {
		res.Projects = projSet[:limit]
		res.More = true
	} else {
		res.Projects = projSet
		res.More = false
	}
	ctx.SetCacheValue(&res, cacheKey)
	return &res
}
