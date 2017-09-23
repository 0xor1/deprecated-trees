package timeLogs

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"bitbucket.org/0xor1/task/server/region/api/v1/private"
	"bitbucket.org/0xor1/task/server/region/api/v1/project"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func Test_system(t *testing.T) {
	shards := map[int]isql.ReplicaSet{0: isql.NewReplicaSet("mysql", "tc_rc_trees:T@sk-C3n-T3r-Tr335@tcp(127.0.0.1:3306)/trees?parseTime=true&loc=UTC&multiStatements=true", nil)}
	maxProcessEntityCount := 100
	privateApi := private.New(shards, maxProcessEntityCount)
	projectApi := project.New(shards, maxProcessEntityCount)
	api := New(shards, maxProcessEntityCount)

}
