package private

import (
	. "bitbucket.org/0xor1/task/server/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"bitbucket.org/0xor1/task/server/config"
	"net/http/httptest"
)

func Test_system(t *testing.T) {
	staticResources := config.Config("", "", NewClient, Endpoints)
	testServer := httptest.NewServer(staticResources)
	region := "lcl"
	staticResources.RegionalV1PrivateClient = NewClient(map[string]string{
		region: testServer.URL,
	})
	client := staticResources.RegionalV1PrivateClient

	aliId := NewId()
	orgId := NewId()
	client.CreateAccount(region, orgId, aliId, "ali", nil)
	bob := AddMemberPrivate{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := AddMemberPrivate{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfOnlySpecificProjects
	client.AddMembers(region, 0, orgId, aliId, []*AddMemberPrivate{&bob, &cat})
	val, err := client.MemberIsOnlyAccountOwner(region, 0, orgId, aliId)
	assert.Nil(t, err)
	assert.True(t, val)
	client.SetMemberName(region, 0, orgId, aliId, "aliNew")
	val, err = client.MemberIsAccountOwner(region, 0, orgId, aliId)
	assert.Nil(t, err)
	assert.True(t, val)
	client.RemoveMembers(region, 0, orgId, aliId, []Id{bob.Id})
	client.DeleteAccount(region, 0, orgId, aliId)
}
