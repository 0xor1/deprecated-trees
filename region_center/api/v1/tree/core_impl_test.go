package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_newApi_nilRegionsPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newInternalApi(nil, nil)
}

func Test_newApi_nilLogPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	regions := map[string]singularInternalApi{"us": &mockInternalApi{}}
	newInternalApi(regions, nil)
}

func Test_newApi_success(t *testing.T) {
	regions := map[string]singularInternalApi{"us": &mockInternalApi{}}
	api := newInternalApi(regions, NewLog(nil))

	assert.NotNil(t, api)
}

func Test_apiIsValidRegion(t *testing.T) {
	regions := map[string]singularInternalApi{"us": &mockInternalApi{}}
	api := newInternalApi(regions, NewLog(nil))

	assert.False(t, api.IsValidRegion("ch"))
	assert.True(t, api.IsValidRegion("us"))
}

func Test_apiGetRegions(t *testing.T) {
	regions := map[string]singularInternalApi{"us": &mockInternalApi{}, "au": &mockInternalApi{}}
	api := newInternalApi(regions, NewLog(nil))

	regionsSlice := api.GetRegions()
	assert.Contains(t, regionsSlice, "us")
	assert.Contains(t, regionsSlice, "au")
}

func Test_apiCreatePersonalTaskCenter(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

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
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

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
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

	publicErr, err := api.DeleteTaskCenter("ch", 2, orgId, userId)
	assert.Nil(t, publicErr)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("deleteTaskCenter", 2, orgId, userId).Return(testErr, testErr2)
	publicErr, err = api.DeleteTaskCenter("us", 2, orgId, userId)
	assert.Equal(t, testErr, publicErr)
	assert.Equal(t, testErr2, err)
}

func Test_apiAddMembers(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	members := make([]*NamedEntity, 2, 2)
	api := newInternalApi(regions, NewLog(nil))

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
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	members := make([]Id, 2, 2)
	api := newInternalApi(regions, NewLog(nil))

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
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

	err := api.SetMemberDeleted("ch", 2, orgId, userId)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("setMemberDeleted", 2, orgId, userId).Return(testErr)
	err = api.SetMemberDeleted("us", 2, orgId, userId)
	assert.Equal(t, testErr, err)
}

func Test_apiRenameMember(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

	err := api.RenameMember("ch", 2, orgId, userId, "ali")
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("renameMember", 2, orgId, userId, "ali").Return(testErr)
	err = api.RenameMember("us", 2, orgId, userId, "ali")
	assert.Equal(t, testErr, err)
}

func Test_apiUserCanRenameOrg(t *testing.T) {
	iApi := &mockInternalApi{}
	regions := map[string]singularInternalApi{"us": iApi}
	userId, _ := NewId()
	orgId, _ := NewId()
	api := newInternalApi(regions, NewLog(nil))

	can, err := api.UserCanRenameOrg("ch", 2, orgId, userId)
	assert.False(t, can)
	assert.Equal(t, invalidRegionErr, err)

	iApi.On("userCanRenameOrg", 2, orgId, userId).Return(true, testErr)
	can, err = api.UserCanRenameOrg("us", 2, orgId, userId)
	assert.True(t, can)
	assert.Equal(t, testErr, err)
}

func Test_newIApi_nilStorePanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newSingularInternalApi(nil)
}

func Test_newIApi_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)
	assert.NotNil(t, iApi)
}

func Test_iApiCreatePersonalTaskCenter_storeCreateTaskSetErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	store.On("createTaskSet", &taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: myId,
				},
			},
			Created: nowTestVal,
		},
	}).Return(0, testErr)

	shard, err := iApi.createPersonalTaskCenter(myId)
	assert.Zero(t, shard)
	assert.Equal(t, testErr, err)
}

func Test_iApiCreatePersonalTaskCenter_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	store.On("createTaskSet", &taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: myId,
				},
			},
			Created: nowTestVal,
		},
	}).Return(5, nil)

	shard, err := iApi.createPersonalTaskCenter(myId)
	assert.Equal(t, 5, shard)
	assert.Nil(t, err)
}

func Test_iApiCreateOrgTaskCenter_storeCreateTaskSetErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("createTaskSet", &taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: orgId,
				},
			},
			Created: nowTestVal,
		},
	}).Return(0, testErr)

	shard, err := iApi.createOrgTaskCenter(orgId, myId, "ali")
	assert.Equal(t, 0, shard)
	assert.Equal(t, testErr, err)
}

func Test_iApiCreateOrgTaskCenter_storeCreateMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("createTaskSet", &taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: orgId,
				},
			},
			Created: nowTestVal,
		},
	}).Return(5, nil)
	store.On("createMember", 5, orgId, &member{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: myId,
			},
			Name: "ali",
		},
	}).Return(testErr)
	store.On("deleteAccount", 5, orgId).Return(nil)

	shard, err := iApi.createOrgTaskCenter(orgId, myId, "ali")
	assert.Equal(t, 0, shard)
	assert.Equal(t, testErr, err)
}

func Test_iApiCreateOrgTaskCenter_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("createTaskSet", &taskSet{
		task: task{
			NamedEntity: NamedEntity{
				Entity: Entity{
					Id: orgId,
				},
			},
			Created: nowTestVal,
		},
	}).Return(5, nil)
	store.On("createMember", 5, orgId, &member{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: myId,
			},
			Name: "ali",
		},
	}).Return(nil)

	shard, err := iApi.createOrgTaskCenter(orgId, myId, "ali")
	assert.Equal(t, 5, shard)
	assert.Nil(t, err)
}

func Test_iApiDeleteTaskCenter_storeGetMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(nil, testErr)

	publicErr, err := iApi.deleteTaskCenter(5, orgId, myId)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiDeleteTaskCenter_insufficientPermissionErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)

	publicErr, err := iApi.deleteTaskCenter(5, orgId, myId)
	assert.Equal(t, insufficientPermissionErr, publicErr)
	assert.Nil(t, err)
}

func Test_iApiDeleteTaskCenter_storeDeleteAccountErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)
	store.On("deleteAccount", 5, orgId).Return(testErr)

	publicErr, err := iApi.deleteTaskCenter(5, orgId, myId)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiDeleteTaskCenter_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)
	store.On("deleteAccount", 5, orgId).Return(nil)

	publicErr, err := iApi.deleteTaskCenter(5, orgId, myId)
	assert.Nil(t, publicErr)
	assert.Nil(t, err)
}

func Test_iApiAddMembers_storeGetMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []*NamedEntity{}
	store.On("getMember", 5, orgId, myId).Return(nil, testErr)

	publicErr, err := iApi.addMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiAddMembers_insufficientPermissionErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []*NamedEntity{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Writer}, nil)

	publicErr, err := iApi.addMembers(5, orgId, myId, members)
	assert.Equal(t, insufficientPermissionErr, publicErr)
	assert.Nil(t, err)
}

func Test_iApiAddMembers_storeAddMembersErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []*NamedEntity{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)
	store.On("addMembers", 5, orgId, members).Return(testErr)

	publicErr, err := iApi.addMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiAddMembers_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []*NamedEntity{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)
	store.On("addMembers", 5, orgId, members).Return(nil)

	publicErr, err := iApi.addMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Nil(t, err)
}

func Test_iApiRemoveMembers_storeGetMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(nil, testErr)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiRemoveMembers_insufficientPermissionErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Writer}, nil)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Equal(t, insufficientPermissionErr, publicErr)
	assert.Nil(t, err)
}

func Test_iApiRemoveMembers_owner_storeGetTotalOrgOwnerCountErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)
	store.On("getTotalOrgOwnerCount", 5, orgId).Return(0, testErr)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiRemoveMembers_owner_storeGetOwnerCountInRemoveSetErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)
	store.On("getTotalOrgOwnerCount", 5, orgId).Return(3, nil)
	store.On("getOwnerCountInRemoveSet", 5, orgId, members).Return(0, testErr)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiRemoveMembers_owner_zeroOwnerCountErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)
	store.On("getTotalOrgOwnerCount", 5, orgId).Return(3, nil)
	store.On("getOwnerCountInRemoveSet", 5, orgId, members).Return(3, nil)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Equal(t, zeroOwnerCountErr, publicErr)
	assert.Nil(t, err)
}

func Test_iApiRemoveMembers_admin_storeGetOwnerCountInRemoveSetErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)
	store.On("getOwnerCountInRemoveSet", 5, orgId, members).Return(0, testErr)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiRemoveMembers_admin_insufficientPermissionErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)
	store.On("getOwnerCountInRemoveSet", 5, orgId, members).Return(1, nil)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Equal(t, insufficientPermissionErr, publicErr)
	assert.Nil(t, err)
}

func Test_iApiRemoveMembers_storeSetMembersInactiveErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := []Id{}
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)
	store.On("getOwnerCountInRemoveSet", 5, orgId, members).Return(0, nil)
	store.On("setMembersInactive", 5, orgId, members).Return(testErr)

	publicErr, err := iApi.removeMembers(5, orgId, myId, members)
	assert.Nil(t, publicErr)
	assert.Equal(t, testErr, err)
}

func Test_iApiSetMemberDeleted_storeSetMemberInactiveAndDeletedErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("setMemberInactiveAndDeleted", 5, orgId, myId).Return(testErr)

	err := iApi.setMemberDeleted(5, orgId, myId)
	assert.Equal(t, testErr, err)
}

func Test_iApiSetMemberDeleted_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("setMemberInactiveAndDeleted", 5, orgId, myId).Return(nil)

	err := iApi.setMemberDeleted(5, orgId, myId)
	assert.Nil(t, err)
}

func Test_iApiRenameMember_storeRenameMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("renameMember", 5, orgId, myId, "ali").Return(testErr)

	err := iApi.renameMember(5, orgId, myId, "ali")
	assert.Equal(t, testErr, err)
}

func Test_iApiUserCanRenameOrg_storeGetMemberErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(nil, testErr)

	can, err := iApi.userCanRenameOrg(5, orgId, myId)
	assert.False(t, can)
	assert.Equal(t, testErr, err)
}

func Test_iApiUserCanRenameOrg_true(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Owner}, nil)

	can, err := iApi.userCanRenameOrg(5, orgId, myId)
	assert.True(t, can)
	assert.Nil(t, err)
}

func Test_iApiUserCanRenameOrg_false(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getMember", 5, orgId, myId).Return(&member{Role: Admin}, nil)

	can, err := iApi.userCanRenameOrg(5, orgId, myId)
	assert.False(t, can)
	assert.Nil(t, err)
}

func Test_iApiUserCanRenameOrg_invalidTaskCenterTypeErr(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()

	can, err := iApi.userCanRenameOrg(5, myId, myId)
	assert.False(t, can)
	assert.Equal(t, invalidTaskCenterTypeErr, err)
}

func Test_iApiRenameMember_success(t *testing.T) {
	store := &mockStore{}
	iApi := newSingularInternalApi(store)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("renameMember", 5, orgId, myId, "ali").Return(nil)

	err := iApi.renameMember(5, orgId, myId, "ali")
	assert.Nil(t, err)
}

var (
	testErr    = errors.New("test")
	testErr2   = errors.New("test2")
	nowTestVal = time.Now().UTC()
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

func (m *mockInternalApi) deleteTaskCenter(shard int, account, owner Id) (error, error) {
	args := m.Called(shard, account, owner)
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

type mockStore struct {
	mock.Mock
}

func (m *mockStore) createTaskSet(ts *taskSet) (int, error) {
	ts.Created = nowTestVal
	args := m.Called(ts)
	return args.Int(0), args.Error(1)
}

func (m *mockStore) createMember(shard int, org Id, member *member) error {
	args := m.Called(shard, org, member)
	return args.Error(0)
}

func (m *mockStore) deleteAccount(shard int, account Id) error {
	args := m.Called(shard, account)
	return args.Error(0)
}

func (m *mockStore) getMember(shard int, org, memberId Id) (*member, error) {
	args := m.Called(shard, org, memberId)
	mem := args.Get(0)
	if mem != nil {
		return mem.(*member), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockStore) addMembers(shard int, org Id, members []*NamedEntity) error {
	args := m.Called(shard, org, members)
	return args.Error(0)
}

func (m *mockStore) getTotalOrgOwnerCount(shard int, org Id) (int, error) {
	args := m.Called(shard, org)
	return args.Int(0), args.Error(1)
}

func (m *mockStore) getOwnerCountInRemoveSet(shard int, org Id, members []Id) (int, error) {
	args := m.Called(shard, org, members)
	return args.Int(0), args.Error(1)
}

func (m *mockStore) setMembersInactive(shard int, org Id, members []Id) error {
	args := m.Called(shard, org, members)
	return args.Error(0)
}

func (m *mockStore) setMemberInactiveAndDeleted(shard int, org Id, member Id) error {
	args := m.Called(shard, org, member)
	return args.Error(0)
}

func (m *mockStore) renameMember(shard int, org Id, member Id, newName string) error {
	args := m.Called(shard, org, member, newName)
	return args.Error(0)
}
