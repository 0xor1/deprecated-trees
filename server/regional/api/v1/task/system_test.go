package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func Test_system(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "t_r_trees:T@sk-Tr335@tcp(127.0.0.1:3307)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	privateApi := private.New(shards, maxProcessEntityCount)
	projectApi := project.New(shards, maxProcessEntityCount)
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
	start := Now()
	end := start.Add(5*24*time.Hour)
	project := projectApi.CreateProject(0, orgId, ali.Id, "proj", "projDesc", &start, &end, true, false, []*AddProjectMember{{Id: ali.Id, Role: ProjectAdmin}, {Id: bob.Id, Role: ProjectAdmin}, {Id: cat.Id, Role: ProjectWriter}, {Id: dan.Id, Role: ProjectReader}})

	falseVal := false
	//create node needs extensive testing to test every avenue of the stored procedure
	firstNode := api.CreateNode(0, orgId, project.Id, project.Id, ali.Id, nil, "first", "firstDesc", true, &falseVal, nil, nil)
	assert.Equal(t, true, firstNode.IsAbstract)
	assert.Equal(t, "first", firstNode.Name)
	assert.Equal(t, "firstDesc", firstNode.Description)
	assert.InDelta(t, Now().Unix(), firstNode.CreatedOn.Unix(), 100)
	assert.Equal(t, uint64(0), firstNode.TotalRemainingTime)
	assert.Equal(t, uint64(0), firstNode.TotalLoggedTime)
	assert.Equal(t, uint64(0), firstNode.MinimumRemainingTime)
	assert.Equal(t, uint64(0), firstNode.LinkedFileCount)
	assert.Equal(t, uint64(0), firstNode.ChatCount)
	assert.Equal(t, uint64(0), *firstNode.ChildCount)
	assert.Equal(t, uint64(0), *firstNode.DescendantCount)
	assert.Equal(t, false, *firstNode.IsParallel)
	assert.Nil(t, firstNode.Member)

	api.SetName(0, orgId, project.Id, firstNode.Id, ali.Id, "firstRenamed")
	api.SetDescription(0, orgId, project.Id, firstNode.Id, ali.Id, "firstChangedDesc")
	api.SetIsParallel(0, orgId, project.Id, firstNode.Id, ali.Id, true)

	zero := uint64(0)
	secondNode := api.CreateNode(0, orgId, project.Id, firstNode.Id, ali.Id, nil, "second", "secondDesc", false, nil, nil, &zero)
	api.SetMember(0, orgId, project.Id, secondNode.Id, ali.Id, &bob.Id)
	api.SetTimeRemaining(0, orgId, project.Id, secondNode.Id, ali.Id, 7600)

	api.DeleteNode(0, orgId, project.Id, firstNode.Id, ali.Id)
	privateApi.DeleteAccount(0, orgId, ali.Id)
}
