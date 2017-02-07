package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
)

var (
	invalidRegionErr = &Error{Code: 24, Msg: ""}
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
	if a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}

}

func (a *api) CreateOrgTaskCenter(region string, org, owner Id, ownerName string) (int, error) {
	if a.IsValidRegion(region) {
		return 0, invalidRegionErr
	}
}

func (a *api) DeleteTaskCenter(region string, shard int, account Id) error {
	if a.IsValidRegion(region) {
		return invalidRegionErr
	}
}

func (a *api) AddMember(region string, shard int, org, member Id, memberName string) error {
	if a.IsValidRegion(region) {
		return invalidRegionErr
	}
}

func (a *api) RemoveMember(region string, shard int, org, member Id) error {
	if a.IsValidRegion(region) {
		return invalidRegionErr
	}
}

func (a *api) RenameMember(region string, shard int, org, member Id, newName string) error {
	if a.IsValidRegion(region) {
		return invalidRegionErr
	}
}

func (a *api) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	if a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
}

func (a *api) UserCanMigrateOrg(region string, shard int, org, user Id) (bool, error) {
	if a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
}

func (a *api) UserCanDeleteOrg(region string, shard int, org, user Id) (bool, error) {
	if a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
}

func (a *api) UserCanManageMembers(region string, shard int, org, user Id) (bool, error) {
	if a.IsValidRegion(region) {
		return false, invalidRegionErr
	}
}

type internalApi interface {
	CreatePersonalTaskCenter(user Id) (int, error)
	CreateOrgTaskCenter(org, owner Id, ownerName string) (int, error)
	DeleteTaskCenter(shard int, account Id) error
	AddMember(shard int, org, member Id, memberName string) error
	RemoveMember(shard int, org, member Id) error
	RenameMember(shard int, org, member Id, newName string) error
	UserCanRenameOrg(shard int, org, user Id) (bool, error)
	UserCanMigrateOrg(shard int, org, user Id) (bool, error)
	UserCanDeleteOrg(shard int, org, user Id) (bool, error)
	UserCanManageMembers(shard int, org, user Id) (bool, error)
}
