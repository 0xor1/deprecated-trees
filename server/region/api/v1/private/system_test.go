package private

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_system(t *testing.T) {
	maxProcessEntityCount := 100
	api := New(map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}, maxProcessEntityCount)

	aliId := NewId()
	orgId := NewId()
	api.CreateAccount(orgId, aliId, "ali")
	bob := AddMemberPrivate{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := AddMemberPrivate{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfOnlySpecificProjects
	api.AddMembers(0, orgId, aliId, []*AddMemberPrivate{&bob, &cat})
	assert.True(t, api.MemberIsOnlyAccountOwner(0, orgId, aliId))
	api.RenameMember(0, orgId, aliId, "aliNew")
	assert.True(t, api.MemberIsAccountOwner(0, orgId, aliId))
	api.RemoveMembers(0, orgId, aliId, []Id{bob.Id})
	api.DeleteAccount(0, orgId, aliId)
}
