package task

import (
	"bitbucket.org/0xor1/trees/server/api/v1/account"
	"bitbucket.org/0xor1/trees/server/api/v1/centralaccount"
	"bitbucket.org/0xor1/trees/server/api/v1/private"
	"bitbucket.org/0xor1/trees/server/api/v1/project"
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/server"
	"bitbucket.org/0xor1/trees/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
	"bitbucket.org/0xor1/trees/server/util/field"
)

func Test_system(t *testing.T) {
	SR := static.Config("", private.NewClient)
	serv := server.New(SR, centralaccount.Endpoints, private.Endpoints, account.Endpoints, project.Endpoints, Endpoints)
	testServer := httptest.NewServer(serv)
	aliCss := clientsession.New()
	centralClient := centralaccount.NewClient(testServer.URL)
	accountClient := account.NewClient(testServer.URL)
	projectClient := project.NewClient(testServer.URL)
	client := NewClient(testServer.URL)
	region := cnst.EUWRegion
	SR.RegionalV1PrivateClient = private.NewTestClient(testServer.URL)

	aliDisplayName := "Ali O'Mally"
	centralClient.Register(region, "ali", "ali@ali.com", "al1-Pwd-W00", "en", &aliDisplayName, cnst.DarkTheme)
	activationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)
	centralClient.Activate("ali@ali.com", activationCode)
	aliInitInfo, err := centralClient.Authenticate(aliCss, "ali@ali.com", "al1-Pwd-W00")
	aliId := aliInitInfo.Me.Id
	assert.Nil(t, err)
	bobDisplayName := "Fat Bob"
	centralClient.Register(region, "bob", "bob@bob.com", "8ob-Pwd-W00", "en", &bobDisplayName, cnst.LightTheme)
	catDisplayName := "Lap Cat"
	centralClient.Register(region, "cat", "cat@cat.com", "c@t-Pwd-W00", "de", &catDisplayName, cnst.ColorBlindTheme)
	danDisplayName := "Dan the Man"
	centralClient.Register(region, "dan", "dan@dan.com", "d@n-Pwd-W00", "en", &danDisplayName, cnst.DarkTheme)
	bobActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	centralClient.Activate("bob@bob.com", bobActivationCode)
	bobCss := clientsession.New()
	bobInitInfo, err := centralClient.Authenticate(bobCss, "bob@bob.com", "8ob-Pwd-W00")
	bobId := bobInitInfo.Me.Id
	catActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	centralClient.Activate("cat@cat.com", catActivationCode)
	catCss := clientsession.New()
	catInitInfo, err := centralClient.Authenticate(catCss, "cat@cat.com", "c@t-Pwd-W00")
	catId := catInitInfo.Me.Id
	danActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "dan@dan.com").Scan(&danActivationCode)
	centralClient.Activate("dan@dan.com", danActivationCode)
	danCss := clientsession.New()
	danInitInfo, err := centralClient.Authenticate(danCss, "dan@dan.com", "d@n-Pwd-W00")
	danId := danInitInfo.Me.Id

	org, err := centralClient.CreateAccount(aliCss, region, "org", nil)
	bob := centralaccount.AddMember{}
	bob.Id = bobId
	bob.Role = cnst.AccountAdmin
	cat := centralaccount.AddMember{}
	cat.Id = catId
	cat.Role = cnst.AccountMemberOfAllProjects
	dan := centralaccount.AddMember{}
	dan.Id = danId
	dan.Role = cnst.AccountMemberOfOnlySpecificProjects
	centralClient.AddMembers(aliCss, org.Id, []*centralaccount.AddMember{&bob, &cat, &dan})

	start := time.Now()
	end := start.Add(5 * 24 * time.Hour)
	desc := "desc"
	proj, err := projectClient.Create(aliCss, region, 0, org.Id, "proj", &desc, 8, 5, &start, &end, true, false, []*project.AddProjectMember{{Id: aliId, Role: cnst.ProjectAdmin}, {Id: bobId, Role: cnst.ProjectAdmin}, {Id: catId, Role: cnst.ProjectWriter}, {Id: danId, Role: cnst.ProjectReader}})

	oneVal := uint64(1)
	twoVal := uint64(2)
	threeVal := uint64(3)
	fourVal := uint64(4)
	falseVal := false
	trueVal := true
	//create task needs extensive testing to test every avenue of the stored procedure
	taskA, err := client.Create(aliCss, region, 0, org.Id, proj.Id, proj.Id, nil, "A", &desc, true, &falseVal, nil, nil)
	assert.Equal(t, true, taskA.IsAbstract)
	assert.Equal(t, "A", taskA.Name)
	assert.Equal(t, "desc", *taskA.Description)
	assert.InDelta(t, time.Now().Unix(), taskA.CreatedOn.Unix(), 100)
	assert.Equal(t, uint64(0), taskA.TotalRemainingTime)
	assert.Equal(t, uint64(0), taskA.TotalLoggedTime)
	assert.Equal(t, uint64(0), *taskA.MinimumRemainingTime)
	assert.Equal(t, uint64(0), taskA.LinkedFileCount)
	assert.Equal(t, uint64(0), taskA.ChatCount)
	assert.Equal(t, uint64(0), *taskA.ChildCount)
	assert.Equal(t, uint64(0), *taskA.DescendantCount)
	assert.Equal(t, false, *taskA.IsParallel)
	assert.Nil(t, taskA.Member)
	client.Create(aliCss, region, 0, org.Id, proj.Id, proj.Id, &taskA.Id, "B", &desc, true, &falseVal, nil, nil)
	taskC, err := client.Create(aliCss, region, 0, org.Id, proj.Id, proj.Id, nil, "C", &desc, true, &trueVal, nil, nil)
	taskD, err := client.Create(aliCss, region, 0, org.Id, proj.Id, proj.Id, &taskA.Id, "D", &desc, false, nil, &aliId, &fourVal)
	taskE, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskC.Id, nil, "E", &desc, false, nil, &aliId, &twoVal)
	taskF, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskC.Id, &taskE.Id, "F", &desc, false, nil, &aliId, &oneVal)
	taskG, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskC.Id, nil, "G", &desc, false, nil, &aliId, &fourVal)
	taskH, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskC.Id, &taskE.Id, "H", &desc, false, nil, &aliId, &threeVal)
	taskI, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskA.Id, nil, "I", &desc, false, nil, &aliId, &twoVal)
	taskJ, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskA.Id, &taskI.Id, "J", &desc, false, nil, &aliId, &oneVal)
	taskK, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskA.Id, nil, "K", &desc, false, nil, &aliId, &fourVal)
	taskL, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskA.Id, &taskI.Id, "L", &desc, true, &trueVal, nil, nil)
	taskM, err := client.Create(aliCss, region, 0, org.Id, proj.Id, taskL.Id, nil, "M", &desc, false, nil, &aliId, &threeVal)

	client.SetName(aliCss, region, 0, org.Id, proj.Id, proj.Id, "PROJ")
	client.SetName(aliCss, region, 0, org.Id, proj.Id, taskA.Id, "AAA")
	client.SetDescription(aliCss, region, 0, org.Id, proj.Id, taskA.Id, nil)
	client.SetIsParallel(aliCss, region, 0, org.Id, proj.Id, taskA.Id, true)
	client.SetIsParallel(aliCss, region, 0, org.Id, proj.Id, proj.Id, false)
	client.SetMember(aliCss, region, 0, org.Id, proj.Id, taskM.Id, &bob.Id)
	client.SetMember(aliCss, region, 0, org.Id, proj.Id, taskM.Id, &cat.Id)
	client.SetMember(aliCss, region, 0, org.Id, proj.Id, taskM.Id, nil)
	client.SetMember(aliCss, region, 0, org.Id, proj.Id, taskM.Id, &cat.Id)
	client.SetRemainingTime(catCss, region, 0, org.Id, proj.Id, taskG.Id, 1)

	client.Move(aliCss, region, 0, org.Id, proj.Id, taskG.Id, taskA.Id, nil)
	client.Move(aliCss, region, 0, org.Id, proj.Id, taskG.Id, taskA.Id, &taskK.Id)
	client.Move(aliCss, region, 0, org.Id, proj.Id, taskG.Id, taskA.Id, &taskJ.Id)
	client.Move(aliCss, region, 0, org.Id, proj.Id, taskG.Id, taskL.Id, &taskM.Id)

	ancestors, err := client.GetAncestors(aliCss, region, 0, org.Id, proj.Id, taskM.Id, 100)
	assert.Equal(t, 3, len(ancestors.Ancestors))
	assert.True(t, proj.Id.Equal(ancestors.Ancestors[0].Id))
	assert.True(t, taskA.Id.Equal(ancestors.Ancestors[1].Id))
	assert.True(t, taskL.Id.Equal(ancestors.Ancestors[2].Id))

	//test setting project as public and try getting info without a session
	ancestors, err = client.GetAncestors(nil, region, 0, org.Id, proj.Id, taskM.Id, 100)
	assert.Nil(t, ancestors)
	assert.NotNil(t, err)
	accountClient.Edit(aliCss, region, 0, org.Id, account.Fields{PublicProjectsEnabled:&field.Bool{true}})
	projectClient.Edit(aliCss, region, 0, org.Id, proj.Id, project.Fields{IsPublic: &field.Bool{true}})
	ancestors, err = client.GetAncestors(nil, region, 0, org.Id, proj.Id, taskM.Id, 100)
	assert.Equal(t, 3, len(ancestors.Ancestors))
	assert.True(t, proj.Id.Equal(ancestors.Ancestors[0].Id))
	assert.True(t, taskA.Id.Equal(ancestors.Ancestors[1].Id))
	assert.True(t, taskL.Id.Equal(ancestors.Ancestors[2].Id))
	projectClient.Edit(aliCss, region, 0, org.Id, proj.Id, project.Fields{IsPublic: &field.Bool{false}})
	ancestors, err = client.GetAncestors(nil, region, 0, org.Id, proj.Id, taskM.Id, 100)
	assert.Nil(t, ancestors)
	assert.NotNil(t, err)

	client.Delete(aliCss, region, 0, org.Id, proj.Id, taskA.Id)
	client.Delete(aliCss, region, 0, org.Id, proj.Id, taskD.Id)

	task, err := client.Get(aliCss, region, 0, org.Id, proj.Id, taskH.Id)
	assert.NotNil(t, 2, task)
	res, err := client.GetChildren(aliCss, region, 0, org.Id, proj.Id, taskC.Id, nil, 100)
	assert.Equal(t, 3, len(res.Children))
	res, err = client.GetChildren(aliCss, region, 0, org.Id, proj.Id, taskC.Id, nil, 2)
	assert.Equal(t, 2, len(res.Children))
	res, err = client.GetChildren(aliCss, region, 0, org.Id, proj.Id, taskC.Id, &taskE.Id, 100)
	assert.Equal(t, 2, len(res.Children))
	res, err = client.GetChildren(aliCss, region, 0, org.Id, proj.Id, taskC.Id, &taskH.Id, 100)
	assert.Equal(t, 1, len(res.Children))
	res, err = client.GetChildren(aliCss, region, 0, org.Id, proj.Id, taskC.Id, &taskF.Id, 100)
	assert.Equal(t, 0, len(res.Children))

	centralClient.DeleteAccount(aliCss, org.Id)
	centralClient.DeleteAccount(aliCss, aliId)
	centralClient.DeleteAccount(bobCss, bobId)
	centralClient.DeleteAccount(catCss, catId)
	centralClient.DeleteAccount(danCss, danId)
	SR.AvatarClient.DeleteAll()
	cnn := SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}
