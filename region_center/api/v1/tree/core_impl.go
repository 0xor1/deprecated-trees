package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
	"time"
)

var (
	invalidRegionErr          = &Error{Code: 18, Msg: "invalid region"}
	insufficientPermissionErr = &Error{Code: 19, Msg: "insufficient permission"}
	zeroOwnerCountErr         = &Error{Code: 20, Msg: "zero owner count"}
	invalidTaskCenterTypeErr  = &Error{Code: 21, Msg: "invalid task center type"}
)

func newInternalApi(regions map[string]singularInternalApi, log Log) InternalApi {
	if regions == nil {
		NilCriticalParamPanic("regions")
	}
	if log == nil {
		NilCriticalParamPanic("logs")
	}
	return &internalApi{
		regions: regions,
		log:     log,
	}
}

type internalApi struct {
	regions map[string]singularInternalApi
	log     Log
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
		return 0, a.log.InfoErr(invalidRegionErr)
	}
	shard, err := a.regions[region].createPersonalTaskCenter(user)
	if err != nil {
		a.log.InfoErr(err)
	}
	return shard, err
}

func (a *internalApi) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, a.log.InfoErr(invalidRegionErr)
	}
	shard, err := a.regions[region].createOrgTaskCenter(org, owner, ownerName)
	if err != nil {
		a.log.InfoErr(err)
	}
	return shard, err
}

func (a *internalApi) DeleteTaskCenter(region string, shard int, account, owner Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, a.log.InfoErr(invalidRegionErr)
	}
	publicErr, err := a.regions[region].deleteTaskCenter(shard, account, owner)
	if err != nil {
		a.log.InfoErr(err)
	}
	return publicErr, err
}

func (a *internalApi) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, a.log.InfoErr(invalidRegionErr)
	}
	publicErr, err := a.regions[region].addMembers(shard, org, admin, members)
	if err != nil {
		a.log.InfoErr(err)
	}
	return publicErr, err
}

func (a *internalApi) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, a.log.InfoErr(invalidRegionErr)
	}
	publicErr, err := a.regions[region].removeMembers(shard, org, admin, members)
	if err != nil {
		a.log.InfoErr(err)
	}
	return publicErr, err
}

func (a *internalApi) SetMemberDeleted(region string, shard int, org, member Id) error {
	if !a.IsValidRegion(region) {
		return a.log.InfoErr(invalidRegionErr)
	}
	err := a.regions[region].setMemberDeleted(shard, org, member)
	if err != nil {
		a.log.InfoErr(err)
	}
	return err
}

func (a *internalApi) RenameMember(region string, shard int, org, member Id, newName string) error {
	if !a.IsValidRegion(region) {
		return a.log.InfoErr(invalidRegionErr)
	}
	err := a.regions[region].renameMember(shard, org, member, newName)
	if err != nil {
		a.log.InfoErr(err)
	}
	return err
}

func (a *internalApi) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, a.log.InfoErr(invalidRegionErr)
	}
	can, err := a.regions[region].userCanRenameOrg(shard, org, user)
	if err != nil {
		a.log.InfoErr(err)
	}
	return can, err
}

func newSingularInternalApi(store internalStore) singularInternalApi {
	if store == nil {
		NilCriticalParamPanic("store")
	}
	return &sIApi{
		store: store,
	}
}

type sIApi struct {
	store internalStore
}

func (a *sIApi) createPersonalTaskCenter(user Id) (int, error) {
	return a.store.createTaskSet(&taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: user,
				},
			},
			Created: time.Now().UTC(),
		},
	})
}

func (a *sIApi) createOrgTaskCenter(org, owner Id, ownerName string) (int, error) {
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
		return shard, err
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
		return 0, err
	}

	return shard, err
}

func (a *sIApi) deleteTaskCenter(shard int, account, owner Id) (error, error) {
	if !owner.Equal(account) {
		member, err := a.store.getMember(shard, account, owner)
		if err != nil {
			return nil, err
		}
		if member == nil || member.Role != Owner {
			return insufficientPermissionErr, nil
		}
	}
	return nil, a.store.deleteAccount(shard, account)
}

func (a *sIApi) addMembers(shard int, org, admin Id, members []*NamedEntity) (error, error) {
	member, err := a.store.getMember(shard, org, admin)
	if err != nil {
		return nil, err
	}
	if member == nil || (member.Role != Owner && member.Role != Admin) {
		return insufficientPermissionErr, nil
	}
	return nil, a.store.addMembers(shard, org, members)
}

func (a *sIApi) removeMembers(shard int, org, admin Id, members []Id) (error, error) {
	member, err := a.store.getMember(shard, org, admin)
	if err != nil {
		return nil, err
	}

	switch member.Role {
	case Owner:
		totalOrgOwnerCount, err := a.store.getTotalOrgOwnerCount(shard, org)
		if err != nil {
			return nil, err
		}

		ownerCountInRemoveSet, err := a.store.getOwnerCountInRemoveSet(shard, org, members)
		if err != nil {
			return nil, err
		}

		if totalOrgOwnerCount == ownerCountInRemoveSet {
			return zeroOwnerCountErr, nil
		}

	case Admin:
		ownerCountInRemoveSet, err := a.store.getOwnerCountInRemoveSet(shard, org, members)
		if err != nil {
			return nil, err
		}
		if ownerCountInRemoveSet > 0 {
			return insufficientPermissionErr, nil
		}

	default:
		return insufficientPermissionErr, nil
	}

	return nil, a.store.setMembersInactive(shard, org, members)
}

func (a *sIApi) setMemberDeleted(shard int, org, member Id) error {
	return a.store.setMemberInactiveAndDeleted(shard, org, member)
}

func (a *sIApi) renameMember(shard int, org, member Id, newName string) error {
	return a.store.renameMember(shard, org, member, newName)
}

func (a *sIApi) userCanRenameOrg(shard int, org, user Id) (bool, error) {
	if !user.Equal(org) {
		member, err := a.store.getMember(shard, org, user)
		if err != nil {
			return false, err
		}
		if member != nil && member.Role == Owner {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, invalidTaskCenterTypeErr
}

type singularInternalApi interface {
	createPersonalTaskCenter(user Id) (int, error)
	createOrgTaskCenter(org, owner Id, ownerName string) (int, error)
	deleteTaskCenter(shard int, account, owner Id) (public error, private error)
	addMembers(shard int, org, admin Id, members []*NamedEntity) (public error, private error)
	removeMembers(shard int, org, admin Id, members []Id) (public error, private error)
	setMemberDeleted(shard int, org, member Id) error
	renameMember(shard int, org, member Id, newName string) error
	userCanRenameOrg(shard int, org, user Id) (bool, error)
}

type internalStore interface {
	createTaskSet(*taskSet) (int, error)
	createMember(shard int, org Id, member *member) error
	deleteAccount(shard int, account Id) error
	getMember(shard int, org, member Id) (*member, error)
	addMembers(shard int, org Id, members []*NamedEntity) error
	getTotalOrgOwnerCount(shard int, org Id) (int, error)
	getOwnerCountInRemoveSet(shard int, org Id, members []Id) (int, error)
	setMembersInactive(shard int, org Id, members []Id) error
	setMemberInactiveAndDeleted(shard int, org Id, member Id) error
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
	AccessTask         Id     `json:"accessTask"`
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
	Role               role   `json:"role"`
	IsActive           bool   `json:"isActive"`
	IsDeleted          bool   `json:"isDeleted"`
}
