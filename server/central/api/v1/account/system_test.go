package account

import (
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	. "bitbucket.org/0xor1/task/server/util"
	"encoding/base64"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func Test_system(t *testing.T) {
	accountsDb := isql.NewReplicaSet("mysql", "t_c_accounts:T@sk-@cc-0unt5@tcp(127.0.0.1:3307)/accounts?parseTime=true&loc=UTC&multiStatements=true", nil)
	endpointServer := Config()
	endpoints := make([]*Endpoint, 0, len(Endpoints)+len(private.Endpoints))
	endpoints = append(endpoints, Endpoints...)
	endpoints = append(endpoints, private.Endpoints...)
	endpointServer.AddEndpoints(endpoints)
	ts := endpointServer.StartTest()
	aliClient := NewClient("http", strings.TrimPrefix(ts.URL, "http://"))
	region := "lcl"

	regions, err := aliClient.GetRegions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, region, regions[0])
	aliDisplayName := "Ali O'Mally"
	aliClient.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, DarkTheme)

	aliClient.ResendActivationEmail("ali@ali.com")
	activationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)

	aliClient.Activate("ali@ali.com", activationCode)
	aliId, err := aliClient.Authenticate("ali@ali.com", "al1-Pwd-W00")

	err = aliClient.SetMyEmail("aliNew@aliNew.com")

	err = aliClient.ResendMyNewEmailConfirmationEmail()
	newEmailConfirmationCode := ""
	accountsDb.QueryRow(`SELECT newEmailConfirmationCode FROM personalAccounts`).Scan(&newEmailConfirmationCode)

	aliClient.ConfirmNewEmail("ali@ali.com", "aliNew@aliNew.com", newEmailConfirmationCode)

	aliClient.ResetPwd("aliNew@aliNew.com")
	resetPwdCode := ""
	accountsDb.QueryRow(`SELECT resetPwdCode FROM personalAccounts`).Scan(&resetPwdCode)

	aliClient.SetNewPwdFromPwdReset("al1-Pwd-W00-2", "aliNew@aliNew.com", resetPwdCode)

	acc, err := aliClient.GetAccount("ali")
	assert.True(t, acc.Id.Equal(aliId))
	assert.Equal(t, "ali", acc.Name)
	assert.Equal(t, aliDisplayName, *acc.DisplayName)
	assert.InDelta(t, Now().Unix(), acc.CreatedOn.Unix(), 5)
	assert.Equal(t, false, acc.HasAvatar)
	assert.Equal(t, true, acc.IsPersonal)
	assert.Nil(t, acc.NewRegion)
	assert.Equal(t, region, acc.Region)
	assert.Equal(t, 0, acc.Shard)

	accs, err := aliClient.GetAccounts([]Id{aliId})
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "ali", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.InDelta(t, Now().Unix(), accs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, accs[0].HasAvatar)
	assert.Equal(t, true, accs[0].IsPersonal)
	assert.Nil(t, accs[0].NewRegion)
	assert.Equal(t, region, accs[0].Region)
	assert.Equal(t, 0, accs[0].Shard)

	me, err := aliClient.GetMe()
	assert.True(t, me.Id.Equal(aliId))
	assert.Equal(t, "ali", me.Name)
	assert.Equal(t, aliDisplayName, *me.DisplayName)
	assert.InDelta(t, Now().Unix(), me.CreatedOn.Unix(), 5)
	assert.Equal(t, false, me.HasAvatar)
	assert.Equal(t, true, me.IsPersonal)
	assert.Nil(t, me.NewRegion)
	assert.Equal(t, region, me.Region)
	assert.Equal(t, 0, me.Shard)
	assert.Equal(t, DarkTheme, me.Theme)
	assert.Equal(t, "en", me.Language)

	aliClient.SetMyPwd("al1-Pwd-W00-2", "al1-Pwd-W00")
	aliId2, err := aliClient.Authenticate("aliNew@aliNew.com", "al1-Pwd-W00")
	assert.True(t, aliId.Equal(aliId2))

	err = aliClient.SetAccountName(aliId, "aliNew")
	aliDisplayName = "ZZZ ali ZZZ"
	err = aliClient.SetAccountDisplayName(aliId, &aliDisplayName)
	err = aliClient.SetAccountAvatar(aliId, ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))))

	err = aliClient.MigrateAccount(aliId, "usw")

	orgDisplayName := "Big Corp"
	org, err := aliClient.CreateAccount("org", region, &orgDisplayName)
	assert.Equal(t, "org", org.Name)
	assert.Equal(t, orgDisplayName, *org.DisplayName)
	assert.InDelta(t, Now().Unix(), org.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org.HasAvatar)
	assert.Equal(t, false, org.IsPersonal)
	assert.Nil(t, org.NewRegion)
	assert.Equal(t, region, org.Region)
	assert.Equal(t, 0, org.Shard)
	orgDisplayName2 := "Big Corp 2"
	org2, err := aliClient.CreateAccount("zorg2", region, &orgDisplayName2)
	assert.Equal(t, "zorg2", org2.Name)
	assert.Equal(t, orgDisplayName2, *org2.DisplayName)
	assert.InDelta(t, Now().Unix(), org2.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org2.HasAvatar)
	assert.Equal(t, false, org2.IsPersonal)
	assert.Nil(t, org2.NewRegion)
	assert.Equal(t, region, org2.Region)
	assert.Equal(t, 0, org2.Shard)

	myAccsRes, err := aliClient.GetMyAccounts(nil, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.True(t, myAccsRes.More)
	assert.Equal(t, "org", myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	myAccsRes, err = aliClient.GetMyAccounts(&org.Id, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.False(t, myAccsRes.More)
	assert.Equal(t, "zorg2", myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName2, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	bobDisplayName := "Fat Bob"

	aliClient.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, LightTheme)
	catDisplayName := "Lap Cat"
	aliClient.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, ColorBlindTheme)

	bobActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	aliClient.Activate("bob@bob.com", bobActivationCode)
	bobId, err := aliClient.Authenticate("bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	aliClient.Activate("cat@cat.com", catActivationCode)
	catId, err := aliClient.Authenticate("cat@cat.com", "c@t-Pwd-W00")

	addBob := AddMemberPublic{}
	addBob.Id = bobId
	addBob.Role = AccountAdmin
	addCat := AddMemberPublic{}
	addCat.Id = catId
	addCat.Role = AccountMemberOfOnlySpecificProjects
	aliClient.AddMembers(org.Id, []*AddMemberPublic{&addBob, &addCat})

	accs, err = aliClient.SearchAccounts("org")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(org.Id))
	assert.Equal(t, "org", accs[0].Name)
	assert.Equal(t, orgDisplayName, *accs[0].DisplayName)
	assert.Equal(t, false, accs[0].IsPersonal)

	accs, err = aliClient.SearchAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = aliClient.SearchAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = aliClient.SearchAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = aliClient.SearchPersonalAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = aliClient.SearchPersonalAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, err = aliClient.SearchPersonalAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	wd, err := os.Getwd()
	os.RemoveAll(path.Join(wd, "avatar"))
	aliClient.DeleteAccount(org.Id)
	aliClient.DeleteAccount(org2.Id)
	aliClient.DeleteAccount(aliId)
	//aliClient.DeleteAccount(bobId, bobId)
	//aliClient.DeleteAccount(catId, catId)
}
