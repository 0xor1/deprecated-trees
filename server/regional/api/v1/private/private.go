package private

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bytes"
	"math/rand"
	"strings"
)

var (
	zeroOwnerCountErr = &AppError{Code: "r_v1_pr_zoc", Message: "zero owner count", Public: true}
	noSuchRegionErr   = &AppError{Code: "r_v1_pr_nsr", Message: "no such region", Public: true}
)

type createAccountArgs struct {
	AccountId     Id      `json:"accountId"`
	MyId          Id      `json:"myId"`
	MyName        string  `json:"myName"`
	MyDisplayName *string `json:"myDisplayName"`
}

var createAccount = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/createAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		return dbCreateAccount(ctx, args.AccountId, args.MyId, args.MyName, args.MyDisplayName)
	},
}

type deleteAccountArgs struct {
	Shard     int `json:"shard"`
	AccountId Id  `json:"accountId"`
	MyId      Id  `json:"myId"`
}

var deleteAccount = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/deleteAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		if !args.MyId.Equal(args.AccountId) {
			ctx.Validate().MemberHasAccountOwnerAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId))
		}
		dbDeleteAccount(ctx, args.Shard, args.AccountId)
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard     int                 `json:"shard"`
	AccountId Id                  `json:"accountId"`
	MyId      Id                  `json:"myId"`
	Members   []*AddMemberPrivate `json:"members"`
}

var addMembers = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/addMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		ctx.Validate().EntityCount(len(args.Members))
		if args.AccountId.Equal(args.MyId) {
			InvalidOperationErr.Panic()
		}
		accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
		ctx.Validate().MemberHasAccountAdminAccess(accountRole)

		allIds := make([]Id, 0, len(args.Members))
		newMembersMap := map[string]*AddMemberPrivate{}
		for _, mem := range args.Members { //loop over all the new entries and check permissions and build up useful id map and allIds slice
			mem.Role.Validate()
			if mem.Role == AccountOwner {
				ctx.Validate().MemberHasAccountOwnerAccess(accountRole)
			}
			newMembersMap[mem.Id.String()] = mem
			allIds = append(allIds, mem.Id)
		}

		inactiveMemberIds := dbGetAllInactiveMemberIdsFromInputSet(ctx, args.Shard, args.AccountId, allIds)
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
			dbAddMembers(ctx, args.Shard, args.AccountId, newMembers)
		}
		if len(inactiveMembers) > 0 {
			dbUpdateMembersAndSetActive(ctx, args.Shard, args.AccountId, inactiveMembers) //has to be AddMemberPrivate in case the member changed their name whilst they were inactive on the account
		}
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.AccountId, args.MyId, allIds, "added")
		return nil
	},
}

type removeMembersArgs struct {
	Shard     int  `json:"shard"`
	AccountId Id   `json:"accountId"`
	MyId      Id   `json:"myId"`
	Members   []Id `json:"members"`
}

var removeMembers = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/removeMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		ctx.Validate().EntityCount(len(args.Members))
		if args.AccountId.Equal(args.MyId) {
			InvalidOperationErr.Panic()
		}

		accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
		if accountRole == nil {
			InsufficientPermissionErr.Panic()
		}

		switch *accountRole {
		case AccountOwner:
			totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.AccountId)
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, args.Members)
			if totalOwnerCount == ownerCountInRemoveSet {
				zeroOwnerCountErr.Panic()
			}

		case AccountAdmin:
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, args.Members)
			if ownerCountInRemoveSet > 0 {
				InsufficientPermissionErr.Panic()
			}
		default:
			if len(args.Members) != 1 || !args.Members[0].Equal(args.MyId) { //any member can remove themselves
				InsufficientPermissionErr.Panic()
			}
		}

		dbSetMembersInactive(ctx, args.Shard, args.AccountId, args.Members)
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.AccountId, args.MyId, args.Members, "removed")
		return nil
	},
}

type memberIsOnlyAccountOwnerArgs struct {
	Shard     int `json:"shard"`
	AccountId Id  `json:"accountId"`
	MyId      Id  `json:"myId"`
}

var memberIsOnlyAccountOwner = &Endpoint{
	Method:    GET,
	Path:      "/api/v1/private/memberIsOnlyAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsOnlyAccountOwnerArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*memberIsOnlyAccountOwnerArgs)
		if args.AccountId.Equal(args.MyId) {
			return true
		}
		totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.AccountId)
		ownerCount := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, []Id{args.MyId})
		return totalOwnerCount == 1 && ownerCount == 1
	},
}

type setMemberNameArgs struct {
	Shard     int    `json:"shard"`
	AccountId Id     `json:"accountId"`
	MyId      Id     `json:"myId"`
	NewName   string `json:"newName"`
}

var setMemberName = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/setMemberName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberNameArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*setMemberNameArgs)
		dbSetMemberName(ctx, args.Shard, args.AccountId, args.MyId, args.NewName)
		return nil
	},
}

type setMemberDisplayNameArgs struct {
	Shard          int     `json:"shard"`
	AccountId      Id      `json:"accountId"`
	MyId           Id      `json:"myId"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setMemberDisplayName = &Endpoint{
	Method:    POST,
	Path:      "/api/v1/private/setMemberDisplayName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberNameArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*setMemberDisplayNameArgs)
		dbSetMemberDisplayName(ctx, args.Shard, args.AccountId, args.MyId, args.NewDisplayName)
		return nil
	},
}

type memberIsAccountOwnerArgs struct {
	Shard     int `json:"shard"`
	AccountId Id  `json:"accountId"`
	MyId      Id  `json:"myId"`
}

var memberIsAccountOwner = &Endpoint{
	Method:    GET,
	Path:      "/api/v1/private/memberIsAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsAccountOwnerArgs{}
	},
	CtxHandler: func(ctx RegionalCtx, a interface{}) interface{} {
		args := a.(*memberIsAccountOwnerArgs)
		if !args.MyId.Equal(args.AccountId) {
			accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
			if accountRole != nil && *accountRole == AccountOwner {
				return true
			} else {
				return false
			}
		}
		return true
	},
}

var Endpoints = []*Endpoint{
	createAccount,
	deleteAccount,
	addMembers,
	removeMembers,
	memberIsOnlyAccountOwner,
	setMemberName,
	setMemberDisplayName,
	memberIsAccountOwner,
}

func dbCreateAccount(ctx RegionalCtx, id Id, myId Id, myName string, myDisplayName *string) int {
	shardId := rand.Intn(ctx.Db().TreeShardCount())
	_, err := ctx.Db().Tree(shardId).Exec(`CALL registerAccount(?, ?, ?, ?)`, []byte(id), []byte(myId), myName, myDisplayName)
	PanicIf(err)
	return shardId
}

func dbDeleteAccount(ctx RegionalCtx, shard int, account Id) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL deleteAccount(?)`, []byte(account))
	PanicIf(err)
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
	PanicIf(err)
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
	if IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
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
	PanicIf(err)
}

func dbUpdateMembersAndSetActive(ctx RegionalCtx, shard int, accountId Id, members []*AddMemberPrivate) {
	for _, mem := range members {
		_, err := ctx.Db().Tree(shard).Exec(`CALL updateMembersAndSetActive(?, ?, ?, ?, ?)`, []byte(accountId), []byte(mem.Id), mem.Name, mem.DisplayName, mem.Role)
		PanicIf(err)
	}
}

func dbGetTotalOwnerCount(ctx RegionalCtx, shard int, accountId Id) int {
	count := 0
	IsSqlErrNoRowsElsePanicIf(ctx.Db().Tree(shard).QueryRow(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, []byte(accountId)).Scan(&count))
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
	IsSqlErrNoRowsElsePanicIf(ctx.Db().Tree(shard).QueryRow(query.String(), queryArgs...).Scan(&count))
	return count
}

func dbSetMembersInactive(ctx RegionalCtx, shard int, accountId Id, members []Id) {
	accountIdBytes := []byte(accountId)
	for _, mem := range members {
		_, err := ctx.Db().Tree(shard).Exec(`CALL setAccountMemberInactive(?, ?)`, accountIdBytes, []byte(mem))
		PanicIf(err)
	}
}

func dbSetMemberName(ctx RegionalCtx, shard int, accountId Id, member Id, newName string) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL setMemberName(?, ?, ?)`, []byte(accountId), []byte(member), newName)
	PanicIf(err)
}

func dbSetMemberDisplayName(ctx RegionalCtx, shard int, accountId Id, member Id, newDisplayName *string) {
	_, err := ctx.Db().Tree(shard).Exec(`CALL setMemberDisplayName(?, ?, ?)`, []byte(accountId), []byte(member), newDisplayName)
	PanicIf(err)
}

func dbLogAccountBatchAddOrRemoveMembersActivity(ctx RegionalCtx, shard int, accountId, member Id, members []Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	now := Now()
	args = append(args, []byte(accountId), now, []byte(member), []byte(members[0]), "member", action, nil, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, []byte(accountId), now, []byte(member), []byte(memId), "member", action, nil, nil)
	}
	_, err := ctx.Db().Tree(shard).Exec(query.String(), args...)
	PanicIf(err)
}

func NewClient(regions map[string]string) *client {
	lowerRegionsMap := map[string]string{}
	for k, v := range regions {
		regions[strings.ToLower(k)] = v
	}
	return &client{
		regions: lowerRegionsMap,
	}
}

type client struct {
	regions map[string]string
}

func (c *client) getHost(region string) string {
	host, exists := c.regions[strings.ToLower(region)]
	if !exists {
		noSuchRegionErr.Panic()
	}
	return host
}

func (c *client) GetRegions() []string {
	regions := make([]string, 0, len(c.regions))
	for r := range c.regions {
		regions = append(regions, r)
	}
	return regions
}

func (c *client) IsValidRegion(region string) bool {
	_, exists := c.regions[strings.ToLower(region)]
	return exists
}

func (c *client) CreateAccount(region string, account, myId Id, myName string, myDisplayName *string) (int, error) {
	respVal := 0
	val, err := createAccount.DoRequest(c.getHost(region), &createAccountArgs{
		AccountId:     account,
		MyId:          myId,
		MyName:        myName,
		MyDisplayName: myDisplayName,
	}, nil, &respVal)
	if val != nil {
		return *val.(*int), err
	}
	return 0, err
}

func (c *client) DeleteAccount(region string, shard int, account, myId Id) error {
	_, err := deleteAccount.DoRequest(c.getHost(region), &deleteAccountArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, nil)
	return err
}

func (c *client) AddMembers(region string, shard int, account, myId Id, members []*AddMemberPrivate) error {
	_, err := addMembers.DoRequest(c.getHost(region), &addMembersArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
		Members:   members,
	}, nil, nil)
	return err
}

func (c *client) RemoveMembers(region string, shard int, account, myId Id, members []Id) error {
	_, err := removeMembers.DoRequest(c.getHost(region), &removeMembersArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, nil)
	return err
}

func (c *client) MemberIsOnlyAccountOwner(region string, shard int, account, myId Id) (bool, error) {
	respVal := false
	val, err := memberIsOnlyAccountOwner.DoRequest(c.getHost(region), &memberIsOnlyAccountOwnerArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), err
	}
	return false, err
}

func (c *client) SetMemberName(region string, shard int, account, myId Id, newName string) error {
	_, err := setMemberName.DoRequest(c.getHost(region), &setMemberNameArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
		NewName:   newName,
	}, nil, nil)
	return err
}

func (c *client) SetMemberDisplayName(region string, shard int, account, myId Id, newDisplayName *string) error {
	_, err := setMemberDisplayName.DoRequest(c.getHost(region), &setMemberDisplayNameArgs{
		Shard:          shard,
		AccountId:      account,
		MyId:           myId,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return err
}

func (c *client) MemberIsAccountOwner(region string, shard int, account, myId Id) (bool, error) {
	respVal := false
	val, err := memberIsAccountOwner.DoRequest(c.getHost(region), &memberIsAccountOwnerArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), err
	}
	return false, err
}
