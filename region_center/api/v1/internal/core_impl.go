package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
)

var (
	invalidRegionErr = &Error{Code: 18, Msg: ""}
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

func (a *api) IsValidRegion(region string) bool {
	return a.regions[region] != nil
}

func (a *api) CreatePersonalTaskCenter(region string, user Id) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return 0, NotImplementedErr
}

func (a *api) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if !a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
	return 0, NotImplementedErr
}

func (a *api) DeleteTaskCenter(region string, shard int, owner, account Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return nil, NotImplementedErr
}

func (a *api) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return nil, NotImplementedErr
}

func (a *api) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	if !a.IsValidRegion(region) {
		return nil, invalidRegionErr
	}
	return nil, NotImplementedErr
}

func (a *api) SetMemberDeleted(region string, shard int, org, member Id) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return NotImplementedErr
}

func (a *api) RenameMember(region string, shard int, org, member Id, newName string) error {
	if !a.IsValidRegion(region) {
		return invalidRegionErr
	}
	return NotImplementedErr
}

func (a *api) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	if !a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
	return false, NotImplementedErr
}

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
