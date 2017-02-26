package tree

import (
	. "bitbucket.org/robsix/task_center/misc"
	"sync"
	"time"
)

func newMemStore(memCore *memCore) internalStore {
	panic(NotImplementedErr)
}

type memCore struct {
	tasks       []*task
	taskSets    []*taskSet
	parents     []*parentRecord
	firstChild  []*firstChildRecord
	nextSibling []*nextSiblingRecord
	permissions []*permissionRecord
	mtx         *sync.RWMutex
}

type parentRecord struct {
	Org                  string
	Node                 string
	Parent               string
	NodeArchivedDateTime *time.Time
	NodeIsTask           bool
}

type firstChildRecord struct {
	Org              string
	Node             string
	FirstChild       string
	FirstChildIsTask bool
}

type nextSiblingRecord struct {
	Org               string
	Node              string
	NextSibling       string
	NextSiblingIsTask bool
}

type permissionRecord struct {
	Org  string
	Node string
	User string
	Role role
}
