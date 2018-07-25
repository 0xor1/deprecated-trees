package systemtest

import (
	"bitbucket.org/0xor1/trees/server/api/v1/centralaccount"
	"bitbucket.org/0xor1/trees/server/api/v1/private"
	"bitbucket.org/0xor1/trees/server/util/clientsession"
	"bitbucket.org/0xor1/trees/server/util/cnst"
	"bitbucket.org/0xor1/trees/server/util/endpoint"
	"bitbucket.org/0xor1/trees/server/util/server"
	"bitbucket.org/0xor1/trees/server/util/static"
	"context"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Run(t *testing.T, systemTesting func(b *Base), endpointSets ...[]*endpoint.Endpoint) {
	var err error
	st := &Base{
		T:      t,
		SR:     static.Config("", nil),
		Region: cnst.EUWRegion,
	}
	endpointSets = append(endpointSets, centralaccount.Endpoints, private.Endpoints)
	serv := server.New(st.SR, endpointSets...)
	st.TestServer = httptest.NewServer(serv)
	defer st.TestServer.Close()

	st.SR.RegionalV1PrivateClient = private.NewTestClient(st.TestServer.URL)
	st.CentralClient = centralaccount.NewClient(st.TestServer.URL)

	st.Ali.CSS = clientsession.New()
	st.Bob.CSS = clientsession.New()
	st.Cat.CSS = clientsession.New()
	st.Dan.CSS = clientsession.New()

	aliDisplayName := "Ali O'Mally"
	assert.Nil(t, st.CentralClient.Register(st.Region, "ali", "ali@ali.com", "al1-Pwd-W00", "en", &aliDisplayName, cnst.DarkTheme))

	bobDisplayName := "Fat Bob"
	assert.Nil(t, st.CentralClient.Register(st.Region, "bob", "bob@bob.com", "8ob-Pwd-W00", "en", &bobDisplayName, cnst.LightTheme))

	catDisplayName := "Lap Cat"
	assert.Nil(t, st.CentralClient.Register(st.Region, "cat", "cat@cat.com", "c@t-Pwd-W00", "de", &catDisplayName, cnst.ColorBlindTheme))

	danDisplayName := "Dan the Man"
	assert.Nil(t, st.CentralClient.Register(st.Region, "dan", "dan@dan.com", "d@n-Pwd-W00", "en", &danDisplayName, cnst.DarkTheme))

	activationCode := ""
	st.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "ali@ali.com").Scan(&activationCode)
	assert.Nil(t, st.CentralClient.Activate("ali@ali.com", activationCode))
	st.Ali.Info, err = st.CentralClient.Authenticate(st.Ali.CSS, "ali@ali.com", "al1-Pwd-W00")
	assert.Nil(t, err)

	st.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "bob@bob.com").Scan(&activationCode)
	assert.Nil(t, st.CentralClient.Activate("bob@bob.com", activationCode))
	st.Bob.Info, err = st.CentralClient.Authenticate(st.Bob.CSS, "bob@bob.com", "8ob-Pwd-W00")
	assert.Nil(t, err)

	st.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "cat@cat.com").Scan(&activationCode)
	assert.Nil(t, st.CentralClient.Activate("cat@cat.com", activationCode))
	st.Cat.Info, err = st.CentralClient.Authenticate(st.Cat.CSS, "cat@cat.com", "c@t-Pwd-W00")
	assert.Nil(t, err)

	st.SR.AccountDb.QueryRowContext(context.TODO(), `SELECT activationCode FROM personalAccounts WHERE email=?`, "dan@dan.com").Scan(&activationCode)
	assert.Nil(t, st.CentralClient.Activate("dan@dan.com", activationCode))
	st.Dan.Info, err = st.CentralClient.Authenticate(st.Dan.CSS, "dan@dan.com", "d@n-Pwd-W00")
	assert.Nil(t, err)

	st.Org, err = st.CentralClient.CreateAccount(st.Ali.CSS, st.Region, "org", nil)
	assert.Nil(t, err)
	st.CentralClient.AddMembers(st.Ali.CSS, st.Org.Id, []*centralaccount.AddMember{
		{Id: st.Bob.Info.Me.Id, Role: cnst.AccountAdmin},
		{Id: st.Cat.Info.Me.Id, Role: cnst.AccountMemberOfAllProjects},
		{Id: st.Dan.Info.Me.Id, Role: cnst.AccountMemberOfOnlySpecificProjects},
	})

	defer tearDown(st)
	systemTesting(st)

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
	TestServer    *httptest.Server
}

type User struct {
	CSS  *clientsession.Store
	Info *centralaccount.AuthenticateResult
}
