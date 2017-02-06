package account

import (
	. "bitbucket.org/robsix/task_center/misc"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_newMemStore_success(t *testing.T) {
	store := newMemStore()

	assert.NotNil(t, store)
}

func Test_memStore_createOrgAndMembership(t *testing.T) {
	store, userId, _ := setup(t, newMemStore())

	org := &org{}
	org.Id, _ = NewId()
	err := store.createOrgAndMembership(userId, org)
	assert.Nil(t, err)
}

func Test_memStore_accountWithCiNameExists(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	exists, err := store.accountWithCiNameExists("ali")
	assert.True(t, exists)
	assert.Nil(t, err)

	exists, err = store.accountWithCiNameExists("bob")
	assert.True(t, exists)
	assert.Nil(t, err)

	exists, err = store.accountWithCiNameExists("cat")
	assert.False(t, exists)
	assert.Nil(t, err)
}

func Test_memStore_getUserByCiName(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	acc, err := store.getUserByCiName("ali")
	assert.Equal(t, "ali", acc.Name)
	assert.True(t, acc.IsUser)
	assert.Nil(t, err)

	acc, err = store.getUserByCiName("bob")
	assert.Nil(t, acc)
	assert.Nil(t, err)
}

func Test_memStore_getUserByEmail(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	acc, err := store.getUserByEmail("ali@ali.com")
	assert.Equal(t, "ali", acc.Name)
	assert.True(t, acc.IsUser)
	assert.Nil(t, err)

	acc, err = store.getUserByEmail("bob@bob.com")
	assert.Nil(t, acc)
	assert.Nil(t, err)
}

func Test_memStore_getUserById(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	acc, err := store.getUserById(userId)
	assert.Equal(t, "ali", acc.Name)
	assert.True(t, acc.IsUser)
	assert.Nil(t, err)

	acc, err = store.getUserById(orgId)
	assert.Nil(t, acc)
	assert.Nil(t, err)
}

func Test_memStore_getPwdInfo(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	pwdInfo, err := store.getPwdInfo(userId)
	assert.Equal(t, []byte("pwd"), pwdInfo.Pwd)
	assert.Equal(t, []byte("salt"), pwdInfo.Salt)
	assert.Nil(t, err)

	pwdInfo, err = store.getPwdInfo(orgId)
	assert.Nil(t, pwdInfo)
	assert.Nil(t, err)
}

func Test_memStore_updateUser(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	user, _ := store.getUserByCiName("ali")
	user.Name = "cat"
	err := store.updateUser(user)
	assert.Nil(t, err)

	user, _ = store.getUserById(user.Id)
	assert.Equal(t, "cat", user.Name)
}

func Test_memStore_updatePwdInfo(t *testing.T) {
	store, userId, _ := setup(t, newMemStore())

	pwdInfo, _ := store.getPwdInfo(userId)
	pwdInfo.Pwd = []byte("pwd2")
	store.updatePwdInfo(userId, pwdInfo)

	pwdInfo, _ = store.getPwdInfo(userId)
	assert.Equal(t, []byte("pwd2"), pwdInfo.Pwd)
}

func Test_memStore_deleteUser(t *testing.T) {
	store, userId, _ := setup(t, newMemStore())

	err := store.deleteUser(userId)
	assert.Nil(t, err)

	user, err := store.getUserById(userId)
	assert.Nil(t, user)
	assert.Nil(t, err)
}

func Test_memStore_getUsers(t *testing.T) {
	store, userId, _ := setup(t, newMemStore())

	users, err := store.getUsers([]Id{userId})
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "ali", users[0].Name)
	assert.Nil(t, err)
}

func Test_memStore_searchUsers(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	users, total, err := store.searchUsers("li", 0, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "ali", users[0].Name)
	assert.Nil(t, err)

	users, total, err = store.searchUsers("li", 1, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 0, len(users))
	assert.Nil(t, err)

	users, total, err = store.searchUsers("ob", 0, 100)
	assert.Equal(t, 0, len(users))
	assert.Nil(t, err)
}

func Test_memStore_getOrgById(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	org, err := store.getOrgById(orgId)
	assert.Equal(t, "bob", org.Name)
	assert.Nil(t, err)

	org, err = store.getOrgById(userId)
	assert.Nil(t, org)
	assert.Nil(t, err)
}

func Test_memStore_getOrgByName(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	org, err := store.getOrgByName("bob")
	assert.Equal(t, "bob", org.Name)
	assert.Nil(t, err)

	org, err = store.getOrgByName("ali")
	assert.Nil(t, org)
	assert.Nil(t, err)
}

func Test_memStore_updateOrg(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	org, _ := store.getOrgByName("bob")
	org.Name = "cat"
	err := store.updateOrg(org)
	assert.Nil(t, err)

	org, err = store.getOrgById(org.Id)
	assert.Equal(t, "cat", org.Name)
	assert.Nil(t, err)
}

func Test_memStore_deleteOrgAndAllAssociatedMemberships(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	err := store.deleteOrgAndAllAssociatedMemberships(orgId)
	assert.Nil(t, err)

	exists, err := store.membershipExists(userId, orgId)
	assert.False(t, exists)
	assert.Nil(t, err)

	org, err := store.getOrgById(orgId)
	assert.Nil(t, org)
	assert.Nil(t, err)
}

func Test_memStore_getOrgs(t *testing.T) {
	store, _, orgId := setup(t, newMemStore())

	orgs, err := store.getOrgs([]Id{orgId})
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, "bob", orgs[0].Name)
	assert.Nil(t, err)
}

func Test_memStore_searchOrgs(t *testing.T) {
	store, _, _ := setup(t, newMemStore())

	orgs, total, err := store.searchOrgs("ob", 0, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, "bob", orgs[0].Name)
	assert.Nil(t, err)

	orgs, total, err = store.searchOrgs("ob", 1, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 0, len(orgs))
	assert.Nil(t, err)

	orgs, total, err = store.searchOrgs("li", 0, 100)
	assert.Equal(t, 0, len(orgs))
	assert.Nil(t, err)
}

func Test_memStore_getUsersOrgs(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	orgs, total, err := store.getUsersOrgs(userId, 0, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, "bob", orgs[0].Name)
	assert.Nil(t, err)

	orgs, total, err = store.getUsersOrgs(userId, 1, 100)
	assert.Equal(t, 1, total)
	assert.Equal(t, 0, len(orgs))
	assert.Nil(t, err)

	orgs, total, err = store.getUsersOrgs(orgId, 0, 100)
	assert.Equal(t, 0, len(orgs))
	assert.Nil(t, err)
}

func Test_memStore_memberships(t *testing.T) {
	store, userId, orgId := setup(t, newMemStore())

	user := &fullUserInfo{}
	user.Id, _ = NewId()
	store.createUser(user, &pwdInfo{})
	err := store.createMembership(user.Id, orgId)
	assert.Nil(t, err)

	exists, err := store.membershipExists(user.Id, orgId)
	assert.True(t, exists)
	assert.Nil(t, err)

	err = store.deleteMembership(user.Id, orgId)
	assert.Nil(t, err)

	exists, err = store.membershipExists(user.Id, orgId)
	assert.False(t, exists)
	assert.Nil(t, err)

	org := &org{}
	org.Id, _ = NewId()
	err = store.createOrgAndMembership(userId, org)
	assert.Nil(t, err)

	err = store.createMembership(user.Id, org.Id)
	assert.Nil(t, err)
}

func setup(t *testing.T, store store) (store, Id, Id) {
	newRegion := "new"
	newEmail := "new@new.com"
	newEmailConfirmationCode := "newEmailConfirmationCode"
	activationCode := "activationCode"
	activated := time.Now().UTC()
	resetPwdCode := "resetPwdCode"

	user := &fullUserInfo{}
	user.Id, _ = NewId()
	user.Name = "ali"
	user.Email = "ali@ali.com"
	user.NewRegion = &newRegion
	user.NewEmail = &newEmail
	user.NewEmailConfirmationCode = &newEmailConfirmationCode
	user.ActivationCode = &activationCode
	user.Activated = &activated
	user.ResetPwdCode = &resetPwdCode
	user.IsUser = true
	pwdInfo := &pwdInfo{
		Pwd:    []byte("pwd"),
		Salt:   []byte("salt"),
		N:      1,
		R:      2,
		P:      3,
		KeyLen: 4,
	}
	err := store.createUser(user, pwdInfo)
	assert.Nil(t, err)

	org := &org{}
	org.Id, _ = NewId()
	org.Name = "bob"
	org.NewRegion = &newRegion
	org.IsUser = false
	err = store.createOrgAndMembership(user.Id, org)
	assert.Nil(t, err)

	exists, err := store.membershipExists(user.Id, org.Id)
	assert.True(t, exists)
	assert.Nil(t, err)

	return store, user.Id, org.Id
}
