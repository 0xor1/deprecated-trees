package centralaccount

import (
	"bitbucket.org/0xor1/task/server/api/v1/private"
	"bitbucket.org/0xor1/task/server/util/clientsession"
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"bitbucket.org/0xor1/task/server/util/time"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_system(t *testing.T) {
	SR := static.Config("", private.NewClient)
	serv := server.New(SR, Endpoints, private.Endpoints)
	testServer := httptest.NewServer(serv)
	aliCss := clientsession.New()
	client := NewClient(testServer.URL)
	region := "lcl"
	SR.RegionalV1PrivateClient = private.NewClient(map[string]string{
		region: testServer.URL,
	})

	regions, err := client.GetRegions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, region, regions[0])
	aliDisplayName := "Ali O'Mally"
	client.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, cnst.DarkTheme)

	client.ResendActivationEmail("ali@ali.com")
	activationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)

	client.Activate("ali@ali.com", activationCode)
	aliId, err := client.Authenticate(aliCss, "ali@ali.com", "al1-Pwd-W00")

	err = client.SetMyEmail(aliCss, "aliNew@aliNew.com")

	err = client.ResendMyNewEmailConfirmationEmail(aliCss)
	newEmailConfirmationCode := ""
	SR.AccountDb.QueryRow(`SELECT newEmailConfirmationCode FROM personalAccounts`).Scan(&newEmailConfirmationCode)

	client.ConfirmNewEmail("ali@ali.com", "aliNew@aliNew.com", newEmailConfirmationCode)

	client.ResetPwd("aliNew@aliNew.com")
	resetPwdCode := ""
	SR.AccountDb.QueryRow(`SELECT resetPwdCode FROM personalAccounts`).Scan(&resetPwdCode)

	client.SetNewPwdFromPwdReset("al1-Pwd-W00-2", "aliNew@aliNew.com", resetPwdCode)

	acc, err := client.GetAccount("ali")
	assert.True(t, acc.Id.Equal(aliId))
	assert.Equal(t, "ali", acc.Name)
	assert.Equal(t, aliDisplayName, *acc.DisplayName)
	assert.InDelta(t, time.Now().Unix(), acc.CreatedOn.Unix(), 5)
	assert.Equal(t, false, acc.HasAvatar)
	assert.Equal(t, true, acc.IsPersonal)
	assert.Nil(t, acc.NewRegion)
	assert.Equal(t, region, acc.Region)
	assert.Equal(t, 0, acc.Shard)

	accs, err := client.GetAccounts([]id.Id{aliId})
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "ali", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), accs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, accs[0].HasAvatar)
	assert.Equal(t, true, accs[0].IsPersonal)
	assert.Nil(t, accs[0].NewRegion)
	assert.Equal(t, region, accs[0].Region)
	assert.Equal(t, 0, accs[0].Shard)

	me, err := client.GetMe(aliCss)
	assert.True(t, me.Id.Equal(aliId))
	assert.Equal(t, "ali", me.Name)
	assert.Equal(t, aliDisplayName, *me.DisplayName)
	assert.InDelta(t, time.Now().Unix(), me.CreatedOn.Unix(), 5)
	assert.Equal(t, false, me.HasAvatar)
	assert.Equal(t, true, me.IsPersonal)
	assert.Nil(t, me.NewRegion)
	assert.Equal(t, region, me.Region)
	assert.Equal(t, 0, me.Shard)
	assert.Equal(t, cnst.DarkTheme, me.Theme)
	assert.Equal(t, "en", me.Language)

	client.SetMyPwd(aliCss, "al1-Pwd-W00-2", "al1-Pwd-W00")
	aliId2, err := client.Authenticate(aliCss, "aliNew@aliNew.com", "al1-Pwd-W00")
	assert.True(t, aliId.Equal(aliId2))

	err = client.SetAccountName(aliCss, aliId, "aliNew")
	aliDisplayName = "ZZZ ali ZZZ"
	err = client.SetAccountDisplayName(aliCss, aliId, &aliDisplayName)
	err = client.SetAccountAvatar(aliCss, aliId, ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))))

	//err = client.MigrateAccount(aliCss, aliId, "usw")

	orgDisplayName := "Big Corp"
	org, err := client.CreateAccount(aliCss, "org", region, &orgDisplayName)
	assert.Equal(t, "org", org.Name)
	assert.Equal(t, orgDisplayName, *org.DisplayName)
	assert.InDelta(t, time.Now().Unix(), org.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org.HasAvatar)
	assert.Equal(t, false, org.IsPersonal)
	assert.Nil(t, org.NewRegion)
	assert.Equal(t, region, org.Region)
	assert.Equal(t, 0, org.Shard)
	orgDisplayName2 := "Big Corp 2"
	org2, err := client.CreateAccount(aliCss, "zorg2", region, &orgDisplayName2)
	assert.Equal(t, "zorg2", org2.Name)
	assert.Equal(t, orgDisplayName2, *org2.DisplayName)
	assert.InDelta(t, time.Now().Unix(), org2.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org2.HasAvatar)
	assert.Equal(t, false, org2.IsPersonal)
	assert.Nil(t, org2.NewRegion)
	assert.Equal(t, region, org2.Region)
	assert.Equal(t, 0, org2.Shard)

	myAccsRes, err := client.GetMyAccounts(aliCss, nil, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.True(t, myAccsRes.More)
	assert.Equal(t, "org", myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	myAccsRes, err = client.GetMyAccounts(aliCss, &org.Id, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.False(t, myAccsRes.More)
	assert.Equal(t, "zorg2", myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName2, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	bobDisplayName := "Fat Bob"

	client.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, cnst.LightTheme)
	catDisplayName := "Lap Cat"
	client.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, cnst.ColorBlindTheme)

	bobActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	client.Activate("bob@bob.com", bobActivationCode)
	bobCss := clientsession.New()
	bobId, err := client.Authenticate(bobCss, "bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	SR.AccountDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	client.Activate("cat@cat.com", catActivationCode)
	catCss := clientsession.New()
	catId, err := client.Authenticate(catCss, "cat@cat.com", "c@t-Pwd-W00")

	addBob := AddMember{}
	addBob.Id = bobId
	addBob.Role = cnst.AccountAdmin
	addCat := AddMember{}
	addCat.Id = catId
	addCat.Role = cnst.AccountMemberOfOnlySpecificProjects
	client.AddMembers(aliCss, org.Id, []*AddMember{&addBob, &addCat})

	accs, err = client.SearchAccounts("org")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(org.Id))
	assert.Equal(t, "org", accs[0].Name)
	assert.Equal(t, orgDisplayName, *accs[0].DisplayName)
	assert.Equal(t, false, accs[0].IsPersonal)

	accs, err = client.SearchAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = client.SearchAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = client.SearchAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = client.SearchPersonalAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = client.SearchPersonalAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = client.SearchPersonalAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	SR.AvatarClient.DeleteAll()
	client.DeleteAccount(aliCss, org.Id)
	client.DeleteAccount(aliCss, org2.Id)
	client.DeleteAccount(aliCss, aliId)
	client.DeleteAccount(bobCss, bobId)
	client.DeleteAccount(catCss, catId)
	cnn := SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}
