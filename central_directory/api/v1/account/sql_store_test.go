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
	personalAccount1 := &fullPersonalAccountInfo{}
	personalAccount1.Id = NewId()
	personalAccount1.Name = "ali"
	personalAccount1.CreatedOn = time.Now().UTC()
	personalAccount1.Region = "use"
	personalAccount1.NewRegion = nil
	personalAccount1.Shard = 3
	personalAccount1.HasAvatar = true
	personalAccount1.IsPersonal = true
	personalAccount1.Email = "ali@ali.com"
	personalAccount1.Language = "en"
	personalAccount1.Theme = LightTheme
	personalAccount1.NewEmail = &str
	personalAccount1.activationCode = &str
	personalAccount1.activated = &now
	personalAccount1.newEmailConfirmationCode = &str
	personalAccount1.resetPwdCode = &str

	pwdInfo1 := &pwdInfo{}
	pwdInfo1.salt = []byte("salt")
	pwdInfo1.pwd = []byte("pwd")
	pwdInfo1.n = 10
	pwdInfo1.r = 10
	pwdInfo1.p = 10
	pwdInfo1.keyLen = 10

	store.createPersonalAccount(personalAccount1, pwdInfo1)

	val := store.accountWithCiNameExists("ali")
	assert.True(t, val)

	personalAccount1Dup1 := store.getAccountByCiName("ali")
	assert.Equal(t, personalAccount1.Id, personalAccount1Dup1.Id)
	assert.Equal(t, personalAccount1.Name, personalAccount1Dup1.Name)
	assert.Equal(t, personalAccount1.CreatedOn.Unix(), personalAccount1Dup1.CreatedOn.Unix())
	assert.Equal(t, personalAccount1.Region, personalAccount1Dup1.Region)
	assert.Equal(t, personalAccount1.NewRegion, personalAccount1Dup1.NewRegion)
	assert.Equal(t, personalAccount1.Shard, personalAccount1Dup1.Shard)
	assert.Equal(t, personalAccount1.HasAvatar, personalAccount1Dup1.HasAvatar)
	assert.Equal(t, personalAccount1.IsPersonal, personalAccount1Dup1.IsPersonal)

	personalAccount1Dup3 := store.getPersonalAccountByEmail("ali@ali.com")
	assert.Equal(t, personalAccount1.Id, personalAccount1Dup3.Id)
	assert.Equal(t, personalAccount1.Name, personalAccount1Dup3.Name)
	assert.Equal(t, personalAccount1.CreatedOn.Unix(), personalAccount1Dup3.CreatedOn.Unix())
	assert.Equal(t, personalAccount1.Region, personalAccount1Dup3.Region)
	assert.Equal(t, personalAccount1.NewRegion, personalAccount1Dup3.NewRegion)
	assert.Equal(t, personalAccount1.Shard, personalAccount1Dup3.Shard)
	assert.Equal(t, personalAccount1.HasAvatar, personalAccount1Dup3.HasAvatar)
	assert.Equal(t, personalAccount1.IsPersonal, personalAccount1Dup3.IsPersonal)
	assert.Equal(t, personalAccount1.Email, personalAccount1Dup3.Email)
	assert.Equal(t, personalAccount1.Language, personalAccount1Dup3.Language)
	assert.Equal(t, personalAccount1.Theme, personalAccount1Dup3.Theme)
	assert.Equal(t, personalAccount1.NewEmail, personalAccount1Dup3.NewEmail)
	assert.Equal(t, personalAccount1.activationCode, personalAccount1Dup3.activationCode)
	assert.Equal(t, personalAccount1.activated.Unix(), personalAccount1Dup3.activated.Unix())
	assert.Equal(t, personalAccount1.newEmailConfirmationCode, personalAccount1Dup3.newEmailConfirmationCode)
	assert.Equal(t, personalAccount1.resetPwdCode, personalAccount1Dup3.resetPwdCode)

	personalAccount1Dup4 := store.getPersonalAccountById(personalAccount1.Id)
	assert.Equal(t, personalAccount1.Id, personalAccount1Dup4.Id)
	assert.Equal(t, personalAccount1.Name, personalAccount1Dup4.Name)
	assert.Equal(t, personalAccount1.CreatedOn.Unix(), personalAccount1Dup4.CreatedOn.Unix())
	assert.Equal(t, personalAccount1.Region, personalAccount1Dup4.Region)
	assert.Equal(t, personalAccount1.NewRegion, personalAccount1Dup4.NewRegion)
	assert.Equal(t, personalAccount1.Shard, personalAccount1Dup4.Shard)
	assert.Equal(t, personalAccount1.HasAvatar, personalAccount1Dup4.HasAvatar)
	assert.Equal(t, personalAccount1.IsPersonal, personalAccount1Dup4.IsPersonal)
	assert.Equal(t, personalAccount1.Email, personalAccount1Dup4.Email)
	assert.Equal(t, personalAccount1.Language, personalAccount1Dup4.Language)
	assert.Equal(t, personalAccount1.Theme, personalAccount1Dup4.Theme)
	assert.Equal(t, personalAccount1.NewEmail, personalAccount1Dup4.NewEmail)
	assert.Equal(t, personalAccount1.activationCode, personalAccount1Dup4.activationCode)
	assert.Equal(t, personalAccount1.activated.Unix(), personalAccount1Dup4.activated.Unix())
	assert.Equal(t, personalAccount1.newEmailConfirmationCode, personalAccount1Dup4.newEmailConfirmationCode)
	assert.Equal(t, personalAccount1.resetPwdCode, personalAccount1Dup4.resetPwdCode)

	pwdInfo1Dup1 := store.getPwdInfo(personalAccount1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup1.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup1.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup1.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup1.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup1.keyLen)

	personalAccount1.Name = "bob"
	personalAccount1.Email = "bob@bob.com"
	store.updatePersonalAccount(personalAccount1)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup1.salt)

	personalAccount1Dup5 := store.getPersonalAccountByEmail("bob@bob.com")
	assert.Equal(t, personalAccount1.Id, personalAccount1Dup5.Id)
	assert.Equal(t, personalAccount1.Name, personalAccount1Dup5.Name)
	assert.Equal(t, personalAccount1.CreatedOn.Unix(), personalAccount1Dup5.CreatedOn.Unix())
	assert.Equal(t, personalAccount1.Region, personalAccount1Dup5.Region)
	assert.Equal(t, personalAccount1.NewRegion, personalAccount1Dup5.NewRegion)
	assert.Equal(t, personalAccount1.Shard, personalAccount1Dup5.Shard)
	assert.Equal(t, personalAccount1.HasAvatar, personalAccount1Dup5.HasAvatar)
	assert.Equal(t, personalAccount1.IsPersonal, personalAccount1Dup5.IsPersonal)
	assert.Equal(t, personalAccount1.Email, personalAccount1Dup5.Email)
	assert.Equal(t, personalAccount1.Language, personalAccount1Dup5.Language)
	assert.Equal(t, personalAccount1.Theme, personalAccount1Dup5.Theme)
	assert.Equal(t, personalAccount1.NewEmail, personalAccount1Dup5.NewEmail)
	assert.Equal(t, personalAccount1.activationCode, personalAccount1Dup5.activationCode)
	assert.Equal(t, personalAccount1.activated.Unix(), personalAccount1Dup5.activated.Unix())
	assert.Equal(t, personalAccount1.newEmailConfirmationCode, personalAccount1Dup5.newEmailConfirmationCode)
	assert.Equal(t, personalAccount1.resetPwdCode, personalAccount1Dup5.resetPwdCode)

	pwdInfo1.salt = []byte("salt_update")
	pwdInfo1.pwd = []byte("pwd_update")
	pwdInfo1.n = 5
	pwdInfo1.r = 5
	pwdInfo1.p = 5
	pwdInfo1.keyLen = 5
	store.updatePwdInfo(personalAccount1.Id, pwdInfo1)

	pwdInfo1Dup2 := store.getPwdInfo(personalAccount1.Id)
	assert.Equal(t, pwdInfo1.salt, pwdInfo1Dup2.salt)
	assert.Equal(t, pwdInfo1.pwd, pwdInfo1Dup2.pwd)
	assert.Equal(t, pwdInfo1.n, pwdInfo1Dup2.n)
	assert.Equal(t, pwdInfo1.r, pwdInfo1Dup2.r)
	assert.Equal(t, pwdInfo1.p, pwdInfo1Dup2.p)
	assert.Equal(t, pwdInfo1.keyLen, pwdInfo1Dup2.keyLen)

	personalAccounts1 := store.getPersonalAccounts([]Id{personalAccount1.Id})
	assert.Equal(t, 1, len(personalAccounts1))
	assert.Equal(t, personalAccount1.Id, personalAccounts1[0].Id)
	assert.Equal(t, personalAccount1.Name, personalAccounts1[0].Name)
	assert.Equal(t, personalAccount1.CreatedOn.Unix(), personalAccounts1[0].CreatedOn.Unix())
	assert.Equal(t, personalAccount1.Region, personalAccounts1[0].Region)
	assert.Equal(t, personalAccount1.NewRegion, personalAccounts1[0].NewRegion)
	assert.Equal(t, personalAccount1.Shard, personalAccounts1[0].Shard)
	assert.Equal(t, personalAccount1.HasAvatar, personalAccounts1[0].HasAvatar)
	assert.Equal(t, personalAccount1.IsPersonal, personalAccounts1[0].IsPersonal)

	acc1 := &account{}
	acc1.Id = NewId()
	acc1.Name = "acc1"
	acc1.CreatedOn = time.Now().UTC()
	acc1.Region = "use"
	acc1.NewRegion = nil
	acc1.Shard = 4
	acc1.HasAvatar = true
	acc1.IsPersonal = false
	store.createGroupAccountAndMembership(acc1, personalAccount1.Id)

	acc1Dup1 := store.getAccount(acc1.Id)
	assert.Equal(t, acc1.Id, acc1Dup1.Id)
	assert.Equal(t, acc1.Name, acc1Dup1.Name)
	assert.Equal(t, acc1.CreatedOn.Unix(), acc1Dup1.CreatedOn.Unix())
	assert.Equal(t, acc1.Region, acc1Dup1.Region)
	assert.Equal(t, acc1.NewRegion, acc1Dup1.NewRegion)
	assert.Equal(t, acc1.Shard, acc1Dup1.Shard)
	assert.Equal(t, acc1.HasAvatar, acc1Dup1.HasAvatar)
	assert.Equal(t, acc1.IsPersonal, acc1Dup1.IsPersonal)

	acc1.Name = "acc1_updated"
	store.updateAccount(acc1)

	accs1 := store.getAccounts([]Id{acc1.Id})
	assert.Equal(t, 1, len(accs1))
	assert.Equal(t, acc1.Id, accs1[0].Id)
	assert.Equal(t, acc1.Name, accs1[0].Name)
	assert.Equal(t, acc1.CreatedOn.Unix(), accs1[0].CreatedOn.Unix())
	assert.Equal(t, acc1.Region, accs1[0].Region)
	assert.Equal(t, acc1.NewRegion, accs1[0].NewRegion)
	assert.Equal(t, acc1.Shard, accs1[0].Shard)
	assert.Equal(t, acc1.HasAvatar, accs1[0].HasAvatar)
	assert.Equal(t, acc1.IsPersonal, accs1[0].IsPersonal)

	accs2, total := store.getGroupAccounts(personalAccount1.Id, 0, 50)
	assert.Equal(t, 1, len(accs2))
	assert.Equal(t, 1, total)
	assert.Equal(t, acc1.Id, accs2[0].Id)
	assert.Equal(t, acc1.Name, accs2[0].Name)
	assert.Equal(t, acc1.CreatedOn.Unix(), accs2[0].CreatedOn.Unix())
	assert.Equal(t, acc1.Region, accs2[0].Region)
	assert.Equal(t, acc1.NewRegion, accs2[0].NewRegion)
	assert.Equal(t, acc1.Shard, accs2[0].Shard)
	assert.Equal(t, acc1.HasAvatar, accs2[0].HasAvatar)
	assert.Equal(t, acc1.IsPersonal, accs2[0].IsPersonal)

	personalAccount2 := &fullPersonalAccountInfo{}
	personalAccount2.Id = NewId()
	personalAccount2.Name = "cat"
	personalAccount2.CreatedOn = time.Now().UTC()
	personalAccount2.Region = "use"
	personalAccount2.NewRegion = nil
	personalAccount2.Shard = 3
	personalAccount2.HasAvatar = false
	personalAccount2.IsPersonal = true
	personalAccount2.Email = "cat@cat.com"
	personalAccount2.Language = "de"
	personalAccount2.Theme = DarkTheme
	personalAccount2.NewEmail = &str
	personalAccount2.activationCode = &str
	personalAccount2.activated = &now
	personalAccount2.newEmailConfirmationCode = &str
	personalAccount2.resetPwdCode = &str
	store.createPersonalAccount(personalAccount2, pwdInfo1)

	personalAccount3 := &fullPersonalAccountInfo{}
	personalAccount3.Id = NewId()
	personalAccount3.Name = "dan"
	personalAccount3.CreatedOn = time.Now().UTC()
	personalAccount3.Region = "use"
	personalAccount3.NewRegion = nil
	personalAccount3.Shard = 3
	personalAccount3.HasAvatar = true
	personalAccount3.IsPersonal = true
	personalAccount3.Email = "dan@dan.com"
	personalAccount3.Language = "fr"
	personalAccount3.Theme = ColorBlindTheme
	personalAccount3.NewEmail = &str
	personalAccount3.activationCode = &str
	personalAccount3.activated = &now
	personalAccount3.newEmailConfirmationCode = &str
	personalAccount3.resetPwdCode = &str
	store.createPersonalAccount(personalAccount3, pwdInfo1)

	acc2 := &account{}
	acc2.Id = NewId()
	acc2.Name = "acc2"
	acc2.CreatedOn = time.Now().UTC()
	acc2.Region = "use"
	acc2.NewRegion = nil
	acc2.Shard = 4
	acc2.HasAvatar = true
	acc2.IsPersonal = false
	store.createGroupAccountAndMembership(acc2, personalAccount1.Id)

	store.createMemberships(acc2.Id, []Id{personalAccount2.Id, personalAccount3.Id})

	accs3, total := store.getGroupAccounts(personalAccount2.Id, 0, 50)
	assert.Equal(t, 1, len(accs3))
	assert.Equal(t, 1, total)
	assert.Equal(t, acc2.Id, accs3[0].Id)
	assert.Equal(t, acc2.Name, accs3[0].Name)
	assert.Equal(t, acc2.CreatedOn.Unix(), accs3[0].CreatedOn.Unix())
	assert.Equal(t, acc2.Region, accs3[0].Region)
	assert.Equal(t, acc2.NewRegion, accs3[0].NewRegion)
	assert.Equal(t, acc2.Shard, accs3[0].Shard)
	assert.Equal(t, acc2.HasAvatar, accs3[0].HasAvatar)
	assert.Equal(t, acc2.IsPersonal, accs3[0].IsPersonal)

	store.deleteMemberships(acc2.Id, []Id{personalAccount2.Id, personalAccount3.Id})

	accs4, total := store.getGroupAccounts(personalAccount2.Id, 0, 50)
	assert.Equal(t, 0, len(accs4))
	assert.Equal(t, 0, total)

	store.deleteAccountAndAllAssociatedMemberships(acc2.Id)

	accs5, total := store.getGroupAccounts(personalAccount1.Id, 0, 50)
	assert.Equal(t, 1, len(accs5))
	assert.Equal(t, 1, total)

	store.deleteAccountAndAllAssociatedMemberships(personalAccount1.Id)

	accs6, total := store.getGroupAccounts(personalAccount1.Id, 0, 50)
	assert.Equal(t, 0, len(accs6))
	assert.Equal(t, 0, total)

	// clean up left over test data
	store.deleteAccountAndAllAssociatedMemberships(personalAccount2.Id)
	store.deleteAccountAndAllAssociatedMemberships(personalAccount3.Id)
	store.deleteAccountAndAllAssociatedMemberships(acc1.Id)
}
