package private

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bytes"
	"math/rand"
)

var (
	zeroOwnerCountErr = &AppError{Code: "r_v1_pr_zoc", Message: "zero owner count", Public: true}
)

type Api interface {
	CreateAccount(ctx RegionalCtx, accountId, myId Id, myName string, myDisplayName *string) int
	DeleteAccount(ctx RegionalCtx, shard int, accountId, myId Id)
	AddMembers(ctx RegionalCtx, shard int, accountId, myId Id, members []*AddMemberPrivate)
	RemoveMembers(ctx RegionalCtx, shard int, accountId, myId Id, members []Id)
	MemberIsOnlyAccountOwner(ctx RegionalCtx, shard int, accountId, memberId Id) bool
	SetMemberName(ctx RegionalCtx, shard int, accountId, memberId Id, newName string)
	SetMemberDisplayName(ctx RegionalCtx, shard int, accountId, memberId Id, newDisplayName *string)
	MemberIsAccountOwner(ctx RegionalCtx, shard int, accountId, memberId Id) bool
}

func New() Api {
	return &api{}
}

type api struct{}

func (a *api) CreateAccount(ctx RegionalCtx, accountId, myId Id, myName string, myDisplayName *string) int {
	return dbCreateAccount(ctx, accountId, myId, myName, myDisplayName)
}

func (a *api) DeleteAccount(ctx RegionalCtx, shard int, accountId, myId Id) {
	if !myId.Equal(accountId) {
		ctx.Validate().MemberHasAccountOwnerAccess(dbGetAccountRole(ctx, shard, accountId, myId))
	}
	dbDeleteAccount(ctx, shard, accountId)
	//TODO delete s3 data, uploaded files etc
}

func (a *api) AddMembers(ctx RegionalCtx, shard int, accountId, myId Id, members []*AddMemberPrivate) {
	ctx.Validate().EntityCount(len(members))
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}
	accountRole := dbGetAccountRole(ctx, shard, accountId, myId)
	ctx.Validate().MemberHasAccountAdminAccess(accountRole)

	allIds := make([]Id, 0, len(members))
	newMembersMap := map[string]*AddMemberPrivate{}
	for _, mem := range members { //loop over all the new entries and check permissions and build up useful id map and allIds slice
		mem.Role.Validate()
		if mem.Role == AccountOwner {
			ctx.Validate().MemberHasAccountOwnerAccess(accountRole)
		}
		newMembersMap[mem.Id.String()] = mem
		allIds = append(allIds, mem.Id)
	}

	inactiveMemberIds := dbGetAllInactiveMemberIdsFromInputSet(ctx, shard, accountId, allIds)
	inactiveMembers := make([]*AddMemberPrivate, 0, len(inactiveMemberIds))
	for _, inactiveMemberId := range inactiveMemberIds {
		idStr := inactiveMemberId.String()
		inactiveMembers = append(inactiveMembers, newMembersMap[idStr])
		delete(newMembersMap, idStr)
	}

	newMembers := make([]*AddMemberPrivate, 0, len(newMembersMap))
	for _, newMem := range newMembersMap {
		newMembers = append(newMembers, newMem)
	}

	if len(newMembers) > 0 {
		dbAddMembers(ctx, shard, accountId, newMembers)
	}
	if len(inactiveMembers) > 0 {
		dbUpdateMembersAndSetActive(ctx, shard, accountId, inactiveMembers) //has to be AddMemberPrivate in case the member changed their name whilst they were inactive on the account
	}
	dbLogAccountBatchAddOrRemoveMembersActivity(ctx, shard, accountId, myId, allIds, "added")
}

func (a *api) RemoveMembers(ctx RegionalCtx, shard int, accountId, myId Id, members []Id) {
	ctx.Validate().EntityCount(len(members))
	if accountId.Equal(myId) {
		InvalidOperationErr.Panic()
	}

	accountRole := dbGetAccountRole(ctx, shard, accountId, myId)
	if accountRole == nil {
		InsufficientPermissionErr.Panic()
	}

	switch *accountRole {
	case AccountOwner:
		totalOwnerCount := dbGetTotalOwnerCount(ctx, shard, accountId)
		ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, shard, accountId, members)
		if totalOwnerCount == ownerCountInRemoveSet {
			zeroOwnerCountErr.Panic()
		}

	case AccountAdmin:
		ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, shard, accountId, members)
		if ownerCountInRemoveSet > 0 {
			InsufficientPermissionErr.Panic()
		}
	default:
		if len(members) != 1 || !members[0].Equal(myId) { //any member can remove themselves
			InsufficientPermissionErr.Panic()
		}
	}

	dbSetMembersInactive(ctx, shard, accountId, members)
	dbLogAccountBatchAddOrRemoveMembersActivity(ctx, shard, accountId, myId, members, "removed")
}

func (a *api) MemberIsOnlyAccountOwner(ctx RegionalCtx, shard int, accountId, myId Id) bool {
	if accountId.Equal(myId) {
		return true
	}
	totalOwnerCount := dbGetTotalOwnerCount(ctx, shard, accountId)
	ownerCount := dbGetOwnerCountInSet(ctx, shard, accountId, []Id{myId})
	return totalOwnerCount == 1 && ownerCount == 1
}

func (a *api) SetMemberName(ctx RegionalCtx, shard int, accountId, myId Id, newName string) {
	dbSetMemberName(ctx, shard, accountId, myId, newName)
}

func (a *api) SetMemberDisplayName(ctx RegionalCtx, shard int, accountId, myId Id, newDisplayName *string) {
	dbSetMemberDisplayName(ctx, shard, accountId, myId, newDisplayName)
}

func (a *api) MemberIsAccountOwner(ctx RegionalCtx, shard int, accountId, myId Id) bool {
	if !myId.Equal(accountId) {
		accountRole := dbGetAccountRole(ctx, shard, accountId, myId)
		if accountRole != nil && *accountRole == AccountOwner {
			return true
		} else {
			return false
		}
	}
	return true
}

func dbCreateAccount(ctx RegionalCtx, id Id, myId Id, myName string, myDisplayName *string) int {
	shardId := rand.Intn(ctx.Db().TreeShardCount())
	_, err := ctx.Db().Tree(shardId).Exec(`CALL registerAccount(?, ?, ?, ?)`, []byte(id), []byte(myId), myName, myDisplayName)
	ctx.Error().PanicIf(err)
	return shardId
}

func dbDeleteAccount(ctx RegionalCtx, shard int, account Id) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL deleteAccount(?)`, []byte(account))
	ctx.Error().PanicIf(err)
}

func dbGetAllInactiveMemberIdsFromInputSet(ctx RegionalCtx, shard int, accountId Id, members []Id) []Id {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`SELECT id FROM accountMembers WHERE account=? AND isActive=false AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`)`)
	res := make([]Id, 0, len(members))
	rows, err := ctx.Db().Tree(shard).Query(query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	ctx.Error().PanicIf(err)
	for rows.Next() {
		id := make([]byte, 0, 16)
		rows.Scan(&id)
		res = append(res, Id(id))
	}
	return res
}

func dbGetAccountRole(ctx RegionalCtx, shard int, accountId, memberId Id) *AccountRole {
	row := ctx.Db().Tree(shard).QueryRow(`SELECT role FROM accountMembers WHERE account=? AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if ctx.Error().IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func dbAddMembers(ctx RegionalCtx, shard int, accountId Id, members []*AddMemberPrivate) {
	queryArgs := make([]interface{}, 0, 3*len(members))
	queryArgs = append(queryArgs, []byte(accountId), []byte(members[0].Id), members[0].Name, members[0].DisplayName, members[0].Role)
	query := bytes.NewBufferString(`INSERT INTO accountMembers (account, id, name, displayName, role) VALUES (?,?,?,?,?)`)
	for _, mem := range members[1:] {
		query.WriteString(`,(?,?,?,?,?)`)
		queryArgs = append(queryArgs, []byte(accountId), []byte(mem.Id), mem.Name, mem.DisplayName, mem.Role)
	}
	_, err := ctx.Db().Tree(shard).Exec(query.String(), queryArgs...)
	ctx.Error().PanicIf(err)
}

func dbUpdateMembersAndSetActive(ctx RegionalCtx, shard int, accountId Id, members []*AddMemberPrivate) {
	for _, mem := range members {
		_, err := ctx.Db().Tree(shard).Exec(`CALL updateMembersAndSetActive(?, ?, ?, ?, ?)`, []byte(accountId), []byte(mem.Id), mem.Name, mem.DisplayName, mem.Role)
		ctx.Error().PanicIf(err)
	}
}

func dbGetTotalOwnerCount(ctx RegionalCtx, shard int, accountId Id) int {
	count := 0
	ctx.Error().IsSqlErrNoRowsElsePanicIf(ctx.Db().Tree(shard).QueryRow(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, []byte(accountId)).Scan(&count))
	return count
}

func dbGetOwnerCountInSet(ctx RegionalCtx, shard int, accountId Id, members []Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, []byte(accountId), []byte(members[0]))
	query := bytes.NewBufferString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, []byte(mem))
	}
	query.WriteString(`)`)
	count := 0
	ctx.Error().IsSqlErrNoRowsElsePanicIf(ctx.Db().Tree(shard).QueryRow(query.String(), queryArgs...).Scan(&count))
	return count
}

func dbSetMembersInactive(ctx RegionalCtx, shard int, accountId Id, members []Id) {
	accountIdBytes := []byte(accountId)
	for _, mem := range members {
		_, err := ctx.Db().Tree(shard).Exec(`CALL setAccountMemberInactive(?, ?)`, accountIdBytes, []byte(mem))
		ctx.Error().PanicIf(err)
	}
}

func dbSetMemberName(ctx RegionalCtx, shard int, accountId Id, member Id, newName string) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL setMemberName(?, ?, ?)`, []byte(accountId), []byte(member), newName)
	ctx.Error().PanicIf(err)
}

func dbSetMemberDisplayName(ctx RegionalCtx, shard int, accountId Id, member Id, newDisplayName *string) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL setMemberDisplayName(?, ?, ?)`, []byte(accountId), []byte(member), newDisplayName)
	ctx.Error().PanicIf(err)
}

func dbLogAccountBatchAddOrRemoveMembersActivity(ctx RegionalCtx, shard int, accountId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	now := ctx.Time().Now()
	args = append(args, []byte(accountId), now, []byte(member), []byte(members[0]), "member", action, nil, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, []byte(accountId), now, []byte(member), []byte(memId), "member", action, nil, nil)
	}
	_, err := ctx.Db().Tree(shard).Exec(query.String(), args...)
	ctx.Error().PanicIf(err)
}
