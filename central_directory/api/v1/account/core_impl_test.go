package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_newApi_nilStorePanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newApi(nil, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_nilInternalRegionApiPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	store := &mockStore{}
	newApi(store, nil, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_nilLinkMailerPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	store, internalRegionApi := &mockStore{}, &mockInternalRegionApi{}
	newApi(store, internalRegionApi, nil, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_nilNewIdPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	store, internalRegionApi, linkMailer := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}
	newApi(store, internalRegionApi, linkMailer, nil, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_nilCryptoHelperPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	store, internalRegionApi, linkMailer, miscFuncs := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}
	newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, nil, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_nilLogPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}
	newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, nil)
}

func Test_newApi_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	assert.NotNil(t, api)
}

func Test_api_GetRegiosn_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	expectedRegions := []string{"us-w", "us-e", "ch", "au"}
	internalRegionApi.On("GetRegions").Return(expectedRegions)
	regions := api.GetRegions()
	assert.Equal(t, expectedRegions, regions)
}

func Test_api_Register_invalidNameParam(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("a", "email@email.email", "P@ss-W0rd", "us").(*InvalidStringParamErr)
	assert.Equal(t, "name", err.ParamPurpose)
	assert.Equal(t, "name must be between 3 and 20 utf8 characters long and match all regexs []", err.Error())
}

func Test_api_Register_invalidEmailParam(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "invalidEmail", "P@ss-W0rd", "us").(*InvalidStringParamErr)
	assert.Equal(t, "email", err.ParamPurpose)
}

func Test_api_Register_invalidPwdParam(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	err := api.Register("ali", "email@email.com", "p", "us").(*InvalidStringParamErr)
	assert.Equal(t, "password", err.ParamPurpose)
}

func Test_api_Register_invalidRegionParam(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_Register_storeAccountWithNameExistsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeAccountWithNameExists_true(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(true, nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_Register_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_storeGetUserByEmailNoneNilUser_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{me: me{Email: "email@email.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "email@email.com").Return(nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Nil(t, err)
}

func Test_api_Register_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
	store.On("getUserByEmail", "email@email.com").Return(nil, nil)
	salt := []byte("salt")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), salt, 16384, 8, 1, 32).Return(nil, testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
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
					NamedEntity: NamedEntity{
						Entity: Entity{
							Id: id,
						},
						Name: "ali",
					},
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser:  true,
				},
				Email: "email@email.com",
			},
			activationCode: &activationCode,
		},
		&pwdInfo{
			salt:   salt,
			pwd:    pwd,
			n:      16384,
			r:      8,
			p:      1,
			keyLen: 32,
		}).Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
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
					NamedEntity: NamedEntity{
						Entity: Entity{
							Id: id,
						},
						Name: "ali",
					},
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser:  true,
				},
				Email: "email@email.com",
			},
			activationCode: &activationCode,
		},
		&pwdInfo{
			salt:   salt,
			pwd:    pwd,
			n:      16384,
			r:      8,
			p:      1,
			keyLen: 32,
		}).Return(nil)
	linkMailer.On("sendActivationLink", "email@email.com", activationCode).Return(testErr)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Register_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "ali").Return(false, nil)
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
					NamedEntity: NamedEntity{
						Entity: Entity{
							Id: id,
						},
						Name: "ali",
					},
					Region:  "us",
					Shard:   -1,
					Created: timeNowReplacement,
					IsUser:  true,
				},
				Email: "email@email.com",
			},
			activationCode: &activationCode,
		},
		&pwdInfo{
			salt:   salt,
			pwd:    pwd,
			n:      16384,
			r:      8,
			p:      1,
			keyLen: 32,
		}).Return(nil)
	linkMailer.On("sendActivationLink", "email@email.com", activationCode).Return(nil)

	err := api.Register("ali", "email@email.com", "P@ss-W0rd", "us")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_storeGetUseByEmailActivatedUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activated := time.Now().UTC()
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{activated: &activated}, nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_ResendActivationEmail_linkMailerSendActivationLinkErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{activationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(testErr)

	err := api.ResendActivationEmail("email@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendActivationEmail_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	code := "test"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{activationCode: &code}, nil)
	linkMailer.On("sendActivationLink", "email@email.com", code).Return(nil)

	err := api.ResendActivationEmail("email@email.com")
	assert.Nil(t, err)
}

func Test_api_Activate_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_storeGetUserByActivationCodeNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.Equal(t, invalidActivationAttemptErr, err)
}

func Test_api_Activate_storeGetUserByActivationCode_activationCodeMismatch(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationCode := "not-the-activation-code"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{activationCode: &activationCode}, nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.Equal(t, invalidActivationAttemptErr, err)
}

func Test_api_Activate_invalidUserRegion(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationCode := "activationCode"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{activationCode: &activationCode, me: me{user: user{Region: "us"}}}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	id, err := api.Activate("email@email.com", activationCode)
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_internalRegionalApiCreatePersonalTaskCenterErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				NamedEntity: NamedEntity{
					Entity: Entity{
						Id: userId,
					},
				},
				Shard: -1,
			},
		},
		activationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", userId).Return(-1, testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				NamedEntity: NamedEntity{
					Entity: Entity{
						Id: userId,
					},
				},
				Shard: -1,
			},
		},
		activationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", userId).Return(3, nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.Activate("email@email.com", "test")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Activate_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	userId, _ := NewId()
	activationCode := "test"
	user := &fullUserInfo{
		me: me{
			user: user{
				Region: "us",
				NamedEntity: NamedEntity{
					Entity: Entity{
						Id: userId,
					},
				},
				Shard: -1,
			},
		},
		activationCode: &activationCode,
	}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", userId).Return(3, nil)
	store.On("updateUser", user).Return(nil)

	id, err := api.Activate("email@email.com", "test")
	assert.Equal(t, userId, id)
	assert.Nil(t, err)
	assert.Nil(t, user.activationCode)
	assert.NotNil(t, user.activated)
	assert.Equal(t, user.Shard, 3)
}

func Test_api_Authenticate_storeGetUserByNameErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByCiName", "name").Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_storeGetUserByNameNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByCiName", "name").Return(nil, nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetPwdInfoErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{activated: &activationTime}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{activated: &activationTime}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_incorrectPwdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	user := &fullUserInfo{activated: &activationTime}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("N0t-P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, nameOrPwdIncorrectErr, err)
}

func Test_api_Authenticate_storeGetUserByName_userNotActivatedErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	user := &fullUserInfo{}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.Equal(t, userNotActivatedErr, err)
}

func Test_api_Authenticate_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{activated: &activationTime, resetPwdCode: &resetPwdCode}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{activated: &activationTime, resetPwdCode: &resetPwdCode}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_genScryptKey2Err(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{activated: &activationTime, resetPwdCode: &resetPwdCode}
	store.On("getUserByCiName", "name").Return(user, nil)
	store.On("getPwdInfo", Id(nil)).Return(&pwdInfo{pwd: []byte("P@ss-W0rd")}, nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte(nil), 0, 0, 0, 0).Return([]byte("P@ss-W0rd"), nil)
	store.On("updateUser", user).Return(nil)
	cryptoHelper.On("Bytes", 128).Return([]byte("test"), nil)
	cryptoHelper.On("ScryptKey", []byte("P@ss-W0rd"), []byte("test"), 16384, 8, 1, 32).Return(nil, testErr)

	id, err := api.Authenticate("name", "P@ss-W0rd")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_Authenticate_storeUpdatePwdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	user := &fullUserInfo{activated: &activationTime, resetPwdCode: &resetPwdCode}
	store.On("getUserByCiName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{pwd: []byte("P@ss-W0rd")}
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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	activationTime := time.Now().UTC()
	resetPwdCode := "resetPwdCode"
	id, _ := NewId()
	user := &fullUserInfo{activated: &activationTime, resetPwdCode: &resetPwdCode, me: me{user: user{NamedEntity: NamedEntity{Entity: Entity{Id: id}}}}}
	store.On("getUserByCiName", "name").Return(user, nil)
	pwdInfo := &pwdInfo{pwd: []byte("P@ss-W0rd")}
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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "currentEmail@currentEmail.com").Return(nil, testErr)

	resultId, err := api.ConfirmNewEmail("currentEmail@currentEmail.com", "email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmaiNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "currentEmail@currentEmail.com").Return(nil, nil)

	resultId, err := api.ConfirmNewEmail("currentEmail@currentEmail.com", "email@email.com", "confirmationCode")
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_mismatchNewEmail(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, "not-new@email.com", confirmationCode)
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_mismatchConfirmationCode(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, "not-confirmation-code")
	assert.Nil(t, resultId)
	assert.Equal(t, invalidNewEmailConfirmationAttemptErr, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmail2Err(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, testErr)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_storeGetUserByEmail2NoneNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(&fullUserInfo{}, nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.Equal(t, emailAlreadyInUseErr, err)
}

func Test_api_ConfirmNewEmail_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(testErr)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Nil(t, resultId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ConfirmNewEmail_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	confirmationCode := "confirmationCode"
	newEmail := "new@email.com"
	currentEmail := "currentEmail@currentEmail.com"
	id, _ := NewId()
	user := &fullUserInfo{newEmailConfirmationCode: &confirmationCode, me: me{NewEmail: &newEmail, Email: currentEmail, user: user{NamedEntity: NamedEntity{Entity: Entity{Id: id}}}}}
	store.On("getUserByEmail", currentEmail).Return(user, nil)
	store.On("getUserByEmail", newEmail).Return(nil, nil)
	store.On("updateUser", user).Return(nil)

	resultId, err := api.ConfirmNewEmail(currentEmail, newEmail, confirmationCode)
	assert.Equal(t, id, resultId)
	assert.Nil(t, err)
	assert.Equal(t, newEmail, user.Email)
	assert.Nil(t, user.NewEmail)
	assert.Nil(t, user.newEmailConfirmationCode)
}

func Test_api_ResetPwd_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(nil, testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_storeGetUserByEmailNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(nil, nil)

	err := api.ResetPwd(email)
	assert.Nil(t, err)
}

func Test_api_ResetPwd_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	store.On("getUserByEmail", email).Return(&fullUserInfo{}, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("", testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	email := "email@email.com"
	user := &fullUserInfo{}
	store.On("getUserByEmail", email).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("test", nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ResetPwd(email)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResetPwd_linkMailerSendPwdResetLinkErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, err := api.SetNewPwdFromPwdReset("yo", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &InvalidStringParamErr{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmailNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getUserByEmail", "email@email.com").Return(nil, nil)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.Equal(t, invalidResetPwdAttemptErr, err)
}

func Test_api_SetNewPwdFromPwdReset_storeGetUserByEmail_resetPwdCodeMismatch(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "not-the-correct-code"
	store.On("getUserByEmail", "email@email.com").Return(&fullUserInfo{resetPwdCode: &resetCode}, nil)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.Equal(t, invalidResetPwdAttemptErr, err)
}

func Test_api_SetNewPwdFromPwdReset_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{resetPwdCode: &resetCode}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset("P@ss-W0rd", "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{resetPwdCode: &resetCode}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(nil, testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	user := &fullUserInfo{resetPwdCode: &resetCode, activationCode: &resetCode, me: me{user: user{Region: "us"}}}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_internalRegionApiCreatePersonalTaskCenterErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	myId, _ := NewId()
	user := &fullUserInfo{resetPwdCode: &resetCode, activationCode: &resetCode, me: me{user: user{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: myId}}}}}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", myId).Return(0, testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	myId, _ := NewId()
	user := &fullUserInfo{resetPwdCode: &resetCode, activationCode: &resetCode, me: me{user: user{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: myId}}}}}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", myId).Return(3, nil)
	store.On("updateUser", user).Return(testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, 3, user.Shard)
	assert.Nil(t, user.resetPwdCode)
	assert.Nil(t, user.activationCode)
}

func Test_api_SetNewPwdFromPwdReset_storeUpdatePwdInfoErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	myId, _ := NewId()
	user := &fullUserInfo{resetPwdCode: &resetCode, activationCode: &resetCode, me: me{user: user{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: myId}}}}}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", myId).Return(3, nil)
	store.On("updateUser", user).Return(nil)
	store.On("updatePwdInfo", user.Id, &pwdInfo{pwd: newPwd, salt: salt, n: 16384, r: 8, p: 1, keyLen: 32}).Return(testErr)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Nil(t, id)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_SetNewPwdFromPwdReset_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	resetCode := "resetCode"
	myId, _ := NewId()
	user := &fullUserInfo{resetPwdCode: &resetCode, activationCode: &resetCode, me: me{user: user{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: myId}}}}}
	store.On("getUserByEmail", "email@email.com").Return(user, nil)
	salt := []byte("salt")
	newPwd := []byte("P@ss-W0rd")
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("CreatePersonalTaskCenter", "us", myId).Return(3, nil)
	store.On("updateUser", user).Return(nil)
	store.On("updatePwdInfo", user.Id, &pwdInfo{pwd: newPwd, salt: salt, n: 16384, r: 8, p: 1, keyLen: 32}).Return(nil)

	id, err := api.SetNewPwdFromPwdReset(string(newPwd), "email@email.com", "resetCode")
	assert.Equal(t, myId, id)
	assert.Nil(t, err)
}

func Test_api_GetAccount_storeGetAccountErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	store.On("getAccountByCiName", "ali").Return(nil, testErr)

	account, err := api.GetAccount("ali")
	assert.Nil(t, account)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetAccount_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	acc := &account{}
	store.On("getAccountByCiName", "ali").Return(acc, nil)

	account, err := api.GetAccount("ali")
	assert.Equal(t, acc, account)
	assert.Nil(t, err)
}

func Test_api_GetUsers_maxEntityCountExceededErri(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	ids := make([]Id, 101, 101)

	users, err := api.GetUsers(ids)
	assert.Nil(t, users)
	assert.Equal(t, maxEntityCountExceededErr, err)
}

func Test_api_GetUsers_storeGetUsersErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	store.On("getUsers", ids).Return(nil, testErr)

	users, err := api.GetUsers(ids)
	assert.Nil(t, users)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetUsers_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	res := []*user{}
	store.On("getUsers", ids).Return(res, nil)

	users, err := api.GetUsers(ids)
	assert.Equal(t, res, users)
	assert.Nil(t, err)
}

func Test_api_GetOrgs_maxEntityCountExceededErri(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	ids := make([]Id, 101, 101)

	orgs, err := api.GetOrgs(ids)
	assert.Nil(t, orgs)
	assert.Equal(t, maxEntityCountExceededErr, err)
}

func Test_api_GetOrgs_storeGetOrgsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	store.On("getOrgs", ids).Return(nil, testErr)

	orgs, err := api.GetOrgs(ids)
	assert.Nil(t, orgs)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetOrgs_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id1, _ := NewId()
	id2, _ := NewId()
	ids := []Id{id1, id2}
	res := []*org{}
	store.On("getOrgs", ids).Return(res, nil)

	orgs, err := api.GetOrgs(ids)
	assert.Equal(t, res, orgs)
	assert.Nil(t, err)
}

func Test_api_ChangeMyName_invalidStringParamErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()

	err := api.ChangeMyName(id, "yo")
	assert.IsType(t, &InvalidStringParamErr{}, err)
}

func Test_api_ChangeMyName_storeAccountWithCiNameExistsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("accountWithCiNameExists", "test").Return(false, testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_storeSccountWithCiNameExistsTrue(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("accountWithCiNameExists", "test").Return(true, nil)

	err := api.ChangeMyName(id, "test")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_ChangeMyName_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(nil, testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(nil, nil)

	err := api.ChangeMyName(id, "test")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyName_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	user := &fullUserInfo{}
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, "test", user.Name)
}

func Test_api_ChangeMyName_storeGetUsersOrgsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	user := &fullUserInfo{}
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(nil)
	store.On("getUsersOrgs", id, 0, 100).Return(nil, 0, testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, "test", user.Name)
}

func Test_api_ChangeMyName_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	user := &fullUserInfo{}
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(nil)
	store.On("getUsersOrgs", id, 0, 100).Return([]*org{&org{Region: "us"}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_internalRegionApiRenameMemberErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	orgId, _ := NewId()
	user := &fullUserInfo{}
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(nil)
	store.On("getUsersOrgs", id, 0, 100).Return([]*org{&org{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: orgId}}}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("RenameMember", "us", 0, orgId, id, "test").Return(testErr)

	err := api.ChangeMyName(id, "test")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyName_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	id, _ := NewId()
	orgId, _ := NewId()
	user := &fullUserInfo{}
	store.On("accountWithCiNameExists", "test").Return(false, nil)
	store.On("getUserById", id).Return(user, nil)
	store.On("updateUser", user).Return(nil)
	store.On("getUsersOrgs", id, 0, 100).Return([]*org{&org{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: orgId}}}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("RenameMember", "us", 0, orgId, id, "test").Return(nil)

	err := api.ChangeMyName(id, "test")
	assert.Nil(t, err)
}

func Test_api_ChangeMyPwd_invalidStringParam(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "yo")
	assert.IsType(t, &InvalidStringParamErr{}, err)
}

func Test_api_ChangeMyPwd_storeGetPwdInfoErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getPwdInfo", myId).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_storeGetPwdInfoNilPwdInfo(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getPwdInfo", myId).Return(nil, nil)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyPwd_cryptoHelperScryptKeyErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	pwdInfo := &pwdInfo{}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("0ld-P@ss-W0rd"), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, "0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_incorrectPwdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	pwdInfo := &pwdInfo{}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", []byte("N0t-0ld-P@ss-W0rd"), pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return([]byte("0ld-P@ss-W0rd"), nil)

	err := api.ChangeMyPwd(myId, "N0t-0ld-P@ss-W0rd", "P@ss-W0rd")
	assert.Equal(t, incorrectPwdErr, err)
}

func Test_api_ChangeMyPwd_cryptoHelperBytesErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	pwdInfo := &pwdInfo{pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), "P@ss-W0rd")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_cryptoHelperScryptKey2Err(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(nil, testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyPwd_storeUpdatePwdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updatePwdInfo", myId, pwdInfo).Return(testErr)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, newPwd, pwdInfo.pwd)
	assert.Equal(t, salt, pwdInfo.salt)
	assert.Equal(t, 16384, pwdInfo.n)
	assert.Equal(t, 8, pwdInfo.r)
	assert.Equal(t, 1, pwdInfo.p)
	assert.Equal(t, 32, pwdInfo.keyLen)
}

func Test_api_ChangeMyPwd_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	oldPwd := []byte("0ld-P@ss-W0rd")
	newPwd := []byte("P@ss-W0rd")
	salt := []byte("salt")
	pwdInfo := &pwdInfo{pwd: oldPwd}
	store.On("getPwdInfo", myId).Return(pwdInfo, nil)
	cryptoHelper.On("ScryptKey", oldPwd, pwdInfo.salt, pwdInfo.n, pwdInfo.r, pwdInfo.p, pwdInfo.keyLen).Return(oldPwd, nil)
	cryptoHelper.On("Bytes", 128).Return(salt, nil)
	cryptoHelper.On("ScryptKey", newPwd, salt, 16384, 8, 1, 32).Return(newPwd, nil)
	store.On("updatePwdInfo", myId, pwdInfo).Return(nil)

	err := api.ChangeMyPwd(myId, string(oldPwd), string(newPwd))
	assert.Nil(t, err)
}

func Test_api_ChangeMyEmail_invalidStringParamErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.ChangeMyEmail(myId, "new-invalid-email")
	assert.IsType(t, &InvalidStringParamErr{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(&fullUserInfo{me: me{Email: "test@expected.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "test@expected.com").Return(testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByEmailNoneNilUser_linkMailerSendMultipleAccountPolicyEmailSuccess(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(&fullUserInfo{me: me{Email: "test@expected.com"}}, nil)
	linkMailer.On("sendMultipleAccountPolicyEmail", "test@expected.com").Return(nil)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.Nil(t, err)
}

func Test_api_ChangeMyEmail_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(nil, testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(nil, nil)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ChangeMyEmail_cryptoHelperUrlSafeStringErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ChangeMyEmail_storeUpdateUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{}
	store.On("getUserByEmail", "new@email.com").Return(nil, nil)
	store.On("getUserById", myId).Return(user, nil)
	cryptoHelper.On("UrlSafeString", 40).Return("confirmationCode", nil)
	store.On("updateUser", user).Return(testErr)

	err := api.ChangeMyEmail(myId, "new@email.com")
	assert.IsType(t, &ErrorRef{}, err)
	assert.Equal(t, "new@email.com", *user.NewEmail)
	assert.Equal(t, "confirmationCode", *user.newEmailConfirmationCode)
}

func Test_api_ChangeMyEmail_linkMailerSendNewEmailConfirmationLinkErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

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
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, testErr)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_noNewEmailRegisteredEmailErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(&fullUserInfo{}, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Equal(t, noNewEmailRegisteredErr, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_emailConfirmationCodeErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail}}, nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_linkMailerSendNewEmailConfirmationLinkErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	confirmationCode := "confirmationCode"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail, Email: "email@email.com"}, newEmailConfirmationCode: &confirmationCode}, nil)
	linkMailer.On("sendNewEmailConfirmationLink", "email@email.com", "new@email.com", "confirmationCode").Return(testErr)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_ResendMyNewEmailConfirmationEmail_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	newEmail := "new@email.com"
	confirmationCode := "confirmationCode"
	store.On("getUserById", myId).Return(&fullUserInfo{me: me{NewEmail: &newEmail, Email: "email@email.com"}, newEmailConfirmationCode: &confirmationCode}, nil)
	linkMailer.On("sendNewEmailConfirmationLink", "email@email.com", "new@email.com", "confirmationCode").Return(nil)

	err := api.ResendMyNewEmailConfirmationEmail(myId)
	assert.Nil(t, err)
}

func Test_api_MigrateMe_NotImplementedErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	err := api.MigrateMe(myId, "us")
	assert.Equal(t, NotImplementedErr, err)
}

func Test_api_GetMe_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, testErr)

	me, err := api.GetMe(myId)
	assert.Nil(t, me)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetMe_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, nil)

	me, err := api.GetMe(myId)
	assert.Nil(t, me)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_GetMe_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{NamedEntity: NamedEntity{Entity: Entity{Id: myId}}}}}
	store.On("getUserById", myId).Return(user, nil)

	me, err := api.GetMe(myId)
	assert.Equal(t, &user.me, me)
	assert.Nil(t, err)
}

func Test_api_DeleteMe_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUserById", myId).Return(nil, nil)

	err := api.DeleteMe(myId)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_DeleteMe_storeGetUsersOrgsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	orgId, _ := NewId()
	store.On("getUsersOrgs", myId, 0, 100).Return([]*org{&org{Region: "us", NamedEntity: NamedEntity{Entity: Entity{Id: orgId}}}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_internalRegionApiRemoveMemberErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	orgId, _ := NewId()
	store.On("getUsersOrgs", myId, 0, 100).Return([]*org{&org{Region: "us", Shard: 4, NamedEntity: NamedEntity{Entity: Entity{Id: orgId}}}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("SetMemberDeleted", "us", 4, orgId, myId).Return(testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_storeDeleteMembershipErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	orgId, _ := NewId()
	store.On("getUsersOrgs", myId, 0, 100).Return([]*org{&org{Region: "us", Shard: 4, NamedEntity: NamedEntity{Entity: Entity{Id: orgId}}}}, 1, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("SetMemberDeleted", "us", 4, orgId, myId).Return(nil)
	store.On("deleteMemberships", orgId, []Id{myId}).Return(testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_regionGoneErr2(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_internalRegionApiDeleteTaskCenterErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 2, myId, myId).Return(nil, testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_internalRegionApiDeleteTaskCenterPublicErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 2, myId, myId).Return(testErr, nil)

	err := api.DeleteMe(myId)
	assert.Equal(t, testErr, err)
}

func Test_api_DeleteMe_storeDeleteUserErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 2, myId, myId).Return(nil, nil)
	store.On("deleteUserAndAllAssociatedMemberships", myId).Return(testErr)

	err := api.DeleteMe(myId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteMe_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	user := &fullUserInfo{me: me{user: user{Shard: 2, Region: "us"}}}
	store.On("getUserById", myId).Return(user, nil)
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 2, myId, myId).Return(nil, nil)
	store.On("deleteUserAndAllAssociatedMemberships", myId).Return(nil)

	err := api.DeleteMe(myId)
	assert.Nil(t, err)
}

func Test_api_CreateOrg_invalidStringParamErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()

	org, err := api.CreateOrg(myId, "yo", "us")
	assert.Nil(t, org)
	assert.IsType(t, &InvalidStringParamErr{}, err)
}

func Test_api_CreateOrg_noSuchRegionErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.Equal(t, noSuchRegionErr, err)
}

func Test_api_CreateOrg_storeAccountWithNameExistsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeAccountWithNameExists_true(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(true, nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_CreateOrg_newIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(nil, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeCreateOrgAndMembershipErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeGetUserByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	store.On("getUserById", myId).Return(nil, testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_storeGetUserByIdNilUser(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	store.On("getUserById", myId).Return(nil, nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.Equal(t, noSuchUserErr, err)
}

func Test_api_CreateOrg_internalRegionApiCreateOrgTaskCenterErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApi.On("CreateOrgTaskCenter", "us", orgId, myId, "bob").Return(0, testErr)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_deleteOrgAndAllAssociatedMembershipsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApi.On("CreateOrgTaskCenter", "us", orgId, myId, "bob").Return(8, testErr)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_updateOrgErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApi.On("CreateOrgTaskCenter", "us", orgId, myId, "bob").Return(8, nil)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(nil)
	store.On("updateOrg", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   8,
		Created: timeNowReplacement,
		IsUser:  false,
	}).Return(testErr)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, org)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_CreateOrg_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	miscFuncs.On("newId").Return(orgId, nil)
	store.On("createOrgAndMembership", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   -1,
		Created: timeNowReplacement,
		IsUser:  false,
	}, myId).Return(nil)
	user := &fullUserInfo{}
	user.Id = myId
	user.Name = "bob"
	store.On("getUserById", myId).Return(user, nil)
	internalRegionApi.On("CreateOrgTaskCenter", "us", orgId, myId, "bob").Return(8, nil)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(nil)
	store.On("updateOrg", &org{
		NamedEntity: NamedEntity{
			Entity: Entity{
				Id: orgId,
			},
			Name: "newOrg",
		},
		Region:  "us",
		Shard:   8,
		Created: timeNowReplacement,
		IsUser:  false,
	}).Return(nil)

	org, err := api.CreateOrg(myId, "newOrg", "us")
	assert.Nil(t, err)
	assert.Equal(t, "us", org.Region)
	assert.Equal(t, 8, org.Shard)
	assert.Equal(t, "newOrg", org.Name)
	assert.Equal(t, false, org.IsUser)
}

func Test_api_RenameOrg_storeAccountWithNameExistsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, testErr)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RenameOrg_storeAccountWithNameExistsTrue(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(true, nil)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.Equal(t, accountNameAlreadyInUseErr, err)
}

func Test_api_RenameOrg_storeGetOrgByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(nil, testErr)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RenameOrg_storeGetOrgByIdNilOrg(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(nil, nil)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.Equal(t, noSuchOrgErr, err)
}

func Test_api_RenameOrg_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RenameOrg_internalRegionApiUserCanRenameOrgErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("UserCanRenameOrg", "us", 0, orgId, myId).Return(false, testErr)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RenameOrg_insufficientPermissionsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("UserCanRenameOrg", "us", 0, orgId, myId).Return(false, nil)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.Equal(t, insufficientPermissionsErr, err)
}

func Test_api_RenameOrg_storeUpdateOrgErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	org := &org{Region: "us"}
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(org, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("UserCanRenameOrg", "us", 0, orgId, myId).Return(true, nil)
	store.On("updateOrg", org).Return(testErr)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RenameOrg_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	org := &org{Region: "us"}
	store.On("accountWithCiNameExists", "newOrg").Return(false, nil)
	store.On("getOrgById", orgId).Return(org, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("UserCanRenameOrg", "us", 0, orgId, myId).Return(true, nil)
	store.On("updateOrg", org).Return(nil)

	err := api.RenameOrg(myId, orgId, "newOrg")
	assert.Nil(t, err)
	assert.Equal(t, "newOrg", org.Name)
}

func Test_api_MigrateOrg_notImplementedErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()

	err := api.MigrateOrg(myId, orgId, "newOrg")
	assert.Equal(t, NotImplementedErr, err)
}

func Test_api_GetMyOrgs_storeGetUsersOrgsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUsersOrgs", myId, 0, 100).Return(nil, 0, testErr)

	orgs, total, err := api.GetMyOrgs(myId, -1, -1)
	assert.Nil(t, orgs)
	assert.Equal(t, 0, total)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_GetMyOrgs_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	store.On("getUsersOrgs", myId, 0, 100).Return([]*org{&org{}, &org{}}, 2, nil)

	orgs, total, err := api.GetMyOrgs(myId, -1, -1)
	assert.NotNil(t, orgs)
	assert.Equal(t, 2, total)
	assert.Nil(t, err)
}

func Test_api_DeleteOrg_storeGetOrgByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, testErr)

	err := api.DeleteOrg(myId, orgId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteOrg_storeGetOrgByIdNilOrg(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, nil)

	err := api.DeleteOrg(myId, orgId)
	assert.Equal(t, noSuchOrgErr, err)
}

func Test_api_DeleteOrg_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)

	err := api.DeleteOrg(myId, orgId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteOrg_internalRegionApiDeleteTaskCenterErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 0, orgId, myId).Return(nil, testErr)

	err := api.DeleteOrg(myId, orgId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteOrg_internalRegionApiDeleteTaskCenterPublicErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 0, orgId, myId).Return(testErr, nil)

	err := api.DeleteOrg(myId, orgId)
	assert.Equal(t, testErr, err)
}

func Test_api_DeleteOrg_storeDeleteOrgAndAllAssociatedMembershipsErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 0, orgId, myId).Return(nil, nil)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(testErr)

	err := api.DeleteOrg(myId, orgId)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_DeleteOrg_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	internalRegionApi.On("DeleteTaskCenter", "us", 0, orgId, myId).Return(nil, nil)
	store.On("deleteOrgAndAllAssociatedMemberships", orgId).Return(nil)

	err := api.DeleteOrg(myId, orgId)
	assert.Nil(t, err)
}

func Test_api_AddMembers_maxEntityCountExceededErri(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := make([]Id, 101, 101)

	err := api.AddMembers(myId, orgId, members)
	assert.Equal(t, maxEntityCountExceededErr, err)
}

func Test_api_AddMembers_storeGetOrgByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, testErr)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.AddMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_AddMembers_storeGetOrgByIdNilOrg(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, nil)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.AddMembers(myId, orgId, members)
	assert.Equal(t, noSuchOrgErr, err)
}

func Test_api_AddMembers_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.AddMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_AddMembers_storeGetUsersErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	store.On("getUsers", members).Return(nil, testErr)

	err := api.AddMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_AddMembers_internalRegionApiAddMemberErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("AddMembers", "us", 0, orgId, myId, []*NamedEntity{&NamedEntity{Name: "test1", Entity: Entity{Id: m1}}, &NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}).Return(nil, testErr)

	err := api.AddMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_AddMembers_internalRegionApiAddMemberPublicErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("AddMembers", "us", 0, orgId, myId, []*NamedEntity{&NamedEntity{Name: "test1", Entity: Entity{Id: m1}}, &NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}).Return(testErr, nil)

	err := api.AddMembers(myId, orgId, members)
	assert.Equal(t, testErr, err)
}

func Test_api_AddMembers_storeCreateMembershipErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("AddMembers", "us", 0, orgId, myId, []*NamedEntity{&NamedEntity{Name: "test1", Entity: Entity{Id: m1}}, &NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}).Return(nil, nil)
	store.On("createMemberships", orgId, members).Return(testErr)

	err := api.AddMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_AddMembers_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("AddMembers", "us", 0, orgId, myId, []*NamedEntity{&NamedEntity{Name: "test1", Entity: Entity{Id: m1}}, &NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}).Return(nil, nil)
	store.On("createMemberships", orgId, members).Return(nil)

	err := api.AddMembers(myId, orgId, members)
	assert.Nil(t, err)
}

func Test_api_RemoveMembers_maxEntityCountExceededErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	members := make([]Id, 101, 101)

	err := api.RemoveMembers(myId, orgId, members)
	assert.Equal(t, maxEntityCountExceededErr, err)
}

func Test_api_RemoveMembers_storeGetOrgByIdErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, testErr)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.RemoveMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RemoveMembers_storeGetOrgByIdNilOrg(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(nil, nil)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.RemoveMembers(myId, orgId, members)
	assert.Equal(t, noSuchOrgErr, err)
}

func Test_api_RemoveMembers_regionGoneErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(false)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}

	err := api.RemoveMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RemoveMembers_internalRegionApiRemoveMemberErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("RemoveMembers", "us", 0, orgId, myId, members).Return(nil, testErr)

	err := api.RemoveMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RemoveMembers_internalRegionApiRemoveMemberPublicErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("RemoveMembers", "us", 0, orgId, myId, members).Return(testErr, nil)

	err := api.RemoveMembers(myId, orgId, members)
	assert.IsType(t, testErr, err)
}

func Test_api_RemoveMembers_storeDeleteMembershipErr(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("RemoveMembers", "us", 0, orgId, myId, members).Return(nil, nil)
	store.On("deleteMemberships", orgId, members).Return(testErr)

	err := api.RemoveMembers(myId, orgId, members)
	assert.IsType(t, &ErrorRef{}, err)
}

func Test_api_RemoveMembers_success(t *testing.T) {
	store, internalRegionApi, linkMailer, miscFuncs, cryptoHelper, log := &mockStore{}, &mockInternalRegionApi{}, &mockLinkMailer{}, &mockMiscFuncs{}, &mockCryptoHelper{}, NewLog(nil)
	api := newApi(store, internalRegionApi, linkMailer, miscFuncs.newId, cryptoHelper, nil, nil, 3, 20, 3, 20, 100, 40, 128, 16384, 8, 1, 32, log)

	myId, _ := NewId()
	orgId, _ := NewId()
	store.On("getOrgById", orgId).Return(&org{Region: "us"}, nil)
	internalRegionApi.On("IsValidRegion", "us").Return(true)
	m1, _ := NewId()
	m2, _ := NewId()
	members := []Id{m1, m2}
	users := []*user{&user{NamedEntity: NamedEntity{Name: "test1", Entity: Entity{Id: m1}}}, &user{NamedEntity: NamedEntity{Name: "test2", Entity: Entity{Id: m2}}}}
	store.On("getUsers", members).Return(users, nil)
	internalRegionApi.On("RemoveMembers", "us", 0, orgId, myId, members).Return(nil, nil)
	store.On("deleteMemberships", orgId, members).Return(nil)

	err := api.RemoveMembers(myId, orgId, members)
	assert.Nil(t, err)
}

//helpers
var (
	testErr            = errors.New("test")
	timeNowReplacement = time.Now().UTC()
)

type mockStore struct {
	mock.Mock
}

func (m *mockStore) accountWithCiNameExists(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *mockStore) getAccountByCiName(name string) (*account, error) {
	args := m.Called(name)
	acc := args.Get(0)
	if acc == nil {
		return nil, args.Error(1)
	}
	return acc.(*account), args.Error(1)
}

func (m *mockStore) createUser(user *fullUserInfo, pwdInfo *pwdInfo) error {
	user.Created = timeNowReplacement
	args := m.Called(user, pwdInfo)
	return args.Error(0)
}

func (m *mockStore) getUserByCiName(name string) (*fullUserInfo, error) {
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

func (m *mockStore) deleteUserAndAllAssociatedMemberships(id Id) error {
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

func (m *mockStore) createOrgAndMembership(org *org, user Id) error {
	org.Created = timeNowReplacement
	args := m.Called(org, user)
	return args.Error(0)
}

func (m *mockStore) getOrgById(id Id) (*org, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*org), args.Error(1)
}

func (m *mockStore) updateOrg(org *org) error {
	args := m.Called(org)
	return args.Error(0)
}

func (m *mockStore) deleteOrgAndAllAssociatedMemberships(id Id) error {
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

func (m *mockStore) getUsersOrgs(userId Id, offset, limit int) ([]*org, int, error) {
	args := m.Called(userId, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*org), args.Int(1), args.Error(2)
}

func (m *mockStore) createMemberships(org Id, users []Id) error {
	args := m.Called(org, users)
	return args.Error(0)
}

func (m *mockStore) deleteMemberships(org Id, users []Id) error {
	args := m.Called(org, users)
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

func (m *mockInternalRegionApi) GetRegions() []string {
	args := m.Called()
	regions := args.Get(0)
	if regions != nil {
		return regions.([]string)
	}
	return nil
}

func (m *mockInternalRegionApi) IsValidRegion(region string) bool {
	args := m.Called(region)
	return args.Bool(0)
}

func (m *mockInternalRegionApi) CreatePersonalTaskCenter(region string, userId Id) (int, error) {
	args := m.Called(region, userId)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) CreateOrgTaskCenter(region string, orgId, ownerId Id, ownerName string) (int, error) {
	args := m.Called(region, orgId, ownerId, ownerName)
	return args.Int(0), args.Error(1)
}

func (m *mockInternalRegionApi) DeleteTaskCenter(region string, shard int, account, owner Id) (error, error) {
	args := m.Called(region, shard, account, owner)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalRegionApi) AddMembers(region string, shard int, org, admin Id, members []*NamedEntity) (error, error) {
	args := m.Called(region, shard, org, admin, members)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalRegionApi) RemoveMembers(region string, shard int, org, admin Id, members []Id) (error, error) {
	args := m.Called(region, shard, org, admin, members)
	return args.Error(0), args.Error(1)
}

func (m *mockInternalRegionApi) SetMemberDeleted(region string, shard int, org, member Id) error {
	args := m.Called(region, shard, org, member)
	return args.Error(0)
}

func (m *mockInternalRegionApi) RenameMember(region string, shard int, org, member Id, newName string) error {
	args := m.Called(region, shard, org, member, newName)
	return args.Error(0)
}

func (m *mockInternalRegionApi) UserCanRenameOrg(region string, shard int, org, user Id) (bool, error) {
	args := m.Called(region, shard, org, user)
	return args.Bool(0), args.Error(1)
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
