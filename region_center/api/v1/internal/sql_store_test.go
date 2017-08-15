package internal

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newSqlStore_NilCriticalParamErr(t *testing.T) {
	defer func() {
		err := recover().(error)
		assert.IsType(t, &Error{}, err)
	}()
	newSqlStore(nil)
}

func Test_newSqlStore_success(t *testing.T) {
	store := newSqlStore(map[int]isql.ReplicaSet{0: &isql.MockDB{}})
	assert.NotNil(t, store)
}

//this test tests everything using a real sql db, comment/uncomment as necessary
func Test_sqlStore_adHoc(t *testing.T) {
	treeDb, _ := isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)
	store := newSqlStore(map[int]isql.ReplicaSet{0: treeDb})

	aliId := NewId()
	aliShard := store.createAccount(aliId, aliId, "ali")
	assert.Equal(t, 0, aliShard)

	groupAccountId := NewId()
	groupAccountShard := store.createAccount(groupAccountId, aliId, "ali")
	assert.Equal(t, 0, groupAccountShard)

	bob := &AddMemberInternal{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := &AddMemberInternal{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfOnlySpecificProjects
	store.addMembers(groupAccountShard, groupAccountId, []*AddMemberInternal{bob, cat})

	cat.Role = AccountOwner
	store.updateMembersAndSetActive(groupAccountShard, groupAccountId, []*AddMemberInternal{cat})

	totalOwnerCount := store.getTotalOwnerCount(groupAccountShard, groupAccountId)
	assert.Equal(t, 2, totalOwnerCount)

	ownerCountInSet := store.getOwnerCountInSet(groupAccountShard, groupAccountId, []Id{aliId})
	assert.Equal(t, 1, ownerCountInSet)

	ownerCountInSet = store.getOwnerCountInSet(groupAccountShard, groupAccountId, []Id{bob.Id})
	assert.Equal(t, 0, ownerCountInSet)

	store.setMembersInactive(groupAccountShard, groupAccountId, []Id{cat.Id})
	totalOwnerCount = store.getTotalOwnerCount(groupAccountShard, groupAccountId)
	assert.Equal(t, 1, totalOwnerCount)

	store.renameMember(groupAccountShard, groupAccountId, bob.Id, "jimbob")
	inactiveMembers := store.getAllInactiveMemberIdsFromInputSet(groupAccountShard, groupAccountId, []Id{cat.Id, bob.Id})
	assert.Equal(t, 1, len(inactiveMembers))
	assert.True(t, inactiveMembers[0].Equal(cat.Id))

	store.logActivity(groupAccountShard, groupAccountId, bob.Id, cat.Id, "member", "added")

	store.deleteAccount(groupAccountShard, groupAccountId)
	totalOwnerCount = store.getTotalOwnerCount(groupAccountShard, groupAccountId)
	assert.Equal(t, 0, totalOwnerCount) // no owners can only occur if there is no account at all
}
