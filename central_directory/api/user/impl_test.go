package user

import (
	. "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"github.com/uber-go/zap"
)

func Test_NewApi_NilUserStoreErr(t *testing.T) {
	api, err := NewApi(nil, nil, nil, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, nil)
	assert.Nil(t, api)
	assert.Equal(t, err, NilUserStoreErr)
}

func Test_NewApi_NilPwdStoreErr(t *testing.T) {
	api, err := NewApi(&mockUserStore{}, nil, nil, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, nil)
	assert.Nil(t, api)
	assert.Equal(t, err, NilPwdStoreErr)
}

func Test_NewApi_NilLinkMailerErr(t *testing.T) {
	api, err := NewApi(&mockUserStore{}, &mockPwdStore{}, nil, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, nil)
	assert.Nil(t, api)
	assert.Equal(t, err, NilLinkMailerErr)
}

func Test_NewApi_NilLogErr(t *testing.T) {
	api, err := NewApi(&mockUserStore{}, &mockPwdStore{}, &mockLinkMailer{}, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, nil)
	assert.Nil(t, api)
	assert.Equal(t, err, NilLogErr)
}

func Test_NewApi_Success(t *testing.T) {
	api, err := NewApi(&mockUserStore{}, &mockPwdStore{}, &mockLinkMailer{}, nil, nil, 3, 20, 3, 20, 3, 100, 40, 128, 16384, 8, 1, 32, zap.New(zap.NewTextEncoder()))
	assert.NotNil(t, api)
	assert.Nil(t, err)
}

//helpers

type mockUserStore struct {
	mock.Mock
}

func (m *mockUserStore) Create(user *FullUserInfo) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserStore) GetByUsername(username string) (*FullUserInfo, error) {
	args := m.Called(username)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetByEmail(email string) (*FullUserInfo, error) {
	args := m.Called(email)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetById(id UUID) (*FullUserInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetByActivationCode(activationCode string) (*FullUserInfo, error) {
	args := m.Called(activationCode)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetByNewEmailConfirmationCode(confirmationCode string) (*FullUserInfo, error) {
	args := m.Called(confirmationCode)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetByResetPwdCode(resetPwdCode string) (*FullUserInfo, error) {
	args := m.Called(resetPwdCode)
	return args.Get(0).(*FullUserInfo), args.Error(1)
}

func (m *mockUserStore) GetByIds(ids []UUID) ([]*User, error) {
	args := m.Called(ids)
	return args.Get(0).([]*User), args.Error(1)
}

func (m *mockUserStore) Search(search string, limit int) ([]*User, error) {
	args := m.Called(search)
	return args.Get(0).([]*User), args.Error(1)
}

func (m *mockUserStore) Update(user *FullUserInfo) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserStore) Delete(id UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockPwdStore struct {
	mock.Mock
}

func (m *mockPwdStore) Create(userId UUID, pwdInfo *PwdInfo) error {
	args := m.Called(userId, pwdInfo)
	return args.Error(0)
}

func (m *mockPwdStore) Get(userId UUID) (*PwdInfo, error) {
	args := m.Called(userId)
	return args.Get(0).(*PwdInfo), args.Error(1)
}

func (m *mockPwdStore) Update(userId UUID, pwdInfo *PwdInfo) error {
	args := m.Called(userId, pwdInfo)
	return args.Error(0)
}

func (m *mockPwdStore) Delete(userId UUID) error {
	args := m.Called(userId)
	return args.Error(0)
}

type mockLinkMailer struct {
	mock.Mock
}

func (m *mockLinkMailer) SendActivationLink(address, activationCode string) error {
	args := m.Called(address, activationCode)
	return args.Error(0)
}

func (m *mockLinkMailer) SendPwdResetLink(address, resetCode string) error {
	args := m.Called(address, resetCode)
	return args.Error(0)
}

func (m *mockLinkMailer) SendNewEmailConfirmationLink(address, confirmationCode string) error {
	args := m.Called(address, confirmationCode)
	return args.Error(0)
}
