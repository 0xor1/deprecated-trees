package util

import (
	"encoding/json"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"io"
	"net/http"
	"regexp"
	"time"
)

type Ctx interface {
	TryMyId() *Id
	MyId() Id
	Cache() cache
	Validate() validate
	Email() email
	Error() err
	Crypt() crypt
	Log() log
}

type CentralCtx interface {
	Ctx
	PrivateRegionClient() privateRegionClient
	Avatar() avatar
	CentralDb() centralDb
}

type RegionalCtx interface {
	Ctx
	Db() regionalDb
}

type cache interface {
	GetValue(res interface{}, key string, dlmKeys []string, args ...interface{}) (cacheHit bool)
	SetValue(res interface{}, key string, dlmKeys []string, args ...interface{})
	DeleteKeys(keys []string)
	SetDlmsForUpdate(dlmKeys []string)
	DlmKeyForAccountMaster(id Id) string
	DlmKeyForAccount(id Id) string
	DlmKeyForAccountActivities(id Id) string
	DlmKeyForAccountMember(id Id) string
	DlmKeyForProjectMaster(id Id) string
	DlmKeyForProject(id Id) string
	DlmKeyForProjectActivities(id Id) string
	DlmKeyForProjectMember(id Id) string
	DlmKeyForTask(id Id) string
	DlmKeyForTasks(ids []Id) []string
	doCacheUpdate()
}

type validate interface {
	Limit(limit int) int
	EntityCount(count int)
	Name(name string)
	Pwd(pwd string)
	Email(email string)
}

type email interface {
	Send(sendTo []string, content string)
}

type err interface {
	PanicIf(error)
}

type crypt interface {
	CreatePwdSalt() []byte
	CreateUrlSafeString() string
	ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte
	ScryptN() int
	ScryptR() int
	ScryptP() int
	ScryptKeyLen() int
}

type log interface {
	Error(error)
}

type privateRegionClient interface {
	GetRegions() []string
	IsValidRegion(region string) bool
	CreateAccount(region string, account, myId Id, myName string, myDisplayName *string) int
	DeleteAccount(region string, shard int, account, myId Id)
	AddMembers(region string, shard int, account, myId Id, members []*AddMemberPrivate)
	RemoveMembers(region string, shard int, account, myId Id, members []Id)
	MemberIsOnlyAccountOwner(region string, shard int, account, myId Id) bool
	SetMemberName(region string, shard int, account, myId Id, newName string)
	SetMemberDisplayName(region string, shard int, account, myId Id, newDisplayName *string)
	MemberIsAccountOwner(region string, shard int, account, myId Id) bool
}

type avatar interface {
	MaxAvatarDim() uint
	Save(key string, mimeType string, data io.Reader)
	Delete(key string)
	DeleteAll()
}

type centralDb interface {
	Account() isql.ReplicaSet
	Pwd() isql.ReplicaSet
}

type regionalDb interface {
	Tree(int) isql.ReplicaSet
}

type ctx struct {
	myId                   *Id
	requestStartUnixMillis int64
	req                    *http.Request
	resp                   http.ResponseWriter
	retrievedDlms          map[string]time.Time
	dlmsToUpdate           map[string]interface{}
	cacheItemsToUpdate     map[string]interface{}
	cacheKeysToDelete      map[string]interface{}
	statics                *statics
}

type statics struct {
	env                   string
	region                string
	server                string
	version               string
	redisPool             iredis.Pool
	centralAccountDb      isql.ReplicaSet
	centralPwdDb          isql.ReplicaSet
	regionalTreeDbs       map[int]isql.ReplicaSet
	privateRegionClient   privateRegionClient
	masterCacheKey        string
	maxProcessEntityCount int
	nameRegexMatchers     []*regexp.Regexp
	pwdRegexMatchers      []*regexp.Regexp
	maxAvatarDim          uint
	nameMinRuneCount      int
	nameMaxRuneCount      int
	pwdMinRuneCount       int
	pwdMaxRuneCount       int
	cryptoCodeLen         int
	saltLen               int
	scryptN               int
	scryptR               int
	scryptP               int
	scryptKeyLen          int
}

func (c *ctx) TryMyId() *Id {
	return c.myId
}

func (c *ctx) MyId() Id {
	if c.myId == nil {
		unauthorizedErr.Panic()
	}
	return *c.myId
}

func (c *ctx) Cache() cache {
	return c
}

func (c *ctx) Validate() validate {
	return c
}

func (c *ctx) Email() email {
	return c
}

func (c *ctx) Error() err {
	return c
}

func (c *ctx) Crypt() crypt {
	return c
}

func (c *ctx) Log() log {
	return c
}

func (c *ctx) Avatar() avatar {
	return c
}

func (c *ctx) PrivateRegionClient() privateRegionClient {
	return c
}

func (c *ctx) CentralDb() centralDb {
	return c
}

func (c *ctx) Db() regionalDb {
	return c
}

func (c *ctx) GetValue(val interface{}, key string, dlmKeys []string, args ...interface{}) bool {
	dlmsToFetch := make([]interface{}, 0, len(dlmKeys))
	latestDlm := time.Time{}
	for _, dlmKey := range dlmKeys {
		t, exists := c.retrievedDlms[dlmKey]
		if !exists {
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && t.After(latestDlm) {
			latestDlm = t
		}
	}
	cnn := c.statics.redisPool.Get()
	defer cnn.Close()
	cnn.Do("MGET", dlmsToFetch...)
	return false
}

func (c *ctx) SetValue(val interface{}, key string, dlmKeys []string, args ...interface{}) {
	if val == nil || key == "" {
		InvalidArgumentsErr.Panic()
	}
	keyResBytes, err := json.Marshal(val)
	if err != nil {
		return
	}
	if len(args) > 0 {
		keyArgsBytes, err := json.Marshal(args)
		if err != nil {
			return
		}
		key += string(keyArgsBytes)
	}
	for _, dlmKey := range dlmKeys {
		c.dlmsToUpdate[dlmKey] = nil
	}
	c.cacheItemsToUpdate[key] = keyResBytes
}

func (c *ctx) DeleteKeys(keys []string) {
	for _, key := range keys {
		c.cacheKeysToDelete[key] = nil
	}
}

func (c *ctx) SetDlmsForUpdate(dlmKeys []string) {
	for _, key := range dlmKeys {
		c.dlmsToUpdate[key] = nil
	}
}

func (c *ctx) DlmKeyForAccountMaster(id Id) string {

}

func (c *ctx) DlmKeyForAccount(id Id) string {

}

func (c *ctx) DlmKeyForAccountActivities(id Id) string {

}

func (c *ctx) DlmKeyForAccountMember(id Id) string {

}

func (c *ctx) DlmKeyForProjectMaster(id Id) string {

}

func (c *ctx) DlmKeyForProject(id Id) string {

}

func (c *ctx) DlmKeyForProjectActivities(id Id) string {

}

func (c *ctx) DlmKeyForProjectMember(id Id) string {

}

func (c *ctx) DlmKeyForTask(id Id) string {

}

func (c *ctx) DlmKeyForTasks(ids []Id) []string {

}

func (c *ctx) doCacheUpdate() {

}
