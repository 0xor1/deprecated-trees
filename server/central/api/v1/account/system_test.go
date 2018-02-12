package account

import (
	. "bitbucket.org/0xor1/task/server/util"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"encoding/base64"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_system(t *testing.T) {
	accountsDb := isql.NewReplicaSet("mysql", "t_c_accounts:T@sk-@cc-0unt5@tcp(127.0.0.1:3307)/accounts?parseTime=true&loc=UTC&multiStatements=true", nil)
	pwdsDb := isql.NewReplicaSet("mysql", "t_c_pwds:T@sk-Pwd5@tcp(127.0.0.1:3307)/pwds?parseTime=true&loc=UTC&multiStatements=true", nil)
	avatarStore := NewLocalAvatarStore("avatars").(*localAvatarStore)
	region := "use" //US-East
	maxProcessEntityCount := 100
	api := New(
		accountsDb,
		pwdsDb,
		private.NewClient(map[string]private.Api{region: private.New(map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "t_r_trees:T@sk-Tr335@tcp(127.0.0.1:3307)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}, maxProcessEntityCount)}),
		NewLogLinkMailer(),
		avatarStore,
		[]string{},
		[]string{},
		250,
		3,
		50,
		8,
		200,
		100,
		100,
		64,
		16384,
		8,
		1,
		32,
	)

	regions := api.GetRegions()
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, region, regions[0])
	aliDisplayName := "Ali O'Mally"
	api.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", &aliDisplayName, DarkTheme)

	api.ResendActivationEmail("ali@ali.com")
	activationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)

	api.Activate("ali@ali.com", activationCode)
	aliId := api.Authenticate("ali@ali.com", "al1-Pwd-W00")

	api.SetMyEmail(aliId, "aliNew@aliNew.com")

	api.ResendMyNewEmailConfirmationEmail(aliId)
	newEmailConfirmationCode := ""
	accountsDb.QueryRow(`SELECT newEmailConfirmationCode FROM personalAccounts`).Scan(&newEmailConfirmationCode)

	api.ConfirmNewEmail("ali@ali.com", "aliNew@aliNew.com", newEmailConfirmationCode)

	api.ResetPwd("aliNew@aliNew.com")
	resetPwdCode := ""
	accountsDb.QueryRow(`SELECT resetPwdCode FROM personalAccounts`).Scan(&resetPwdCode)

	api.SetNewPwdFromPwdReset("al1-Pwd-W00-2", "aliNew@aliNew.com", resetPwdCode)

	acc := api.GetAccount("ali")
	assert.True(t, acc.Id.Equal(aliId))
	assert.Equal(t, "ali", acc.Name)
	assert.Equal(t, aliDisplayName, *acc.DisplayName)
	assert.InDelta(t, Now().Unix(), acc.CreatedOn.Unix(), 5)
	assert.Equal(t, false, acc.HasAvatar)
	assert.Equal(t, true, acc.IsPersonal)
	assert.Nil(t, acc.NewRegion)
	assert.Equal(t, region, acc.Region)
	assert.Equal(t, 0, acc.Shard)

	accs := api.GetAccounts([]Id{aliId})
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "ali", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.InDelta(t, Now().Unix(), accs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, accs[0].HasAvatar)
	assert.Equal(t, true, accs[0].IsPersonal)
	assert.Nil(t, accs[0].NewRegion)
	assert.Equal(t, region, accs[0].Region)
	assert.Equal(t, 0, accs[0].Shard)

	me := api.GetMe(aliId)
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

	api.SetMyPwd(aliId, "al1-Pwd-W00-2", "al1-Pwd-W00")
	aliId2 := api.Authenticate("aliNew@aliNew.com", "al1-Pwd-W00")
	assert.True(t, aliId.Equal(aliId2))

	api.SetAccountName(aliId, aliId, "aliNew")
	aliDisplayName = "ZZZ ali ZZZ"
	api.SetAccountDisplayName(aliId, aliId, &aliDisplayName)
	api.SetAccountAvatar(aliId, aliId, ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))))

	api.MigrateAccount(aliId, aliId, "usw")

	orgDisplayName := "Big Corp"
	org := api.CreateAccount(aliId, "org", region, &orgDisplayName)
	assert.Equal(t, "org", org.Name)
	assert.Equal(t, orgDisplayName, *org.DisplayName)
	assert.InDelta(t, Now().Unix(), org.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org.HasAvatar)
	assert.Equal(t, false, org.IsPersonal)
	assert.Nil(t, org.NewRegion)
	assert.Equal(t, region, org.Region)
	assert.Equal(t, 0, org.Shard)
	orgDisplayName2 := "Big Corp 2"
	org2 := api.CreateAccount(aliId, "zorg2", region, &orgDisplayName2)
	assert.Equal(t, "zorg2", org2.Name)
	assert.Equal(t, orgDisplayName2, *org2.DisplayName)
	assert.InDelta(t, Now().Unix(), org2.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org2.HasAvatar)
	assert.Equal(t, false, org2.IsPersonal)
	assert.Nil(t, org2.NewRegion)
	assert.Equal(t, region, org2.Region)
	assert.Equal(t, 0, org2.Shard)

	myAccs, more := api.GetMyAccounts(aliId, nil, 1)
	assert.Equal(t, 1, len(myAccs))
	assert.True(t, more)
	assert.Equal(t, "org", myAccs[0].Name)
	assert.Equal(t, orgDisplayName, *myAccs[0].DisplayName)
	assert.InDelta(t, Now().Unix(), myAccs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccs[0].HasAvatar)
	assert.Equal(t, false, myAccs[0].IsPersonal)
	assert.Nil(t, myAccs[0].NewRegion)
	assert.Equal(t, region, myAccs[0].Region)
	assert.Equal(t, 0, myAccs[0].Shard)

	myAccs, more = api.GetMyAccounts(aliId, &org.Id, 1)
	assert.Equal(t, 1, len(myAccs))
	assert.False(t, more)
	assert.Equal(t, "zorg2", myAccs[0].Name)
	assert.Equal(t, orgDisplayName2, *myAccs[0].DisplayName)
	assert.InDelta(t, Now().Unix(), myAccs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccs[0].HasAvatar)
	assert.Equal(t, false, myAccs[0].IsPersonal)
	assert.Nil(t, myAccs[0].NewRegion)
	assert.Equal(t, region, myAccs[0].Region)
	assert.Equal(t, 0, myAccs[0].Shard)

	bobDisplayName := "Fat Bob"
	api.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", &bobDisplayName, LightTheme)
	catDisplayName := "Lap Cat"
	api.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", &catDisplayName, ColorBlindTheme)

	bobActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	api.Activate("bob@bob.com", bobActivationCode)
	bobId := api.Authenticate("bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	api.Activate("cat@cat.com", catActivationCode)
	catId := api.Authenticate("cat@cat.com", "c@t-Pwd-W00")

	addBob := AddMemberPublic{}
	addBob.Id = bobId
	addBob.Role = AccountAdmin
	addCat := AddMemberPublic{}
	addCat.Id = catId
	addCat.Role = AccountMemberOfOnlySpecificProjects
	api.AddMembers(aliId, org.Id, []*AddMemberPublic{&addBob, &addCat})

	accs = api.SearchAccounts("org")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(org.Id))
	assert.Equal(t, "org", accs[0].Name)
	assert.Equal(t, orgDisplayName, *accs[0].DisplayName)
	assert.Equal(t, false, accs[0].IsPersonal)

	accs = api.SearchAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs = api.SearchAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs = api.SearchAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs = api.SearchPersonalAccounts("ali")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "aliNew", accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs = api.SearchPersonalAccounts("bob")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, "bob", accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs = api.SearchPersonalAccounts("cat")
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, "cat", accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	avatarStore.deleteAll()
	api.DeleteAccount(aliId, org.Id)
	api.DeleteAccount(aliId, org2.Id)
	api.DeleteAccount(aliId, aliId)
	api.DeleteAccount(bobId, bobId)
	api.DeleteAccount(catId, catId)
}
