package timelog

import (
	"bitbucket.org/0xor1/task/server/util/id"
	"time"
)

type TimeLog struct {
	Id                 id.Id     `json:"id"`
	Project            id.Id     `json:"project"`
	Task               id.Id     `json:"task"`
	Member             id.Id     `json:"member"`
	LoggedOn           time.Time `json:"loggedOn"`
	TaskHasBeenDeleted bool      `json:"taskHasBeenDeleted"`
	TaskName           string    `json:"taskName"`
	Duration           uint64    `json:"duration"`
	Note               *string   `json:"note"`
}
