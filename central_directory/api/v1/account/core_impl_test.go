package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bytes"
	"errors"
	"github.com/stretchr/testify/mock"
	"io"
	"time"
)

////helpers
var (
	testErr            = errors.New("test")
	timeNowReplacement = time.Now().UTC()
)

type mockStore struct {
	mock.Mock
}

func (m *mockStore) accountWithCiNameExists(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *mockStore) getAccountByCiName(name string) *account {
	args := m.Called(name)
	acc := args.Get(0)
	if acc == nil {
		return nil
	}
	return acc.(*account)
}

func (m *mockStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) {
	user.CreatedOn = timeNowReplacement
	m.Called(user, pwdInfo)
}

func (m *mockStore) getUserByEmail(email string) *fullUserInfo {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*fullUserInfo)
}

func (m *mockStore) getUserById(id Id) *fullUserInfo {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*fullUserInfo)
}

func (m *mockStore) getPwdInfo(id Id) *pwdInfo {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*pwdInfo)
}

func (m *mockStore) updateUser(user *fullUserInfo) {
	m.Called(user)
}

func (m *mockStore) updateAccount(acc *account) {
	m.Called(acc)
}

func (m *mockStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) {
	m.Called(id, pwdInfo)
}

func (m *mockStore) deleteAccountAndAllAssociatedMemberships(id Id) {
	m.Called(id)
}

func (m *mockStore) getAccount(id Id) *account {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*account)
}

func (m *mockStore) getAccounts(ids []Id) []*account {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*account)
}

func (m *mockStore) getUsers(ids []Id) []*account {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*account)
}

func (m *mockStore) createOrgAndMembership(org *account, user Id) {
	org.CreatedOn = timeNowReplacement
	m.Called(org, user)
}

func (m *mockStore) getUsersOrgs(userId Id, offset, limit int) ([]*account, int) {
	args := m.Called(userId, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1)
	}
	return args.Get(0).([]*account), args.Int(1)
}

func (m *mockStore) createMemberships(org Id, users []Id) {
	m.Called(org, users)
}

func (m *mockStore) deleteMemberships(org Id, users []Id) {
	m.Called(org, users)
}

type mockLinkMailer struct {
	mock.Mock
}

func (m *mockLinkMailer) sendMultipleAccountPolicyEmail(address string) {
	m.Called(address)
}

func (m *mockLinkMailer) sendActivationLink(address, activationCode string) {
	m.Called(address, activationCode)
}

func (m *mockLinkMailer) sendPwdResetLink(address, resetCode string) {
	m.Called(address, resetCode)
}

func (m *mockLinkMailer) sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string) {
	m.Called(currentEmail, newEmail, confirmationCode)
}

type mockAvatarStore struct {
	mock.Mock
}

func (m *mockAvatarStore) put(key string, mimeType string, size int64, data io.ReadSeeker) {
	m.Called(key, mimeType, 5, testReadSeaker)
}

func (m *mockAvatarStore) delete(key string) {
	m.Called(key)
}

type mockInternalRegionClient struct {
	mock.Mock
}

func (m *mockInternalRegionClient) GetRegions() []string {
	args := m.Called()
	regions := args.Get(0)
	if regions != nil {
		return regions.([]string)
	}
	return nil
}

func (m *mockInternalRegionClient) IsValidRegion(region string) bool {
	args := m.Called(region)
	return args.Bool(0)
}

func (m *mockInternalRegionClient) CreatePersonalTaskCenter(region string, userId Id) int {
	args := m.Called(region, userId)
	return args.Int(0)
}

func (m *mockInternalRegionClient) CreateOrgTaskCenter(region string, orgId, ownerId Id, ownerName string) int {
	args := m.Called(region, orgId, ownerId, ownerName)
	return args.Int(0)
}

func (m *mockInternalRegionClient) DeleteTaskCenter(region string, shard int, account, owner Id) {
	m.Called(region, shard, account, owner)
}

func (m *mockInternalRegionClient) AddMembers(region string, shard int, org, admin Id, members []*AddMemberInternal) {
	m.Called(region, shard, org, admin, members)
}

func (m *mockInternalRegionClient) RemoveMembers(region string, shard int, org, admin Id, members []Id) {
	m.Called(region, shard, org, admin, members)
}

func (m *mockInternalRegionClient) MemberIsOnlyOwner(region string, shard int, org, member Id) bool {
	args := m.Called(region, shard, org, member)
	return args.Bool(0)
}

func (m *mockInternalRegionClient) RenameMember(region string, shard int, org, member Id, newName string) {
	m.Called(region, shard, org, member, newName)
}

func (m *mockInternalRegionClient) UserIsOrgOwner(region string, shard int, org, user Id) bool {
	args := m.Called(region, shard, org, user)
	return args.Bool(0)
}

type mockMiscFuncs struct {
	mock.Mock
}

func (m *mockMiscFuncs) newCreatedNamedEntity(name string) *CreatedNamedEntity {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*CreatedNamedEntity)
}

type mockCryptoHelper struct {
	mock.Mock
}

func (m *mockCryptoHelper) Bytes(length int) []byte {
	args := m.Called(length)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]byte)
}

func (m *mockCryptoHelper) UrlSafeString(length int) string {
	args := m.Called(length)
	return args.String(0)
}

func (m *mockCryptoHelper) ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte {
	args := m.Called(password, salt, N, r, p, keyLen)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]byte)
}

var testReadSeaker = bytes.NewReader(nil)
