package model

import (
	. "bitbucket.org/robsix/task_center/misc"
	"time"
)

const (
	Owner  = role("owner")
	Admin  = role("admin")
	Writer = role("writer")
	Reader = role("reader")
)

type role string

type Task struct {
	NamedEntity
	User               Id        `json:"user"`
	TotalRemainingTime uint64    `json:"totalRemainingTime"`
	TotalLoggedTime    uint64    `json:"totalLoggedTime"`
	ChatCount          uint64    `json:"chatCount"`
	FileCount          uint64    `json:"fileCount"`
	FileSize           uint64    `json:"fileSize"`
	Created            time.Time `json:"created"`
}

type TaskSet struct {
	Task
	MinimumRemainingTime uint64 `json:"minimumRemainingTime"`
	IsParallel           bool   `json:"isParallel"`
	ChildCount           uint32 `json:"childCount"`
	TaskCount            uint64 `json:"taskCount"`
	SubFileCount         uint64 `json:"subFileCount"`
	SubFileSize          uint64 `json:"subFileSize"`
	ArchivedChildCount   uint32 `json:"archivedChildCount"`
	ArchivedTaskCount    uint64 `json:"archivedTaskCount"`
	ArchivedSubFileCount uint64 `json:"archivedSubFileCount"`
	ArchivedSubFileSize  uint64 `json:"archivedSubFileSize"`
}

type Member struct {
	NamedEntity
	AccessTask         Id     `json:"accessTask"`
	TotalRemainingTime uint64 `json:"totalRemainingTime"`
	TotalLoggedTime    uint64 `json:"totalLoggedTime"`
	Role               role   `json:"role"`
	IsActive           bool   `json:"isActive"`
	IsDeleted          bool   `json:"isDeleted"`
}
