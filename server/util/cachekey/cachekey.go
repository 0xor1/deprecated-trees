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
	Key     string
	DlmKeys []string
}

func NewGet(key string) *Key {
	if key == "" {
		panic(errInvalidCacheKey)
	}
	return &Key{
		isGet:   true,
		Key:     key,
		DlmKeys: make([]string, 0, 10),
	}
}

func NewSet() *Key {
	return &Key{
		isGet:   false,
		DlmKeys: make([]string, 0, 6),
	}
}

func (k *Key) AccountMaster(account id.Id) *Key {
	return k.setKey("amstr", account)
}

func (k *Key) Account(account id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
	}
	return k.setKey("a", account)
}

func (k *Key) AccountActivities(account id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
	}
	return k.setKey("aa", account)
}

func (k *Key) AccountMembersMaster(account id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
	}
	return k.setKey("ammstr", account)
}

func (k *Key) AccountMember(account, member id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account)  //account master
		k.setKey("ammstr", account) //account members master
	}
	return k.setKey("am", member)
}

func (k *Key) AccountMembers(account id.Id, members []id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account)  //account master
		k.setKey("ammstr", account) //account members master
	}
	for _, member := range members {
		k.setKey("am", member)
	}
	return k
}

func (k *Key) AccountProjectsMaster(account id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
	}
	return k.setKey("apmstr", account)
}

func (k *Key) ProjectMaster(account, project id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
	}
	return k.setKey("pmstr", project)
}

func (k *Key) Project(account, project id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
		k.setKey("pmstr", project) //project master
	}
	return k.setKey("p", project)
}

func (k *Key) ProjectActivities(account, project id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
		k.setKey("pmstr", project) //project master
	}
	return k.setKey("pa", project)
}

func (k *Key) ProjectMembersMaster(account, project id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account)  //account master
		k.setKey("ammstr", account) //account members master
		k.setKey("pmstr", project)  //project master
	}
	return k.setKey("pmmstr", project)
}

func (k *Key) ProjectMember(account, project, member id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account)  //account master
		k.setKey("ammstr", account) //account members master
		k.setKey("am", member)      //account member (for name/displayName changes)
		k.setKey("pmstr", project)  //project master
		k.setKey("pmmstr", project) //project members master
	}
	return k.setKey("pm", member)
}

func (k *Key) ProjectMembers(account, project id.Id, members []id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account)  //account master
		k.setKey("ammstr", account) //account members master
		for _, member := range members {
			k.setKey("am", member) //account member (for name/displayName changes)
		}
		k.setKey("pmstr", project)  //project master
		k.setKey("pmmstr", project) //project members master
	}
	for _, member := range members {
		k.setKey("pm", member)
	}
	return k
}

func (k *Key) TaskParent(account, project, parent id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
		k.setKey("pmstr", project) //project master
	}
	k.setKey("t", parent)
	return k.setKey("tp", parent)
}

func (k *Key) Task(account, project, task id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
		k.setKey("pmstr", project) //project master
	}
	if project.Equal(task) {
		k.setKey("p", project)
	}
	return k.setKey("t", task)
}

func (k *Key) Tasks(account, project id.Id, tasks []id.Id) *Key {
	if k.isGet {
		k.setKey("amstr", account) //account master
		k.setKey("pmstr", project) //project master
	}
	for _, task := range tasks {
		if project.Equal(task) {
			k.setKey("p", project)
		}
		k.setKey("t", task)
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
