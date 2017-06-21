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
	store := newSqlStore(map[int]isql.ReplicaSet{0:&isql.MockDB{}})
	assert.NotNil(t, store)
}

//this test tests everything using a real sql db, comment/uncomment as necessary
func Test_sqlStore_adHoc(t *testing.T) {
	treeDb, _ := isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)
	store := newSqlStore(map[int]isql.ReplicaSet{0:treeDb})

	personalId := NewId()
	personalShard := store.registerPersonalAccount(personalId)
	assert.Equal(t, 0, personalShard)

	orgId := NewId()
	orgShard := store.registerOrgAccount(orgId, personalId, "ali")
	assert.Equal(t, 0, orgShard)

	mem1 := store.getMember(orgShard, orgId, personalId)
	assert.Equal(t, personalId, mem1.Id)
	assert.Equal(t, OrgOwner, mem1.Role)
	assert.Equal(t, "ali", mem1.Name)
	assert.Equal(t, uint64(0), mem1.TotalRemainingTime)
	assert.Equal(t, uint64(0), mem1.TotalLoggedTime)
	assert.Equal(t, true, mem1.IsActive)

	bob := &AddMemberInternal{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = OrgAdmin
	cat := &AddMemberInternal{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = OrgMemberOfOnlySpecificProjects
	store.addMembers(orgShard, orgId, []*AddMemberInternal{bob, cat})

	cat.Role = OrgOwner
	store.updateMembersAndSetActive(orgShard, orgId, []*AddMemberInternal{cat})

	totalOwnerCount := store.getTotalOrgOwnerCount(orgShard, orgId)
	assert.Equal(t, 2, totalOwnerCount)

	ownerCountInSet := store.getOwnerCountInSet(orgShard, orgId, []Id{personalId})
	assert.Equal(t, 1, ownerCountInSet)

	ownerCountInSet = store.getOwnerCountInSet(orgShard, orgId, []Id{bob.Id})
	assert.Equal(t, 0, ownerCountInSet)

	store.setMembersInactive(orgShard, orgId, []Id{cat.Id})
	totalOwnerCount = store.getTotalOrgOwnerCount(orgShard, orgId)
	assert.Equal(t, 1, totalOwnerCount)

	store.renameMember(orgShard, orgId, bob.Id, "jimbob")
	mem2 := store.getMember(orgShard, orgId, bob.Id)
	assert.Equal(t, bob.Id, mem2.Id)
	assert.Equal(t, OrgAdmin, mem2.Role)
	assert.Equal(t, "jimbob", mem2.Name)
	assert.Equal(t, uint64(0), mem2.TotalRemainingTime)
	assert.Equal(t, uint64(0), mem2.TotalLoggedTime)
	assert.Equal(t, true, mem2.IsActive)

	mem3 := store.getMember(orgShard, orgId, cat.Id)
	assert.Equal(t, cat.Id, mem3.Id)
	assert.Equal(t, OrgMemberOfOnlySpecificProjects, mem3.Role)
	assert.Equal(t, "cat", mem3.Name)
	assert.Equal(t, uint64(0), mem3.TotalRemainingTime)
	assert.Equal(t, uint64(0), mem3.TotalLoggedTime)
	assert.Equal(t, false, mem3.IsActive)

	store.deleteAccount(orgShard, orgId)
	totalOwnerCount = store.getTotalOrgOwnerCount(orgShard, orgId)
	assert.Equal(t, 0, totalOwnerCount) // no owners can only occur if there is no org at all
}
