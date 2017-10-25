package timeLogs

import (
	. "bitbucket.org/0xor1/task/server/misc"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

type store interface {
}

type node struct {
	CommonNodeProps
	CommonAbstractNodeProps
	Member *Id `json:"member,omitempty"`
}
