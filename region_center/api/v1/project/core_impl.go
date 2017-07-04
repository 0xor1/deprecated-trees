package project

import (
	. "bitbucket.org/0xor1/task_center/misc"
	"time"
)

func newApi(store store) Api {
	if store == nil {
		panic(NilCriticalParamErr)
	}
	return &api{
		store: store,
	}
}

type api struct {
	store store
}

type store interface {

}

type project struct {
	CommonNodeProps
	CommonAbstractNodeProps
	ArchivedOn *time.Time `json:"archivedOn,omitempty"`
	StartOn *time.Time `json:"startOn,omitempty"`
	DueOn *time.Time `json:"dueOn,omitempty"`
	FileCount uint64 `json:"fileCount"`
	FileSize uint64 `json:"fileSize"`
}
