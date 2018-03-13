package account

import (
	"bitbucket.org/0xor1/task/server/central/api/v1/centralaccount"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Test_System(t *testing.T) {
	SR := static.Config("", "", private.NewClient)
	serv := server.New(SR, centralaccount.Endpoints, private.Endpoints, Endpoints)
	testServer := httptest.NewServer(serv)
	aliCss := clientsession.New()
	centralClient := centralaccount.NewClient(testServer.URL)
	client := NewClient(testServer.URL)
	region := "lcl"
	SR.RegionalV1PrivateClient = private.NewClient(map[string]string{
		region: testServer.URL,
	})

	aliDisplayName := "Ali O'Mally"
	centralClient.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, cnst.DarkTheme)
	activationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)
	centralClient.Activate("ali@ali.com", activationCode)
	aliId, err := centralClient.Authenticate(aliCss, "ali@ali.com", "al1-Pwd-W00")
	bobDisplayName := "Fat Bob"
	centralClient.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, cnst.LightTheme)
	catDisplayName := "Lap Cat"
	centralClient.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, cnst.ColorBlindTheme)
	bobActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	centralClient.Activate("bob@bob.com", bobActivationCode)
	bobCss := clientsession.New()
	bobId, err := centralClient.Authenticate(bobCss, "bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	centralClient.Activate("cat@cat.com", catActivationCode)
	catCss := clientsession.New()
	catId, err := centralClient.Authenticate(catCss, "cat@cat.com", "c@t-Pwd-W00")

	org, err := centralClient.CreateAccount(aliCss, "org", region, nil)
	bob := centralaccount.AddMember{}
	bob.Id = bobId
	bob.Role = cnst.AccountAdmin
	cat := centralaccount.AddMember{}
	cat.Id = catId
	cat.Role = cnst.AccountMemberOfOnlySpecificProjects
	centralClient.AddMembers(aliCss, org.Id, []*centralaccount.AddMember{&bob, &cat})

	publicProjectsEnabled, err := client.GetPublicProjectsEnabled(aliCss, 0, org.Id)
	assert.Nil(t, err)
	assert.False(t, publicProjectsEnabled)
	client.SetPublicProjectsEnabled(aliCss, 0, org.Id, true)
	publicProjectsEnabled, err = client.GetPublicProjectsEnabled(aliCss, 0, org.Id)
	assert.Nil(t, err)
	assert.True(t, publicProjectsEnabled)
	client.SetMemberRole(aliCss, 0, org.Id, bob.Id, cnst.AccountMemberOfAllProjects)
	membersRes, err := client.GetMembers(aliCss, 0, org.Id, nil, nil, nil, 2)
	assert.True(t, membersRes.More)
	assert.Equal(t, 2, len(membersRes.Members))
	assert.True(t, aliId.Equal(membersRes.Members[0].Id))
	assert.True(t, bob.Id.Equal(membersRes.Members[1].Id))
	membersRes, err = client.GetMembers(aliCss, 0, org.Id, nil, nil, &membersRes.Members[0].Id, 100)
	assert.False(t, membersRes.More)
	assert.Equal(t, 2, len(membersRes.Members))
	assert.True(t, bob.Id.Equal(membersRes.Members[0].Id))
	assert.True(t, cat.Id.Equal(membersRes.Members[1].Id))
	activities, err := client.GetActivities(aliCss, 0, org.Id, nil, nil, nil, nil, 100)
	assert.Equal(t, 5, len(activities))
	me, err := client.GetMe(bobCss, 0, org.Id)
	assert.Equal(t, cnst.AccountMemberOfAllProjects, me.Role)
	assert.True(t, bob.Id.Equal(me.Id))
	assert.Equal(t, true, me.IsActive)
	centralClient.DeleteAccount(aliCss, org.Id)
	centralClient.DeleteAccount(aliCss, aliId)
	centralClient.DeleteAccount(bobCss, bobId)
	centralClient.DeleteAccount(catCss, catId)
	SR.AvatarClient.DeleteAll()
}
