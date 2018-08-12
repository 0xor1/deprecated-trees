package systemtest

import (
	"context"
	"github.com/0xor1/trees/server/api/v1/centralaccount"
	"github.com/0xor1/trees/server/api/v1/private"
	"github.com/0xor1/trees/server/util/clientsession"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/crypt"
	"github.com/0xor1/trees/server/util/endpoint"
	"github.com/0xor1/trees/server/util/server"
	"github.com/0xor1/trees/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Run(t *testing.T, systemTesting func(b *Base), endpointSets ...[]*endpoint.Endpoint) {
	var err error
	b := &Base{
		T:      t,
		SR:     static.Config("", nil),
		Region: cnst.EUWRegion,
	}
	endpointSets = append(endpointSets, centralaccount.Endpoints, private.Endpoints)
	serv := server.New(b.SR, endpointSets...)
	testServer := httptest.NewServer(serv)
	defer testServer.Close()
	b.TestServerURL = testServer.URL

	b.SR.RegionalV1PrivateClient = private.NewTestClient(b.TestServerURL)
	b.CentralClient = centralaccount.NewClient(b.TestServerURL)

	b.Ali.CSS = clientsession.New()
	b.Bob.CSS = clientsession.New()
	b.Cat.CSS = clientsession.New()
	b.Dan.CSS = clientsession.New()

	uniqueSuffix := crypt.UrlSafeString(5)
	aliDisplayName := "Ali O'Mally"
	assert.Nil(t, b.CentralClient.Register(b.Region, "ali"+uniqueSuffix, "ali@ali.com"+uniqueSuffix, "al1-Pwd-W00", "en", &aliDisplayName, cnst.DarkTheme))

	bobDisplayName := "Fat Bob"
	assert.Nil(t, b.CentralClient.Register(b.Region, "bob"+uniqueSuffix, "bob@bob.com"+uniqueSuffix, "8ob-Pwd-W00", "en", &bobDisplayName, cnst.LightTheme))

	catDisplayName := "Lap Cat"
	assert.Nil(t, b.CentralClient.Register(b.Region, "cat"+uniqueSuffix, "cat@cat.com"+uniqueSuffix, "c@t-Pwd-W00", "de", &catDisplayName, cnst.ColorBlindTheme))

	danDisplayName := "Dan the Man"
	assert.Nil(t, b.CentralClient.Register(b.Region, "dan"+uniqueSuffix, "dan@dan.com"+uniqueSuffix, "d@n-Pwd-W00", "en", &danDisplayName, cnst.DarkTheme))

	activationCode := ""
	b.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com"+uniqueSuffix).Scan(&activationCode)
	assert.Nil(t, b.CentralClient.Activate("ali@ali.com"+uniqueSuffix, activationCode))
	b.Ali.Info, err = b.CentralClient.Authenticate(b.Ali.CSS, "ali@ali.com"+uniqueSuffix, "al1-Pwd-W00")
	assert.Nil(t, err)

	b.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com"+uniqueSuffix).Scan(&activationCode)
	assert.Nil(t, b.CentralClient.Activate("bob@bob.com"+uniqueSuffix, activationCode))
	b.Bob.Info, err = b.CentralClient.Authenticate(b.Bob.CSS, "bob@bob.com"+uniqueSuffix, "8ob-Pwd-W00")
	assert.Nil(t, err)

	b.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com"+uniqueSuffix).Scan(&activationCode)
	assert.Nil(t, b.CentralClient.Activate("cat@cat.com"+uniqueSuffix, activationCode))
	b.Cat.Info, err = b.CentralClient.Authenticate(b.Cat.CSS, "cat@cat.com"+uniqueSuffix, "c@t-Pwd-W00")
	assert.Nil(t, err)

	b.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "dan@dan.com"+uniqueSuffix).Scan(&activationCode)
	assert.Nil(t, b.CentralClient.Activate("dan@dan.com"+uniqueSuffix, activationCode))
	b.Dan.Info, err = b.CentralClient.Authenticate(b.Dan.CSS, "dan@dan.com"+uniqueSuffix, "d@n-Pwd-W00")
	assert.Nil(t, err)

	b.Org, err = b.CentralClient.CreateAccount(b.Ali.CSS, b.Region, "org"+uniqueSuffix, nil)
	assert.Nil(t, err)
	b.CentralClient.AddMembers(b.Ali.CSS, b.Org.Id, []*centralaccount.AddMember{
		{Id: b.Bob.Info.Me.Id, Role: cnst.AccountAdmin},
		{Id: b.Cat.Info.Me.Id, Role: cnst.AccountMemberOfAllProjects},
		{Id: b.Dan.Info.Me.Id, Role: cnst.AccountMemberOfOnlySpecificProjects},
	})

	defer tearDown(b)
	systemTesting(b)

}

func tearDown(b *Base) {
	b.CentralClient.DeleteAccount(b.Ali.CSS, b.Org.Id)
	b.CentralClient.DeleteAccount(b.Ali.CSS, b.Ali.Info.Me.Id)
	b.CentralClient.DeleteAccount(b.Bob.CSS, b.Bob.Info.Me.Id)
	b.CentralClient.DeleteAccount(b.Cat.CSS, b.Cat.Info.Me.Id)
	b.CentralClient.DeleteAccount(b.Dan.CSS, b.Dan.Info.Me.Id)
	b.SR.AvatarClient.DeleteAll()
	cnn := b.SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}

type Base struct {
	Ali           User
	Bob           User
	Cat           User
	Dan           User
	Org           *centralaccount.Account
	CentralClient centralaccount.Client
	Region        cnst.Region
	T             *testing.T
	SR            *static.Resources
	TestServerURL string
}

type User struct {
	CSS  *clientsession.Store
	Info *centralaccount.AuthenticateResult
}
