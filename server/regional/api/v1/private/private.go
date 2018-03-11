package private

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/core"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/time"
	"bitbucket.org/0xor1/task/server/util/validate"
	"bytes"
	"math/rand"
	"strings"
)

var (
	zeroOwnerCountErr = &err.Err{Code: "r_v1_pr_zoc", Message: "zero owner count"}
)

type createAccountArgs struct {
	AccountId     id.Id   `json:"accountId"`
	MyId          id.Id   `json:"myId"`
	MyName        string  `json:"myName"`
	MyDisplayName *string `json:"myDisplayName"`
}

var createAccount = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/createAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		return dbCreateAccount(ctx, args.AccountId, args.MyId, args.MyName, args.MyDisplayName)
	},
}

type deleteAccountArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	MyId      id.Id `json:"myId"`
}

var deleteAccount = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/deleteAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		if !args.MyId.Equal(args.AccountId) {
			validate.MemberHasAccountOwnerAccess(dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId))
		}
		dbDeleteAccount(ctx, args.Shard, args.AccountId)
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard     int                  `json:"shard"`
	AccountId id.Id                `json:"accountId"`
	MyId      id.Id                `json:"myId"`
	Members   []*private.AddMember `json:"members"`
}

var addMembers = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/addMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(args.MyId) {
			panic(err.InvalidOperation)
		}
		accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
		validate.MemberHasAccountAdminAccess(accountRole)

		allIds := make([]id.Id, 0, len(args.Members))
		newMembersMap := map[string]*private.AddMember{}
		for _, mem := range args.Members { //loop over all the new entries and check permissions and build up useful id map and allIds slice
			mem.Role.Validate()
			if mem.Role == cnst.AccountOwner {
				validate.MemberHasAccountOwnerAccess(accountRole)
			}
			newMembersMap[mem.Id.String()] = mem
			allIds = append(allIds, mem.Id)
		}

		inactiveMemberIds := dbGetAllInactiveMemberIdsFromInputSet(ctx, args.Shard, args.AccountId, allIds)
		inactiveMembers := make([]*private.AddMember, 0, len(inactiveMemberIds))
		for _, inactiveMemberId := range inactiveMemberIds {
			idStr := inactiveMemberId.String()
			inactiveMembers = append(inactiveMembers, newMembersMap[idStr])
			delete(newMembersMap, idStr)
		}

		newMembers := make([]*private.AddMember, 0, len(newMembersMap))
		for _, newMem := range newMembersMap {
			newMembers = append(newMembers, newMem)
		}

		if len(newMembers) > 0 {
			dbAddMembers(ctx, args.Shard, args.AccountId, newMembers)
		}
		if len(inactiveMembers) > 0 {
			dbUpdateMembersAndSetActive(ctx, args.Shard, args.AccountId, inactiveMembers) //has to be private.AddMember in case the member changed their name whilst they were inactive on the account
		}
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.AccountId, args.MyId, allIds, "added")
		return nil
	},
}

type removeMembersArgs struct {
	Shard     int     `json:"shard"`
	AccountId id.Id   `json:"accountId"`
	MyId      id.Id   `json:"myId"`
	Members   []id.Id `json:"members"`
}

var removeMembers = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/removeMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		if args.AccountId.Equal(args.MyId) {
			panic(err.InvalidOperation)
		}

		accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
		if accountRole == nil {
			panic(err.InsufficientPermission)
		}

		switch *accountRole {
		case cnst.AccountOwner:
			totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.AccountId)
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, args.Members)
			if totalOwnerCount == ownerCountInRemoveSet {
				panic(zeroOwnerCountErr)
			}

		case cnst.AccountAdmin:
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, args.Members)
			if ownerCountInRemoveSet > 0 {
				panic(err.InsufficientPermission)
			}
		default:
			if len(args.Members) != 1 || !args.Members[0].Equal(args.MyId) { //any member can remove themselves
				panic(err.InsufficientPermission)
			}
		}

		dbSetMembersInactive(ctx, args.Shard, args.AccountId, args.Members)
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.AccountId, args.MyId, args.Members, "removed")
		return nil
	},
}

type memberIsOnlyAccountOwnerArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	MyId      id.Id `json:"myId"`
}

var memberIsOnlyAccountOwner = &core.Endpoint{
	Method:    cnst.GET,
	Path:      "/api/v1/private/memberIsOnlyAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsOnlyAccountOwnerArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*memberIsOnlyAccountOwnerArgs)
		if args.AccountId.Equal(args.MyId) {
			return true
		}
		totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.AccountId)
		ownerCount := dbGetOwnerCountInSet(ctx, args.Shard, args.AccountId, []id.Id{args.MyId})
		return totalOwnerCount == 1 && ownerCount == 1
	},
}

type setMemberNameArgs struct {
	Shard     int    `json:"shard"`
	AccountId id.Id  `json:"accountId"`
	MyId      id.Id  `json:"myId"`
	NewName   string `json:"newName"`
}

var setMemberName = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/setMemberName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberNameArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setMemberNameArgs)
		dbSetMemberName(ctx, args.Shard, args.AccountId, args.MyId, args.NewName)
		return nil
	},
}

type setMemberDisplayNameArgs struct {
	Shard          int     `json:"shard"`
	AccountId      id.Id   `json:"accountId"`
	MyId           id.Id   `json:"myId"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setMemberDisplayName = &core.Endpoint{
	Method:    cnst.POST,
	Path:      "/api/v1/private/setMemberDisplayName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberDisplayNameArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*setMemberDisplayNameArgs)
		dbSetMemberDisplayName(ctx, args.Shard, args.AccountId, args.MyId, args.NewDisplayName)
		return nil
	},
}

type memberIsAccountOwnerArgs struct {
	Shard     int   `json:"shard"`
	AccountId id.Id `json:"accountId"`
	MyId      id.Id `json:"myId"`
}

var memberIsAccountOwner = &core.Endpoint{
	Method:    cnst.GET,
	Path:      "/api/v1/private/memberIsAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsAccountOwnerArgs{}
	},
	CtxHandler: func(ctx *core.Ctx, a interface{}) interface{} {
		args := a.(*memberIsAccountOwnerArgs)
		if !args.MyId.Equal(args.AccountId) {
			accountRole := dbGetAccountRole(ctx, args.Shard, args.AccountId, args.MyId)
			if accountRole != nil && *accountRole == cnst.AccountOwner {
				return true
			} else {
				return false
			}
		}
		return true
	},
}

var Endpoints = []*core.Endpoint{
	createAccount,
	deleteAccount,
	addMembers,
	removeMembers,
	memberIsOnlyAccountOwner,
	setMemberName,
	setMemberDisplayName,
	memberIsAccountOwner,
}

func dbCreateAccount(ctx *core.Ctx, id id.Id, myId id.Id, myName string, myDisplayName *string) int {
	shard := rand.Intn(ctx.TreeShardCount())
	_, e := ctx.TreeExec(shard, `CALL registerAccount(?, ?, ?, ?)`, id, myId, myName, myDisplayName)
	err.PanicIf(e)
	return shard
}

func dbDeleteAccount(ctx *core.Ctx, shard int, account id.Id) {
	_, e := ctx.TreeExec(shard, `CALL deleteAccount(?)`, account)
	err.PanicIf(e)
}

func dbGetAllInactiveMemberIdsFromInputSet(ctx *core.Ctx, shard int, accountId id.Id, members []id.Id) []id.Id {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, accountId, members[0])
	query := bytes.NewBufferString(`SELECT id FROM accountMembers WHERE account=? AND isActive=false AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	res := make([]id.Id, 0, len(members))
	rows, e := ctx.TreeQuery(shard, query.String(), queryArgs...)
	if rows != nil {
		defer rows.Close()
	}
	err.PanicIf(e)
	for rows.Next() {
		i := make([]byte, 0, 16)
		rows.Scan(&i)
		res = append(res, id.Id(i))
	}
	return res
}

func dbGetAccountRole(ctx *core.Ctx, shard int, accountId, memberId id.Id) *cnst.AccountRole {
	row := ctx.TreeQueryRow(shard, `SELECT role FROM accountMembers WHERE account=? AND id=?`, accountId, memberId)
	res := cnst.AccountRole(3)
	if err.IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func dbAddMembers(ctx *core.Ctx, shard int, accountId id.Id, members []*private.AddMember) {
	queryArgs := make([]interface{}, 0, 3*len(members))
	queryArgs = append(queryArgs, accountId, members[0].Id, members[0].Name, members[0].DisplayName, members[0].Role)
	query := bytes.NewBufferString(`INSERT INTO accountMembers (account, id, name, displayName, role) VALUES (?,?,?,?,?)`)
	for _, mem := range members[1:] {
		query.WriteString(`,(?,?,?,?,?)`)
		queryArgs = append(queryArgs, accountId, mem.Id, mem.Name, mem.DisplayName, mem.Role)
	}
	_, e := ctx.TreeExec(shard, query.String(), queryArgs...)
	err.PanicIf(e)
}

func dbUpdateMembersAndSetActive(ctx *core.Ctx, shard int, accountId id.Id, members []*private.AddMember) {
	for _, mem := range members {
		_, e := ctx.TreeExec(shard, `CALL updateMembersAndSetActive(?, ?, ?, ?, ?)`, accountId, mem.Id, mem.Name, mem.DisplayName, mem.Role)
		err.PanicIf(e)
	}
}

func dbGetTotalOwnerCount(ctx *core.Ctx, shard int, accountId id.Id) int {
	count := 0
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, `SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0`, accountId).Scan(&count))
	return count
}

func dbGetOwnerCountInSet(ctx *core.Ctx, shard int, accountId id.Id, members []id.Id) int {
	queryArgs := make([]interface{}, 0, len(members)+1)
	queryArgs = append(queryArgs, accountId, members[0])
	query := bytes.NewBufferString(`SELECT COUNT(*) FROM accountMembers WHERE account=? AND isActive=true AND role=0 AND id IN (?`)
	for _, mem := range members[1:] {
		query.WriteString(`,?`)
		queryArgs = append(queryArgs, mem)
	}
	query.WriteString(`)`)
	count := 0
	err.IsSqlErrNoRowsElsePanicIf(ctx.TreeQueryRow(shard, query.String(), queryArgs...).Scan(&count))
	return count
}

func dbSetMembersInactive(ctx *core.Ctx, shard int, accountId id.Id, members []id.Id) {
	accountIdBytes := accountId
	for _, mem := range members {
		_, e := ctx.TreeExec(shard, `CALL setAccountMemberInactive(?, ?)`, accountIdBytes, mem)
		err.PanicIf(e)
	}
}

func dbSetMemberName(ctx *core.Ctx, shard int, accountId id.Id, member id.Id, newName string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberName(?, ?, ?)`, accountId, member, newName)
	err.PanicIf(e)
}

func dbSetMemberDisplayName(ctx *core.Ctx, shard int, accountId, member id.Id, newDisplayName *string) {
	_, e := ctx.TreeExec(shard, `CALL setMemberDisplayName(?, ?, ?)`, accountId, member, newDisplayName)
	err.PanicIf(e)
}

func dbLogAccountBatchAddOrRemoveMembersActivity(ctx *core.Ctx, shard int, accountId, member id.Id, members []id.Id, action string) {
	query := bytes.NewBufferString(`INSERT INTO accountActivities (account, occurredOn, member, item, itemType, action, itemName, extraInfo) VALUES (?,?,?,?,?,?,?,?)`)
	args := make([]interface{}, 0, len(members)*8)
	now := time.Now()
	args = append(args, accountId, now, member, members[0], "member", action, nil, nil)
	for _, memId := range members[1:] {
		query.WriteString(`,(?,?,?,?,?,?,?,?)`)
		args = append(args, accountId, now, member, memId, "member", action, nil, nil)
	}
	_, e := ctx.TreeExec(shard, query.String(), args...)
	err.PanicIf(e)
}

func NewClient(regions map[string]string) private.V1Client {
	lowerRegionsMap := map[string]string{}
	for k, v := range regions {
		lowerRegionsMap[strings.ToLower(k)] = v
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
		panic(err.NoSuchRegion)
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

func (c *client) CreateAccount(region string, account, myId id.Id, myName string, myDisplayName *string) (int, error) {
	respVal := 0
	val, e := createAccount.DoRequest(nil, c.getHost(region), &createAccountArgs{
		AccountId:     account,
		MyId:          myId,
		MyName:        myName,
		MyDisplayName: myDisplayName,
	}, nil, &respVal)
	if val != nil {
		return *val.(*int), e
	}
	return 0, e
}

func (c *client) DeleteAccount(region string, shard int, account, myId id.Id) error {
	_, e := deleteAccount.DoRequest(nil, c.getHost(region), &deleteAccountArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, nil)
	return e
}

func (c *client) AddMembers(region string, shard int, account, myId id.Id, members []*private.AddMember) error {
	_, e := addMembers.DoRequest(nil, c.getHost(region), &addMembersArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
		Members:   members,
	}, nil, nil)
	return e
}

func (c *client) RemoveMembers(region string, shard int, account, myId id.Id, members []id.Id) error {
	_, err := removeMembers.DoRequest(nil, c.getHost(region), &removeMembersArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
		Members:   members,
	}, nil, nil)
	return err
}

func (c *client) MemberIsOnlyAccountOwner(region string, shard int, account, myId id.Id) (bool, error) {
	respVal := false
	val, e := memberIsOnlyAccountOwner.DoRequest(nil, c.getHost(region), &memberIsOnlyAccountOwnerArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}

func (c *client) SetMemberName(region string, shard int, account, myId id.Id, newName string) error {
	_, err := setMemberName.DoRequest(nil, c.getHost(region), &setMemberNameArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
		NewName:   newName,
	}, nil, nil)
	return err
}

func (c *client) SetMemberDisplayName(region string, shard int, account, myId id.Id, newDisplayName *string) error {
	_, e := setMemberDisplayName.DoRequest(nil, c.getHost(region), &setMemberDisplayNameArgs{
		Shard:          shard,
		AccountId:      account,
		MyId:           myId,
		NewDisplayName: newDisplayName,
	}, nil, nil)
	return e
}

func (c *client) MemberIsAccountOwner(region string, shard int, account, myId id.Id) (bool, error) {
	respVal := false
	val, e := memberIsAccountOwner.DoRequest(nil, c.getHost(region), &memberIsAccountOwnerArgs{
		Shard:     shard,
		AccountId: account,
		MyId:      myId,
	}, nil, &respVal)
	if val != nil {
		return *val.(*bool), e
	}
	return false, e
}
