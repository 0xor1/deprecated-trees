package task

import (
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	. "bitbucket.org/0xor1/task/server/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"net/http/httptest"
	"bitbucket.org/0xor1/task/server/central/api/v1/centralaccount"
	"bitbucket.org/0xor1/task/server/regional/api/v1/account"
	"bitbucket.org/0xor1/task/server/config"
)

func Test_system(t *testing.T) {
	staticResources := config.Config("", "", private.NewClient, centralaccount.Endpoints, private.Endpoints, account.Endpoints, project.Endpoints, Endpoints)
	testServer := httptest.NewServer(staticResources)
	aliCss := NewClientSessionStore()
	centralClient := centralaccount.NewClient(testServer.URL)
	projectClient := project.NewClient(testServer.URL)
	client := NewClient(testServer.URL)
	region := "lcl"
	staticResources.RegionalV1PrivateClient = private.NewClient(map[string]string{
		region: testServer.URL,
	})

	aliDisplayName := "Ali O'Mally"
	centralClient.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, DarkTheme)
	activationCode := ""
	staticResources.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)
	centralClient.Activate("ali@ali.com", activationCode)
	aliId, err := centralClient.Authenticate(aliCss, "ali@ali.com", "al1-Pwd-W00")
	assert.Nil(t, err)
	bobDisplayName := "Fat Bob"
	centralClient.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, LightTheme)
	catDisplayName := "Lap Cat"
	centralClient.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, ColorBlindTheme)
	danDisplayName := "Dan the Man"
	centralClient.Register("dan", "dan@dan.com", "d@n-Pwd-W00", region, "en", &danDisplayName, DarkTheme)
	bobActivationCode := ""
	staticResources.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	centralClient.Activate("bob@bob.com", bobActivationCode)
	bobCss := NewClientSessionStore()
	bobId, err := centralClient.Authenticate(bobCss, "bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	staticResources.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	centralClient.Activate("cat@cat.com", catActivationCode)
	catCss := NewClientSessionStore()
	catId, err := centralClient.Authenticate(catCss, "cat@cat.com", "c@t-Pwd-W00")
	danActivationCode := ""
	staticResources.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "dan@dan.com").Scan(&danActivationCode)
	centralClient.Activate("dan@dan.com", danActivationCode)
	danCss := NewClientSessionStore()
	danId, err := centralClient.Authenticate(danCss, "dan@dan.com", "d@n-Pwd-W00")

	org, err := centralClient.CreateAccount(aliCss, "org", region , nil)
	bob := AddMemberPublic{}
	bob.Id = bobId
	bob.Role = AccountAdmin
	cat := AddMemberPublic{}
	cat.Id = catId
	cat.Role = AccountMemberOfAllProjects
	dan := AddMemberPublic{}
	dan.Id = danId
	dan.Role = AccountMemberOfOnlySpecificProjects
	centralClient.AddMembers(aliCss, org.Id, []*AddMemberPublic{&bob, &cat, &dan})


	start := Now()
	end := start.Add(5 * 24 * time.Hour)
	desc := "desc"
	project, err := projectClient.CreateProject(aliCss, 0, org.Id, "proj", &desc, &start, &end, true, false, []*AddProjectMember{{Id: aliId, Role: ProjectAdmin}, {Id: bobId, Role: ProjectAdmin}, {Id: catId, Role: ProjectWriter}, {Id: danId, Role: ProjectReader}})

	oneVal := uint64(1)
	twoVal := uint64(2)
	threeVal := uint64(3)
	fourVal := uint64(4)
	falseVal := false
	trueVal := true
	//create task needs extensive testing to test every avenue of the stored procedure
	taskA, err := client.CreateTask(aliCss, 0, org.Id, project.Id, project.Id, nil, "A", &desc, true, &falseVal, nil, nil)
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
	client.CreateTask(aliCss, 0, org.Id, project.Id, project.Id, &taskA.Id, "B", &desc, true, &falseVal, nil, nil)
	taskC, err := client.CreateTask(aliCss, 0, org.Id, project.Id, project.Id, nil, "C", &desc, true, &trueVal, nil, nil)
	taskD, err := client.CreateTask(aliCss, 0, org.Id, project.Id, project.Id, &taskA.Id, "D", &desc, false, nil, &aliId, &fourVal)
	taskE, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskC.Id, nil, "E", &desc, false, nil, &aliId, &twoVal)
	taskF, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskC.Id, &taskE.Id, "F", &desc, false, nil, &aliId, &oneVal)
	taskG, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskC.Id, nil, "G", &desc, false, nil, &aliId, &fourVal)
	taskH, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskC.Id, &taskE.Id, "H", &desc, false, nil, &aliId, &threeVal)
	taskI, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskA.Id, nil, "I", &desc, false, nil, &aliId, &twoVal)
	taskJ, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskA.Id, &taskI.Id, "J", &desc, false, nil, &aliId, &oneVal)
	taskK, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskA.Id, nil, "K", &desc, false, nil, &aliId, &fourVal)
	taskL, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskA.Id, &taskI.Id, "L", &desc, true, &trueVal, nil, nil)
	taskM, err := client.CreateTask(aliCss, 0, org.Id, project.Id, taskL.Id, nil, "M", &desc, false, nil, &aliId, &threeVal)

	client.SetName(aliCss, 0, org.Id, project.Id, project.Id, "PROJ")
	client.SetName(aliCss, 0, org.Id, project.Id, taskA.Id, "AAA")
	client.SetDescription(aliCss, 0, org.Id, project.Id, taskA.Id, nil)
	client.SetIsParallel(aliCss, 0, org.Id, project.Id, taskA.Id, true)
	client.SetIsParallel(aliCss, 0, org.Id, project.Id, project.Id, false)
	client.SetMember(aliCss, 0, org.Id, project.Id, taskM.Id, &bob.Id)
	client.SetMember(aliCss, 0, org.Id, project.Id, taskM.Id, &cat.Id)
	client.SetMember(aliCss, 0, org.Id, project.Id, taskM.Id, nil)
	client.SetMember(aliCss, 0, org.Id, project.Id, taskM.Id, &cat.Id)
	client.SetRemainingTime(catCss, 0, org.Id, project.Id, taskG.Id, 1)
	note := "word up!"
	tl, err := client.SetRemainingTimeAndLogTime(catCss, 0, org.Id, project.Id, taskG.Id, 30, 40, &note)
	assert.Equal(t, uint64(40), tl.Duration)

	client.MoveTask(aliCss, 0, org.Id, project.Id, taskG.Id, taskA.Id, nil)
	client.MoveTask(aliCss, 0, org.Id, project.Id, taskG.Id, taskA.Id, &taskK.Id)
	client.MoveTask(aliCss, 0, org.Id, project.Id, taskG.Id, taskA.Id, &taskJ.Id)
	client.MoveTask(aliCss, 0, org.Id, project.Id, taskG.Id, taskL.Id, &taskM.Id)

	client.DeleteTask(aliCss, 0, org.Id, project.Id, taskA.Id)
	client.DeleteTask(aliCss, 0, org.Id, project.Id, taskD.Id)

	res, err := client.GetTasks(aliCss, 0, org.Id, project.Id, []Id{taskC.Id, taskH.Id})
	assert.Equal(t, 2, len(res))
	res, err = client.GetChildTasks(aliCss, 0, org.Id, project.Id, taskC.Id, nil, 100)
	assert.Equal(t, 3, len(res))
	res, err = client.GetChildTasks(aliCss, 0, org.Id, project.Id, taskC.Id, nil, 2)
	assert.Equal(t, 2, len(res))
	res, err = client.GetChildTasks(aliCss, 0, org.Id, project.Id, taskC.Id, &taskE.Id, 100)
	assert.Equal(t, 2, len(res))
	res, err = client.GetChildTasks(aliCss, 0, org.Id, project.Id, taskC.Id, &taskH.Id, 100)
	assert.Equal(t, 1, len(res))
	res, err = client.GetChildTasks(aliCss, 0, org.Id, project.Id, taskC.Id, &taskF.Id, 100)
	assert.Equal(t, 0, len(res))
	centralClient.DeleteAccount(aliCss, org.Id)
	centralClient.DeleteAccount(aliCss, aliId)
	centralClient.DeleteAccount(bobCss, bobId)
	centralClient.DeleteAccount(catCss, catId)
	centralClient.DeleteAccount(danCss, danId)
	staticResources.AvatarClient.DeleteAll()
}
