package node

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"bitbucket.org/0xor1/task_center/region_center/api/v1/private"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"bitbucket.org/0xor1/task_center/region_center/api/v1/project"
)

func Test_system(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	privateApi := private.New(shards, maxProcessEntityCount)
	projectApi := project.New(shards, maxProcessEntityCount)
	api := New(shards, maxProcessEntityCount)

}
