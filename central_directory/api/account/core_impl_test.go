package account

import (
	"bitbucket.org/robsix/task_center/misc"
	"errors"
	. "github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_newApi_nilStoreErr(t *testing.T) {
	api, err := newApi(nil, nil, nil, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilStoreErr)
}

func Test_newApi_nilinternalRegionApisErr(t *testing.T) {
	store := &mockStore{}
	api, err := newApi(store, nil, nil, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilInternalRegionApisErr)
}

func Test_newApi_nilLinkMailerErr(t *testing.T) {
	store, internalRegionApis := &mockStore{}, map[string]internalRegionApi{}
	api, err := newApi(store, internalRegionApis, nil, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLinkMailerErr)
}

func Test_newApi_nilGenNewIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}
	api, err := newApi(store, internalRegionApis, linkMailer, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilGenNewIdErr)
}

func Test_newApi_nilGenCryptoBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilGenCryptoBytesErr)
}

func Test_newApi_nilGenCryptoUrlSafeStringErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilGenCryptoUrlSafeStringErr)
}

func Test_newApi_nilGenScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilGenScryptKeyErr)
}

func Test_newApi_nilLogErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLogErr)
}

func Test_newApi_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	assert.NotNil(t, api)
	assert.Nil(t, err)
}

func Test_api_Register_invalidNameParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("a", "email@email.email", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.IsType(t, "name", err.paramPurpose)
	assert.Equal(t, "name must be between 3 and 20 utf8 characters long and match all regexs []", err.Error())
}

func Test_api_Register_invalidEmailParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "invalidEmail", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.Equal(t, "email", err.paramPurpose)
}

func Test_api_Register_invalidPwdParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "p", "us").(*invalidStringParamErr)
	assert.Equal(t, "password", err.paramPurpose)
}

func Test_api_Register_invalidRegionParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_Register_storeGetAccountByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_storeGetAccountByNameNoneNilAccount(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(&account{}, nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_Register_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{}, nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, emailAlreadyInUseErr, err)
}

func Test_api_Register_genCryptoBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	miscFuncs.On("GenCryptoBytes", 128).Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_genScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), salt, 16384, 8, 1, 32).Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_genCryptoUrlSafeStringErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	miscFuncs.On("GenCryptoUrlSafeString", 40).Return("", expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_genNewIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	miscFuncs.On("GenCryptoUrlSafeString", 40).Return("test", nil)
	miscFuncs.On("GenNewId").Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_storeCreateNewUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	miscFuncs.On("GenCryptoUrlSafeString", 40).Return(activationCode, nil)
	id := NewUUID()
	miscFuncs.On("GenNewId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: misc.Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
				},
				Email: "email@email.com",
			},
			ActivationCode: &activationCode,
		},
		&pwdInfo{
			Salt:   salt,
			Pwd:    pwd,
			N:      16384,
			R:      8,
			P:      1,
			KeyLen: 32,
		}).Return(expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	miscFuncs.On("GenCryptoUrlSafeString", 40).Return(activationCode, nil)
	id := NewUUID()
	miscFuncs.On("GenNewId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: misc.Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
				},
				Email: "email@email.com",
			},
			ActivationCode: &activationCode,
		},
		&pwdInfo{
			Salt:   salt,
			Pwd:    pwd,
			N:      16384,
			R:      8,
			P:      1,
			KeyLen: 32,
		}).Return(nil)
	linkMailer.On("sendActivationLink", "email@email.com", activationCode).Return(expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, expectedErr, err)
}

func Test_api_Register_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	miscFuncs.On("GenCryptoBytes", 128).Return(salt, nil)
	miscFuncs.On("GenScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	miscFuncs.On("GenCryptoUrlSafeString", 40).Return(activationCode, nil)
	id := NewUUID()
	miscFuncs.On("GenNewId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: misc.Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
				},
				Email: "email@email.com",
			},
			ActivationCode: &activationCode,
		},
		&pwdInfo{
			Salt:   salt,
			Pwd:    pwd,
			N:      16384,
			R:      8,
			P:      1,
			KeyLen: 32,
		}).Return(nil)
	linkMailer.On("sendActivationLink", "email@email.com", activationCode).Return(nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Nil(t, err)
}

//helpers
var (
	expectedErr        = errors.New("test")
	timeNowReplacement = time.Now().UTC()
)

type mockStore struct {
	mock.Mock
}

func (m *mockStore) getAccountByName(name string) (*account, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account), args.Error(1)
}

func (m *mockStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	user.Created = timeNowReplacement
	args := m.Called(user, pwdInfo)
	return args.Error(0)
}

func (m *mockStore) getUserByName(name string) (*fullUserInfo, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByEmail(email string) (*fullUserInfo, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserById(id UUID) (*fullUserInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByActivationCode(activationCode string) (*fullUserInfo, error) {
	args := m.Called(activationCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByNewEmailConfirmationCode(confirmationCode string) (*fullUserInfo, error) {
	args := m.Called(confirmationCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getUserByResetPwdCode(resetPwdCode string) (*fullUserInfo, error) {
	args := m.Called(resetPwdCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getPwdInfo(id UUID) (*pwdInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user), args.Error(1)
}

func (m *mockStore) searchUsers(search string, limit int) ([]*user, error) {
	args := m.Called(search, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user), args.Error(1)
}

func (m *mockStore) createOrg(org *org) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *mockStore) getOrgById(id UUID) (*org, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*org), args.Error(1)
}

func (m *mockStore) getOrgByName(name string) (*org, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*org), args.Error(1)
}

func (m *mockStore) searchOrgs(search string, limit int) ([]*org, error) {
	args := m.Called(search, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*org), args.Error(1)
}

func (m *mockStore) getUsersOrgs(userId UUID, limit int) ([]*org, error) {
	args := m.Called(userId, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

type mockMiscFuncs struct {
	mock.Mock
}

func (m *mockMiscFuncs) GenNewId() (UUID, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(UUID), args.Error(1)
}

func (m *mockMiscFuncs) GenCryptoBytes(length int) ([]byte, error) {
	args := m.Called(length)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockMiscFuncs) GenCryptoUrlSafeString(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *mockMiscFuncs) GenScryptKey(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	args := m.Called(password, salt, N, r, p, keyLen)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
