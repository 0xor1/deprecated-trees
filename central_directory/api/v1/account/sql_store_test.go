package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"database/sql"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
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

	rawAccountsDb, _ := sql.Open("mysql", "tc_cd_accounts:T@sk-C3n-T3r-@cc-0unt5@tcp(127.0.0.1:3306)/accounts?parseTime=true&loc=UTC&multiStatements=true")
	rawPwdsDb, _ := sql.Open("mysql", "tc_cd_pwds:T@sk-C3n-T3r-Pwd5@tcp(127.0.0.1:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	accountsDb, pwdsDb := isql.NewDB(rawAccountsDb), isql.NewDB(rawPwdsDb)
	store := newSqlStore(accountsDb, pwdsDb)

	str := "str"
	now := time.Now()
	user1 := &fullUserInfo{}
	user1.Id = NewId()
	user1.Name = "ali"
	user1.CreatedOn = time.Now().UTC()
	user1.Region = "use"
	user1.NewRegion = nil
	user1.Shard = 3
	user1.HasAvatar = true
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

	store.createUser(user1, pwdInfo1)

	val := store.accountWithCiNameExists("ali")
	assert.True(t, val)

	user1Dup1 := store.getAccountByCiName("ali")
	assert.Equal(t, user1.Id, user1Dup1.Id)
	assert.Equal(t, user1.Name, user1Dup1.Name)
	assert.Equal(t, user1.CreatedOn.Unix(), user1Dup1.CreatedOn.Unix())
	assert.Equal(t, user1.Region, user1Dup1.Region)
	assert.Equal(t, user1.NewRegion, user1Dup1.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup1.Shard)
	assert.Equal(t, user1.HasAvatar, user1Dup1.HasAvatar)
	assert.Equal(t, user1.IsUser, user1Dup1.IsUser)

	user1Dup3 := store.getUserByEmail("ali@ali.com")
	assert.Equal(t, user1.Id, user1Dup3.Id)
	assert.Equal(t, user1.Name, user1Dup3.Name)
	assert.Equal(t, user1.CreatedOn.Unix(), user1Dup3.CreatedOn.Unix())
	assert.Equal(t, user1.Region, user1Dup3.Region)
	assert.Equal(t, user1.NewRegion, user1Dup3.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup3.Shard)
	assert.Equal(t, user1.HasAvatar, user1Dup3.HasAvatar)
	assert.Equal(t, user1.IsUser, user1Dup3.IsUser)
	assert.Equal(t, user1.Email, user1Dup3.Email)
	assert.Equal(t, user1.NewEmail, user1Dup3.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup3.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup3.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup3.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup3.resetPwdCode)

	user1Dup4 := store.getUserById(user1.Id)
	assert.Equal(t, user1.Id, user1Dup4.Id)
	assert.Equal(t, user1.Name, user1Dup4.Name)
	assert.Equal(t, user1.CreatedOn.Unix(), user1Dup4.CreatedOn.Unix())
	assert.Equal(t, user1.Region, user1Dup4.Region)
	assert.Equal(t, user1.NewRegion, user1Dup4.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup4.Shard)
	assert.Equal(t, user1.HasAvatar, user1Dup4.HasAvatar)
	assert.Equal(t, user1.IsUser, user1Dup4.IsUser)
	assert.Equal(t, user1.Email, user1Dup4.Email)
	assert.Equal(t, user1.NewEmail, user1Dup4.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup4.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup4.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup4.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup4.resetPwdCode)

	pwdInfo1Dup1 := store.getPwdInfo(user1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup1.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup1.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup1.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup1.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup1.keyLen)

	user1.Name = "bob"
	user1.Email = "bob@bob.com"
	store.updateUser(user1)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)

	user1Dup5 := store.getUserByEmail("bob@bob.com")
	assert.Equal(t, user1.Id, user1Dup5.Id)
	assert.Equal(t, user1.Name, user1Dup5.Name)
	assert.Equal(t, user1.CreatedOn.Unix(), user1Dup5.CreatedOn.Unix())
	assert.Equal(t, user1.Region, user1Dup5.Region)
	assert.Equal(t, user1.NewRegion, user1Dup5.NewRegion)
	assert.Equal(t, user1.Shard, user1Dup5.Shard)
	assert.Equal(t, user1.HasAvatar, user1Dup5.HasAvatar)
	assert.Equal(t, user1.IsUser, user1Dup5.IsUser)
	assert.Equal(t, user1.Email, user1Dup5.Email)
	assert.Equal(t, user1.NewEmail, user1Dup5.NewEmail)
	assert.Equal(t, user1.activationCode, user1Dup5.activationCode)
	assert.Equal(t, user1.activated.Unix(), user1Dup5.activated.Unix())
	assert.Equal(t, user1.newEmailConfirmationCode, user1Dup5.newEmailConfirmationCode)
	assert.Equal(t, user1.resetPwdCode, user1Dup5.resetPwdCode)

	pwdInfo1.salt = []byte("salt_update")
	pwdInfo1.pwd = []byte("pwd_update")
	pwdInfo1.n = 5
	pwdInfo1.r = 5
	pwdInfo1.p = 5
	pwdInfo1.keyLen = 5
	store.updatePwdInfo(user1.Id, pwdInfo1)

	pwdInfo1Dup2 := store.getPwdInfo(user1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup2.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup2.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup2.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup2.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup2.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup2.keyLen)

	users1 := store.getUsers([]Id{user1.Id})
	assert.Equal(t, 1, len(users1))
	assert.Equal(t, user1.Id, users1[0].Id)
	assert.Equal(t, user1.Name, users1[0].Name)
	assert.Equal(t, user1.CreatedOn.Unix(), users1[0].CreatedOn.Unix())
	assert.Equal(t, user1.Region, users1[0].Region)
	assert.Equal(t, user1.NewRegion, users1[0].NewRegion)
	assert.Equal(t, user1.Shard, users1[0].Shard)
	assert.Equal(t, user1.HasAvatar, users1[0].HasAvatar)
	assert.Equal(t, user1.IsUser, users1[0].IsUser)

	org1 := &account{}
	org1.Id = NewId()
	org1.Name = "org1"
	org1.CreatedOn = time.Now().UTC()
	org1.Region = "use"
	org1.NewRegion = nil
	org1.Shard = 4
	org1.HasAvatar = true
	org1.IsUser = false
	store.createOrgAndMembership(org1, user1.Id)

	org1Dup1 := store.getAccount(org1.Id)
	assert.Equal(t, org1.Id, org1Dup1.Id)
	assert.Equal(t, org1.Name, org1Dup1.Name)
	assert.Equal(t, org1.CreatedOn.Unix(), org1Dup1.CreatedOn.Unix())
	assert.Equal(t, org1.Region, org1Dup1.Region)
	assert.Equal(t, org1.NewRegion, org1Dup1.NewRegion)
	assert.Equal(t, org1.Shard, org1Dup1.Shard)
	assert.Equal(t, org1.HasAvatar, org1Dup1.HasAvatar)
	assert.Equal(t, org1.IsUser, org1Dup1.IsUser)

	org1.Name = "org1_updated"
	store.updateAccount(org1)

	orgs1 := store.getAccounts([]Id{org1.Id})
	assert.Equal(t, 1, len(orgs1))
	assert.Equal(t, org1.Id, orgs1[0].Id)
	assert.Equal(t, org1.Name, orgs1[0].Name)
	assert.Equal(t, org1.CreatedOn.Unix(), orgs1[0].CreatedOn.Unix())
	assert.Equal(t, org1.Region, orgs1[0].Region)
	assert.Equal(t, org1.NewRegion, orgs1[0].NewRegion)
	assert.Equal(t, org1.Shard, orgs1[0].Shard)
	assert.Equal(t, org1.HasAvatar, orgs1[0].HasAvatar)
	assert.Equal(t, org1.IsUser, orgs1[0].IsUser)

	orgs2, total := store.getUsersOrgs(user1.Id, 0, 50)
	assert.Equal(t, 1, len(orgs2))
	assert.Equal(t, 1, total)
	assert.Equal(t, org1.Id, orgs2[0].Id)
	assert.Equal(t, org1.Name, orgs2[0].Name)
	assert.Equal(t, org1.CreatedOn.Unix(), orgs2[0].CreatedOn.Unix())
	assert.Equal(t, org1.Region, orgs2[0].Region)
	assert.Equal(t, org1.NewRegion, orgs2[0].NewRegion)
	assert.Equal(t, org1.Shard, orgs2[0].Shard)
	assert.Equal(t, org1.HasAvatar, orgs2[0].HasAvatar)
	assert.Equal(t, org1.IsUser, orgs2[0].IsUser)

	user2 := &fullUserInfo{}
	user2.Id = NewId()
	user2.Name = "cat"
	user2.CreatedOn = time.Now().UTC()
	user2.Region = "use"
	user2.NewRegion = nil
	user2.Shard = 3
	user2.HasAvatar = false
	user2.IsUser = true
	user2.Email = "cat@cat.com"
	user2.NewEmail = &str
	user2.activationCode = &str
	user2.activated = &now
	user2.newEmailConfirmationCode = &str
	user2.resetPwdCode = &str
	store.createUser(user2, pwdInfo1)

	user3 := &fullUserInfo{}
	user3.Id = NewId()
	user3.Name = "dan"
	user3.CreatedOn = time.Now().UTC()
	user3.Region = "use"
	user3.NewRegion = nil
	user3.Shard = 3
	user3.HasAvatar = true
	user3.IsUser = true
	user3.Email = "dan@dan.com"
	user3.NewEmail = &str
	user3.activationCode = &str
	user3.activated = &now
	user3.newEmailConfirmationCode = &str
	user3.resetPwdCode = &str
	store.createUser(user3, pwdInfo1)

	org2 := &account{}
	org2.Id = NewId()
	org2.Name = "org2"
	org2.CreatedOn = time.Now().UTC()
	org2.Region = "use"
	org2.NewRegion = nil
	org2.Shard = 4
	org2.HasAvatar = true
	org2.IsUser = false
	store.createOrgAndMembership(org2, user1.Id)

	store.createMemberships(org2.Id, []Id{user2.Id, user3.Id})

	orgs3, total := store.getUsersOrgs(user2.Id, 0, 50)
	assert.Equal(t, 1, len(orgs3))
	assert.Equal(t, 1, total)
	assert.Equal(t, org2.Id, orgs3[0].Id)
	assert.Equal(t, org2.Name, orgs3[0].Name)
	assert.Equal(t, org2.CreatedOn.Unix(), orgs3[0].CreatedOn.Unix())
	assert.Equal(t, org2.Region, orgs3[0].Region)
	assert.Equal(t, org2.NewRegion, orgs3[0].NewRegion)
	assert.Equal(t, org2.Shard, orgs3[0].Shard)
	assert.Equal(t, org2.HasAvatar, orgs3[0].HasAvatar)
	assert.Equal(t, org2.IsUser, orgs3[0].IsUser)

	store.deleteMemberships(org2.Id, []Id{user2.Id, user3.Id})

	orgs4, total := store.getUsersOrgs(user2.Id, 0, 50)
	assert.Equal(t, 0, len(orgs4))
	assert.Equal(t, 0, total)

	store.deleteAccountAndAllAssociatedMemberships(org2.Id)

	orgs5, total := store.getUsersOrgs(user1.Id, 0, 50)
	assert.Equal(t, 1, len(orgs5))
	assert.Equal(t, 1, total)

	store.deleteAccountAndAllAssociatedMemberships(user1.Id)

	orgs6, total := store.getUsersOrgs(user1.Id, 0, 50)
	assert.Equal(t, 0, len(orgs6))
	assert.Equal(t, 0, total)

	// clean up left over test data
	store.deleteAccountAndAllAssociatedMemberships(user2.Id)
	store.deleteAccountAndAllAssociatedMemberships(user3.Id)
	store.deleteAccountAndAllAssociatedMemberships(org1.Id)
}
