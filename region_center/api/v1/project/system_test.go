package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bitbucket.org/0xor1/task_center/region_center/api/v1/private"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_system(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	internalApi := private.New(shards, maxProcessEntityCount)
	projectApi := New(shards, maxProcessEntityCount)

	orgId := NewId()
	ali := AddMemberInternal{}
	ali.Id = NewId()
	ali.Name = "ali"
	ali.Role = AccountOwner
	bob := AddMemberInternal{}
	bob.Id = NewId()
	bob.Name = "bob"
	bob.Role = AccountAdmin
	cat := AddMemberInternal{}
	cat.Id = NewId()
	cat.Name = "cat"
	cat.Role = AccountMemberOfOnlySpecificProjects
	dan := AddMemberInternal{}
	dan.Id = NewId()
	dan.Name = "dan"
	dan.Role = AccountMemberOfOnlySpecificProjects
	internalApi.CreateAccount(orgId, ali.Id, ali.Name)
	internalApi.AddMembers(0, orgId, ali.Id, []*AddMemberInternal{&bob, &cat, &dan})

	project1 := projectApi.CreateProject(0, orgId, ali.Id, "p1", "p1_desc", nil, nil, true, false, []*addMember{})

	//clean up the database
	internalApi.DeleteAccount(0, orgId, ali.Id)
}
