package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

var (
	invalidRegionErr          = &Error{Code: 19, Msg: "invalid region", IsPublic: false}
	insufficientPermissionErr = &Error{Code: 20, Msg: "insufficient permission", IsPublic: true}
	zeroOwnerCountErr         = &Error{Code: 21, Msg: "zero owner count", IsPublic: true}
	invalidTaskCenterTypeErr  = &Error{Code: 22, Msg: "invalid task center type", IsPublic: true}
)

func newInternalRegionClient(regions map[string]Api) InternalRegionClient {
	if regions == nil {
		panic(NilCriticalParamErr)
	}
	return &internalRegionClient{
		regions: regions,
	}
}

type internalRegionClient struct {
	regions map[string]Api
}

func (c *internalRegionClient) GetRegions() []string {
	regions := make([]string, 0, len(c.regions))
	for k := range c.regions {
		regions = append(regions, k)
	}
	return regions
}

func (c *internalRegionClient) getRegion(region string) Api {
	if !c.IsValidRegion(region) {
		panic(invalidRegionErr)
	}
	return c.getRegion(region)
}

func (c *internalRegionClient) IsValidRegion(region string) bool {
	return c.getRegion(region) != nil
}

func (c *internalRegionClient) CreatePersonalTaskCenter(region string, user Id) int {
	return c.getRegion(region).CreatePersonalTaskCenter(user)
}

func (c *internalRegionClient) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) int {
	return c.getRegion(region).CreateOrgTaskCenter(org, owner, ownerName)
}

func (c *internalRegionClient) DeleteTaskCenter(region string, shard int, account, owner Id) {
	c.getRegion(region).DeleteTaskCenter(shard, account, owner)
}

func (c *internalRegionClient) AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal) {
	c.getRegion(region).AddMembers(shard, org, admin, members)
}

func (c *internalRegionClient) RemoveMembers(region string, shard int, org, admin Id, members []Id) {
	c.getRegion(region).RemoveMembers(shard, org, admin, members)
}

func (c *internalRegionClient) MemberIsOnlyOwner(region string, shard int, org, member Id) bool {
	return c.getRegion(region).MemberIsOnlyOwner(shard, org, member)
}

func (c *internalRegionClient) RenameMember(region string, shard int, org, member Id, newName string) {
	c.getRegion(region).RenameMember(shard, org, member, newName)
}

func (c *internalRegionClient) UserIsOrgOwner(region string, shard int, org, user Id) bool {
	return c.getRegion(region).UserIsOrgOwner(shard, org, user)
}

func newInternalApi(store store) Api {
	if store == nil {
		panic(NilCriticalParamErr)
	}
	return &internalApi{
		store: store,
	}
}

type internalApi struct {
	store store
}

func (a *internalApi) CreatePersonalTaskCenter(user Id) int {
	return a.store.registerPersonalAccount(user)
}

func (a *internalApi) CreateOrgTaskCenter(org, owner Id, ownerName string) int {
	return a.store.registerOrgAccount(org, owner, ownerName)
}

func (a *internalApi) DeleteTaskCenter(shard int, account, owner Id) {
	if !owner.Equal(account) {
		member := a.store.getMember(shard, account, owner)
		if member == nil || member.Role != Owner {
			panic(insufficientPermissionErr)
		}
	}
}

func (a *internalApi) AddMembers(shard int, org, actorId Id, members []*AddMemberInternal) {
	actor := a.store.getMember(shard, org, actorId)
	if actor == nil || (actor.Role != Owner && actor.Role != Admin) {
		panic(insufficientPermissionErr) //only owners and admins can add new org members
	}

	pastMembers := make([]*AddMemberInternal, 0, len(members))
	newMembers := make([]*AddMemberInternal, 0, len(members))
	for _, mem := range members {
		if mem.Role != Owner && mem.Role != Admin {
			mem.Role = Reader //at org level users are either owners, admins or nothing, we use reader to signify this as the lowest permission level
		}
		if mem.Role == Owner && actor.Role != Owner {
			panic(insufficientPermissionErr) //only owners can add owners
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
	}
	if len(pastMembers) > 0 {
		a.store.updateMembers(shard, org, pastMembers) //has to be AddMemberInternal in case the user changed their name whilst they were inactive on the org, or if they were
	}
}

func (a *internalApi) RemoveMembers(shard int, org, admin Id, members []Id) {
	member := a.store.getMember(shard, org, admin)

	switch member.Role {
	case Owner:
		totalOrgOwnerCount := a.store.getTotalOrgOwnerCount(shard, org)
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, org, members)
		if totalOrgOwnerCount == ownerCountInRemoveSet {
			panic(zeroOwnerCountErr)
		}

	case Admin:
		ownerCountInRemoveSet := a.store.getOwnerCountInSet(shard, org, members)
		if ownerCountInRemoveSet > 0 {
			panic(insufficientPermissionErr)
		}

	default:
		panic(insufficientPermissionErr)
	}

	a.store.setMembersInactive(shard, org, members)
}

func (a *internalApi) MemberIsOnlyOwner(shard int, org, member Id) bool {
	totalOrgOwnerCount := a.store.getTotalOrgOwnerCount(shard, org)
	ownerCount := a.store.getOwnerCountInSet(shard, org, []Id{member})
	return totalOrgOwnerCount == 1 && ownerCount == 1
}

func (a *internalApi) RenameMember(shard int, org, member Id, newName string) {
	a.store.renameMember(shard, org, member, newName)
}

func (a *internalApi) UserIsOrgOwner(shard int, org, user Id) bool {
	if !user.Equal(org) {
		member := a.store.getMember(shard, org, user)
		if member != nil && member.Role == Owner {
			return true
		} else {
			return false
		}
	}
	panic(invalidTaskCenterTypeErr)
}

type store interface {
	registerPersonalAccount(id Id) int
	registerOrgAccount(id, ownerId Id, ownerName string) int
	deleteAccount(shard int, account Id)
	getMember(shard int, org, member Id) *orgMember
	addMembers(shard int, org Id, members []*AddMemberInternal)
	updateMembers(shard int, org Id, members []*AddMemberInternal)
	getTotalOrgOwnerCount(shard int, org Id) int
	getOwnerCountInSet(shard int, org Id, members []Id) int
	setMembersInactive(shard int, org Id, members []Id)
	renameMember(shard int, org Id, member Id, newName string)
}

//type task struct {
//	NamedEntity
//	Org                Id     `json:"org"`
//	User               Id     `json:"user"`
//	TotalRemainingTime uint64 `json:"totalRemainingTime"`
//	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
//	ChatCount          uint64 `json:"chatCount"`
//	FileCount          uint64 `json:"fileCount"`
//	FileSize           uint64 `json:"fileSize"`
//	IsAbstractTask     bool   `json:"isAbstractTask"`
//}
//
//type abstractTask struct {
//	task
//	MinimumRemainingTime    uint64 `json:"minimumRemainingTime"`
//	IsParallel              bool   `json:"isParallel"`
//	ChildCount              uint64 `json:"childCount"`
//	DescendantCount         uint64 `json:"descendantCount"`
//	LeafCount               uint64 `json:"leafCount"`
//	SubFileCount            uint64 `json:"subFileCount"`
//	SubFileSize             uint64 `json:"subFileSize"`
//	ArchivedChildCount      uint64 `json:"archivedChildCount"`
//	ArchivedDescendantCount uint64 `json:"archivedDescendantCount"`
//	ArchivedLeafCount       uint64 `json:"archivedLeafCount"`
//	ArchivedSubFileCount    uint64 `json:"archivedSubFileCount"`
//	ArchivedSubFileSize     uint64 `json:"archivedSubFileSize"`
//}

type orgMember struct {
	AddMemberInternal
	Org                Id     `json:"org"`
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
	IsActive           bool   `json:"isActive"`
}
