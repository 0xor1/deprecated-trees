package centralaccount

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/0xor1/trees/server/api/v1/private"
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/crypt"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/server"
	"github.com/0xor1/trees/server/util/static"
	"github.com/0xor1/trees/server/util/time"
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
	region := cnst.EUWRegion
	SR.RegionalV1PrivateClient = private.NewTestClient(testServer.URL)

	aliDisplayName := "Ali O'Mally"
	aliName := "A" + crypt.UrlSafeString(5)
	aliEmail := fmt.Sprintf("%s@%s.com", aliName, aliName)
	client.Register(region, aliName, aliEmail, "al1-Pwd-W00", "en", &aliDisplayName, cnst.DarkTheme)

	client.ResendActivationEmail(aliEmail)
	activationCode := ""
	SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, aliEmail).Scan(&activationCode)

	client.Activate(aliEmail, activationCode)
	aliInitInfo, _ := client.Authenticate(aliCss, aliEmail, "al1-Pwd-W00")
	aliId := aliInitInfo.Me.Id

	aliEmail += "New"
	client.SetMyEmail(aliCss, aliEmail)

	client.ResendMyNewEmailConfirmationEmail(aliCss)
	newEmailConfirmationCode := ""
	SR.AccountDb.QueryRowContext(context.TODO(), `SELECT newEmailConfirmationCode FROM personalAccounts WHERE email=?`, strings.TrimSuffix(aliEmail, "New")).Scan(&newEmailConfirmationCode)

	client.ConfirmNewEmail(strings.TrimSuffix(aliEmail, "New"), aliEmail, newEmailConfirmationCode)

	client.ResetPwd(aliEmail)
	resetPwdCode := ""
	SR.AccountDb.QueryRowContext(context.TODO(), `SELECT resetPwdCode FROM personalAccounts WHERE email=?`, aliEmail).Scan(&resetPwdCode)

	client.SetNewPwdFromPwdReset("al1-Pwd-W00-2", aliEmail, resetPwdCode)

	acc, _ := client.GetAccount(aliName)
	assert.True(t, acc.Id.Equal(aliId))
	assert.Equal(t, aliName, acc.Name)
	assert.Equal(t, aliDisplayName, *acc.DisplayName)
	assert.InDelta(t, time.Now().Unix(), acc.CreatedOn.Unix(), 5)
	assert.Equal(t, false, acc.HasAvatar)
	assert.Equal(t, true, acc.IsPersonal)
	assert.Nil(t, acc.NewRegion)
	assert.Equal(t, region, acc.Region)
	assert.Equal(t, 0, acc.Shard)

	accs, _ := client.GetAccounts([]id.Id{aliId})
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, aliName, accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), accs[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, accs[0].HasAvatar)
	assert.Equal(t, true, accs[0].IsPersonal)
	assert.Nil(t, accs[0].NewRegion)
	assert.Equal(t, region, accs[0].Region)
	assert.Equal(t, 0, accs[0].Shard)

	me, _ := client.GetMe(aliCss)
	assert.True(t, me.Id.Equal(aliId))
	assert.Equal(t, aliName, me.Name)
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
	aliInitInfo2, _ := client.Authenticate(aliCss, aliEmail, "al1-Pwd-W00")
	aliId2 := aliInitInfo2.Me.Id
	assert.True(t, aliId.Equal(aliId2))

	aliName += "New"
	client.SetAccountName(aliCss, aliId, aliName)
	aliDisplayName = "ZZZ ali ZZZ"
	client.SetAccountDisplayName(aliCss, aliId, &aliDisplayName)
	client.SetAccountAvatar(aliCss, aliId, ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))))

	//err = client.MigrateAccount(aliCss, aliId, "usw")

	orgName := "O" + crypt.UrlSafeString(5)
	orgDisplayName := "Big Corp"
	org, _ := client.CreateAccount(aliCss, region, orgName, &orgDisplayName)
	assert.Equal(t, orgName, org.Name)
	assert.Equal(t, orgDisplayName, *org.DisplayName)
	assert.InDelta(t, time.Now().Unix(), org.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org.HasAvatar)
	assert.Equal(t, false, org.IsPersonal)
	assert.Nil(t, org.NewRegion)
	assert.Equal(t, region, org.Region)
	assert.Equal(t, 0, org.Shard)
	orgName2 := "Z" + crypt.UrlSafeString(5)
	orgDisplayName2 := "Big Corp 2"
	org2, _ := client.CreateAccount(aliCss, region, orgName2, &orgDisplayName2)
	assert.Equal(t, orgName2, org2.Name)
	assert.Equal(t, orgDisplayName2, *org2.DisplayName)
	assert.InDelta(t, time.Now().Unix(), org2.CreatedOn.Unix(), 5)
	assert.Equal(t, false, org2.HasAvatar)
	assert.Equal(t, false, org2.IsPersonal)
	assert.Nil(t, org2.NewRegion)
	assert.Equal(t, region, org2.Region)
	assert.Equal(t, 0, org2.Shard)

	myAccsRes, _ := client.GetMyAccounts(aliCss, nil, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.True(t, myAccsRes.More)
	assert.Equal(t, orgName, myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	myAccsRes, _ = client.GetMyAccounts(aliCss, &org.Id, 1)
	assert.Equal(t, 1, len(myAccsRes.Accounts))
	assert.False(t, myAccsRes.More)
	assert.Equal(t, orgName2, myAccsRes.Accounts[0].Name)
	assert.Equal(t, orgDisplayName2, *myAccsRes.Accounts[0].DisplayName)
	assert.InDelta(t, time.Now().Unix(), myAccsRes.Accounts[0].CreatedOn.Unix(), 5)
	assert.Equal(t, false, myAccsRes.Accounts[0].HasAvatar)
	assert.Equal(t, false, myAccsRes.Accounts[0].IsPersonal)
	assert.Nil(t, myAccsRes.Accounts[0].NewRegion)
	assert.Equal(t, region, myAccsRes.Accounts[0].Region)
	assert.Equal(t, 0, myAccsRes.Accounts[0].Shard)

	bobName := "B" + crypt.UrlSafeString(5)
	bobEmail := fmt.Sprintf("%s@%s.com", bobName, bobName)
	bobDisplayName := "Fat Bob"
	client.Register(region, bobName, bobEmail, "8ob-Pwd-W00", "en", &bobDisplayName, cnst.LightTheme)
	catName := "C" + crypt.UrlSafeString(5)
	catEmail := fmt.Sprintf("%s@%s.com", catName, catName)
	catDisplayName := "Lap Cat"
	client.Register(region, catName, catEmail, "c@t-Pwd-W00", "de", &catDisplayName, cnst.ColorBlindTheme)

	bobActivationCode := ""
	SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, bobEmail).Scan(&bobActivationCode)
	client.Activate(bobEmail, bobActivationCode)
	bobCss := clientsession.New()
	bobInitInfo, _ := client.Authenticate(bobCss, bobEmail, "8ob-Pwd-W00")
	bobId := bobInitInfo.Me.Id
	catActivationCode := ""
	SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, catEmail).Scan(&catActivationCode)
	client.Activate(catEmail, catActivationCode)
	catCss := clientsession.New()
	catInitInfo, _ := client.Authenticate(catCss, catEmail, "c@t-Pwd-W00")
	catId := catInitInfo.Me.Id

	addBob := AddMember{}
	addBob.Id = bobId
	addBob.Role = cnst.AccountAdmin
	addCat := AddMember{}
	addCat.Id = catId
	addCat.Role = cnst.AccountMemberOfOnlySpecificProjects
	client.AddMembers(aliCss, org.Id, []*AddMember{&addBob, &addCat})

	accs, _ = client.SearchAccounts(orgName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(org.Id))
	assert.Equal(t, orgName, accs[0].Name)
	assert.Equal(t, orgDisplayName, *accs[0].DisplayName)
	assert.Equal(t, false, accs[0].IsPersonal)

	accs, _ = client.SearchAccounts(aliName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, aliName, accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, _ = client.SearchAccounts(bobName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, bobName, accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, _ = client.SearchAccounts(catName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, catName, accs[0].Name)
	assert.Equal(t, catDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, _ = client.SearchPersonalAccounts(aliName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(aliId))
	assert.Equal(t, aliName, accs[0].Name)
	assert.Equal(t, aliDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, _ = client.SearchPersonalAccounts(bobName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(bobId))
	assert.Equal(t, bobName, accs[0].Name)
	assert.Equal(t, bobDisplayName, *accs[0].DisplayName)
	assert.Equal(t, true, accs[0].IsPersonal)

	accs, _ = client.SearchPersonalAccounts(catName)
	assert.Equal(t, 1, len(accs))
	assert.True(t, accs[0].Id.Equal(catId))
	assert.Equal(t, catName, accs[0].Name)
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
