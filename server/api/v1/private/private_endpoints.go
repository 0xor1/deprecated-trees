package private

import (
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/ctx"
	"bitbucket.org/0xor1/trees/server/util/db"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/err"
	"bitbucket.org/0xor1/trees/server/util/id"
	"bitbucket.org/0xor1/trees/server/util/private"
	"bitbucket.org/0xor1/trees/server/util/validate"
	"github.com/0xor1/panic"
	"net/http"
)

var (
	zeroOwnerCountErr = &err.Err{Code: "r_v1_pr_zoc", Message: "zero owner count"}
)

type createAccountArgs struct {
	Account       id.Id   `json:"account"`
	Me            id.Id   `json:"me"`
	MyName        string  `json:"myName"`
	MyDisplayName *string `json:"myDisplayName"`
	HasAvatar     bool    `json:"hasAvatar"`
}

var createAccount = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/createAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &createAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*createAccountArgs)
		return dbCreateAccount(ctx, args.Account, args.Me, args.MyName, args.MyDisplayName, args.HasAvatar)
	},
}

type deleteAccountArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Me      id.Id `json:"me"`
}

var deleteAccount = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/deleteAccount",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &deleteAccountArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*deleteAccountArgs)
		if !args.Me.Equal(args.Account) {
			validate.MemberHasAccountOwnerAccess(db.GetAccountRole(ctx, args.Shard, args.Account, args.Me))
		}
		dbDeleteAccount(ctx, args.Shard, args.Account)
		//TODO delete s3 data, uploaded files etc
		return nil
	},
}

type addMembersArgs struct {
	Shard   int                  `json:"shard"`
	Account id.Id                `json:"account"`
	Me      id.Id                `json:"me"`
	Members []*private.AddMember `json:"members"`
}

var addMembers = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/addMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &addMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*addMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		panic.IfTrueWith(args.Account.Equal(args.Me), err.InvalidOperation)
		accountRole := db.GetAccountRole(ctx, args.Shard, args.Account, args.Me)
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

		inactiveMemberIds := dbGetAllInactiveMembersFromInputSet(ctx, args.Shard, args.Account, allIds)
		inactiveMembers := make([]*private.AddMember, 0, len(inactiveMemberIds))
		for _, inactiveMember := range inactiveMemberIds {
			idStr := inactiveMember.String()
			inactiveMembers = append(inactiveMembers, newMembersMap[idStr])
			delete(newMembersMap, idStr)
		}

		newMembers := make([]*private.AddMember, 0, len(newMembersMap))
		for _, newMem := range newMembersMap {
			newMembers = append(newMembers, newMem)
		}

		if len(newMembers) > 0 {
			dbAddMembers(ctx, args.Shard, args.Account, newMembers)
		}
		if len(inactiveMembers) > 0 {
			dbUpdateMembersAndSetActive(ctx, args.Shard, args.Account, inactiveMembers) //has to be private.AddMember in case the member changed their name whilst they were inactive on the account
		}
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.Account, args.Me, allIds, "added")
		return nil
	},
}

type removeMembersArgs struct {
	Shard   int     `json:"shard"`
	Account id.Id   `json:"account"`
	Me      id.Id   `json:"me"`
	Members []id.Id `json:"members"`
}

var removeMembers = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/removeMembers",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &removeMembersArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*removeMembersArgs)
		validate.EntityCount(len(args.Members), ctx.MaxProcessEntityCount())
		panic.IfTrueWith(args.Account.Equal(args.Me), err.InvalidOperation)

		accountRole := db.GetAccountRole(ctx, args.Shard, args.Account, args.Me)
		panic.IfTrueWith(accountRole == nil, err.InsufficientPermission)

		switch *accountRole {
		case cnst.AccountOwner:
			totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.Account)
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.Account, args.Members)
			panic.IfTrueWith(totalOwnerCount == ownerCountInRemoveSet, zeroOwnerCountErr)
		case cnst.AccountAdmin:
			ownerCountInRemoveSet := dbGetOwnerCountInSet(ctx, args.Shard, args.Account, args.Members)
			panic.IfTrueWith(ownerCountInRemoveSet > 0, err.InsufficientPermission)
		default:
			panic.IfTrueWith(len(args.Members) != 1 || !args.Members[0].Equal(args.Me), err.InsufficientPermission)
		}

		dbSetMembersInactive(ctx, args.Shard, args.Account, args.Members)
		dbLogAccountBatchAddOrRemoveMembersActivity(ctx, args.Shard, args.Account, args.Me, args.Members, "removed")
		return nil
	},
}

type memberIsOnlyAccountOwnerArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Me      id.Id `json:"me"`
}

var memberIsOnlyAccountOwner = &endpoint.Endpoint{
	Method:    http.MethodGet,
	Path:      "/api/v1/private/memberIsOnlyAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsOnlyAccountOwnerArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*memberIsOnlyAccountOwnerArgs)
		if args.Account.Equal(args.Me) {
			return true
		}
		totalOwnerCount := dbGetTotalOwnerCount(ctx, args.Shard, args.Account)
		ownerCount := dbGetOwnerCountInSet(ctx, args.Shard, args.Account, []id.Id{args.Me})
		return totalOwnerCount == 1 && ownerCount == 1
	},
}

type setMemberNameArgs struct {
	Shard   int    `json:"shard"`
	Account id.Id  `json:"account"`
	Me      id.Id  `json:"me"`
	NewName string `json:"newName"`
}

var setMemberName = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/setMemberName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberNameArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberNameArgs)
		dbSetMemberName(ctx, args.Shard, args.Account, args.Me, args.NewName)
		return nil
	},
}

type setMemberDisplayNameArgs struct {
	Shard          int     `json:"shard"`
	Account        id.Id   `json:"account"`
	Me             id.Id   `json:"me"`
	NewDisplayName *string `json:"newDisplayName"`
}

var setMemberDisplayName = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/setMemberDisplayName",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberDisplayNameArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberDisplayNameArgs)
		dbSetMemberDisplayName(ctx, args.Shard, args.Account, args.Me, args.NewDisplayName)
		return nil
	},
}

type setMemberHasAvatarArgs struct {
	Shard     int   `json:"shard"`
	Account   id.Id `json:"account"`
	Me        id.Id `json:"me"`
	HasAvatar bool  `json:"hasAvatar"`
}

var setMemberHasAvatar = &endpoint.Endpoint{
	Method:    http.MethodPost,
	Path:      "/api/v1/private/setMemberHasAvatar",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &setMemberHasAvatarArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*setMemberHasAvatarArgs)
		dbSetMemberHasAvatar(ctx, args.Shard, args.Account, args.Me, args.HasAvatar)
		return nil
	},
}

type memberIsAccountOwnerArgs struct {
	Shard   int   `json:"shard"`
	Account id.Id `json:"account"`
	Me      id.Id `json:"me"`
}

var memberIsAccountOwner = &endpoint.Endpoint{
	Method:    http.MethodGet,
	Path:      "/api/v1/private/memberIsAccountOwner",
	IsPrivate: true,
	GetArgsStruct: func() interface{} {
		return &memberIsAccountOwnerArgs{}
	},
	CtxHandler: func(ctx ctx.Ctx, a interface{}) interface{} {
		args := a.(*memberIsAccountOwnerArgs)
		if !args.Me.Equal(args.Account) {
			accountRole := db.GetAccountRole(ctx, args.Shard, args.Account, args.Me)
			if accountRole != nil && *accountRole == cnst.AccountOwner {
				return true
			} else {
				return false
			}
		}
		return true
	},
}

var Endpoints = []*endpoint.Endpoint{
	createAccount,
	deleteAccount,
	addMembers,
	removeMembers,
	memberIsOnlyAccountOwner,
	setMemberName,
	setMemberDisplayName,
	setMemberHasAvatar,
	memberIsAccountOwner,
}
