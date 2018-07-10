package cachekey

import (
	"bitbucket.org/0xor1/trees/server/util/id"
	"github.com/0xor1/panic"
	"sort"
)

type Key struct {
	isGet   bool
	Key     string
	Args    []interface{}
	DlmKeys map[string]bool
}

func NewGet(key string, args ...interface{}) *Key {
	return &Key{
		isGet:   true,
		Key:     key,
		Args:    args,
		DlmKeys: map[string]bool{},
	}
}

func NewSetDlms() *Key {
	return &Key{
		isGet:   false,
		DlmKeys: map[string]bool{},
	}
}

func (k *Key) SortedDlmKeys() []string {
	sorted := make([]string, 0, len(k.DlmKeys))
	for key := range k.DlmKeys {
		idx := sort.SearchStrings(sorted, key)
		if idx == len(sorted) {
			sorted = append(sorted, key)
		} else if sorted[idx] != key {
			sorted = append(sorted, "")
			copy(sorted[idx+1:], sorted[idx:])
			sorted[idx] = key
		}
	}
	return sorted
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
		k.AccountMembersSet(account) //need to check this in case a member changed their name/displayName/hasAvatar
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

func (k *Key) CombinedTaskAndTaskChildrenSets(account, project id.Id, tasks []id.Id) *Key {
	for _, task := range tasks {
		k.Task(account, project, task)
		k.setKey("tcs", task)
	}
	return k
}

func (k *Key) ProjectTimeLogSet(account, project id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	}
	k.setKey("ptls", project)
	return k
}

func (k *Key) ProjectMemberTimeLogSet(account, project, member id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
	} else {
		k.ProjectTimeLogSet(account, project)
	}
	k.setKey("pmtls", member)
	return k
}

func (k *Key) TaskTimeLogSet(account, project, task id.Id, member *id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
		if member != nil {
			k.ProjectMemberTimeLogSet(account, project, *member)
		}
	} else if member != nil {
		k.ProjectMemberTimeLogSet(account, project, *member)
	} else {
		panic.If(true, "missing member in taskTimeLogSet dlm")
	}
	k.setKey("ttls", task)
	return k
}

func (k *Key) TimeLog(account, project, timeLog id.Id, task, member *id.Id) *Key {
	if k.isGet {
		k.ProjectMaster(account, project)
		if task != nil {
			k.TaskTimeLogSet(account, project, *task, member)
		}
	} else if task != nil {
		k.TaskTimeLogSet(account, project, *task, member)
	} else {
		panic.If(true, "missing tasl in taskTimeLog dlm")
	}
	k.setKey("tl", timeLog)
	return k
}

func (k *Key) setKey(typeKey string, id id.Id) *Key {
	var key string
	if id == nil {
		key = typeKey
	} else {
		key = typeKey + ":" + id.String()
	}
	k.DlmKeys[key] = true
	return k
}
