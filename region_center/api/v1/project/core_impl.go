package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

type api struct {
	store             store
	maxGetEntityCount int
}

type store interface {
}

type addProjectMember struct{
	Entity
	Role ProjectRole `json:"role"`
}

type projectMember struct{
	Entity
	CommonTimeProps
	Role ProjectRole `json:"role"`
}

type project struct {
	CommonNodeProps
	CommonAbstractNodeProps
	ArchivedOn *time.Time `json:"archivedOn,omitempty"`
	StartOn    *time.Time `json:"startOn,omitempty"`
	DueOn      *time.Time `json:"dueOn,omitempty"`
	FileCount  uint64     `json:"fileCount"`
	FileSize   uint64     `json:"fileSize"`
	IsPublic   bool       `json:"isPublic"`
}
