package ctx

import (
	"bitbucket.org/0xor1/task/server/util/avatar"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/mail"
	"bitbucket.org/0xor1/task/server/util/private"
	"database/sql"
	"github.com/0xor1/isql"
	"regexp"
)

// per request ctx
type Ctx interface {
	//session
	TryMe() *id.Id
	Me() id.Id
	//error logging
	Log(err error)
	//db access
	AccountExec(query string, args ...interface{}) (sql.Result, error)
	AccountQuery(query string, args ...interface{}) (isql.Rows, error)
	AccountQueryRow(query string, args ...interface{}) isql.Row
	PwdExec(query string, args ...interface{}) (sql.Result, error)
	PwdQuery(query string, args ...interface{}) (isql.Rows, error)
	PwdQueryRow(query string, args ...interface{}) isql.Row
	TreeShardCount() int
	TreeExec(shard int, query string, args ...interface{}) (sql.Result, error)
	TreeQuery(shard int, query string, args ...interface{}) (isql.Rows, error)
	TreeQueryRow(shard int, query string, args ...interface{}) isql.Row
	//cache access
	GetCacheValue(val interface{}, key string, dlmKeys []string, args ...interface{}) bool
	SetCacheValue(val interface{}, key string, dlmKeys []string, args ...interface{})
	DeleteDlmKeys(keys []string)
	//basic static values
	NameRegexMatchers() []*regexp.Regexp
	PwdRegexMatchers() []*regexp.Regexp
	NameMinRuneCount() int
	NameMaxRuneCount() int
	PwdMinRuneCount() int
	PwdMaxRuneCount() int
	MaxProcessEntityCount() int
	CryptCodeLen() int
	SaltLen() int
	ScryptN() int
	ScryptR() int
	ScryptP() int
	ScryptKeyLen() int
	RegionalV1PrivateClient() private.V1Client
	MailClient() mail.Client
	AvatarClient() avatar.Client
}
