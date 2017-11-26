package project

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bitbucket.org/0xor1/task/server/regional/api/v1/account"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_system(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "t_r_trees:T@sk-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	privateApi := private.New(shards, maxProcessEntityCount)
	accountApi := account.New(shards, maxProcessEntityCount)
	api := New(shards, maxProcessEntityCount)

	orgId := NewId()
	ali := AddMemberPrivate{}
	ali.Id = NewId()
	ali.Name = "ali"
	ali.Role = AccountOwner
	bob := AddMemberPrivate{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := AddMemberPrivate{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfAllProjects
	dan := AddMemberPrivate{}
	dan.Id = NewId()
	dan.Name = "dan"
	dan.Role = AccountMemberOfOnlySpecificProjects
	privateApi.CreateAccount(orgId, ali.Id, ali.Name, nil)
	privateApi.AddMembers(0, orgId, ali.Id, []*AddMemberPrivate{&bob, &cat, &dan})
	accountApi.SetPublicProjectsEnabled(0, orgId, ali.Id, true)

	proj := api.CreateProject(0, orgId, ali.Id, "a-p1", "p1_desc", nil, nil, true, false, nil)
	proj2 := api.CreateProject(0, orgId, ali.Id, "b-p2", "p2_desc", nil, nil, true, false, nil)
	proj3 := api.CreateProject(0, orgId, ali.Id, "c-p3", "p3_desc", nil, nil, true, false, nil)
	api.SetName(0, orgId, proj.Id, ali.Id, "a-p1_new")
	api.SetDescription(0, orgId, proj.Id, ali.Id, "a-p1_desc_new")
	api.SetIsParallel(0, orgId, proj.Id, ali.Id, false)
	api.SetIsPublic(0, orgId, proj.Id, ali.Id, true)
	proj = api.GetProject(0, orgId, proj.Id, ali.Id)
	assert.Equal(t, "a-p1_new", proj.Name)
	assert.Equal(t, "a-p1_desc_new", proj.Description)
	assert.Equal(t, false, proj.IsParallel)
	assert.Equal(t, true, proj.IsPublic)
	projs, more := api.GetProjects(0, orgId, ali.Id, nil, nil, nil, nil, nil, nil, nil, false, SortByCreatedOn, SortDirAsc, nil, 1)
	assert.Equal(t, 1, len(projs))
	assert.True(t, more)
	assert.Equal(t, proj.Name, projs[0].Name)
	projs, more = api.GetProjects(0, orgId, ali.Id, nil, nil, nil, nil, nil, nil, nil, false, SortByCreatedOn, SortDirAsc, &proj.Id, 100)
	assert.Equal(t, 2, len(projs))
	assert.False(t, more)
	assert.Equal(t, proj2.Name, projs[0].Name)
	assert.Equal(t, proj3.Name, projs[1].Name)
	api.ArchiveProject(0, orgId, proj.Id, ali.Id)
	projs, more = api.GetProjects(0, orgId, ali.Id, nil, nil, nil, nil, nil, nil, nil, true, SortByCreatedOn, SortDirAsc, nil, 100)
	assert.Equal(t, 1, len(projs))
	assert.False(t, more)
	api.UnarchiveProject(0, orgId, proj.Id, ali.Id)
	aliP := &addMember{}
	aliP.Id = ali.Id
	aliP.Role = ProjectAdmin
	bobP := &addMember{}
	bobP.Id = bob.Id
	bobP.Role = ProjectWriter
	catP := &addMember{}
	catP.Id = cat.Id
	catP.Role = ProjectReader
	api.AddMembers(0, orgId, proj.Id, ali.Id, []*addMember{aliP, bobP, catP})
	api.SetMemberRole(0, orgId, proj.Id, ali.Id, bobP.Id, ProjectReader)
	mems, more := api.GetMembers(0, orgId, proj.Id, ali.Id, nil, nil, nil, 100)
	assert.Equal(t, 3, len(mems))
	assert.False(t, more)
	assert.True(t, mems[0].Id.Equal(ali.Id))
	assert.True(t, mems[1].Id.Equal(bob.Id))
	assert.True(t, mems[2].Id.Equal(cat.Id))
	bobMe := api.GetMe(0, orgId, proj.Id, bob.Id)
	assert.True(t, bobMe.Id.Equal(bob.Id))
	activities := api.GetActivities(0, orgId, proj.Id, ali.Id, nil, nil, nil, nil, 100)
	assert.Equal(t, 11, len(activities))
	api.RemoveMembers(0, orgId, proj.Id, ali.Id, []Id{bob.Id, cat.Id})
	api.DeleteProject(0, orgId, proj.Id, ali.Id)

	privateApi.DeleteAccount(0, orgId, ali.Id)
}
