package private

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

var (
	invalidRegionErr  = &Error{Code: "rc_v1_i_ir", Msg: "invalid region", IsPublic: false}
	zeroOwnerCountErr = &Error{Code: "rc_v1_i_zoc", Msg: "zero owner count", IsPublic: true}
)

type client struct {
	regions map[string]Api
}

func (c *client) GetRegions() []string {
	regions := make([]string, 0, len(c.regions))
	for k := range c.regions {
		regions = append(regions, k)
	}
	return regions
}

func (c *client) getRegion(region string) Api {
	if !c.IsValidRegion(region) {
		panic(invalidRegionErr)
	}
	return c.regions[region]
}

func (c *client) IsValidRegion(region string) bool {
	return c.regions[region] != nil
}

func (c *client) CreateAccount(region string, accountId, myId Id, myName string) int {
	return c.getRegion(region).CreateAccount(accountId, myId, myName)
}

func (c *client) DeleteAccount(region string, shard int, accountId, myId Id) {
	c.getRegion(region).DeleteAccount(shard, accountId, myId)
}

func (c *client) AddMembers(region string, shard int, accountId, myId Id, members []*AddMemberPrivate) {
	c.getRegion(region).AddMembers(shard, accountId, myId, members)
}

func (c *client) RemoveMembers(region string, shard int, accountId, myId Id, members []Id) {
	c.getRegion(region).RemoveMembers(shard, accountId, myId, members)
}

func (c *client) MemberIsOnlyAccountOwner(region string, shard int, accountId, myId Id) bool {
	return c.getRegion(region).MemberIsOnlyAccountOwner(shard, accountId, myId)
}

func (c *client) RenameMember(region string, shard int, accountId, myId Id, newName string) {
	c.getRegion(region).RenameMember(shard, accountId, myId, newName)
}

func (c *client) MemberIsAccountOwner(region string, shard int, accountId, myId Id) bool {
	return c.getRegion(region).MemberIsAccountOwner(shard, accountId, myId)
}

type api struct {
	store                 store
	maxProcessEntityCount int
}

func (a *api) CreateAccount(accountId, myId Id, myName string) int {
	shard := a.store.createAccount(accountId, myId, myName)
	a.store.logActivity(shard, accountId, myId, accountId, "account", "created")
	return shard
}

func (a *api) DeleteAccount(shard int, accountId, myId Id) {
	if !myId.Equal(accountId) {
		ValidateMemberHasAccountOwnerAccess(a.store.getAccountRole(shard, accountId, myId))
	}
	a.store.deleteAccount(shard, accountId)
	//TODO delete s3 data, uploaded files etc
}

func (a *api) AddMembers(shard int, accountId, myId Id, members []*AddMemberPrivate) {
	if len(members) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}
	accountRole := a.store.getAccountRole(shard, accountId, myId)
	ValidateMemberHasAccountAdminAccess(accountRole)

	allIds := make([]Id, 0, len(members))
	newMembersMap := map[string]*AddMemberPrivate{}
	for _, mem := range members { //loop over all the new entries and check permissions and build up useful id map and allIds slice
		mem.Role.Validate()
		if mem.Role == AccountOwner {
			ValidateMemberHasAccountOwnerAccess(accountRole)
		}
		newMembersMap[mem.Id.String()] = mem
		allIds = append(allIds, mem.Id)
	}

	inactiveMemberIds := a.store.getAllInactiveMemberIdsFromInputSet(shard, accountId, allIds)
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
		a.store.addMembers(shard, accountId, newMembers)
	}
	if len(inactiveMembers) > 0 {
		a.store.updateMembersAndSetActive(shard, accountId, inactiveMembers) //has to be AddMemberPrivate in case the member changed their name whilst they were inactive on the account
	}
	a.store.logAccountBatchAddOrRemoveMembersActivity(shard, accountId, myId, allIds, "added")
}

func (a *api) RemoveMembers(shard int, accountId, myId Id, members []Id) {
	if len(members) > a.maxProcessEntityCount {
		panic(MaxEntityCountExceededErr)
	}
	if accountId.Equal(myId) {
		panic(InvalidOperationErr)
	}

	accountRole := a.store.getAccountRole(shard, accountId, myId)
	if accountRole == nil {
		panic(InsufficientPermissionErr)
	}

	switch *accountRole {
	case AccountOwner:
		totalOwnerCount := a.store.getTotalOwnerCount(shard, accountId)
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, accountId, members)
		if totalOwnerCount == ownerCountInRemoveSet {
			panic(zeroOwnerCountErr)
		}

	case AccountAdmin:
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, accountId, members)
		if ownerCountInRemoveSet > 0 {
			panic(InsufficientPermissionErr)
		}
	default:
		if len(members) != 1 || !members[0].Equal(myId) { //any member can remove themselves
			panic(InsufficientPermissionErr)
		}
	}

	a.store.setMembersInactive(shard, accountId, members)
	a.store.logAccountBatchAddOrRemoveMembersActivity(shard, accountId, myId, members, "removed")
}

func (a *api) MemberIsOnlyAccountOwner(shard int, accountId, myId Id) bool {
	if accountId.Equal(myId) {
		return true
	}
	totalOwnerCount := a.store.getTotalOwnerCount(shard, accountId)
	ownerCount := a.store.getOwnerCountInSet(shard, accountId, []Id{myId})
	return totalOwnerCount == 1 && ownerCount == 1
}

func (a *api) RenameMember(shard int, accountId, myId Id, newName string) {
	a.store.renameMember(shard, accountId, myId, newName)
}

func (a *api) MemberIsAccountOwner(shard int, accountId, myId Id) bool {
	if !myId.Equal(accountId) {
		accountRole := a.store.getAccountRole(shard, accountId, myId)
		if accountRole != nil && *accountRole == AccountOwner {
			return true
		} else {
			return false
		}
	}
	return true
}

type store interface {
	createAccount(id, ownerId Id, ownerName string) int
	deleteAccount(shard int, account Id)
	getAllInactiveMemberIdsFromInputSet(shard int, accountId Id, members []Id) []Id
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	addMembers(shard int, accountId Id, members []*AddMemberPrivate)
	updateMembersAndSetActive(shard int, accountId Id, members []*AddMemberPrivate)
	getTotalOwnerCount(shard int, accountId Id) int
	getOwnerCountInSet(shard int, accountId Id, members []Id) int
	setMembersInactive(shard int, accountId Id, members []Id)
	renameMember(shard int, accountId Id, member Id, newName string)
	logActivity(shard int, accountId Id, member, item Id, itemType, action string)
	logAccountBatchAddOrRemoveMembersActivity(shard int, accountId, member Id, members []Id, action string)
}
