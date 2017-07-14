package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

var (
	invalidRegionErr         = &Error{Code: 18, Msg: "invalid region", IsPublic: false}
	zeroOwnerCountErr        = &Error{Code: 19, Msg: "zero owner count", IsPublic: true}
	invalidTaskCenterTypeErr = &Error{Code: 20, Msg: "invalid task center type", IsPublic: true}
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

func (c *client) CreateTaskCenter(region string, org, owner Id, ownerName string) int {
	return c.getRegion(region).CreateTaskCenter(org, owner, ownerName)
}

func (c *client) DeleteTaskCenter(region string, shard int, account, owner Id) {
	c.getRegion(region).DeleteTaskCenter(shard, account, owner)
}

func (c *client) AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal) {
	c.getRegion(region).AddMembers(shard, org, admin, members)
}

func (c *client) RemoveMembers(region string, shard int, org, admin Id, members []Id) {
	c.getRegion(region).RemoveMembers(shard, org, admin, members)
}

func (c *client) MemberIsOnlyOwner(region string, shard int, org, member Id) bool {
	return c.getRegion(region).MemberIsOnlyOwner(shard, org, member)
}

func (c *client) RenameMember(region string, shard int, org, member Id, newName string) {
	c.getRegion(region).RenameMember(shard, org, member, newName)
}

func (c *client) UserIsOrgOwner(region string, shard int, org, user Id) bool {
	return c.getRegion(region).UserIsOrgOwner(shard, org, user)
}

type api struct {
	store store
}

func (a *api) CreateTaskCenter(org, owner Id, ownerName string) int {
	shard := a.store.registerAccount(org, owner, ownerName)
	a.store.logActivity(shard, org, Now(), org, owner, "org", "created")
	return shard
}

func (a *api) DeleteTaskCenter(shard int, account, owner Id) {
	if !owner.Equal(account) {
		member := a.store.getMember(shard, account, owner)
		if member == nil || member.Role != OrgOwner {
			panic(InsufficientPermissionErr)
		}
	}
	//TODO delete s3 data, uploaded files etc
	a.store.deleteAccount(shard, account)
}

func (a *api) AddMembers(shard int, org, actorId Id, members []*AddMemberInternal) {
	actor := a.store.getMember(shard, org, actorId)
	if actor == nil || (actor.Role != OrgOwner && actor.Role != OrgAdmin) {
		panic(InsufficientPermissionErr) //only owners and admins can add new org members
	}

	pastMembers := make([]*AddMemberInternal, 0, len(members))
	newMembers := make([]*AddMemberInternal, 0, len(members))
	for _, mem := range members {
		if mem.Role == OrgOwner && actor.Role != OrgOwner {
			panic(InsufficientPermissionErr) //only owners can add owners
		}
		existingMember := a.store.getMember(shard, org, mem.Id)
		if existingMember == nil {
			newMembers = append(newMembers, mem)
		} else if !existingMember.IsActive || existingMember.Role != mem.Role {
			pastMembers = append(pastMembers, mem)
		}
	}

	if len(newMembers) > 0 {
		a.store.addMembers(shard, org, newMembers)
		for _, mem := range newMembers {
			a.store.logActivity(shard, org, Now(), mem.Id, actorId, "member", "added")
		}
	}
	if len(pastMembers) > 0 {
		a.store.updateMembersAndSetActive(shard, org, pastMembers) //has to be AddMemberInternal in case the user changed their name whilst they were inactive on the org, or if they were
		for _, mem := range pastMembers {
			a.store.logActivity(shard, org, Now(), mem.Id, actorId, "member", "added")
		}
	}
}

func (a *api) RemoveMembers(shard int, org, admin Id, members []Id) {
	member := a.store.getMember(shard, org, admin)

	switch member.Role {
	case OrgOwner:
		totalOrgOwnerCount := a.store.getTotalOrgOwnerCount(shard, org)
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, org, members)
		if totalOrgOwnerCount == ownerCountInRemoveSet {
			panic(zeroOwnerCountErr)
		}

	case OrgAdmin:
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, org, members)
		if ownerCountInRemoveSet > 0 {
			panic(InsufficientPermissionErr)
		}
	default:
		if len(members) != 1 || !members[0].Equal(admin) { //user can remove themselves
			panic(InsufficientPermissionErr)
		}
	}

	a.store.setMembersInactive(shard, org, members)
	for _, mem := range members {
		a.store.logActivity(shard, org, Now(), mem, admin, "member", "removed")
	}
}

func (a *api) MemberIsOnlyOwner(shard int, org, member Id) bool {
	totalOrgOwnerCount := a.store.getTotalOrgOwnerCount(shard, org)
	ownerCount := a.store.getOwnerCountInSet(shard, org, []Id{member})
	return totalOrgOwnerCount == 1 && ownerCount == 1
}

func (a *api) RenameMember(shard int, org, member Id, newName string) {
	a.store.renameMember(shard, org, member, newName)
}

func (a *api) UserIsOrgOwner(shard int, org, user Id) bool {
	if !user.Equal(org) {
		member := a.store.getMember(shard, org, user)
		if member != nil && member.Role == OrgOwner {
			return true
		} else {
			return false
		}
	}
	panic(invalidTaskCenterTypeErr)
}

type store interface {
	registerAccount(id, ownerId Id, ownerName string) int
	deleteAccount(shard int, account Id)
	getMember(shard int, org, member Id) *Member
	addMembers(shard int, org Id, members []*AddMemberInternal)
	updateMembersAndSetActive(shard int, org Id, members []*AddMemberInternal)
	getTotalOrgOwnerCount(shard int, org Id) int
	getOwnerCountInSet(shard int, org Id, members []Id) int
	setMembersInactive(shard int, org Id, members []Id)
	renameMember(shard int, org Id, member Id, newName string)
	logActivity(shard int, org Id, occurredOn time.Time, item, member Id, itemType, action string)
}
