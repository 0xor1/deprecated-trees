package project

import (
	"bitbucket.org/0xor1/task/server/api/v1/account"
	"bitbucket.org/0xor1/task/server/api/v1/centralaccount"
	"bitbucket.org/0xor1/task/server/api/v1/private"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Test_system(t *testing.T) {
	SR := static.Config("", private.NewClient)
	serv := server.New(SR, centralaccount.Endpoints, private.Endpoints, account.Endpoints, Endpoints)
	testServer := httptest.NewServer(serv)
	aliCss := clientsession.New()
	centralClient := centralaccount.NewClient(testServer.URL)
	accountClient := account.NewClient(testServer.URL)
	client := NewClient(testServer.URL)
	region := cnst.LclEnv
	SR.RegionalV1PrivateClient = private.NewClient(map[string]string{
		region: testServer.URL,
	})

	aliDisplayName := "Ali O'Mally"
	centralClient.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, cnst.DarkTheme)
	activationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)
	centralClient.Activate("ali@ali.com", activationCode)
	aliInitInfo, err := centralClient.Authenticate(aliCss, "ali@ali.com", "al1-Pwd-W00")
	aliId := aliInitInfo.Me.Id
	assert.Nil(t, err)
	bobDisplayName := "Fat Bob"
	centralClient.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, cnst.LightTheme)
	catDisplayName := "Lap Cat"
	centralClient.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, cnst.ColorBlindTheme)
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

	org, err := centralClient.CreateAccount(aliCss, "org", region, nil)
	bob := centralaccount.AddMember{}
	bob.Id = bobId
	bob.Role = cnst.AccountAdmin
	cat := centralaccount.AddMember{}
	cat.Id = catId
	cat.Role = cnst.AccountMemberOfOnlySpecificProjects
	centralClient.AddMembers(aliCss, org.Id, []*centralaccount.AddMember{&bob, &cat})

	accountClient.SetPublicProjectsEnabled(aliCss, 0, org.Id, true)

	p1Desc := "p1_desc"
	p2Desc := "p2_desc"
	p3Desc := "p3_desc"
	proj, err := client.Create(aliCss, 0, org.Id, "a-p1", &p1Desc, nil, nil, true, false, nil)
	proj2, err := client.Create(aliCss, 0, org.Id, "b-p2", &p2Desc, nil, nil, true, false, nil)
	proj3, err := client.Create(aliCss, 0, org.Id, "c-p3", &p3Desc, nil, nil, true, false, nil)
	client.SetIsPublic(aliCss, 0, org.Id, proj.Id, true)
	proj, err = client.Get(aliCss, 0, org.Id, proj.Id)
	assert.Equal(t, "a-p1", proj.Name)
	assert.Equal(t, "p1_desc", *proj.Description)
	assert.Equal(t, true, proj.IsParallel)
	assert.Equal(t, true, proj.IsPublic)
	projRes, err := client.GetSet(aliCss, 0, org.Id, nil, nil, nil, nil, nil, nil, nil, false, cnst.SortByCreatedOn, cnst.SortDirAsc, nil, 1)
	assert.Equal(t, 1, len(projRes.Projects))
	assert.True(t, projRes.More)
	assert.Equal(t, proj.Name, projRes.Projects[0].Name)
	projRes, err = client.GetSet(aliCss, 0, org.Id, nil, nil, nil, nil, nil, nil, nil, false, cnst.SortByCreatedOn, cnst.SortDirAsc, &proj.Id, 100)
	assert.Equal(t, 2, len(projRes.Projects))
	assert.False(t, projRes.More)
	assert.Equal(t, proj2.Name, projRes.Projects[0].Name)
	assert.Equal(t, proj3.Name, projRes.Projects[1].Name)
	client.SetIsArchived(aliCss, 0, org.Id, proj.Id, true)
	projRes, err = client.GetSet(aliCss, 0, org.Id, nil, nil, nil, nil, nil, nil, nil, true, cnst.SortByCreatedOn, cnst.SortDirAsc, nil, 100)
	assert.Equal(t, 1, len(projRes.Projects))
	assert.False(t, projRes.More)
	client.SetIsArchived(aliCss, 0, org.Id, proj.Id, false)
	aliP := &AddProjectMember{}
	aliP.Id = aliId
	aliP.Role = cnst.ProjectAdmin
	bobP := &AddProjectMember{}
	bobP.Id = bob.Id
	bobP.Role = cnst.ProjectWriter
	catP := &AddProjectMember{}
	catP.Id = cat.Id
	catP.Role = cnst.ProjectReader
	client.AddMembers(aliCss, 0, org.Id, proj.Id, []*AddProjectMember{aliP, bobP, catP})
	accountClient.SetMemberRole(aliCss, 0, org.Id, bobP.Id, cnst.AccountMemberOfOnlySpecificProjects)
	client.SetMemberRole(aliCss, 0, org.Id, proj.Id, bobP.Id, cnst.ProjectReader)
	memRes, err := client.GetMembers(aliCss, 0, org.Id, proj.Id, nil, nil, nil, 100)
	assert.Equal(t, 3, len(memRes.Members))
	assert.False(t, memRes.More)
	assert.True(t, memRes.Members[0].Id.Equal(aliId))
	assert.True(t, memRes.Members[1].Id.Equal(bob.Id))
	assert.True(t, memRes.Members[2].Id.Equal(cat.Id))
	bobMe, err := client.GetMe(bobCss, 0, org.Id, proj.Id)
	assert.True(t, bobMe.Id.Equal(bob.Id))
	activities, err := client.GetActivities(aliCss, 0, org.Id, proj.Id, nil, nil, nil, nil, 100)
	assert.Equal(t, 8, len(activities))
	client.RemoveMembers(aliCss, 0, org.Id, proj.Id, []id.Id{bob.Id, cat.Id})
	client.Delete(aliCss, 0, org.Id, proj.Id)

	centralClient.DeleteAccount(aliCss, aliId)
	centralClient.DeleteAccount(aliCss, org.Id)
	centralClient.DeleteAccount(bobCss, bobId)
	centralClient.DeleteAccount(catCss, catId)
	SR.AvatarClient.DeleteAll()
	cnn := SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}
