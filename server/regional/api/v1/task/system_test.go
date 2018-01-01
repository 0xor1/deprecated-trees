package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	end := start.Add(5 * 24 * time.Hour)
	desc := "desc"
	project := projectApi.CreateProject(0, orgId, ali.Id, "proj", &desc, &start, &end, true, false, []*AddProjectMember{{Id: ali.Id, Role: ProjectAdmin}, {Id: bob.Id, Role: ProjectAdmin}, {Id: cat.Id, Role: ProjectWriter}, {Id: dan.Id, Role: ProjectReader}})

	oneVal := uint64(1)
	twoVal := uint64(2)
	threeVal := uint64(3)
	fourVal := uint64(4)
	falseVal := false
	trueVal := true
	//create node needs extensive testing to test every avenue of the stored procedure
	nodeA := api.CreateNode(0, orgId, project.Id, project.Id, ali.Id, nil, "A", &desc, true, &falseVal, nil, nil)
	assert.Equal(t, true, nodeA.IsAbstract)
	assert.Equal(t, "A", nodeA.Name)
	assert.Equal(t, "desc", *nodeA.Description)
	assert.InDelta(t, Now().Unix(), nodeA.CreatedOn.Unix(), 100)
	assert.Equal(t, uint64(0), nodeA.TotalRemainingTime)
	assert.Equal(t, uint64(0), nodeA.TotalLoggedTime)
	assert.Equal(t, uint64(0), *nodeA.MinimumRemainingTime)
	assert.Equal(t, uint64(0), nodeA.LinkedFileCount)
	assert.Equal(t, uint64(0), nodeA.ChatCount)
	assert.Equal(t, uint64(0), *nodeA.ChildCount)
	assert.Equal(t, uint64(0), *nodeA.DescendantCount)
	assert.Equal(t, false, *nodeA.IsParallel)
	assert.Nil(t, nodeA.Member)
	api.CreateNode(0, orgId, project.Id, project.Id, ali.Id, &nodeA.Id, "B", &desc, true, &falseVal, nil, nil)
	nodeC := api.CreateNode(0, orgId, project.Id, project.Id, ali.Id, nil, "C", &desc, true, &trueVal, nil, nil)
	api.CreateNode(0, orgId, project.Id, project.Id, ali.Id, &nodeA.Id, "D", &desc, false, nil, &ali.Id, &fourVal)
	nodeE := api.CreateNode(0, orgId, project.Id, nodeC.Id, ali.Id, nil, "E", &desc, false, nil, &ali.Id, &twoVal)
	api.CreateNode(0, orgId, project.Id, nodeC.Id, ali.Id, &nodeE.Id, "F", &desc, false, nil, &ali.Id, &oneVal)
	nodeG := api.CreateNode(0, orgId, project.Id, nodeC.Id, ali.Id, nil, "G", &desc, false, nil, &ali.Id, &fourVal)
	api.CreateNode(0, orgId, project.Id, nodeC.Id, ali.Id, &nodeE.Id, "H", &desc, false, nil, &ali.Id, &threeVal)
	nodeI := api.CreateNode(0, orgId, project.Id, nodeA.Id, ali.Id, nil, "I", &desc, false, nil, &ali.Id, &twoVal)
	api.CreateNode(0, orgId, project.Id, nodeA.Id, ali.Id, &nodeI.Id, "J", &desc, false, nil, &ali.Id, &oneVal)
	api.CreateNode(0, orgId, project.Id, nodeA.Id, ali.Id, nil, "K", &desc, false, nil, &ali.Id, &fourVal)
	nodeL := api.CreateNode(0, orgId, project.Id, nodeA.Id, ali.Id, &nodeI.Id, "L", &desc, true, &trueVal, nil, nil)
	nodeM := api.CreateNode(0, orgId, project.Id, nodeL.Id, ali.Id, nil, "M", &desc, false, nil, &ali.Id, &threeVal)

	api.SetName(0, orgId, project.Id, project.Id, ali.Id, "Proj-renamed")
	api.SetName(0, orgId, project.Id, nodeA.Id, ali.Id, "A-renamed")
	api.SetDescription(0, orgId, project.Id, nodeA.Id, ali.Id, nil)
	api.SetIsParallel(0, orgId, project.Id, nodeA.Id, ali.Id, true)
	api.SetIsParallel(0, orgId, project.Id, project.Id, ali.Id, false)
	api.SetMember(0, orgId, project.Id, nodeM.Id, ali.Id, &bob.Id)
	api.SetMember(0, orgId, project.Id, nodeM.Id, ali.Id, &cat.Id)
	api.SetMember(0, orgId, project.Id, nodeM.Id, ali.Id, nil)
	api.SetMember(0, orgId, project.Id, nodeM.Id, ali.Id, &cat.Id)
	api.SetTimeRemaining(0, orgId, project.Id, nodeG.Id, cat.Id, 1)
	note := "word up!"
	tl := api.SetTimeRemainingAndLogTime(0, orgId, project.Id, nodeG.Id, 30, cat.Id, 40, &note)
	assert.Equal(t, uint64(40), tl.Duration)
	privateApi.DeleteAccount(0, orgId, ali.Id)
}
