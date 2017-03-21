package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
	"time"
)

var (
	invalidRegionErr          = &Error{Code: 19, Msg: "invalid region"}
	insufficientPermissionErr = &Error{Code: 20, Msg: "insufficient permission"}
	zeroOwnerCountErr         = &Error{Code: 21, Msg: "zero owner count"}
	invalidTaskCenterTypeErr  = &Error{Code: 22, Msg: "invalid task center type"}
)

func newInternalApi(regions map[string]SingularInternalApi) InternalApi {
	if regions == nil {
		NilCriticalParamPanic("regions")
	}
	return &internalApi{
		regions: regions,
	}
}

type internalApi struct {
	regions map[string]SingularInternalApi
}

func (a *internalApi) GetRegions() []string {
	regions := make([]string, 0, len(a.regions))
	for k := range a.regions {
		regions = append(regions, k)
	}
	return regions
}

func (a *internalApi) IsValidRegion(region string) bool {
	return a.regions[region] != nil
}

func (a *internalApi) CreatePersonalTaskCenter(region string, user Id) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].CreatePersonalTaskCenter(user)
}

func (a *internalApi) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].CreateOrgTaskCenter(org, owner, ownerName)
}

func (a *internalApi) DeleteTaskCenter(region string, shard int, account, owner Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].DeleteTaskCenter(shard, account, owner)
}

func (a *internalApi) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].AddMembers(shard, org, admin, members)
}

func (a *internalApi) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].RemoveMembers(shard, org, admin, members)
}

func (a *internalApi) SetMemberDeleted(region string, shard int, org, member Id) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].SetMemberDeleted(shard, org, member)
}

func (a *internalApi) MemberIsOnlyOwner(region string, shard int, org, member Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}

	return a.regions[region].MemberIsOnlyOwner(shard, org, member)
}

func (a *internalApi) RenameMember(region string, shard int, org, member Id, newName string) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].RenameMember(shard, org, member, newName)
}

func (a *internalApi) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
	return a.regions[region].UserCanRenameOrg(shard, org, user)
}

func newSingularInternalApi(store store, log Log) SingularInternalApi {
	if store == nil {
		NilCriticalParamPanic("store")
	}
	if log == nil {
		NilCriticalParamPanic("log")
	}
	return &sIApi{
		store: store,
		log:   log,
	}
}

type sIApi struct {
	store store
	log   Log
}

func (a *sIApi) CreatePersonalTaskCenter(user Id) (int, error) {
	shard, err := a.store.createTaskSet(&taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: user,
				},
			},
			Created: time.Now().UTC(),
		},
	})
	return shard, a.log.InfoErr(err)
}

func (a *sIApi) CreateOrgTaskCenter(org, owner Id, ownerName string) (int, error) {
	shard, err := a.store.createTaskSet(&taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: org,
				},
			},
			Created: time.Now().UTC(),
		},
	})
	if err != nil {
		return shard, a.log.InfoErr(err)
	}

	err = a.store.createMember(shard, org, &member{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: owner,
			},
			Name: ownerName,
		},
	})
	if err != nil {
		a.store.deleteAccount(shard, org)
		return 0, a.log.InfoErr(err)
	}

	return shard, nil
}

func (a *sIApi) DeleteTaskCenter(shard int, account, owner Id) (error, error) {
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

func (a *sIApi) AddMembers(shard int, org, admin Id, members []*NamedEntity) (error, error) {
	member, err := a.store.getMember(shard, org, admin)
	if err != nil {
		return nil, a.log.InfoErr(err)
	}
	if member == nil || (member.Role != Owner && member.Role != Admin) {
		return insufficientPermissionErr, nil
	}

	existingMembers := make([]*NamedEntity, 0, len(members))
	newMembers := make([]*NamedEntity, 0, len(members))
	for _, mem := range members {
		existingMember, err := a.store.getMember(shard, org, mem.Id)
		if err != nil {
			return nil, a.log.InfoErr(err)
		}
		if existingMember == nil {
			newMembers = append(newMembers, mem)
		} else {
			existingMembers = append(existingMembers, mem)
		}
	}

	if len(newMembers) > 0 {
		if err := a.store.addMembers(shard, org, newMembers); err != nil {
			return nil, a.log.InfoErr(err)
		}
	}
	if len(existingMembers) > 0 {
		if err := a.store.setMembersActive(shard, org, existingMembers); err != nil {
			return nil, a.log.InfoErr(err)
		}
	}

	return nil, nil
}

func (a *sIApi) RemoveMembers(shard int, org, admin Id, members []Id) (error, error) {
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

func (a *sIApi) SetMemberDeleted(shard int, org, member Id) error {
	return a.log.InfoErr(a.store.setMemberDeleted(shard, org, member))
}

func (a *sIApi) MemberIsOnlyOwner(shard int, org, member Id) (bool, error) {
	is, err := a.store.memberIsOnlyOwner(shard, org, member)
	return is, a.log.InfoErr(err)
}

func (a *sIApi) RenameMember(shard int, org, member Id, newName string) error {
	return a.log.InfoErr(a.store.renameMember(shard, org, member, newName))
}

func (a *sIApi) UserCanRenameOrg(shard int, org, user Id) (bool, error) {
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

type role string

type store interface {
	createTaskSet(*taskSet) (int, error)
	createMember(shard int, org Id, member *member) error
	deleteAccount(shard int, account Id) error
	getMember(shard int, org, member Id) (*member, error)
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
	User               Id        `json:"user"`
	TotalRemainingTime uint64    `json:"totalRemainingTime"`
	TotalLoggedTime    uint64    `json:"totalLoggedTime"`
	ChatCount          uint64    `json:"chatCount"`
	FileCount          uint64    `json:"fileCount"`
	FileSize           uint64    `json:"fileSize"`
	Created            time.Time `json:"created"`
}

type taskSet struct {
	task
	MinimumRemainingTime uint64 `json:"minimumRemainingTime"`
	IsParallel           bool   `json:"isParallel"`
	ChildCount           uint32 `json:"childCount"`
	TaskCount            uint64 `json:"taskCount"`
	SubFileCount         uint64 `json:"subFileCount"`
	SubFileSize          uint64 `json:"subFileSize"`
	ArchivedChildCount   uint32 `json:"archivedChildCount"`
	ArchivedTaskCount    uint64 `json:"archivedTaskCount"`
	ArchivedSubFileCount uint64 `json:"archivedSubFileCount"`
	ArchivedSubFileSize  uint64 `json:"archivedSubFileSize"`
}

type member struct {
	NamedEntity
	Added              time.Time  `json:"added"`
	Deactivated        *time.Time `json:"deactivated,omitempty"`
	Deleted            *time.Time `json:"deleted,omitempty"`
	AccessTask         Id         `json:"accessTask"`
	TotalRemainingTime uint64     `json:"totalRemainingTime"`
	TotalLoggedTime    uint64     `json:"totalLoggedTime"`
	Role               role       `json:"role"`
}
