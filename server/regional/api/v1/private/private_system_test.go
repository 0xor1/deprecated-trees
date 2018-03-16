package private

import (
	"bitbucket.org/0xor1/task/server/util/cnst"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/server"
	"bitbucket.org/0xor1/task/server/util/static"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func Test_system(t *testing.T) {
	SR := static.Config("", "", NewClient)
	serv := server.New(SR, Endpoints)
	testServer := httptest.NewServer(serv)
	region := "lcl"
	SR.RegionalV1PrivateClient = NewClient(map[string]string{
		region: testServer.URL,
	})
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
}
