package task

import (
	. "bitbucket.org/0xor1/task/server/misc"
	"time"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

type store interface {
}

type node struct {
	Id                   Id        `json:"id"`
	Name                 string    `json:"name"`
	CreatedOn            time.Time `json:"createdOn"`
	TotalRemainingTime   uint64    `json:"totalRemainingTime"`
	TotalLoggedTime      uint64    `json:"totalLoggedTime"`
	IsAbstract           bool      `json:"isAbstract"`
	Description          string    `json:"description"`
	LinkedFileCount      uint64    `json:"linkedFileCount"`
	ChatCount            uint64    `json:"chatCount"`
	MinimumRemainingTime *uint64   `json:"minimumRemainingTime,omitempty"`
	IsParallel           *bool     `json:"isParallel,omitempty"`
	Member               *Id       `json:"member,omitempty"`
}
