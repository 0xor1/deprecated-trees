package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
)

var (
	invalidRegionErr          = &Error{Code: 19, Msg: "invalid region"}
	insufficientPermissionErr = &Error{Code: 20, Msg: "insufficient permission"}
	zeroOwnerCountErr         = &Error{Code: 21, Msg: "zero owner count"}
	invalidTaskCenterTypeErr  = &Error{Code: 22, Msg: "invalid task center type"}
)

func newInternalApiClient(regions map[string]InternalApi) InternalApiClient {
	if regions == nil {
		NilCriticalParamPanic("regions")
	}
	return &internalApiClient{
		regions: regions,
	}
}

type internalApiClient struct {
	regions map[string]InternalApi
}

func (a *internalApiClient) GetRegions() []string {
	regions := make([]string, 0, len(a.regions))
	for k := range a.regions {
		regions = append(regions, k)
	}
	return regions
}

func (a *internalApiClient) IsValidRegion(region string) bool {
	return a.regions[region] != nil
}

func (a *internalApiClient) CreatePersonalTaskCenter(region string, user Id) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].CreatePersonalTaskCenter(user)
}

func (a *internalApiClient) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].CreateOrgTaskCenter(org, owner, ownerName)
}

func (a *internalApiClient) DeleteTaskCenter(region string, shard int, account, owner Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].DeleteTaskCenter(shard, account, owner)
}

func (a *internalApiClient) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].AddMembers(shard, org, admin, members)
}

func (a *internalApiClient) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].RemoveMembers(shard, org, admin, members)
}

func (a *internalApiClient) SetMemberDeleted(region string, shard int, org, member Id) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].SetMemberDeleted(shard, org, member)
}

func (a *internalApiClient) MemberIsOnlyOwner(region string, shard int, org, member Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}

	return a.regions[region].MemberIsOnlyOwner(shard, org, member)
}

func (a *internalApiClient) RenameMember(region string, shard int, org, member Id, newName string) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].RenameMember(shard, org, member, newName)
}

func (a *internalApiClient) UserIsOrgOwner(region string, shard int, org, user Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
	return a.regions[region].UserIsOrgOwner(shard, org, user)
}

func newInternalApi(store store, log Log) InternalApi {
	if store == nil {
		NilCriticalParamPanic("store")
	}
	if log == nil {
		NilCriticalParamPanic("log")
	}
	return &internalApi{
		store: store,
		log:   log,
	}
}

type internalApi struct {
	store store
	log   Log
}

func (a *internalApi) CreatePersonalTaskCenter(user Id) (int, error) {
	a.log.Location()

	shard, err := a.store.registerPersonalAccount(user)

	return shard, a.log.InfoErr(err)
}

func (a *internalApi) CreateOrgTaskCenter(org, owner Id, ownerName string) (int, error) {
	a.log.Location()

	shard, err := a.store.registerOrgAccount(org, owner, ownerName)

	return shard, a.log.InfoErr(err)
}

func (a *internalApi) DeleteTaskCenter(shard int, account, owner Id) (error, error) {
	a.log.Location()

	if !owner.Equal(account) {
		member, err := a.store.getMember(shard, account, owner)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}
		if member == nil || member.Role != Owner {
			return insufficientPermissionErr, nil
		}
	}
	return nil, a.log.InfoErr(a.store.deleteAccount(shard, account))
}

func (a *internalApi) AddMembers(shard int, org, admin Id, members []*NamedEntity) (error, error) {
	a.log.Location()

	member, err := a.store.getMember(shard, org, admin)
	if err != nil {
		return nil, a.log.InfoErr(err)
	}
	if member == nil || (member.Role != Owner && member.Role != Admin) {
		return insufficientPermissionErr, nil
	}

	pastMembers := make([]*NamedEntity, 0, len(members))
	newMembers := make([]*NamedEntity, 0, len(members))
	for _, mem := range members {
		existingMember, err := a.store.getMember(shard, org, mem.Id)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}
		if existingMember == nil {
			newMembers = append(newMembers, mem)
		} else if !existingMember.IsActive {
			pastMembers = append(pastMembers, mem)
		}
	}

	if len(newMembers) > 0 {
		if err := a.store.addMembers(shard, org, newMembers); err != nil {
			return nil, a.log.InfoErr(err)
		}
	}
	if len(pastMembers) > 0 {
		if err := a.store.setMembersActive(shard, org, pastMembers); err != nil { //has to be NamedEntity in case the user changed their name whilst they were inactive on the org
			return nil, a.log.InfoErr(err)
		}
	}

	return nil, nil
}

func (a *internalApi) RemoveMembers(shard int, org, admin Id, members []Id) (error, error) {
	a.log.Location()

	member, err := a.store.getMember(shard, org, admin)
	if err != nil {
		return nil, a.log.InfoErr(err)
	}

	switch member.Role {
	case Owner:
		totalOrgOwnerCount, err := a.store.getTotalOrgOwnerCount(shard, org)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}

		ownerCountInRemoveSet, err := a.store.getOwnerCountInSet(shard, org, members)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}

		if totalOrgOwnerCount == ownerCountInRemoveSet {
			return zeroOwnerCountErr, nil
		}

	case Admin:
		ownerCountInRemoveSet, err := a.store.getOwnerCountInSet(shard, org, members)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}
		if ownerCountInRemoveSet > 0 {
			return insufficientPermissionErr, nil
		}

	default:
		return insufficientPermissionErr, nil
	}

	return nil, a.log.InfoErr(a.store.setMembersInactive(shard, org, members))
}

func (a *internalApi) SetMemberDeleted(shard int, org, member Id) error {
	a.log.Location()

	return a.log.InfoErr(a.store.setMemberDeleted(shard, org, member))
}

func (a *internalApi) MemberIsOnlyOwner(shard int, org, member Id) (bool, error) {
	a.log.Location()

	is, err := a.store.memberIsOnlyOwner(shard, org, member)
	return is, a.log.InfoErr(err)
}

func (a *internalApi) RenameMember(shard int, org, member Id, newName string) error {
	a.log.Location()

	return a.log.InfoErr(a.store.renameMember(shard, org, member, newName))
}

func (a *internalApi) UserIsOrgOwner(shard int, org, user Id) (bool, error) {
	a.log.Location()

	if !user.Equal(org) {
		member, err := a.store.getMember(shard, org, user)
		if err != nil {
			return false, a.log.InfoErr(err)
		}
		if member != nil && member.Role == Owner {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, invalidTaskCenterTypeErr
}

type store interface {
	registerPersonalAccount(id Id) (int, error)
	registerOrgAccount(id, ownerId Id, ownerName string) (int, error)
	deleteAccount(shard int, account Id) error
	getMember(shard int, org, member Id) (*orgMember, error)
	addMembers(shard int, org Id, members []*NamedEntity) error
	setMembersActive(shard int, org Id, members []*NamedEntity) error
	getTotalOrgOwnerCount(shard int, org Id) (int, error)
	getOwnerCountInSet(shard int, org Id, members []Id) (int, error)
	setMembersInactive(shard int, org Id, members []Id) error
	memberIsOnlyOwner(shard int, org, member Id) (bool, error)
	setMemberDeleted(shard int, org Id, member Id) error
	renameMember(shard int, org Id, member Id, newName string) error
}

type task struct {
	NamedEntity
	Org                Id     `json:"org"`
	User               Id     `json:"user"`
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
	ChatCount          uint64 `json:"chatCount"`
	FileCount          uint64 `json:"fileCount"`
	FileSize           uint64 `json:"fileSize"`
	IsAbstractTask     bool   `json:"isAbstractTask"`
}

type abstractTask struct {
	task
	MinimumRemainingTime    uint64 `json:"minimumRemainingTime"`
	IsParallel              bool   `json:"isParallel"`
	ChildCount              uint64 `json:"childCount"`
	DescendantCount         uint64 `json:"descendantCount"`
	LeafCount               uint64 `json:"leafCount"`
	SubFileCount            uint64 `json:"subFileCount"`
	SubFileSize             uint64 `json:"subFileSize"`
	ArchivedChildCount      uint64 `json:"archivedChildCount"`
	ArchivedDescendantCount uint64 `json:"archivedDescendantCount"`
	ArchivedLeafCount       uint64 `json:"archivedLeafCount"`
	ArchivedSubFileCount    uint64 `json:"archivedSubFileCount"`
	ArchivedSubFileSize     uint64 `json:"archivedSubFileSize"`
}

type orgMember struct {
	NamedEntity
	Org 		   Id	  `json:"org"`
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
	IsActive           bool   `json:"isActive"`
	IsDeleted          bool   `json:"isDeleted"`
	Role               Role   `json:"role"`
}
