package private

import (
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/private"
	"github.com/0xor1/trees/server/util/server"
	"github.com/0xor1/trees/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Test_system(t *testing.T) {
	SR := static.Config("", NewClient)
	serv := server.New(SR, Endpoints)
	testServer := httptest.NewServer(serv)
	region := cnst.EUWRegion
	SR.RegionalV1PrivateClient = NewTestClient(testServer.URL)
	client := SR.RegionalV1PrivateClient

	aliId := id.New()
	orgId := id.New()
	client.CreateAccount(region, orgId, aliId, "ali", nil, false)
	bob := private.AddMember{}
	bob.Id = id.New()
	bob.Name = "bob"
	bob.Role = cnst.AccountAdmin
	cat := private.AddMember{}
	cat.Id = id.New()
	cat.Name = "cat"
	cat.Role = cnst.AccountMemberOfOnlySpecificProjects
	client.AddMembers(region, 0, orgId, aliId, []*private.AddMember{&bob, &cat})
	val, err := client.MemberIsOnlyAccountOwner(region, 0, orgId, aliId)
	assert.Nil(t, err)
	assert.True(t, val)
	client.SetMemberName(region, 0, orgId, aliId, "aliNew")
	client.SetMemberHasAvatar(region, 0, orgId, aliId, true)
	val, err = client.MemberIsAccountOwner(region, 0, orgId, aliId)
	assert.Nil(t, err)
	assert.True(t, val)
	client.RemoveMembers(region, 0, orgId, aliId, []id.Id{bob.Id})
	client.DeleteAccount(region, 0, orgId, aliId)
	client.DeleteAccount(region, 0, aliId, aliId)
	client.DeleteAccount(region, 0, bob.Id, bob.Id)
	client.DeleteAccount(region, 0, cat.Id, cat.Id)
	SR.AvatarClient.DeleteAll()
	cnn := SR.DlmAndDataRedisPool.Get()
	defer cnn.Close()
	cnn.Do("FLUSHALL")
}
