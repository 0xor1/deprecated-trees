package account

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_System(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "t_r_trees:T@sk-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	privateApi := private.New(shards, maxProcessEntityCount)
	api := New(shards, 100)

	orgId := NewId()
	aliId := NewId()
	privateApi.CreateAccount(orgId, aliId, "ali", nil)
	bob := AddMemberPrivate{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := AddMemberPrivate{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfOnlySpecificProjects
	privateApi.AddMembers(0, orgId, aliId, []*AddMemberPrivate{&bob, &cat})

	assert.False(t, api.GetPublicProjectsEnabled(0, orgId, aliId))
	api.SetPublicProjectsEnabled(0, orgId, aliId, true)
	assert.True(t, api.GetPublicProjectsEnabled(0, orgId, aliId))
	api.SetMemberRole(0, orgId, aliId, bob.Id, AccountMemberOfAllProjects)
	members, more := api.GetMembers(0, orgId, aliId, nil, nil, nil, 2)
	assert.True(t, more)
	assert.Equal(t, 2, len(members))
	assert.True(t, aliId.Equal(members[0].Id))
	assert.True(t, bob.Id.Equal(members[1].Id))
	members, more = api.GetMembers(0, orgId, aliId, nil, nil, &members[0].Id, 100)
	assert.False(t, more)
	assert.Equal(t, 2, len(members))
	assert.True(t, bob.Id.Equal(members[0].Id))
	assert.True(t, cat.Id.Equal(members[1].Id))
	activities := api.GetActivities(0, orgId, aliId, nil, nil, nil, nil, 100)
	assert.Equal(t, 5, len(activities))
	me := api.GetMe(0, orgId, bob.Id)
	assert.Equal(t, AccountMemberOfAllProjects, me.Role)
	assert.True(t, bob.Id.Equal(me.Id))
	assert.Equal(t, true, me.IsActive)
	privateApi.DeleteAccount(0, orgId, aliId)
}
