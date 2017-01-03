package account

import (
	"errors"
	. "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/zap"
	"testing"
)

func Test_newApi_nilStoreErr(t *testing.T) {
	api, err := newApi(nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilStoreErr)
}

func Test_newApi_nilInternalRegionalApisErr(t *testing.T) {
	store := &mockStore{}
	api, err := newApi(store, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilInternalRegionalApisErr)
}

func Test_newApi_nilLinkMailerErr(t *testing.T) {
	store, internalRegionalApiProvider := &mockStore{}, map[string]internalRegionalApi{}
	api, err := newApi(store, internalRegionalApiProvider, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLinkMailerErr)
}

func Test_newApi_nilLogErr(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}
	api, err := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLogErr)
}

func Test_newApi_success(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, err := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	assert.NotNil(t, api)
	assert.Nil(t, err)
}

func Test_api_Register_invalidNameParam(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, _ := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("a", "email@email.email", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.IsType(t, "name", err.paramPurpose)
}

func Test_api_Register_invalidEmailParam(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, _ := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "invalidEmail", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.Equal(t, "email", err.paramPurpose)
}

func Test_api_Register_invalidPwdParam(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, _ := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "p", "us").(*invalidStringParamErr)
	assert.Equal(t, "password", err.paramPurpose)
}

func Test_api_Register_invalidRegionParam(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, _ := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_Register_storeGetAccountByNameErr(t *testing.T) {
	store, internalRegionalApiProvider, linkMailer, log := &mockStore{}, map[string]internalRegionalApi{"us": nil}, &mockLinkMailer{}, zap.New(zap.NewTextEncoder())
	api, _ := newApi(store, internalRegionalApiProvider, linkMailer, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)
	expectedErr := errors.New("test")
	var nilAccount *account
	store.On("getAccountByName", "ali").Return(nilAccount, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

//helpers

type mockStore struct {
	mock.Mock
}

func (m *mockStore) getAccountByName(name string) (*account, error) {
	args := m.Called(name)
	return args.Get(0).(*account), args.Error(1)
}

func (m *mockStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	args := m.Called(user, pwdInfo)
	return args.Error(0)
}

func (m *mockStore) getUserByName(name string) (*fullUserInfo, error) {
	args := m.Called(name)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByEmail(email string) (*fullUserInfo, error) {
	args := m.Called(email)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserById(id UUID) (*fullUserInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByActivationCode(activationCode string) (*fullUserInfo, error) {
	args := m.Called(activationCode)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error) {
	args := m.Called(confirmationCode)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByResetPwdCode(resetPwdCode string) (*fullUserInfo, error) {
	args := m.Called(resetPwdCode)
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getPwdInfo(id UUID) (*pwdInfo, error) {
	args := m.Called(id)
	return args.Get(0).(*pwdInfo), args.Error(1)
}

func (m *mockStore) updateUser(user *fullUserInfo) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockStore) updatePwdInfo(id UUID, pwdInfo *pwdInfo) error {
	args := m.Called(id, pwdInfo)
	return args.Error(0)
}

func (m *mockStore) deleteUser(id UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockStore) getUsers(ids []UUID) ([]*user, error) {
	args := m.Called(ids)
	return args.Get(0).([]*user), args.Error(1)
}

func (m *mockStore) searchUsers(search string, limit int) ([]*user, error) {
	args := m.Called(search, limit)
	return args.Get(0).([]*user), args.Error(1)
}

func (m *mockStore) createOrg(org *org) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *mockStore) getOrgById(id UUID) (*org, error) {
	args := m.Called(id)
	return args.Get(0).(*org), args.Error(1)
}

func (m *mockStore) getOrgByName(name string) (*org, error) {
	args := m.Called(name)
	return args.Get(0).(*org), args.Error(1)
}

func (m *mockStore) updateOrg(org *org) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *mockStore) deleteOrg(id UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockStore) getOrgs(ids []UUID) ([]*org, error) {
	args := m.Called(ids)
	return args.Get(0).([]*org), args.Error(1)
}

func (m *mockStore) searchOrgs(search string, limit int) ([]*org, error) {
	args := m.Called(search, limit)
	return args.Get(0).([]*org), args.Error(1)
}

func (m *mockStore) getUsersOrgs(userId UUID, limit int) ([]*org, error) {
	args := m.Called(userId, limit)
	return args.Get(0).([]*org), args.Error(1)
}

type mockLinkMailer struct {
	mock.Mock
}

func (m *mockLinkMailer) sendActivationLink(address, activationCode string) error {
	args := m.Called(address, activationCode)
	return args.Error(0)
}

func (m *mockLinkMailer) sendPwdResetLink(address, resetCode string) error {
	args := m.Called(address, resetCode)
	return args.Error(0)
}

func (m *mockLinkMailer) sendNewEmailConfirmationLink(address, confirmationCode string) error {
	args := m.Called(address, confirmationCode)
	return args.Error(0)
}
