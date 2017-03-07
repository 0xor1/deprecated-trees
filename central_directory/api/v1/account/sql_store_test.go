package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/robsix/isql"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func Test_newSqlStore_nilAccountsDbPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newSqlStore(nil, nil)
}

func Test_newSqlStore_nilPwdsDbPanic(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newSqlStore(&isql.MockDB{}, nil)
}

func Test_newSqlStore_success(t *testing.T) {
	store := newSqlStore(&isql.MockDB{}, &isql.MockDB{})
	assert.NotNil(t, store)
}

//this test tests everything using a real sql db, comment/uncomment as necessary
func Test_sqlStore_adHoc(t *testing.T) {
	rawAccountsDb, _ := sql.Open("mysql", "tc_cd_accounts:T@sk-C3n-T3r@tcp(127.0.0.1:3306)/accounts?parseTime=true&loc=UTC&multiStatements=true")
	rawPwdsDb, _ := sql.Open("mysql", "tc_cd_pwds:T@sk-C3n-T3r-Pwd@tcp(127.0.0.1:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	accountsDb, pwdsDb := isql.NewDB(rawAccountsDb), isql.NewDB(rawPwdsDb)
	store := newSqlStore(accountsDb, pwdsDb)

	str := "str"
	now := time.Now()
	user1 := &fullUserInfo{}
	user1.Id, _ = NewId()
	user1.Name = "ali"
	user1.Created = time.Now().UTC()
	user1.Region = "use"
	user1.NewRegion = nil
	user1.Shard = 3
	user1.IsUser = true
	user1.Email = "ali@ali.com"
	user1.NewEmail = &str
	user1.activationCode = &str
	user1.activated = &now
	user1.newEmailConfirmationCode = &str
	user1.resetPwdCode = &str

	pwdInfo1 := &pwdInfo{}
	pwdInfo1.salt = []byte("salt")
	pwdInfo1.pwd = []byte("pwd")
	pwdInfo1.n = 10
	pwdInfo1.r = 10
	pwdInfo1.p = 10
	pwdInfo1.keyLen = 10

	err := store.createUser(user1, pwdInfo1)
	assert.Nil(t, err)

	val, err := store.accountWithCiNameExists("ali")
	assert.True(t, val)
	assert.Nil(t, err)

	user1Dup1, err := store.getAccountByCiName("ali")
	assert.Equal(t, user1.Id, user1Dup1.Id)
	assert.Equal(t, user1.Name, user1Dup1.Name)
	assert.Equal(t, user1.Created.Unix(), user1Dup1.Created.Unix())
	assert.Equal(t, user1.Region, user1Dup1.Region)
	assert.Equal(t, user1.NewRegion, user1Dup1.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup1.Shard)
	assert.Equal(t, user1.IsUser, user1Dup1.IsUser)
	assert.Nil(t, err)

	user1Dup2, err := store.getUserByCiName("ali")
	assert.Equal(t, user1.Id, user1Dup2.Id)
	assert.Equal(t, user1.Name, user1Dup2.Name)
	assert.Equal(t, user1.Created.Unix(), user1Dup2.Created.Unix())
	assert.Equal(t, user1.Region, user1Dup2.Region)
	assert.Equal(t, user1.NewRegion, user1Dup2.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup2.Shard)
	assert.Equal(t, user1.IsUser, user1Dup2.IsUser)
	assert.Equal(t, user1.Email, user1Dup2.Email)
	assert.Equal(t, user1.NewEmail, user1Dup2.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup2.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup2.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup2.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup2.resetPwdCode)
	assert.Nil(t, err)

	user1Dup3, err := store.getUserByEmail("ali@ali.com")
	assert.Equal(t, user1.Id, user1Dup3.Id)
	assert.Equal(t, user1.Name, user1Dup3.Name)
	assert.Equal(t, user1.Created.Unix(), user1Dup3.Created.Unix())
	assert.Equal(t, user1.Region, user1Dup3.Region)
	assert.Equal(t, user1.NewRegion, user1Dup3.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup3.Shard)
	assert.Equal(t, user1.IsUser, user1Dup3.IsUser)
	assert.Equal(t, user1.Email, user1Dup3.Email)
	assert.Equal(t, user1.NewEmail, user1Dup3.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup3.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup3.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup3.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup3.resetPwdCode)
	assert.Nil(t, err)

	user1Dup4, err := store.getUserById(user1.Id)
	assert.Equal(t, user1.Id, user1Dup4.Id)
	assert.Equal(t, user1.Name, user1Dup4.Name)
	assert.Equal(t, user1.Created.Unix(), user1Dup4.Created.Unix())
	assert.Equal(t, user1.Region, user1Dup4.Region)
	assert.Equal(t, user1.NewRegion, user1Dup4.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup4.Shard)
	assert.Equal(t, user1.IsUser, user1Dup4.IsUser)
	assert.Equal(t, user1.Email, user1Dup4.Email)
	assert.Equal(t, user1.NewEmail, user1Dup4.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup4.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup4.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup4.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup4.resetPwdCode)
	assert.Nil(t, err)

	pwdInfo1Dup1, err := store.getPwdInfo(user1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup1.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup1.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup1.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup1.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup1.keyLen)
	assert.Nil(t, err)

	user1.Name = "bob"
	user1.Email = "bob@bob.com"
	err = store.updateUser(user1)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)
	assert.Nil(t, err)

	user1Dup5, err := store.getUserByEmail("bob@bob.com")
	assert.Equal(t, user1.Id, user1Dup5.Id)
	assert.Equal(t, user1.Name, user1Dup5.Name)
	assert.Equal(t, user1.Created.Unix(), user1Dup5.Created.Unix())
	assert.Equal(t, user1.Region, user1Dup5.Region)
	assert.Equal(t, user1.NewRegion, user1Dup5.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup5.Shard)
	assert.Equal(t, user1.IsUser, user1Dup5.IsUser)
	assert.Equal(t, user1.Email, user1Dup5.Email)
	assert.Equal(t, user1.NewEmail, user1Dup5.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup5.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup5.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup5.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup5.resetPwdCode)
	assert.Nil(t, err)

	pwdInfo1.salt = []byte("salt_update")
	pwdInfo1.pwd = []byte("pwd_update")
	pwdInfo1.n = 5
	pwdInfo1.r = 5
	pwdInfo1.p = 5
	pwdInfo1.keyLen = 5
	err = store.updatePwdInfo(user1.Id, pwdInfo1)
	assert.Nil(t, err)

	pwdInfo1Dup2, err := store.getPwdInfo(user1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup2.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup2.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup2.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup2.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup2.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup2.keyLen)
	assert.Nil(t, err)
	

}
