package node

import (
	. "bitbucket.org/0xor1/task_center/misc"
)

type api struct {
	store                 store
	maxProcessEntityCount int
}

type store interface {
}

type abstractNode struct {
	CommonNodeProps
	CommonAbstractNodeProps
}

type node struct {
	CommonNodeProps
}