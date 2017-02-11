package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_newApi_nilRegionsPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newApi(nil, nil)
}

func Test_newApi_nilLogPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	regions := map[string]internalApi{"us": &mockInternalApi{}}
	newApi(regions, nil)
}

func Test_newApi_success(t *testing.T) {
	regions := map[string]internalApi{"us": &mockInternalApi{}}
	api := newApi(regions, NewLog(nil))

	assert.NotNil(t, api)
}

type mockInternalApi struct {
	mock.Mock
}

func (m *mockInternalApi) createPersonalTaskCenter(user Id) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalApi) createOrgTaskCenter(org, owner Id, ownerName string) (int, error) {
	args := m.Called(org, owner, ownerName)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalApi) deleteTaskCenter(shard int, owner, account Id) (error, error) {
	args := m.Called(shard, owner, account)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalApi) addMembers(shard int, org, admin Id, members []*NamedEntity) (error, error) {
	args := m.Called(shard, org, admin, members)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalApi) removeMembers(shard int, org, admin Id, members []Id) (error, error) {
	args := m.Called(shard, org, admin, members)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalApi) setMemberDeleted(shard int, org, member Id) error {
	args := m.Called(shard, org, member)
	return args.Error(0)
}

func (m *mockInternalApi) renameMember(shard int, org, member Id, newName string) error {
	args := m.Called(shard, org, member, newName)
	return args.Error(0)
}

func (m *mockInternalApi) userCanRenameOrg(shard int, org, user Id) (bool, error) {
	args := m.Called(shard, org, user)
	return args.Bool(0), args.Error(1)
}
