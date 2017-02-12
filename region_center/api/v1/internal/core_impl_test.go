package internal

import (
	. "bitbucket.org/robsix/task_center/misc"
	"errors"
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

func Test_apiIsValidRegion(t *testing.T) {
	regions := map[string]internalApi{"us": &mockInternalApi{}}
	api := newApi(regions, NewLog(nil))

	assert.False(t, api.IsValidRegion("ch"))
	assert.True(t, api.IsValidRegion("us"))
}

func Test_apiGetRegions(t *testing.T) {
	regions := map[string]internalApi{"us": &mockInternalApi{}, "au": &mockInternalApi{}}
	api := newApi(regions, NewLog(nil))

	regionsSlice := api.GetRegions()
	assert.Contains(t, regionsSlice, "us")
	assert.Contains(t, regionsSlice, "au")
}

func Test_apiCreatePersonalTaskCenter(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	shard, err := api.CreatePersonalTaskCenter("ch", userId)
	assert.Equal(t, 0, shard)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("createPersonalTaskCenter", userId).Return(2, testErr)
	shard, err = api.CreatePersonalTaskCenter("us", userId)
	assert.Equal(t, 2, shard)
	assert.Equal(t, testErr, err)
}

func Test_apiCreateOrgTaskCenter(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	shard, err := api.CreateOrgTaskCenter("ch", orgId, userId, "ali")
	assert.Equal(t, 0, shard)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("createOrgTaskCenter", orgId, userId, "ali").Return(2, testErr)
	shard, err = api.CreateOrgTaskCenter("us", orgId, userId, "ali")
	assert.Equal(t, 2, shard)
	assert.Equal(t, testErr, err)
}

func Test_apiDeleteTaskCenter(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	publicErr, err := api.DeleteTaskCenter("ch", 2, userId, orgId)
	assert.Nil(t, publicErr)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("deleteTaskCenter", 2, userId, orgId).Return(testErr, testErr2)
	publicErr, err = api.DeleteTaskCenter("us", 2, userId, orgId)
	assert.Equal(t, testErr, publicErr)
	assert.Equal(t, testErr2, err)
}

func Test_apiAddMembers(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	members := make([]*NamedEntity, 2, 2)
	api := newApi(regions, NewLog(nil))

	publicErr, err := api.AddMembers("ch", 2, orgId, userId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("addMembers", 2, orgId, userId, members).Return(testErr, testErr2)
	publicErr, err = api.AddMembers("us", 2, orgId, userId, members)
	assert.Equal(t, testErr, publicErr)
	assert.Equal(t, testErr2, err)
}

func Test_apiRemoveMembers(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	members := make([]Id, 2, 2)
	api := newApi(regions, NewLog(nil))

	publicErr, err := api.RemoveMembers("ch", 2, orgId, userId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("removeMembers", 2, orgId, userId, members).Return(testErr, testErr2)
	publicErr, err = api.RemoveMembers("us", 2, orgId, userId, members)
	assert.Equal(t, testErr, publicErr)
	assert.Equal(t, testErr2, err)
}

func Test_apiSetMemberDeleted(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	err := api.SetMemberDeleted("ch", 2, orgId, userId)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("setMemberDeleted", 2, orgId, userId).Return(testErr)
	err = api.SetMemberDeleted("us", 2, orgId, userId)
	assert.Equal(t, testErr, err)
}

func Test_apiRenameMember(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	err := api.RenameMember("ch", 2, orgId, userId, "ali")
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("renameMember", 2, orgId, userId, "ali").Return(testErr)
	err = api.RenameMember("us", 2, orgId, userId, "ali")
	assert.Equal(t, testErr, err)
}

func Test_apiUserCanRenameOrg(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]internalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newApi(regions, NewLog(nil))

	can, err := api.UserCanRenameOrg("ch", 2, orgId, userId)
	assert.False(t, can)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("userCanRenameOrg", 2, orgId, userId).Return(true, testErr)
	can, err = api.UserCanRenameOrg("us", 2, orgId, userId)
	assert.True(t, can)
	assert.Equal(t, testErr, err)
}

var (
	testErr  = errors.New("test")
	testErr2 = errors.New("test2")
)

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
