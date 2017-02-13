package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
	"bitbucket.org/robsix/task_center/region_center/model"
	"time"
)

var (
	invalidRegionErr          = &Error{Code: 18, Msg: "invalid region"}
	insufficientPermissionErr = &Error{Code: 19, Msg: "insufficient permission"}
	zeroOwnerCountErr         = &Error{Code: 20, Msg: "zero owner count"}
)

func newApi(regions map[string]internalApi, log Log) Api {
	if regions == nil {
		NilCriticalParamPanic("regions")
	}
	if log == nil {
		NilCriticalParamPanic("logs")
	}
	return &api{
		regions: regions,
		log:     log,
	}
}

type api struct {
	regions map[string]internalApi
	log     Log
}

func (a *api) GetRegions() []string {
	regions := make([]string, 0, len(a.regions))
	for k := range a.regions {
		regions = append(regions, k)
	}
	return regions
}

func (a *api) IsValidRegion(region string) bool {
	return a.regions[region] != nil
}

func (a *api) CreatePersonalTaskCenter(region string, user Id) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].createPersonalTaskCenter(user)
}

func (a *api) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return a.regions[region].createOrgTaskCenter(org, owner, ownerName)
}

func (a *api) DeleteTaskCenter(region string, shard int, owner, account Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].deleteTaskCenter(shard, owner, account)
}

func (a *api) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].addMembers(shard, org, admin, members)
}

func (a *api) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return a.regions[region].removeMembers(shard, org, admin, members)
}

func (a *api) SetMemberDeleted(region string, shard int, org, member Id) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].setMemberDeleted(shard, org, member)
}

func (a *api) RenameMember(region string, shard int, org, member Id, newName string) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return a.regions[region].renameMember(shard, org, member, newName)
}

func (a *api) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
	return a.regions[region].userCanRenameOrg(shard, org, user)
}

//func newIApi(store store) internalApi {
//	return &iApi{
//		store: store,
//	}
//}

type iApi struct {
	store store
}

func (a *iApi) createPersonalTaskCenter(user Id) (int, error) {
	return a.store.createTaskSet(&model.TaskSet{
		Task: model.Task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: user,
				},
			},
			Created: time.Now().UTC(),
		},
	})
}

func (a *iApi) createOrgTaskCenter(org, owner Id, ownerName string) (int, error) {
	shard, err := a.store.createTaskSet(&model.TaskSet{
		Task: model.Task{
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

	err = a.store.createMember(shard, &model.Member{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: owner,
			},
			Name: ownerName,
		},
	})
	if err != nil {
		a.store.deleteAccount(shard, org)
	}

	return shard, err
}

func (a *iApi) deleteTaskCenter(shard int, owner, account Id) (error, error) {
	if !owner.Equal(account) {
		member, err := a.store.getMember(shard, owner)
		if err != nil {
			return nil, err
		}
		if member == nil || member.Role != model.Owner {
			return insufficientPermissionErr, nil
		}
	}
	return nil, a.store.deleteAccount(shard, account)
}

func (a *iApi) addMembers(shard int, org, admin Id, members []*NamedEntity) (error, error) {
	member, err := a.store.getMember(shard, admin)
	if err != nil {
		return nil, err
	}
	if member == nil || member.Role != model.Owner || member.Role != model.Admin {
		return insufficientPermissionErr, nil
	}
	return nil, a.store.addMembers(shard, members)
}

func (a *iApi) removeMembers(shard int, org, admin Id, members []Id) (error, error) {
	member, err := a.store.getMember(shard, admin)
	if err != nil {
		return nil, err
	}
	if member == nil || member.Role != model.Owner || member.Role != model.Admin {
		return insufficientPermissionErr, nil
	}

	if member.Role == model.Owner {
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
	}

	if member.Role == model.Admin {
		ownerCountInRemoveSet, err := a.store.getOwnerCountInRemoveSet(shard, org, members)
		if err != nil {
			return nil, err
		}
		if ownerCountInRemoveSet > 0 {
			return insufficientPermissionErr, nil
		}
	}

	return nil, a.store.setMembersInactive(shard, org, members)
}

//func (a *iApi) setMemberDeleted(shard int, org, member Id) error {
//
//}
//
//func (a *iApi) renameMember(shard int, org, member Id, newName string) error {
//
//}
//
//func (a *iApi) userCanRenameOrg(shard int, org, user Id) (bool, error) {
//
//}

type internalApi interface {
	createPersonalTaskCenter(user Id) (int, error)
	createOrgTaskCenter(org, owner Id, ownerName string) (int, error)
	deleteTaskCenter(shard int, owner, account Id) (public error, private error)
	addMembers(shard int, org, admin Id, members []*NamedEntity) (public error, private error)
	removeMembers(shard int, org, admin Id, members []Id) (public error, private error)
	setMemberDeleted(shard int, org, member Id) error
	renameMember(shard int, org, member Id, newName string) error
	userCanRenameOrg(shard int, org, user Id) (bool, error)
}

type store interface {
	createTaskSet(node Id) (int, error)
	createMember(shard int, member model.Member) error
	deleteAccount(shard int, account Id) error
	getMember(shard int, member Id) (*model.Member, error)
	addMembers(shard int, members []*NamedEntity) error
	getTotalOrgOwnerCount(shard int, org Id) (int, error)
	getOwnerCountInRemoveSet(shard int, org Id, members []Id) (int, error)
	setMembersInactive(shard int, org Id, members []Id) error
}
