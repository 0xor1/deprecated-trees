package org

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
	store := newSqlStore(map[int]isql.ReplicaSet{0:treeDb})

	orgId := NewId()
	treeDb.Exec(`INSERT INTO orgs (id, publicProjectsEnabled) VALUES (?, ?)`, []byte(orgId), false)

	publicProjectsEnabled := store.getPublicProjectsEnabled(0, orgId)
	assert.False(t, publicProjectsEnabled)

	store.setPublicProjectsEnabled(0, orgId, true)
	publicProjectsEnabled = store.getPublicProjectsEnabled(0, orgId)
	assert.True(t, publicProjectsEnabled)

	ali := Member{}
	ali.Id = NewId()
	ali.Name = "ali"
	ali.TotalRemainingTime = 7
	ali.TotalLoggedTime = 3
	ali.IsActive = true
	ali.Role = OrgOwner
	treeDb.Exec(`INSERT INTO orgMembers (org, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), []byte(ali.Id), ali.Name, ali.TotalRemainingTime, ali.TotalLoggedTime, ali.IsActive, ali.Role)

	bob := Member{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.TotalRemainingTime = 5
	bob.TotalLoggedTime = 2
	bob.IsActive = true
	bob.Role = OrgAdmin
	treeDb.Exec(`INSERT INTO orgMembers (org, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), []byte(bob.Id), bob.Name, bob.TotalRemainingTime, bob.TotalLoggedTime, bob.IsActive, bob.Role)

	cat := Member{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.TotalRemainingTime = 8
	cat.TotalLoggedTime = 1
	cat.IsActive = true
	cat.Role = OrgMemberOfAllProjects
	treeDb.Exec(`INSERT INTO orgMembers (org, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), []byte(cat.Id), cat.Name, cat.TotalRemainingTime, cat.TotalLoggedTime, cat.IsActive, cat.Role)

	ali2 := store.getMember(0, orgId, ali.Id)
	assert.True(t, ali.Id.Equal(ali2.Id))
	assert.Equal(t, ali.Name, ali2.Name)
	assert.Equal(t, ali.TotalRemainingTime, ali2.TotalRemainingTime)
	assert.Equal(t, ali.TotalLoggedTime, ali2.TotalLoggedTime)

	members1, count := store.getMembers(0, orgId, nil, nil, 0, 10)
	assert.Equal(t, 3, len(members1))
	assert.Equal(t, 3, count)
	assert.Equal(t, members1[0].Name, ali.Name)
	assert.Equal(t, members1[1].Name, bob.Name)
	assert.Equal(t, members1[2].Name, cat.Name)

	filterRole := OrgAdmin
	members2, count := store.getMembers(0, orgId, &filterRole, nil, 0, 10)
	assert.Equal(t, 1, len(members2))
	assert.Equal(t, 1, count)
	assert.Equal(t, members2[0].Name, bob.Name)

	filterRole = OrgAdmin
	nameContains := "cat"
	members3, count := store.getMembers(0, orgId, &filterRole, &nameContains, 0, 10)
	assert.Equal(t, 0, len(members3))
	assert.Equal(t, 0, count)

	nameContains = "a"
	members4, count := store.getMembers(0, orgId, nil, &nameContains, 0, 10)
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
	treeDb.Exec(`INSERT INTO orgActivities (org, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), activity1.OccurredOn, []byte(activity1.Item), []byte(activity1.Member), activity1.ItemType, activity1.ItemName, activity1.Action)
	activity2 := Activity{}
	activity2.OccurredOn = activity1.OccurredOn.Add(1*time.Second)
	activity2.Item = activity1.Item
	activity2.Member = bob.Id
	activity2.ItemType = "testType2"
	activity2.ItemName = "testName2"
	activity2.Action = "testAction2"
	treeDb.Exec(`INSERT INTO orgActivities (org, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), activity2.OccurredOn, []byte(activity2.Item), []byte(activity2.Member), activity2.ItemType, activity2.ItemName, activity2.Action)
	activity3 := Activity{}
	activity3.OccurredOn = activity2.OccurredOn.Add(1*time.Second)
	activity3.Item = NewId()
	activity3.Member = cat.Id
	activity3.ItemType = "testType3"
	activity3.ItemName = "testName3"
	activity3.Action = "testAction3"
	treeDb.Exec(`INSERT INTO orgActivities (org, occurredOn, item, member, itemType, itemName, action) VALUES (?, ?, ?, ?, ?, ?, ?)`, []byte(orgId), activity3.OccurredOn, []byte(activity3.Item), []byte(activity3.Member), activity3.ItemType, activity3.ItemName, activity3.Action)

	activities1 := store.getActivities(0, orgId, nil, nil, nil, nil, 100)
	assert.Equal(t, 3, len(activities1))
	assert.Equal(t, activity3.ItemName, activities1[0].ItemName)
	assert.Equal(t, activity2.ItemName, activities1[1].ItemName)
	assert.Equal(t, activity1.ItemName, activities1[2].ItemName)

	activities2 := store.getActivities(0, orgId, &activity1.Item, nil, nil, nil, 100)
	assert.Equal(t, 2, len(activities2))
	assert.Equal(t, activity2.ItemName, activities2[0].ItemName)
	assert.Equal(t, activity1.ItemName, activities2[1].ItemName)

	activities3 := store.getActivities(0, orgId, &activity1.Item, &ali.Id, nil, nil, 100)
	assert.Equal(t, 1, len(activities3))
	assert.Equal(t, activity1.ItemName, activities3[0].ItemName)

	activities4 := store.getActivities(0, orgId, nil, nil, &activity1.OccurredOn, nil, 100)
	assert.Equal(t, 2, len(activities4))
	assert.Equal(t, activity2.ItemName, activities4[0].ItemName)
	assert.Equal(t, activity3.ItemName, activities4[1].ItemName)

	activities5 := store.getActivities(0, orgId, nil, nil, nil, &activity3.OccurredOn, 100)
	assert.Equal(t, 2, len(activities5))
	assert.Equal(t, activity2.ItemName, activities5[0].ItemName)
	assert.Equal(t, activity1.ItemName, activities5[1].ItemName)

	treeDb.Exec(`CALL deleteAccount(?)`, []byte(orgId))
}
