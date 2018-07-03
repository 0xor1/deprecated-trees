package account

import (
	"bitbucket.org/0xor1/trees/server/api/v1/centralaccount"
	"bitbucket.org/0xor1/trees/server/api/v1/private"
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/server"
	"bitbucket.org/0xor1/trees/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"bitbucket.org/0xor1/trees/server/util/field"
)

func Test_System(t *testing.T) {
	SR := static.Config("", private.NewClient)
	serv := server.New(SR, centralaccount.Endpoints, private.Endpoints, Endpoints)
	testServer := httptest.NewServer(serv)
	aliCss := clientsession.New()
	centralClient := centralaccount.NewClient(testServer.URL)
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
	bobDisplayName := "Fat Bob"
	centralClient.Register(region, "bob", "bob@bob.com", "8ob-Pwd-W00", "en", &bobDisplayName, cnst.LightTheme)
	catDisplayName := "Lap Cat"
	centralClient.Register(region, "cat", "cat@cat.com", "c@t-Pwd-W00", "de", &catDisplayName, cnst.ColorBlindTheme)
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

	org, err := centralClient.CreateAccount(aliCss, region, "org", nil)
	bob := centralaccount.AddMember{}
	bob.Id = bobId
	bob.Role = cnst.AccountAdmin
	cat := centralaccount.AddMember{}
	cat.Id = catId
	cat.Role = cnst.AccountMemberOfOnlySpecificProjects
	centralClient.AddMembers(aliCss, org.Id, []*centralaccount.AddMember{&bob, &cat})

	acc, err := client.Get(aliCss, region, 0, org.Id)
	assert.Nil(t, err)
	assert.False(t, acc.PublicProjectsEnabled)
	assert.Equal(t, uint8(8), acc.HoursPerDay)
	assert.Equal(t, uint8(5), acc.DaysPerWeek)
	client.Edit(aliCss, region, 0, org.Id, Fields{
		PublicProjectsEnabled: &field.Bool{true},
		HoursPerDay: &field.UInt8{6},
		DaysPerWeek: &field.UInt8{6},
	})
	acc, err = client.Get(aliCss, region, 0, org.Id)
	assert.Nil(t, err)
	assert.True(t, acc.PublicProjectsEnabled)
	assert.Equal(t, uint8(6), acc.HoursPerDay)
	assert.Equal(t, uint8(6), acc.DaysPerWeek)
	client.SetMemberRole(aliCss, region, 0, org.Id, bob.Id, cnst.AccountMemberOfAllProjects)
	membersRes, err := client.GetMembers(aliCss, region, 0, org.Id, nil, nil, nil, 2)
	assert.True(t, membersRes.More)
	assert.Equal(t, 2, len(membersRes.Members))
	assert.True(t, aliId.Equal(membersRes.Members[0].Id))
	assert.True(t, bob.Id.Equal(membersRes.Members[1].Id))
	membersRes, err = client.GetMembers(aliCss, region, 0, org.Id, nil, nil, &membersRes.Members[0].Id, 100)
	assert.False(t, membersRes.More)
	assert.Equal(t, 2, len(membersRes.Members))
	assert.True(t, bob.Id.Equal(membersRes.Members[0].Id))
	assert.True(t, cat.Id.Equal(membersRes.Members[1].Id))
	activities, err := client.GetActivities(aliCss, region, 0, org.Id, nil, nil, nil, nil, 100)
	assert.Equal(t, 7, len(activities))
	me, err := client.GetMe(bobCss, region, 0, org.Id)
	assert.Equal(t, cnst.AccountMemberOfAllProjects, me.Role)
	assert.True(t, bob.Id.Equal(me.Id))
	assert.Equal(t, true, me.IsActive)
	centralClient.DeleteAccount(aliCss, org.Id)
	centralClient.DeleteAccount(aliCss, aliId)
	centralClient.DeleteAccount(bobCss, bobId)
	centralClient.DeleteAccount(catCss, catId)
	SR.AvatarClient.DeleteAll()
	cnn := SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}
