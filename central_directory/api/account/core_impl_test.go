package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_newApi_nilStoreErr(t *testing.T) {
	api, err := newApi(nil, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilStoreErr)
}

func Test_newApi_nilinternalRegionApisErr(t *testing.T) {
	store := &mockStore{}
	api, err := newApi(store, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilInternalRegionApisErr)
}

func Test_newApi_nilLinkMailerErr(t *testing.T) {
	store, internalRegionApis := &mockStore{}, map[string]internalRegionApi{}
	api, err := newApi(store, internalRegionApis, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLinkMailerErr)
}

func Test_newApi_nilnewIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}
	api, err := newApi(store, internalRegionApis, linkMailer, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilNewIdErr)
}

func Test_newApi_nilcryptoHelperErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilCryptoHelperErr)
}

func Test_newApi_nilLogErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)

	assert.Nil(t, api)
	assert.Equal(t, err, nilLogErr)
}

func Test_newApi_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, err := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	assert.NotNil(t, api)
	assert.Nil(t, err)
}

func Test_api_Register_invalidNameParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("a", "email@email.email", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.Equal(t, "name", err.paramPurpose)
	assert.Equal(t, "name must be between 3 and 20 utf8 characters long and match all regexs []", err.Error())
}

func Test_api_Register_invalidEmailParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "invalidEmail", "P@ss-W0rd", "us").(*invalidStringParamErr)
	assert.Equal(t, "email", err.paramPurpose)
}

func Test_api_Register_invalidPwdParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "p", "us").(*invalidStringParamErr)
	assert.Equal(t, "password", err.paramPurpose)
}

func Test_api_Register_invalidRegionParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_Register_storeGetAccountByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeGetAccountByNameNoneNilAccount(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(&account{}, nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_Register_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Nil(t, err)
}

func Test_api_Register_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), salt, 16384, 8, 1, 32).Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("", testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_gennewIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("test", nil)
	miscFuncs.On("newId").Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeCreateNewUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	cryptoHelper.On("UrlSafeString", 40).Return(activationCode, nil)
	id, _ := NewId()
	miscFuncs.On("newId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser: true,
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
		}).Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	cryptoHelper.On("UrlSafeString", 40).Return(activationCode, nil)
	id, _ := NewId()
	miscFuncs.On("newId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser: true,
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
	linkMailer.On("sendActivationLink", "email@email.com", activationCode).Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": nil}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByName", "ali").Return(nil, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	pwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", pwd, salt, 16384, 8, 1, 32).Return(pwd, nil)
	activationCode := "test"
	cryptoHelper.On("UrlSafeString", 40).Return(activationCode, nil)
	id, _ := NewId()
	miscFuncs.On("newId").Return(id, nil)
	store.On(
		"createUser",
		&fullUserInfo{
			me: me{
				user: user{
					Entity: Entity{
						Id: id,
					},
					Name:    "ali",
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser: true,
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
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailActivatedUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activated := time.Now().UTC()
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{Activated: &activated}, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(testErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_Activate_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_storeGetUserByActivationCodeNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.Equal(t, invalidActivationAttemptErr, err)
}

func Test_api_Activate_storeGetUserByActivationCode_activationCodeMismatch(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationCode := "not-the-activation-code"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &activationCode}, nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.Equal(t, invalidActivationAttemptErr, err)
}

func Test_api_Activate_invalidUserRegion(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationCode := "activationCode"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ActivationCode: &activationCode, me: me{user: user{Region: "us"}}}, nil)

	id, err := api.Activate("email@email.com", activationCode)
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_internalRegionalApiCreatePersonalTaskCenterErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(-1, testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(3, nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				Entity: Entity{
					Id: userId,
				},
				Shard: -1,
			},
		},
		ActivationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreatePersonalTaskCenter", userId).Return(3, nil)
	store.On("updateUser", user).Return(nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Equal(t, userId, id)
	assert.Nil(t, err)
	assert.Nil(t, user.ActivationCode)
	assert.NotNil(t, user.Activated)
	assert.Equal(t, user.Shard, 3)
}

func Test_api_Authenticate_storeGetUserByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByName", "name").Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_storeGetUserByNameNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByName", "name").Return(nil, nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetPwdInfoErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_incorrectPwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{Activated: &activationTime}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("N0t-P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetUserByName_userNotActivatedErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	user := &fullUserInfo{}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, userNotActivatedErr, err)
}

func Test_api_Authenticate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_genScryptKey2Err(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{Pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return([]byte("test"), nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_storeUpdatePwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode}
	store.On("getUserByName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{Pwd: []byte("P@ss-W0rd")}
	store.On("getPwdInfo", Id(nil)).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return([]byte("test"), nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return([]byte("P@ss-W0rd"), nil)
	store.On("updatePwdInfo", Id(nil), pwdInfo).Return(testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	id, _ := NewId()
	user := &fullUserInfo{Activated: &activationTime, ResetPwdCode: &resetPwdCode, me: me{user: user{Entity: Entity{Id: id}}}}
	store.On("getUserByName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{Pwd: []byte("P@ss-W0rd")}
	store.On("getPwdInfo", id).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return([]byte("test"), nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return([]byte("P@ss-W0rd"), nil)
	store.On("updatePwdInfo", id, pwdInfo).Return(nil)

	resultId, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Equal(t, id, resultId)
	assert.Nil(t, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "currentEmail@currentEmail.com").Return(nil, testErr)

	resultId, err := api.ConfirmNewEmail("currentEmail@currentEmail.com", "email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmaiNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "currentEmail@currentEmail.com").Return(nil, nil)

	resultId, err := api.ConfirmNewEmail("currentEmail@currentEmail.com", "email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_mismatchNewEmail(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, "not-new@email.com", confirmationCode)
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_mismatchConfirmationCode(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, "not-confirmation-code")
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmail2Err(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, testErr)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmail2NoneNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(&fullUserInfo{}, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.Equal(t, emailAlreadyInUseErr, err)
}

func Test_api_ConfirmNewEmail_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(testErr)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	id, _ := NewId()
	user := &fullUserInfo{NewEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail, user: user{Entity: Entity{Id: id}}}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Equal(t, id, resultId)
	assert.Nil(t, err)
	assert.Equal(t, newEmail, user.Email)
	assert.Nil(t, user.NewEmail)
	assert.Nil(t, user.NewEmailConfirmationCode)
}

func Test_api_ResetPwd_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(nil, testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_storeGetUserByEmailNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(nil, nil)

	err := api.ResetPwd(email)
	assert.Nil(t, err)
}

func Test_api_ResetPwd_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(&fullUserInfo{}, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("", testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	user := &fullUserInfo{}
	store.On("getUserByEmail", email).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("test", nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_linkMailerSendPwdResetLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	user := &fullUserInfo{}
	store.On("getUserByEmail", email).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("test", nil)
	store.On("updateUser", user).Return(nil)
	linkMailer.On("sendPwdResetLink", email, "test").Return(testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	user := &fullUserInfo{}
	store.On("getUserByEmail", email).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("test", nil)
	store.On("updateUser", user).Return(nil)
	linkMailer.On("sendPwdResetLink", email, "test").Return(nil)

	err := api.ResetPwd(email)
	assert.Nil(t, err)
}

func Test_api_SetNewPwdFromPwdReset_invalidNewPwd(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, err := api.SetNewPwdFromPwdReset("yo", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &invalidStringParamErr{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmailNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.Equal(t, invalidResetPwdAttemptErr, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmail_resetPwdCodeMismatch(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "not-the-correct-code"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{ResetPwdCode: &resetCode}, nil)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.Equal(t, invalidResetPwdAttemptErr, err)
}

func Test_api_SetNewPwdFromPwdReset_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{ResetPwdCode: &resetCode}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{ResetPwdCode: &resetCode}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{ResetPwdCode: &resetCode, ActivationCode: &resetCode}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
	assert.Nil(t, user.ResetPwdCode)
	assert.Nil(t, user.ActivationCode)
}

func Test_api_SetNewPwdFromPwdReset_storeUpdatePwdInfoErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{ResetPwdCode: &resetCode, ActivationCode: &resetCode}
	user.Id, _ = NewId()
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updateUser", user).Return(nil)
	store.On("updatePwdInfo", user.Id, &pwdInfo{Pwd: newPwd, Salt: salt, N: 16384, R: 8, P: 1, KeyLen: 32}).Return(testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{ResetPwdCode: &resetCode, ActivationCode: &resetCode}
	user.Id, _ = NewId()
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updateUser", user).Return(nil)
	store.On("updatePwdInfo", user.Id, &pwdInfo{Pwd: newPwd, Salt: salt, N: 16384, R: 8, P: 1, KeyLen: 32}).Return(nil)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.NotNil(t, id)
	assert.Nil(t, err)
	assert.Nil(t, user.ResetPwdCode)
	assert.Nil(t, user.ActivationCode)
}

func Test_api_GetUsers_storeGetUsersErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	store.On("getUsers", ids).Return(nil, testErr)

	users, err := api.GetUsers(ids)
	assert.Nil(t, users)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetUsers_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	res := []*user{}
	store.On("getUsers", ids).Return(res, nil)

	users, err := api.GetUsers(ids)
	assert.Equal(t, res, users)
	assert.Nil(t, err)
}

func Test_api_SearchUsers_invalidStringParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	users, total, err := api.SearchUsers("yo", 0, 5)
	assert.Nil(t, users)
	assert.IsType(t, &invalidStringParamErr{}, err)
	assert.Equal(t, 0, total)
}

func Test_api_SearchUsers_storeSearchUsersErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("searchUsers", "test", 0, 5).Return(nil, 0, testErr)

	users, total, err := api.SearchUsers("test", 0, 5)
	assert.Nil(t, users)
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, 0, total)
}

func Test_api_SearchUsers_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	res := []*user{}
	store.On("searchUsers", "test", 0, 100).Return(res, 8, nil)

	users, total, err := api.SearchUsers("test", -1, -1)
	assert.Equal(t, res, users)
	assert.Nil(t, err)
	assert.Equal(t, 8, total)
}

func Test_api_GetOrgs_storeGetOrgsErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	store.On("getOrgs", ids).Return(nil, testErr)

	orgs, err := api.GetOrgs(ids)
	assert.Nil(t, orgs)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetOrgs_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	res := []*org{}
	store.On("getOrgs", ids).Return(res, nil)

	orgs, err := api.GetOrgs(ids)
	assert.Equal(t, res, orgs)
	assert.Nil(t, err)
}

func Test_api_SearchOrgs_invalidStringParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	orgs, total, err := api.SearchOrgs("yo", 0, 5)
	assert.Nil(t, orgs)
	assert.IsType(t, &invalidStringParamErr{}, err)
	assert.Equal(t, 0, total)
}

func Test_api_SearchOrgs_storeSearchOrgsErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("searchOrgs", "test", 0, 5).Return(nil, 0, testErr)

	orgs, total, err := api.SearchOrgs("test", 0, 5)
	assert.Nil(t, orgs)
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, 0, total)
}

func Test_api_SearchOrgs_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	res := []*org{}
	store.On("searchOrgs", "test", 0, 100).Return(res, 8, nil)

	orgs, total, err := api.SearchOrgs("test", -1, -1)
	assert.Equal(t, res, orgs)
	assert.Nil(t, err)
	assert.Equal(t, 8, total)
}

func Test_api_ChangeMyName_invalidStringParamErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()

	err := api.ChangeMyName(id, "yo")
	assert.IsType(t, &invalidStringParamErr{}, err)
}

func Test_api_ChangeMyName_storeGetUserByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("getUserByName", "test").Return(nil, testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_storeGetUserByNameNoneNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("getUserByName", "test").Return(&fullUserInfo{}, nil)

	err := api.ChangeMyName(id, "test")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_ChangeMyName_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("getUserByName", "test").Return(nil, nil)
	store.On("getUserById", id).Return(nil, testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("getUserByName", "test").Return(nil, nil)
	store.On("getUserById", id).Return(nil, nil)

	err := api.ChangeMyName(id, "test")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyName_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByName", "test").Return(nil, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, "test", user.Name)
}

func Test_api_ChangeMyName_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByName", "test").Return(nil, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(nil)

	err := api.ChangeMyName(id, "test")
	assert.Nil(t, err)
}

func Test_api_ChangeMyPwd_invalidStringParam(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "yo")
	assert.IsType(t, &invalidStringParamErr{}, err)
}

func Test_api_ChangeMyPwd_storeGetPwdInfoErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getPwdInfo", myId).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_storeGetPwdInfoNilPwdInfo(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getPwdInfo", myId).Return(nil, nil)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyPwd_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	pwdInfo := &pwdInfo{}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("0ld-P@ss-W0rd"), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_incorrectPwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	pwdInfo := &pwdInfo{}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("N0t-0ld-P@ss-W0rd"), pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return([]byte("0ld-P@ss-W0rd"), nil)

	err := api.ChangeMyPwd(myId, "N0t-0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.Equal(t, incorrectPwdErr, err)
}

func Test_api_ChangeMyPwd_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	pwdInfo := &pwdInfo{Pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_cryptoHelperScryptKey2Err(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{Pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_storeUpdatePwdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{Pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updatePwdInfo", myId, pwdInfo).Return(testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, newPwd, pwdInfo.Pwd)
	assert.Equal(t, salt, pwdInfo.Salt)
	assert.Equal(t, 16384, pwdInfo.N)
	assert.Equal(t, 8, pwdInfo.R)
	assert.Equal(t, 1, pwdInfo.P)
	assert.Equal(t, 32, pwdInfo.KeyLen)
}

func Test_api_ChangeMyPwd_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{Pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.Salt, pwdInfo.N, pwdInfo.R, pwdInfo.P, pwdInfo.KeyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updatePwdInfo", myId, pwdInfo).Return(nil)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.Nil(t, err)
}

func Test_api_ChangeMyEmail_invalidStringParamErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.ChangeMyEmail(myId, "new-invalid-email")
	assert.IsType(t, &invalidStringParamErr{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(&fullUserInfo{me: me{Email: "test@expected.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "test@expected.com").Return(testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailSuccess(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(&fullUserInfo{me: me{Email: "test@expected.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "test@expected.com").Return(nil)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.Nil(t, err)
}

func Test_api_ChangeMyEmail_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(nil, testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(nil, nil)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyEmail_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, "new@email.com", *user.NewEmail)
	assert.Equal(t, "confirmationCode", *user.NewEmailConfirmationCode)
}

func Test_api_ChangeMyEmail_linkMailerSendNewEmailConfirmationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{Email: "current@email.com"}}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", nil)
	store.On("updateUser", user).Return(nil)
	linkMailer.On("sendNewEmailConfirmationLink", "current@email.com", "new@email.com", "confirmationCode").Return(testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{Email: "current@email.com"}}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", nil)
	store.On("updateUser", user).Return(nil)
	linkMailer.On("sendNewEmailConfirmationLink", "current@email.com", "new@email.com", "confirmationCode").Return(nil)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, testErr)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_noNewEmailRegisteredEmailErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(&fullUserInfo{}, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Equal(t, noNewEmailRegisteredErr, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_emailConfirmationCodeErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail}}, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_linkMailerSendNewEmailConfirmationLinkErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	confirmationCode := "confirmationCode"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail, Email: "email@email.com"}, NewEmailConfirmationCode: &confirmationCode}, nil)
	linkMailer.On("sendNewEmailConfirmationLink", "email@email.com", "new@email.com", "confirmationCode").Return(testErr)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	confirmationCode := "confirmationCode"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail, Email: "email@email.com"}, NewEmailConfirmationCode: &confirmationCode}, nil)
	linkMailer.On("sendNewEmailConfirmationLink", "email@email.com", "new@email.com", "confirmationCode").Return(nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Nil(t, err)
}

func Test_api_MigrateMe_NotImplementedErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.MigrateMe(myId, "us")
	assert.Equal(t, NotImplementedErr, err)
}

func Test_api_GetMe_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, testErr)

	me, err := api.GetMe(myId)
	assert.Nil(t, me)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetMe_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, nil)

	me, err := api.GetMe(myId)
	assert.Nil(t, me)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_GetMe_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Entity: Entity{Id: myId}}}}
	store.On("getUserById", myId).Return(user, nil)

	me, err := api.GetMe(myId)
	assert.Equal(t, &user.me, me)
	assert.Nil(t, err)
}

func Test_api_DeleteMe_storeDeleteUserErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("deleteUser", myId).Return(testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("deleteUser", myId).Return(nil)

	err := api.DeleteMe(myId)
	assert.Nil(t, err)
}

func Test_api_CreateOrg_invalidStringParamErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	org, err := api.CreateOrg(myId, "yo", "us")
	assert.Nil(t, org)
	assert.IsType(t, &invalidStringParamErr{}, err)
}

func Test_api_CreateOrg_noSuchRegionErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	org, err := api.CreateOrg(myId, "newOrg", "not-a-region")
	assert.Nil(t, org)
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_CreateOrg_getAccountByNameErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_getAccountByNameNoneNilAccount(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(&account{}, nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_CreateOrg_newIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(nil, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeCreateOrgAndMembershipErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	store.On("getUserById", myId).Return(nil, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	store.On("getUserById", myId).Return(nil, nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_CreateOrg_internalRegionApiCreateOrgTaskCenterErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreateOrgTaskCenter", myId, orgId, "bob").Return(0, testErr)
	store.On("deleteOrg", orgId).Return(nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_deleteOrgErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreateOrgTaskCenter", myId, orgId, "bob").Return(8, testErr)
	store.On("deleteOrg", orgId).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_updateOrgErr(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreateOrgTaskCenter", myId, orgId, "bob").Return(8, nil)
	store.On("deleteOrg", orgId).Return(nil)
	store.On("updateOrg", &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: 8,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_success(t *testing.T) {
	store, internalRegionApis, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, map[string]internalRegionApi{"us": &mockInternalRegionApi{}}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api, _ := newApi(store, internalRegionApis, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getAccountByName", "newOrg").Return(nil, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", myId, &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: -1,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApis["us"].(*mockInternalRegionApi).On("CreateOrgTaskCenter", myId, orgId, "bob").Return(8, nil)
	store.On("deleteOrg", orgId).Return(nil)
	store.On("updateOrg", &org{
		Entity: Entity{
			Id: orgId,
		},
		Region: "us",
		Shard: 8,
		Created: timeNowReplacement,
		Name: "newOrg",
		IsUser: false,
	}).Return(nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, err)
	assert.Equal(t, "us", org.Region)
	assert.Equal(t, 8, org.Shard)
	assert.Equal(t, "newOrg", org.Name)
	assert.Equal(t, false, org.IsUser)
}

//helpers
var (
	testErr            = errors.New("test")
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

func (m *mockStore) getUserById(id Id) (*fullUserInfo, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fullUserInfo), args.Error(1)
}

func (m *mockStore) getPwdInfo(id Id) (*pwdInfo, error) {
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

func (m *mockStore) updatePwdInfo(id Id, pwdInfo *pwdInfo) error {
	args := m.Called(id, pwdInfo)
	return args.Error(0)
}

func (m *mockStore) deleteUser(id Id) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockStore) getUsers(ids []Id) ([]*user, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user), args.Error(1)
}

func (m *mockStore) searchUsers(search string, offset, limit int) ([]*user, int, error) {
	args := m.Called(search, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*user), args.Int(1), args.Error(2)
}

func (m *mockStore) createOrgAndMembership(user Id, org *org) error {
	org.Created = timeNowReplacement
	args := m.Called(user, org)
	return args.Error(0)
}

func (m *mockStore) getOrgById(id Id) (*org, error) {
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

func (m *mockStore) deleteOrg(id Id) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockStore) getOrgs(ids []Id) ([]*org, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*org), args.Error(1)
}

func (m *mockStore) searchOrgs(search string, offset, limit int) ([]*org, int, error) {
	args := m.Called(search, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*org), args.Int(1), args.Error(2)
}

func (m *mockStore) getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error) {
	args := m.Called(userId, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*org), args.Int(1), args.Error(2)
}

func (m *mockStore) membershipExists(user, org Id) (bool, error) {
	args := m.Called(user, org)
	return args.Bool(0), args.Error(1)
}

func (m *mockStore) createMembership(user, org Id) error {
	args := m.Called(user, org)
	return args.Error(0)
}

func (m *mockStore) deleteMembership(user, org Id) error {
	args := m.Called(user, org)
	return args.Error(0)
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

func (m *mockLinkMailer) sendNewEmailConfirmationLink(currentEmail, newEmail, confirmationCode string) error {
	args := m.Called(currentEmail, newEmail, confirmationCode)
	return args.Error(0)
}

type mockInternalRegionApi struct {
	mock.Mock
}

func (m *mockInternalRegionApi) CreatePersonalTaskCenter(userId Id) (int, error) {
	args := m.Called(userId)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) CreateOrgTaskCenter(ownerId, orgId Id, ownerName string) (int, error) {
	args := m.Called(ownerId, orgId, ownerName)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) RenameMember(memberId, orgId Id, newName string) error {
	args := m.Called(memberId, orgId, newName)
	return args.Error(0)
}

type mockMiscFuncs struct {
	mock.Mock
}

func (m *mockMiscFuncs) newId() (Id, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Id), args.Error(1)
}

type mockCryptoHelper struct {
	mock.Mock
}

func (m *mockCryptoHelper) Bytes(length int) ([]byte, error) {
	args := m.Called(length)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockCryptoHelper) UrlSafeString(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *mockCryptoHelper) ScryptKey(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	args := m.Called(password, salt, N, r, p, keyLen)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
