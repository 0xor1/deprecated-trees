package account

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bitbucket.org/0xor1/task_center/region_center/api/v1/private"
	"encoding/base64"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/zap"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_system(t *testing.T) {
	accountsDb := isql.NewReplicaSet("mysql", "tc_cd_accounts:T@sk-C3n-T3r-@cc-0unt5@tcp(127.0.0.1:3306)/accounts?parseTime=true&loc=UTC&multiStatements=true", nil)
	pwdsDb := isql.NewReplicaSet("mysql", "tc_cd_pwds:T@sk-C3n-T3r-Pwd5@tcp(127.0.0.1:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true", nil)
	avatarStore := NewLocalAvatarStore("avatars")
	region := "use" //US-East
	maxProcessEntityCount := 100
	api := New(
		accountsDb,
		pwdsDb,
		private.NewClient(map[string]private.Api{region: private.New(map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}, maxProcessEntityCount)}),
		NewLogLinkMailer(NewLog(zap.New(zap.NewTextEncoder()))),
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

	api.Register("ali", "ali@ali.com", "al1-Pwd-W00", region, "en", DarkTheme)

	api.ResendActivationEmail("ali@ali.com")
	activationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccountInfo`).Scan(&activationCode)

	api.Activate("ali@ali.com", activationCode)
	aliId := api.Authenticate("ali@ali.com", "al1-Pwd-W00")

	api.SetMyEmail(aliId, "aliNew@aliNew.com")

	api.ResendMyNewEmailConfirmationEmail(aliId)
	newEmailConfirmationCode := ""
	accountsDb.QueryRow(`SELECT newEmailConfirmationCode FROM personalAccountInfo`).Scan(&newEmailConfirmationCode)

	api.ConfirmNewEmail("ali@ali.com", "aliNew@aliNew.com", newEmailConfirmationCode)

	api.ResetPwd("aliNew@aliNew.com")
	resetPwdCode := ""
	accountsDb.QueryRow(`SELECT resetPwdCode FROM personalAccountInfo`).Scan(&resetPwdCode)

	api.SetNewPwdFromPwdReset("al1-Pwd-W00-2", "aliNew@aliNew.com", resetPwdCode)

	acc := api.GetAccount("ali")
	assert.True(t, acc.Id.Equal(aliId))
	assert.Equal(t, "ali", acc.Name)
	assert.InDelta(t, Now().Unix(), acc.CreatedOn.Unix(), 5)
	assert.Equal(t, false, acc.HasAvatar)
	assert.Equal(t, true, acc.IsPersonal)
	assert.Nil(t, acc.NewRegion)
	assert.Equal(t, region, acc.Region)
	assert.Equal(t, 0, acc.Shard)

	accs := api.GetAccounts([]Id{aliId})
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, "ali", accs[0].Name)
	assert.InDelta(t, Now().Unix(), accs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, accs[0].HasAvatar)
	assert.Equal(t, true, accs[0].IsPersonal)
	assert.Nil(t, accs[0].NewRegion)
	assert.Equal(t, region, accs[0].Region)
	assert.Equal(t, 0, accs[0].Shard)

	me := api.GetMe(aliId)
	assert.True(t, me.Id.Equal(aliId))
	assert.Equal(t, "ali", me.Name)
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
	api.SetAccountAvatar(aliId, aliId, ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))))

	api.MigrateAccount(aliId, aliId, "usw")

	org := api.CreateAccount(aliId, "org", region)
	assert.Equal(t, "org", org.Name)
	assert.InDelta(t, Now().Unix(), org.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org.HasAvatar)
	assert.Equal(t, false, org.IsPersonal)
	assert.Nil(t, org.NewRegion)
	assert.Equal(t, region, org.Region)
	assert.Equal(t, 0, org.Shard)

	myAccs, total := api.GetMyAccounts(aliId, 0, 100)
	assert.Equal(t, 1, len(myAccs))
	assert.Equal(t, 1, total)
	assert.Equal(t, "org", myAccs[0].Name)
	assert.InDelta(t, Now().Unix(), myAccs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccs[0].HasAvatar)
	assert.Equal(t, false, myAccs[0].IsPersonal)
	assert.Nil(t, myAccs[0].NewRegion)
	assert.Equal(t, region, myAccs[0].Region)
	assert.Equal(t, 0, myAccs[0].Shard)

	api.Register("bob", "bob@bob.com", "8ob-Pwd-W00", region, "en", LightTheme)
	api.Register("cat", "cat@cat.com", "c@t-Pwd-W00", region, "de", ColorBlindTheme)

	bobActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccountInfo WHERE email=?`, "bob@bob.com").Scan(&bobActivationCode)
	api.Activate("bob@bob.com", bobActivationCode)
	bobId := api.Authenticate("bob@bob.com", "8ob-Pwd-W00")
	catActivationCode := ""
	accountsDb.QueryRow(`SELECT activationCode FROM personalAccountInfo WHERE email=?`, "cat@cat.com").Scan(&catActivationCode)
	api.Activate("cat@cat.com", catActivationCode)
	catId := api.Authenticate("cat@cat.com", "c@t-Pwd-W00")

	addBob := AddMemberExternal{}
	addBob.Id = bobId
	addBob.Role = AccountAdmin
	addCat := AddMemberExternal{}
	addCat.Id = catId
	addCat.Role = AccountMemberOfOnlySpecificProjects
	api.AddMembers(aliId, org.Id, []*AddMemberExternal{&addBob, &addCat})
	api.RemoveMembers(aliId, org.Id, []Id{bobId, catId})

	avatarStore.deleteAll()
	api.DeleteAccount(aliId, org.Id)
	api.DeleteAccount(aliId, aliId)
	api.DeleteAccount(bobId, bobId)
	api.DeleteAccount(catId, catId)
}
