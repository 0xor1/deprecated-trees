package task

import (
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	. "bitbucket.org/0xor1/task/server/util"
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
	//create task needs extensive testing to test every avenue of the stored procedure
	taskA := api.CreateTask(0, orgId, project.Id, project.Id, ali.Id, nil, "A", &desc, true, &falseVal, nil, nil)
	assert.Equal(t, true, taskA.IsAbstract)
	assert.Equal(t, "A", taskA.Name)
	assert.Equal(t, "desc", *taskA.Description)
	assert.InDelta(t, Now().Unix(), taskA.CreatedOn.Unix(), 100)
	assert.Equal(t, uint64(0), taskA.TotalRemainingTime)
	assert.Equal(t, uint64(0), taskA.TotalLoggedTime)
	assert.Equal(t, uint64(0), *taskA.MinimumRemainingTime)
	assert.Equal(t, uint64(0), taskA.LinkedFileCount)
	assert.Equal(t, uint64(0), taskA.ChatCount)
	assert.Equal(t, uint64(0), *taskA.ChildCount)
	assert.Equal(t, uint64(0), *taskA.DescendantCount)
	assert.Equal(t, false, *taskA.IsParallel)
	assert.Nil(t, taskA.Member)
	api.CreateTask(0, orgId, project.Id, project.Id, ali.Id, &taskA.Id, "B", &desc, true, &falseVal, nil, nil)
	taskC := api.CreateTask(0, orgId, project.Id, project.Id, ali.Id, nil, "C", &desc, true, &trueVal, nil, nil)
	taskD := api.CreateTask(0, orgId, project.Id, project.Id, ali.Id, &taskA.Id, "D", &desc, false, nil, &ali.Id, &fourVal)
	taskE := api.CreateTask(0, orgId, project.Id, taskC.Id, ali.Id, nil, "E", &desc, false, nil, &ali.Id, &twoVal)
	taskF := api.CreateTask(0, orgId, project.Id, taskC.Id, ali.Id, &taskE.Id, "F", &desc, false, nil, &ali.Id, &oneVal)
	taskG := api.CreateTask(0, orgId, project.Id, taskC.Id, ali.Id, nil, "G", &desc, false, nil, &ali.Id, &fourVal)
	taskH := api.CreateTask(0, orgId, project.Id, taskC.Id, ali.Id, &taskE.Id, "H", &desc, false, nil, &ali.Id, &threeVal)
	taskI := api.CreateTask(0, orgId, project.Id, taskA.Id, ali.Id, nil, "I", &desc, false, nil, &ali.Id, &twoVal)
	taskJ := api.CreateTask(0, orgId, project.Id, taskA.Id, ali.Id, &taskI.Id, "J", &desc, false, nil, &ali.Id, &oneVal)
	taskK := api.CreateTask(0, orgId, project.Id, taskA.Id, ali.Id, nil, "K", &desc, false, nil, &ali.Id, &fourVal)
	taskL := api.CreateTask(0, orgId, project.Id, taskA.Id, ali.Id, &taskI.Id, "L", &desc, true, &trueVal, nil, nil)
	taskM := api.CreateTask(0, orgId, project.Id, taskL.Id, ali.Id, nil, "M", &desc, false, nil, &ali.Id, &threeVal)

	api.SetName(0, orgId, project.Id, project.Id, ali.Id, "PROJ")
	api.SetName(0, orgId, project.Id, taskA.Id, ali.Id, "AAA")
	api.SetDescription(0, orgId, project.Id, taskA.Id, ali.Id, nil)
	api.SetIsParallel(0, orgId, project.Id, taskA.Id, ali.Id, true)
	api.SetIsParallel(0, orgId, project.Id, project.Id, ali.Id, false)
	api.SetMember(0, orgId, project.Id, taskM.Id, ali.Id, &bob.Id)
	api.SetMember(0, orgId, project.Id, taskM.Id, ali.Id, &cat.Id)
	api.SetMember(0, orgId, project.Id, taskM.Id, ali.Id, nil)
	api.SetMember(0, orgId, project.Id, taskM.Id, ali.Id, &cat.Id)
	api.SetRemainingTime(0, orgId, project.Id, taskG.Id, cat.Id, 1)
	note := "word up!"
	tl := api.SetRemainingTimeAndLogTime(0, orgId, project.Id, taskG.Id, cat.Id, 30, 40, &note)
	assert.Equal(t, uint64(40), tl.Duration)

	api.MoveTask(0, orgId, project.Id, taskG.Id, ali.Id, taskA.Id, nil)
	api.MoveTask(0, orgId, project.Id, taskG.Id, ali.Id, taskA.Id, &taskK.Id)
	api.MoveTask(0, orgId, project.Id, taskG.Id, ali.Id, taskA.Id, &taskJ.Id)
	api.MoveTask(0, orgId, project.Id, taskG.Id, ali.Id, taskL.Id, &taskM.Id)

	api.DeleteTask(0, orgId, project.Id, taskA.Id, ali.Id)
	api.DeleteTask(0, orgId, project.Id, taskD.Id, ali.Id)

	res := api.GetTasks(0, orgId, project.Id, ali.Id, []Id{taskC.Id, taskH.Id})
	assert.Equal(t, 2, len(res))
	res = api.GetChildTasks(0, orgId, project.Id, taskC.Id, ali.Id, nil, 100)
	assert.Equal(t, 3, len(res))
	res = api.GetChildTasks(0, orgId, project.Id, taskC.Id, ali.Id, nil, 2)
	assert.Equal(t, 2, len(res))
	res = api.GetChildTasks(0, orgId, project.Id, taskC.Id, ali.Id, &taskE.Id, 100)
	assert.Equal(t, 2, len(res))
	res = api.GetChildTasks(0, orgId, project.Id, taskC.Id, ali.Id, &taskH.Id, 100)
	assert.Equal(t, 1, len(res))
	res = api.GetChildTasks(0, orgId, project.Id, taskC.Id, ali.Id, &taskF.Id, 100)
	assert.Equal(t, 0, len(res))
	privateApi.DeleteAccount(0, orgId, ali.Id)
}
