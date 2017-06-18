package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"errors"
	"github.com/stretchr/testify/mock"
)

//func Test_newInternalApiClient_nilRegionsPanic(t *testing.T) {
//	defer func() {
//		err := recover().(error)
//		assert.IsType(t, &Error{}, err)
//	}()
//	newInternalRegionClient(nil)
//}
//
//func Test_newInternalApiClient_success(t *testing.T) {
//	regions := map[string]Api{"us": &mockSingluarInternalApi{}}
//	api := newInternalRegionClient(regions)
//
//	assert.NotNil(t, api)
//}
//
//func Test_clientIsValidRegion(t *testing.T) {
//	regions := map[string]Api{"us": &mockSingluarInternalApi{}}
//	api := newInternalRegionClient(regions)
//
//	assert.False(t, api.IsValidRegion("ch"))
//	assert.True(t, api.IsValidRegion("us"))
//}
//
//func Test_clientGetRegions(t *testing.T) {
//	regions := map[string]Api{"us": &mockSingluarInternalApi{}, "au": &mockSingluarInternalApi{}}
//	api := newInternalRegionClient(regions)
//
//	regionsSlice := api.GetRegions()
//	assert.Contains(t, regionsSlice, "us")
//	assert.Contains(t, regionsSlice, "au")
//}
//
//func Test_clientCreatePersonalTaskCenter(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	api := newInternalRegionClient(regions)
//
//	shard, err := api.CreatePersonalTaskCenter("ch", userId)
//	assert.Equal(t, 0, shard)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("CreatePersonalTaskCenter", userId).Return(2, testErr)
//	shard, err = api.CreatePersonalTaskCenter("us", userId)
//	assert.Equal(t, 2, shard)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_clientCreateOrgTaskCenter(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	shard, err := api.CreateOrgTaskCenter("ch", orgId, userId, "ali")
//	assert.Equal(t, 0, shard)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("CreateOrgTaskCenter", orgId, userId, "ali").Return(2, testErr)
//	shard, err = api.CreateOrgTaskCenter("us", orgId, userId, "ali")
//	assert.Equal(t, 2, shard)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_clientDeleteTaskCenter(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	publicErr, err := api.DeleteTaskCenter("ch", 2, orgId, userId)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("DeleteTaskCenter", 2, orgId, userId).Return(testErr, testErr2)
//	publicErr, err = api.DeleteTaskCenter("us", 2, orgId, userId)
//	assert.Equal(t, testErr, publicErr)
//	assert.Equal(t, testErr2, err)
//}
//
//func Test_clientAddMembers(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	members := make([]*NamedEntity, 2, 2)
//	api := newInternalRegionClient(regions)
//
//	publicErr, err := api.AddMembers("ch", 2, orgId, userId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("AddMembers", 2, orgId, userId, members).Return(testErr, testErr2)
//	publicErr, err = api.AddMembers("us", 2, orgId, userId, members)
//	assert.Equal(t, testErr, publicErr)
//	assert.Equal(t, testErr2, err)
//}
//
//func Test_clientRemoveMembers(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	members := make([]Id, 2, 2)
//	api := newInternalRegionClient(regions)
//
//	publicErr, err := api.RemoveMembers("ch", 2, orgId, userId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("RemoveMembers", 2, orgId, userId, members).Return(testErr, testErr2)
//	publicErr, err = api.RemoveMembers("us", 2, orgId, userId, members)
//	assert.Equal(t, testErr, publicErr)
//	assert.Equal(t, testErr2, err)
//}
//
//func Test_clientSetMemberDeleted(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	err := api.SetMemberDeleted("ch", 2, orgId, userId)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("SetMemberDeleted", 2, orgId, userId).Return(testErr)
//	err = api.SetMemberDeleted("us", 2, orgId, userId)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_clientMemberIsOnlyOwner(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	isOnlyOwner, err := api.MemberIsOnlyOwner("ch", 2, orgId, userId)
//	assert.False(t, isOnlyOwner)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("MemberIsOnlyOwner", 2, orgId, userId).Return(true, testErr)
//	isOnlyOwner, err = api.MemberIsOnlyOwner("us", 2, orgId, userId)
//	assert.True(t, isOnlyOwner)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_clientRenameMember(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	err := api.RenameMember("ch", 2, orgId, userId, "ali")
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("RenameMember", 2, orgId, userId, "ali").Return(testErr)
//	err = api.RenameMember("us", 2, orgId, userId, "ali")
//	assert.Equal(t, testErr, err)
//}
//
//func Test_clientUserIsOrgOwner(t *testing.T) {
//	iApi := &mockSingluarInternalApi{}
//	regions := map[string]Api{"us": iApi}
//	userId := NewId()
//	orgId := NewId()
//	api := newInternalRegionClient(regions)
//
//	can, err := api.UserIsOrgOwner("ch", 2, orgId, userId)
//	assert.False(t, can)
//	assert.Equal(t, invalidRegionErr, err)
//
//	iApi.On("UserIsOrgOwner", 2, orgId, userId).Return(true, testErr)
//	can, err = api.UserIsOrgOwner("us", 2, orgId, userId)
//	assert.True(t, can)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_newInternalApi_nilStorePanic(t *testing.T) {
//	defer func() {
//		err := recover().(error)
//		assert.IsType(t, &Error{}, err)
//	}()
//	newInternalApi(nil, nil)
//}
//
//func Test_newInternalApi_nilLogPanic(t *testing.T) {
//	defer func() {
//		err := recover().(error)
//		assert.IsType(t, &Error{}, err)
//	}()
//	store := &mockStore{}
//	newInternalApi(store, nil)
//}
//
//func Test_newInternalApi_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//	assert.NotNil(t, iApi)
//}
//
//func Test_internalApiCreatePersonalTaskCenter_storeCreateAbstractTaskErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	store.On("registerPersonalAccount", myId).Return(0, testErr)
//
//	shard, err := iApi.CreatePersonalTaskCenter(myId)
//	assert.Zero(t, shard)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiCreatePersonalTaskCenter_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	store.On("registerPersonalAccount", myId).Return(5, nil)
//
//	shard, err := iApi.CreatePersonalTaskCenter(myId)
//	assert.Equal(t, 5, shard)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiCreateOrgTaskCenter_storeRegisterOrgOrgErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("registerOrgAccount", orgId, myId, "ali").Return(0, testErr)
//
//	shard, err := iApi.CreateOrgTaskCenter(orgId, myId, "ali")
//	assert.Equal(t, 0, shard)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiCreateOrgTaskCenter_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("registerOrgAccount", orgId, myId, "ali").Return(5, nil)
//
//	shard, err := iApi.CreateOrgTaskCenter(orgId, myId, "ali")
//	assert.Equal(t, 5, shard)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiDeleteTaskCenter_storeGetMemberErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(nil, testErr)
//
//	publicErr, err := iApi.DeleteTaskCenter(5, orgId, myId)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiDeleteTaskCenter_insufficientPermissionErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//
//	publicErr, err := iApi.DeleteTaskCenter(5, orgId, myId)
//	assert.Equal(t, insufficientPermissionErr, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiDeleteTaskCenter_storeDeleteAccountErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//	store.On("deleteAccount", 5, orgId).Return(testErr)
//
//	publicErr, err := iApi.DeleteTaskCenter(5, orgId, myId)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiDeleteTaskCenter_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//	store.On("deleteAccount", 5, orgId).Return(nil)
//
//	publicErr, err := iApi.DeleteTaskCenter(5, orgId, myId)
//	assert.Nil(t, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiAddMembers_storeGetMemberErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []*NamedEntity{}
//	store.On("getMember", 5, orgId, myId).Return(nil, testErr)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiAddMembers_insufficientPermissionErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []*NamedEntity{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Writer}, nil)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Equal(t, insufficientPermissionErr, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiAddMembers_storeGetMemberErr2(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	newMemId := NewId()
//	newMem := &NamedEntity{Entity: Entity{Id: newMemId}, Name: "bob"}
//	members := []*NamedEntity{newMem}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getMember", 5, orgId, newMemId).Return(nil, testErr)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiAddMembers_storeAddMembersErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	newMemId := NewId()
//	newMem := &NamedEntity{Entity: Entity{Id: newMemId}, Name: "bob"}
//	members := []*NamedEntity{newMem}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getMember", 5, orgId, newMemId).Return(nil, nil)
//	store.On("addMembers", 5, orgId, members).Return(testErr)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiAddMembers_storeSetMembersActiveErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	oldMemId := NewId()
//	oldMem := &NamedEntity{Entity: Entity{Id: oldMemId}, Name: "bob"}
//	members := []*NamedEntity{oldMem}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getMember", 5, orgId, oldMemId).Return(&orgMember{}, nil)
//	store.On("setMembersActive", 5, orgId, members).Return(testErr)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiAddMembers_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	newMemId := NewId()
//	oldMemId := NewId()
//	newMem := &NamedEntity{Entity: Entity{Id: newMemId}, Name: "bob"}
//	oldMem := &NamedEntity{Entity: Entity{Id: oldMemId}, Name: "cat"}
//	members := []*NamedEntity{newMem, oldMem}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getMember", 5, orgId, newMemId).Return(nil, nil)
//	store.On("getMember", 5, orgId, oldMemId).Return(&orgMember{}, nil)
//	store.On("addMembers", 5, orgId, []*NamedEntity{newMem}).Return(nil)
//	store.On("setMembersActive", 5, orgId, []*NamedEntity{oldMem}).Return(nil)
//
//	publicErr, err := iApi.AddMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiRemoveMembers_storeGetMemberErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(nil, testErr)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiRemoveMembers_insufficientPermissionErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Writer}, nil)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Equal(t, insufficientPermissionErr, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiRemoveMembers_owner_storeGetTotalOrgOwnerCountErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//	store.On("getTotalOrgOwnerCount", 5, orgId).Return(0, testErr)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiRemoveMembers_owner_storegetOwnerCountInSetErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//	store.On("getTotalOrgOwnerCount", 5, orgId).Return(3, nil)
//	store.On("getOwnerCountInSet", 5, orgId, members).Return(0, testErr)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiRemoveMembers_owner_zeroOwnerCountErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//	store.On("getTotalOrgOwnerCount", 5, orgId).Return(3, nil)
//	store.On("getOwnerCountInSet", 5, orgId, members).Return(3, nil)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Equal(t, zeroOwnerCountErr, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiRemoveMembers_admin_storegetOwnerCountInSetErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getOwnerCountInSet", 5, orgId, members).Return(0, testErr)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiRemoveMembers_admin_insufficientPermissionErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getOwnerCountInSet", 5, orgId, members).Return(1, nil)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Equal(t, insufficientPermissionErr, publicErr)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiRemoveMembers_storeSetMembersInactiveErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	members := []Id{}
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//	store.On("getOwnerCountInSet", 5, orgId, members).Return(0, nil)
//	store.On("setMembersInactive", 5, orgId, members).Return(testErr)
//
//	publicErr, err := iApi.RemoveMembers(5, orgId, myId, members)
//	assert.Nil(t, publicErr)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiSetMemberDeleted_storeSetMemberDeletedErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("setMemberDeleted", 5, orgId, myId).Return(testErr)
//
//	err := iApi.SetMemberDeleted(5, orgId, myId)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiSetMemberDeleted_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("setMemberDeleted", 5, orgId, myId).Return(nil)
//
//	err := iApi.SetMemberDeleted(5, orgId, myId)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiMemberIsOnlyOwner_storeMemberIsOnlyOwnerErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("memberIsOnlyOwner", 5, orgId, myId).Return(true, testErr)
//
//	isOnlyOwner, err := iApi.MemberIsOnlyOwner(5, orgId, myId)
//	assert.True(t, isOnlyOwner)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiMemberIsOnlyOwner_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("memberIsOnlyOwner", 5, orgId, myId).Return(true, nil)
//
//	isOnlyOwner, err := iApi.MemberIsOnlyOwner(5, orgId, myId)
//	assert.True(t, isOnlyOwner)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiRenameMember_storeRenameMemberErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("renameMember", 5, orgId, myId, "ali").Return(testErr)
//
//	err := iApi.RenameMember(5, orgId, myId, "ali")
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiUserIsOrgOwner_storeGetMemberErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(nil, testErr)
//
//	can, err := iApi.UserIsOrgOwner(5, orgId, myId)
//	assert.False(t, can)
//	assert.Equal(t, testErr, err)
//}
//
//func Test_internalApiUserIsOrgOwner_true(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Owner}, nil)
//
//	can, err := iApi.UserIsOrgOwner(5, orgId, myId)
//	assert.True(t, can)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiUserIsOrgOwner_false(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("getMember", 5, orgId, myId).Return(&orgMember{Role: Admin}, nil)
//
//	can, err := iApi.UserIsOrgOwner(5, orgId, myId)
//	assert.False(t, can)
//	assert.Nil(t, err)
//}
//
//func Test_internalApiUserIsOrgOwner_invalidTaskCenterTypeErr(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//
//	can, err := iApi.UserIsOrgOwner(5, myId, myId)
//	assert.False(t, can)
//	assert.Equal(t, invalidTaskCenterTypeErr, err)
//}
//
//func Test_internalApiRenameMember_success(t *testing.T) {
//	store := &mockStore{}
//	iApi := newInternalApi(store, NewLog(nil))
//
//	myId := NewId()
//	orgId := NewId()
//	store.On("renameMember", 5, orgId, myId, "ali").Return(nil)
//
//	err := iApi.RenameMember(5, orgId, myId, "ali")
//	assert.Nil(t, err)
//}

var (
	testErr  = errors.New("test")
	testErr2 = errors.New("test2")
)

type mockSingluarInternalApi struct {
	mock.Mock
}

func (m *mockSingluarInternalApi) CreatePersonalTaskCenter(user Id) int {
	args := m.Called(user)
	return args.Int(0)
}

func (m *mockSingluarInternalApi) CreateOrgTaskCenter(org, owner Id, ownerName string) int {
	args := m.Called(org, owner, ownerName)
	return args.Int(0)
}

func (m *mockSingluarInternalApi) DeleteTaskCenter(shard int, account, owner Id) {
	m.Called(shard, account, owner)
}

func (m *mockSingluarInternalApi) AddMembers(shard int, org, admin Id, members []*AddMemberInternal) {
	m.Called(shard, org, admin, members)
}

func (m *mockSingluarInternalApi) RemoveMembers(shard int, org, admin Id, members []Id) {
	m.Called(shard, org, admin, members)
}

func (m *mockSingluarInternalApi) SetMemberDeleted(shard int, org, member Id) {
	m.Called(shard, org, member)
}

func (m *mockSingluarInternalApi) MemberIsOnlyOwner(shard int, org, member Id) bool {
	args := m.Called(shard, org, member)
	return args.Bool(0)
}

func (m *mockSingluarInternalApi) RenameMember(shard int, org, member Id, newName string) {
	m.Called(shard, org, member, newName)
}

func (m *mockSingluarInternalApi) UserIsOrgOwner(shard int, org, user Id) bool {
	args := m.Called(shard, org, user)
	return args.Bool(0)
}

type mockStore struct {
	mock.Mock
}

func (m *mockStore) registerPersonalAccount(id Id) int {
	args := m.Called(id)
	return args.Int(0)
}

func (m *mockStore) registerOrgAccount(id Id, ownerId Id, ownerName string) int {
	args := m.Called(id, ownerId, ownerName)
	return args.Int(0)
}

func (m *mockStore) deleteAccount(shard int, account Id) {
	m.Called(shard, account)
}

func (m *mockStore) getMember(shard int, org, memberId Id) *orgMember {
	args := m.Called(shard, org, memberId)
	mem := args.Get(0)
	if mem != nil {
		return mem.(*orgMember)
	}
	return nil
}

func (m *mockStore) addMembers(shard int, org Id, members []*AddMemberInternal) {
	m.Called(shard, org, members)
}

func (m *mockStore) setMembersActive(shard int, org Id, members []*AddMemberInternal) {
	m.Called(shard, org, members)
}

func (m *mockStore) getTotalOrgOwnerCount(shard int, org Id) int {
	args := m.Called(shard, org)
	return args.Int(0)
}

func (m *mockStore) getOwnerCountInSet(shard int, org Id, members []Id) int {
	args := m.Called(shard, org, members)
	return args.Int(0)
}

func (m *mockStore) setMembersInactive(shard int, org Id, members []Id) {
	m.Called(shard, org, members)
}

func (m *mockStore) setMemberDeleted(shard int, org Id, member Id) {
	m.Called(shard, org, member)
}

func (m *mockStore) memberIsOnlyOwner(shard int, org, member Id) bool {
	args := m.Called(shard, org, member)
	return args.Bool(0)
}

func (m *mockStore) renameMember(shard int, org Id, member Id, newName string) {
	m.Called(shard, org, member, newName)
}
