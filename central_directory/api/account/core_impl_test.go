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
	assert.Equal(t, "name", err.paramPurpose)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Nil(t, err)
}

func Test_api_Register_genCryptoBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	miscFuncs.On("GenCryptoBytes", 128).Return(nil, expectedErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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
	assert.IsType(t, &misc.ErrorRef{}, err)
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

func Test_api_ResendActivationEmail_storeGetUseByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, expectedErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailActivatedUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activated := time.Now().UTC()
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{Activated: &activated}, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(expectedErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_Activate_storeGetUserByActivationCodeErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByActivationCode", "test").Return(nil, expectedErr)

	id, err := api.Activate("test")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Activate_storeGetUserByActivationCodeNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByActivationCode", "test").Return(nil, nil)

	id, err := api.Activate("test")
	assert.Nil(t, id)
	assert.Equal(t, noSuchActivationCodeErr, err)
}

func Test_api_Activate_invalidUserRegion(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByActivationCode", "test").Return(&fullUserInfo{me: me{user: user{Region: "us"}}}, nil)

	id, err := api.Activate("test")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Activate_internalRegionalApiCreatePersonalTaskCenterErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := misc.NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: misc.Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByActivationCode", activationCode).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(-1, expectedErr)

	id, err := api.Activate("test")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Activate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := misc.NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: misc.Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByActivationCode", activationCode).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(3, nil)
	store.On("updateUser", user).Return(expectedErr)

	id, err := api.Activate("test")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Activate_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := misc.NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: misc.Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByActivationCode", activationCode).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(3, nil)
	store.On("updateUser", user).Return(nil)

	id, err := api.Activate("test")
	assert.Equal(t, userId, id)
	assert.Nil(t, err)
	assert.Nil(t, user.ActivationCode)
	assert.NotNil(t, user.Activated)
	assert.Equal(t, user.Shard, 3)
}

func Test_api_Authenticate_storeGetUserByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByName", "name").Return(nil, expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_storeGetUserByNameNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByName", "name").Return(nil, nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetPwdInfoErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(nil, expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_genScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return(nil, expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_incorrectPwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("N0t-P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetUserByName_userNotActivatedErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	user := &fullUserInfo{}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, userNotActivated, err)
}

func Test_api_Authenticate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_genCryptoBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	miscFuncs.On("GenCryptoBytes", 128).Return(nil, expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_genScryptKey2Err(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", UUID(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	miscFuncs.On("GenCryptoBytes", 128).Return([]byte("test"), nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return(nil, expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_storeUpdatePwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{Pwd: []byte("P@ss-W0rd")}
	store.On("getPwdInfo", UUID(nil)).Return(pwdInfo, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	miscFuncs.On("GenCryptoBytes", 128).Return([]byte("test"), nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return([]byte("P@ss-W0rd"), nil)
	store.On("updatePwdInfo", UUID(nil), pwdInfo).Return(expectedErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_Authenticate_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	id, _ := misc.NewId()
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode, me: me{user: user{Entity: misc.Entity{Id: id}}}}
	store.On("getUserByName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{Pwd: []byte("P@ss-W0rd")}
	store.On("getPwdInfo", id).Return(pwdInfo, nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	miscFuncs.On("GenCryptoBytes", 128).Return([]byte("test"), nil)
	miscFuncs.On("GenScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return([]byte("P@ss-W0rd"), nil)
	store.On("updatePwdInfo", id, pwdInfo).Return(nil)

	resultId, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Equal(t, id, resultId)
	assert.Nil(t, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByNewEmailConfirmationCodeErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByNewEmailConfirmationCode", "confirmationCode").Return(nil, expectedErr)

	resultId, err := api.ConfirmNewEmail("email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByNewEmailConfirmationCodeNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByNewEmailConfirmationCode", "confirmationCode").Return(nil, nil)

	resultId, err := api.ConfirmNewEmail("email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.Equal(t, noSuchNewEmailConfirmationCodeErr, err)
}

func Test_api_ConfirmNewEmail_mismatchInputs(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail}}
	store.On("getUserByNewEmailConfirmationCode", "not-confirmationCode").Return(user, nil)

	resultId, err := api.ConfirmNewEmail("not-new@email.com", "not-confirmationCode")
	assert.Nil(t, resultId)
	assert.Equal(t, newEmailConfirmationErr, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail}}
	store.On("getUserByNewEmailConfirmationCode", confirmationCode).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, expectedErr)

	resultId, err := api.ConfirmNewEmail(newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmailNoneNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail}}
	store.On("getUserByNewEmailConfirmationCode", confirmationCode).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(&fullUserInfo{}, nil)

	resultId, err := api.ConfirmNewEmail(newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.Equal(t, emailAlreadyInUseErr, err)
}

func Test_api_ConfirmNewEmail_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail}}
	store.On("getUserByNewEmailConfirmationCode", confirmationCode).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(expectedErr)

	resultId, err := api.ConfirmNewEmail(newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &misc.ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	id, _ := misc.NewId()
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, user: user{Entity: misc.Entity{Id: id}}}}
	store.On("getUserByNewEmailConfirmationCode", confirmationCode).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(nil)

	resultId, err := api.ConfirmNewEmail(newEmail, confirmationCode)
	assert.Equal(t, id, resultId)
	assert.Nil(t, err)
	assert.Equal(t, newEmail, user.Email)
	assert.Nil(t, user.NewEmail)
	assert.Nil(t, user.NewEmailConfirmationCode)
}

//func Test_api_ResetPwd_storeGetUserByEmailErr(t *testing.T) {
//	store, internalRegionApis, linkMailer, miscFuncs, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, misc.NewLog(nil)
//	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.GenNewId, miscFuncs.GenCryptoBytes, miscFuncs.GenCryptoUrlSafeString, miscFuncs.GenScryptKey, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)
//
//	id, _ := misc.NewId()
//	user := &fullUserInfo{me: me{user: user{Entity: misc.Entity{Id: id}}}}
//	store.On("getUserByEmail", newEmail).Return(nil, nil)
//	store.On("updateUser", user).Return(nil)
//
//	resultId, err := api.ConfirmNewEmail(newEmail, confirmationCode)
//	assert.Equal(t, id, resultId)
//	assert.Nil(t, err)
//	assert.Equal(t, newEmail, user.Email)
//	assert.Nil(t, user.NewEmail)
//	assert.Nil(t, user.NewEmailConfirmationCode)
//}

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

func (m *mockLinkMailer) sendMultipleAccountPolicyEmail(address string) error {
	args := m.Called(address)
	return args.Error(0)
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

type mockInternalRegionApi struct {
	mock.Mock
}

func (m *mockInternalRegionApi) CreatePersonalTaskCenter(userId UUID) (int, error) {
	args := m.Called(userId)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) CreateOrgTaskCenter(ownerId, orgId UUID) (int, error) {
	args := m.Called(ownerId, orgId)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) RenameMember(memberId, orgId UUID, newName string) error {
	args := m.Called(memberId, orgId, newName)
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
