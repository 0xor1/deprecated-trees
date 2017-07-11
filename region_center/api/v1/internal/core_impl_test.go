package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"errors"
	"github.com/stretchr/testify/mock"
)

var (
	testErr  = errors.New("test")
	testErr2 = errors.New("test2")
)

type mockSingluarInternalApi struct {
	mock.Mock
}

func (m *mockSingluarInternalApi) CreateTaskCenter(org, owner Id, ownerName string) int {
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

func (m *mockStore) registerAccount(id Id, ownerId Id, ownerName string) int {
	args := m.Called(id, ownerId, ownerName)
	return args.Int(0)
}

func (m *mockStore) deleteAccount(shard int, account Id) {
	m.Called(shard, account)
}

func (m *mockStore) getMember(shard int, org, memberId Id) *Member {
	args := m.Called(shard, org, memberId)
	mem := args.Get(0)
	if mem != nil {
		return mem.(*Member)
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
