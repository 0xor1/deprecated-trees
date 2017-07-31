package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

var (
	invalidRegionErr         = &Error{Code: "rc_v1_i_ir", Msg: "invalid region", IsPublic: false}
	zeroOwnerCountErr        = &Error{Code: "rc_v1_i_zoc", Msg: "zero owner count", IsPublic: true}
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
	return c.getRegion(region)
}

func (c *client) IsValidRegion(region string) bool {
	return c.getRegion(region) != nil
}

func (c *client) CreateAccount(region string, accountId, myId Id, ownerName string) int {
	return c.getRegion(region).CreateAccount(accountId, myId, ownerName)
}

func (c *client) DeleteAccount(region string, shard int, accountId, myId Id) {
	c.getRegion(region).DeleteAccount(shard, accountId, myId)
}

func (c *client) AddMembers(region string, shard int, accountId, myId Id, members []*AddMemberInternal) {
	c.getRegion(region).AddMembers(shard, accountId, myId, members)
}

func (c *client) RemoveMembers(region string, shard int, accountId, myId Id, members []Id) {
	c.getRegion(region).RemoveMembers(shard, accountId, myId, members)
}

func (c *client) MemberIsOnlyAccountOwner(region string, shard int, accountId, member Id) bool {
	return c.getRegion(region).MemberIsOnlyAccountOwner(shard, accountId, member)
}

func (c *client) RenameMember(region string, shard int, accountId, member Id, newName string) {
	c.getRegion(region).RenameMember(shard, accountId, member, newName)
}

func (c *client) MemberIsAccountOwner(region string, shard int, accountId, myId Id) bool {
	return c.getRegion(region).MemberIsAccountOwner(shard, accountId, myId)
}

type api struct {
	store store
}

func (a *api) CreateAccount(accountId, ownerId Id, ownerName string) int {
	shard := a.store.createAccount(accountId, ownerId, ownerName)
	a.store.logActivity(shard, accountId, ownerId, accountId, "account", "created")
	return shard
}

func (a *api) DeleteAccount(shard int, accountId, ownerId Id) {
	if !ownerId.Equal(accountId) {
		ValidateMemberHasAccountOwnerAccess(a.store.getAccountRole(shard, accountId, ownerId))
	}
	//TODO delete s3 data, uploaded files etc
	a.store.deleteAccount(shard, accountId)
}

func (a *api) AddMembers(shard int, accountId, actorId Id, members []*AddMemberInternal) {
	if accountId.Equal(actorId) {
		panic(InvalidOperationErr)
	}
	accountRole := a.store.getAccountRole(shard, accountId, actorId)
	ValidateMemberHasAccountAdminAccess(accountRole)

	pastMembers := make([]*AddMemberInternal, 0, len(members))
	newMembers := make([]*AddMemberInternal, 0, len(members))
	for _, mem := range members {
		mem.Role.Validate()
		if mem.Role == AccountOwner {
			ValidateMemberHasAccountOwnerAccess(accountRole)
		}
		existingMember := a.store.getMember(shard, accountId, mem.Id)
		if existingMember == nil {
			newMembers = append(newMembers, mem)
		} else if !existingMember.IsActive || existingMember.Role != mem.Role {
			pastMembers = append(pastMembers, mem)
		}
	}

	if len(newMembers) > 0 {
		a.store.addMembers(shard, accountId, newMembers)
		for _, mem := range newMembers {
			a.store.logActivity(shard, accountId, actorId, mem.Id, "member", "added")
		}
	}
	if len(pastMembers) > 0 {
		a.store.updateMembersAndSetActive(shard, accountId, pastMembers) //has to be AddMemberInternal in case the member changed their name whilst they were inactive on the account
		for _, mem := range pastMembers {
			a.store.logActivity(shard, accountId, actorId, mem.Id, "member", "added")
		}
	}
}

func (a *api) RemoveMembers(shard int, accountId, admin Id, members []Id) {
	if accountId.Equal(admin) {
		panic(InvalidOperationErr)
	}
	accountRole := a.store.getAccountRole(shard, accountId, admin)

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
		if len(members) != 1 || !members[0].Equal(admin) { //member can remove themselves
			panic(InsufficientPermissionErr)
		}
	}

	a.store.setMembersInactive(shard, accountId, members)
	for _, mem := range members {
		a.store.logActivity(shard, accountId, admin, mem, "member", "removed")
	}
}

func (a *api) MemberIsOnlyAccountOwner(shard int, accountId, member Id) bool {
	if accountId.Equal(member) {
		return true
	}
	totalOwnerCount := a.store.getTotalOwnerCount(shard, accountId)
	ownerCount := a.store.getOwnerCountInSet(shard, accountId, []Id{member})
	return totalOwnerCount == 1 && ownerCount == 1
}

func (a *api) RenameMember(shard int, accountId, memberId Id, newName string) {
	a.store.renameMember(shard, accountId, memberId, newName)
}

func (a *api) MemberIsAccountOwner(shard int, accountId, myId Id) bool {
	if !myId.Equal(accountId) {
		member := a.store.getMember(shard, accountId, myId)
		if member != nil && member.Role == AccountOwner {
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
	getMember(shard int, accountId, memberId Id) *AccountMember
	getAccountRole(shard int, accountId, memberId Id) *AccountRole
	addMembers(shard int, accountId Id, members []*AddMemberInternal)
	updateMembersAndSetActive(shard int, accountId Id, members []*AddMemberInternal)
	getTotalOwnerCount(shard int, accountId Id) int
	getOwnerCountInSet(shard int, accountId Id, members []Id) int
	setMembersInactive(shard int, accountId Id, members []Id)
	renameMember(shard int, accountId Id, member Id, newName string)
	logActivity(shard int, accountId Id, member, item Id, itemType, action string)
}
