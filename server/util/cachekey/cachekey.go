package cachekey

import (
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"sort"
)

var (
	errInvalidCacheKey = &err.Err{Code: "u_c_ick", Message: "invalid cache key"}
)

type Key struct {
	isGet   bool
	KeyVal  string
	DlmKeys []string
}

func NewGet() *Key {
	return &Key{
		isGet:   true,
		DlmKeys: make([]string, 0, 10),
	}
}

func NewDlms() *Key {
	return &Key{
		isGet:   false,
		DlmKeys: make([]string, 0, 10),
	}
}

func (k *Key) Key(key string) *Key {
	k.KeyVal = key
	return k
}

func (k *Key) AccountMaster(account id.Id) *Key {
	return k.setKey("amstr", account)
}

func (k *Key) Account(account id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	}
	return k.setKey("a", account)
}

func (k *Key) AccountActivities(account id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	}
	return k.setKey("aa", account)
}

func (k *Key) AccountMembersSet(account id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	}
	return k.setKey("ams", account)
}

func (k *Key) AccountMember(account, member id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	} else {
		k.AccountMembersSet(account)
	}
	return k.setKey("am", member)
}

func (k *Key) AccountMembers(account id.Id, members []id.Id) *Key {
	for _, member := range members {
		k.AccountMember(account, member)
	}
	return k
}

func (k *Key) AccountProjectsSet(account id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	}
	return k.setKey("aps", account)
}

func (k *Key) ProjectMaster(account, project id.Id) *Key {
	if k.isGet {
		k.AccountMaster(account)
	}
	return k.setKey("pmstr", project)
}

func (k *Key) Project(account, project id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	} else {
		k.AccountProjectsSet(account)
	}
	return k.setKey("p", project).setKey("t", project) //projects are also tasks
}

func (k *Key) ProjectActivities(account, project id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	}
	return k.setKey("pa", project)
}

func (k *Key) ProjectMembersSet(account, project id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	}
	return k.setKey("pms", project)
}

func (k *Key) ProjectMember(account, project, member id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
		k.AccountMember(account, member) //account member (for name/displayName changes)
	} else {
		k.ProjectMembersSet(account, project)
	}
	return k.setKey("pm", member)
}

func (k *Key) ProjectMembers(account, project id.Id, members []id.Id) *Key {
	for _, member := range members {
		k.ProjectMember(account, project, member)
	}
	return k
}

func (k *Key) TaskChildrenSet(account, project, parent id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	}
	return k.setKey("tcs", parent)
}

func (k *Key) TaskChildrenSets(account, project id.Id, tasks []id.Id) *Key {
	for _, task := range tasks {
		k.TaskChildrenSet(account, project, task)
	}
	return k
}

func (k *Key) Task(account, project, task id.Id) *Key {
	if project.Equal(task) {
		k.Project(account, project) //let project handle project nodes
	} else {
		if k.isGet {
			k.ProjectMaster(account, project)
		}
		k.setKey("t", task)
	}
	return k
}

func (k *Key) Tasks(account, project id.Id, tasks []id.Id) *Key {
	for _, task := range tasks {
		k.Task(account, project, task)
	}
	return k
}

func (k *Key) setKey(typeKey string, id id.Id) *Key {
	var key string
	if id == nil {
		key = typeKey
	} else {
		key = typeKey + ":" + id.String()
	}
	idx := sort.SearchStrings(k.DlmKeys, key)
	if idx == len(k.DlmKeys) {
		k.DlmKeys = append(k.DlmKeys, key)
	} else if k.DlmKeys[idx] != key {
		k.DlmKeys = append(k.DlmKeys, "")
		copy(k.DlmKeys[idx+1:], k.DlmKeys[idx:])
		k.DlmKeys[idx] = key
	}
	return k
}
