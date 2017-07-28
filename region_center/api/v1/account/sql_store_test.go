package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

	accountId := NewId()
	treeDb.Exec(`INSERT INTO accounts (id, publicProjectsEnabled) VALUES (?, ?)`, []byte(accountId), false)

	publicProjectsEnabled := store.getPublicProjectsEnabled(0, accountId)
	assert.False(t, publicProjectsEnabled)

	store.setPublicProjectsEnabled(0, accountId, true)
	publicProjectsEnabled = store.getPublicProjectsEnabled(0, accountId)
	assert.True(t, publicProjectsEnabled)

	ali := AccountMember{}
	ali.Id = NewId()
	ali.Name = "ali"
	ali.TotalRemainingTime = 7
	ali.TotalLoggedTime = 3
	ali.IsActive = true
	ali.Role = AccountOwner
	treeDb.Exec(`INSERT INTO accountMembers (account, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(ali.Id), ali.Name, ali.TotalRemainingTime, ali.TotalLoggedTime, ali.IsActive, ali.Role)

	bob := AccountMember{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.TotalRemainingTime = 5
	bob.TotalLoggedTime = 2
	bob.IsActive = true
	bob.Role = AccountAdmin
	treeDb.Exec(`INSERT INTO accountMembers (account, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(bob.Id), bob.Name, bob.TotalRemainingTime, bob.TotalLoggedTime, bob.IsActive, bob.Role)

	cat := AccountMember{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.TotalRemainingTime = 8
	cat.TotalLoggedTime = 1
	cat.IsActive = true
	cat.Role = AccountMemberOfAllProjects
	treeDb.Exec(`INSERT INTO accountMembers (account, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), []byte(cat.Id), cat.Name, cat.TotalRemainingTime, cat.TotalLoggedTime, cat.IsActive, cat.Role)

	ali2 := store.getMember(0, accountId, ali.Id)
	assert.True(t, ali.Id.Equal(ali2.Id))
	assert.Equal(t, ali.Name, ali2.Name)
	assert.Equal(t, ali.TotalRemainingTime, ali2.TotalRemainingTime)
	assert.Equal(t, ali.TotalLoggedTime, ali2.TotalLoggedTime)

	members1, count := store.getMembers(0, accountId, nil, nil, 0, 10)
	assert.Equal(t, 3, len(members1))
	assert.Equal(t, 3, count)
	assert.Equal(t, members1[0].Name, ali.Name)
	assert.Equal(t, members1[1].Name, bob.Name)
	assert.Equal(t, members1[2].Name, cat.Name)

	filterRole := AccountAdmin
	members2, count := store.getMembers(0, accountId, &filterRole, nil, 0, 10)
	assert.Equal(t, 1, len(members2))
	assert.Equal(t, 1, count)
	assert.Equal(t, members2[0].Name, bob.Name)

	filterRole = AccountAdmin
	nameContains := "cat"
	members3, count := store.getMembers(0, accountId, &filterRole, &nameContains, 0, 10)
	assert.Equal(t, 0, len(members3))
	assert.Equal(t, 0, count)

	nameContains = "a"
	members4, count := store.getMembers(0, accountId, nil, &nameContains, 0, 10)
	assert.Equal(t, 2, len(members4))
	assert.Equal(t, 2, count)
	assert.Equal(t, members4[0].Name, ali.Name)
	assert.Equal(t, members4[1].Name, cat.Name)

	activity1 := Activity{}
	activity1.OccurredOn = Now().Round(time.Second)
	activity1.Item = NewId()
	activity1.Member = ali.Id
	activity1.ItemType = "testType1"
	activity1.ItemName = "testName1"
	activity1.Action = "testAction1"
	treeDb.Exec(`INSERT INTO accountActivities (account, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), activity1.OccurredOn.UnixNano()/1000000, []byte(activity1.Item), []byte(activity1.Member), activity1.ItemType, activity1.ItemName, activity1.Action)
	activity2 := Activity{}
	activity2.OccurredOn = activity1.OccurredOn.Add(1 * time.Second)
	activity2.Item = activity1.Item
	activity2.Member = bob.Id
	activity2.ItemType = "testType2"
	activity2.ItemName = "testName2"
	activity2.Action = "testAction2"
	treeDb.Exec(`INSERT INTO accountActivities (account, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), activity2.OccurredOn.UnixNano()/1000000, []byte(activity2.Item), []byte(activity2.Member), activity2.ItemType, activity2.ItemName, activity2.Action)
	activity3 := Activity{}
	activity3.OccurredOn = activity2.OccurredOn.Add(1 * time.Second)
	activity3.Item = NewId()
	activity3.Member = cat.Id
	activity3.ItemType = "testType3"
	activity3.ItemName = "testName3"
	activity3.Action = "testAction3"
	treeDb.Exec(`INSERT INTO accountActivities (account, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(accountId), activity3.OccurredOn.UnixNano()/1000000, []byte(activity3.Item), []byte(activity3.Member), activity3.ItemType, activity3.ItemName, activity3.Action)

	activities1 := store.getActivities(0, accountId, nil, nil, nil, nil, 100)
	assert.Equal(t, 3, len(activities1))
	assert.Equal(t, activity3.ItemName, activities1[0].ItemName)
	assert.Equal(t, activity2.ItemName, activities1[1].ItemName)
	assert.Equal(t, activity1.ItemName, activities1[2].ItemName)
	assert.Equal(t, activity1.OccurredOn.Unix(), activities1[2].OccurredOn.Unix())

	activities2 := store.getActivities(0, accountId, &activity1.Item, nil, nil, nil, 100)
	assert.Equal(t, 2, len(activities2))
	assert.Equal(t, activity2.ItemName, activities2[0].ItemName)
	assert.Equal(t, activity1.ItemName, activities2[1].ItemName)

	activities3 := store.getActivities(0, accountId, &activity1.Item, &ali.Id, nil, nil, 100)
	assert.Equal(t, 1, len(activities3))
	assert.Equal(t, activity1.ItemName, activities3[0].ItemName)

	activities4 := store.getActivities(0, accountId, nil, nil, &activity1.OccurredOn, nil, 100)
	assert.Equal(t, 2, len(activities4))
	assert.Equal(t, activity2.ItemName, activities4[0].ItemName)
	assert.Equal(t, activity3.ItemName, activities4[1].ItemName)

	activities5 := store.getActivities(0, accountId, nil, nil, nil, &activity3.OccurredOn, 100)
	assert.Equal(t, 2, len(activities5))
	assert.Equal(t, activity2.ItemName, activities5[0].ItemName)
	assert.Equal(t, activity1.ItemName, activities5[1].ItemName)

	store.logActivity(0, accountId, Now(), ali.Id, accountId, "account", "setPublicProjectsEnabled", "true")

	treeDb.Exec(`CALL deleteAccount(?)`, []byte(accountId))
}
