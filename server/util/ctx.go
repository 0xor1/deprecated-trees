package util

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/scrypt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"sync"
	t "time"
)

type Ctx interface {
	TryMyId() *Id
	MyId() Id
	Time() time
	Error() err
	Cache() cache
	Validate() validate
	Mail() mail
	Crypt() crypt
}

type CentralCtx interface {
	Ctx
	RegionalV1PrivateClient() RegionalV1PrivateClient
	Avatar() avatar
	CentralDb() centralDb
}

type RegionalCtx interface {
	Ctx
	Db() regionalDb
}

type time interface {
	Now() t.Time
}

type err interface {
	PanicIf(error)
	IsSqlErrNoRowsElsePanicIf(e error) bool
	Log(error)
}

type errorLog interface {
	Log(error)
}

type cache interface {
	GetValue(res interface{}, key string, dlmKeys []string, args interface{}) (cacheHit bool)
	SetValue(res interface{}, key string, dlmKeys []string, args interface{})
	DeleteKeys(keys []string)
	DlmKeyForSystem() string
	DlmKeyForAccountMaster(id Id) string
	DlmKeyForAccount(id Id) string
	DlmKeyForAccountActivities(id Id) string
	DlmKeyForAccountMember(id Id) string
	DlmKeyForAllAccountMembers(id Id) string
	DlmKeyForProjectMaster(id Id) string
	DlmKeyForProject(id Id) string
	DlmKeyForProjectActivities(id Id) string
	DlmKeyForProjectMember(id Id) string
	DlmKeyForAllProjectMembers(id Id) string
	DlmKeyForTask(id Id) string
	DlmKeyForTasks(ids []Id) []string
}

type validate interface {
	Exists(exists bool)
	Limit(limit int) int
	EntityCount(count int)
	Name(name string)
	Pwd(pwd string)
	Email(email string)
	MemberHasAccountOwnerAccess(accountRole *AccountRole)
	MemberHasAccountAdminAccess(accountRole *AccountRole)
	MemberHasProjectAdminAccess(accountRole *AccountRole, projectRole *ProjectRole)
	MemberHasProjectWriteAccess(accountRole *AccountRole, projectRole *ProjectRole)
	MemberIsAProjectMemberWithWriteAccess(projectRole *ProjectRole)
	MemberHasProjectReadAccess(accountRole *AccountRole, projectRole *ProjectRole, projectIsPublic *bool)
}

type mail interface {
	Send(sendTo []string, content string)
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

type RegionalV1PrivateClient interface {
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
	TreeShardCount() int
	Tree(int) isql.ReplicaSet
	GetAccountRole(shard int, accountId, memberId Id) *AccountRole
	GetAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole)
	GetAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool)
	GetPublicProjectsEnabled(shard int, accountId Id) bool
	MakeChangeHelper(shard int, sql string, args ...interface{})
}

type _ctx struct {
	myId                   *Id
	requestStartUnixMillis int64
	req                    *http.Request
	resp                   http.ResponseWriter
	queryInfosMtx          *sync.RWMutex
	queryInfos             []*queryInfo
	cache                  *_cache
	centralDb              *_centralDb
	regionalDb             *_regionalDb
	statics                *statics
}

type statics struct {
	env                     string
	region                  string
	server                  string
	version                 string
	masterCacheKey          string
	redisPool               iredis.Pool
	regionalV1PrivateClient RegionalV1PrivateClient
	mail                    mail
	avatar                  avatar
	time                    *_time
	err                     *_err
	validate                *_validate
	crypt                   *_crypt
	accountDb               isql.ReplicaSet
	pwdDb                   isql.ReplicaSet
	treeShards              map[int]isql.ReplicaSet
}

func (c *_ctx) TryMyId() *Id {
	return c.myId
}

func (c *_ctx) MyId() Id {
	if c.myId == nil {
		unauthorizedErr.Panic()
	}
	return *c.myId
}

func (c *_ctx) Time() time {
	return c.statics.time
}

func (c *_ctx) Error() err {
	return c.statics.err
}

func (c *_ctx) Cache() cache {
	return c.cache
}

func (c *_ctx) Validate() validate {
	return c.statics.validate
}

func (c *_ctx) Mail() mail {
	return c.statics.mail
}

func (c *_ctx) Crypt() crypt {
	return c.statics.crypt
}

func (c *_ctx) Avatar() avatar {
	return c.statics.avatar
}

func (c *_ctx) RegionalV1PrivateClient() RegionalV1PrivateClient {
	return c.statics.regionalV1PrivateClient
}

func (c *_ctx) CentralDb() centralDb {
	return c.centralDb
}

func (c *_ctx) Db() regionalDb {
	return c.regionalDb
}

func (c *_ctx) writeQueryInfo(qi *queryInfo) {
	c.queryInfosMtx.Lock()
	defer c.queryInfosMtx.Unlock()
	c.queryInfos = append(c.queryInfos, qi)
}

func (c *_ctx) getQueryInfos() []*queryInfo {
	c.queryInfosMtx.RLock()
	defer c.queryInfosMtx.RUnlock()
	cpy := make([]*queryInfo, 0, len(c.queryInfos))
	for _, qi := range c.queryInfos {
		cpy = append(cpy, qi)
	}
	return cpy
}

type _time struct{}

func (t *_time) Now() t.Time {
	return t.Now().UTC()
}

type _err struct {
	log errorLog
}

func (e *_err) PanicIf(err error) {
	panicIf(err)
}

func (e *_err) IsSqlErrNoRowsElsePanicIf(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	e.PanicIf(err)
	return false
}

func (e *_err) Log(err error) {
	e.log.Log(err)
}

type _cache struct {
	ctx                *_ctx
	retrievedDlms      map[string]int64
	dlmsToUpdate       map[string]interface{}
	cacheItemsToUpdate map[string]interface{}
	cacheKeysToDelete  map[string]interface{}
}

func (c *_cache) GetValue(val interface{}, key string, dlmKeys []string, args interface{}) bool {
	if key == "" {
		return false
	}
	dlm, err := c.getDlm(dlmKeys)
	if err != nil {
		c.ctx.Error().Log(err)
		return false
	}
	if dlm > c.ctx.requestStartUnixMillis {
		return false
	}
	jsonBytes, err := json.Marshal(&valueCacheKey{MasterKey: c.ctx.statics.masterCacheKey, Key: key, Args: args})
	if err != nil {
		c.ctx.Error().Log(err)
		return false
	}
	cnn := c.ctx.statics.redisPool.Get()
	defer cnn.Close()
	start := t.Now()
	jsonBytes, err = redis.Bytes(cnn.Do(GET, jsonBytes))
	c.ctx.writeQueryInfo(&queryInfo{Query: GET, Args: jsonBytes, Duration: t.Now().Sub(start)})
	if err != nil {
		c.ctx.Error().Log(err)
		return false
	}
	if len(jsonBytes) == 0 {
		return false
	}
	err = json.Unmarshal(jsonBytes, val)
	if err != nil {
		c.ctx.Error().Log(err)
		return false
	}
	return true
}

func (c *_cache) SetValue(val interface{}, key string, dlmKeys []string, args interface{}) {
	if val == nil || key == "" {
		InvalidArgumentsErr.Panic()
	}
	valBytes, err := json.Marshal(val)
	if err != nil {
		c.ctx.Error().Log(err)
		return
	}
	cacheKeyBytes, err := json.Marshal(&valueCacheKey{MasterKey: c.ctx.statics.masterCacheKey, Key: key, Args: args})
	if err != nil {
		c.ctx.Error().Log(err)
		return
	}
	for _, dlmKey := range dlmKeys {
		c.dlmsToUpdate[dlmKey] = nil
	}
	c.cacheItemsToUpdate[string(cacheKeyBytes)] = valBytes
}

func (c *_cache) DeleteKeys(keys []string) {
	for _, key := range keys {
		c.cacheKeysToDelete[key] = nil
	}
}

func (c *_cache) DlmKeyForSystem() string {
	return "sys"
}

func (c *_cache) DlmKeyForAccountMaster(accountId Id) string {
	return c.dlmKeyFor("amstr", accountId)
}

func (c *_cache) DlmKeyForAccount(accountId Id) string {
	return c.dlmKeyFor("a", accountId)
}

func (c *_cache) DlmKeyForAccountActivities(accountId Id) string {
	return c.dlmKeyFor("aa", accountId)
}

func (c *_cache) DlmKeyForAccountMember(accountId Id) string {
	return c.dlmKeyFor("am", accountId)
}

func (c *_cache) DlmKeyForAllAccountMembers(accountId Id) string {
	return c.dlmKeyFor("ams", accountId)
}

func (c *_cache) DlmKeyForProjectMaster(projectId Id) string {
	return c.dlmKeyFor("pmstr", projectId)
}

func (c *_cache) DlmKeyForProject(projectId Id) string {
	return c.dlmKeyFor("p", projectId)
}

func (c *_cache) DlmKeyForProjectActivities(projectId Id) string {
	return c.dlmKeyFor("pa", projectId)
}

func (c *_cache) DlmKeyForProjectMember(projectMemberId Id) string {
	return c.dlmKeyFor("pm", projectMemberId)
}

func (c *_cache) DlmKeyForAllProjectMembers(projectId Id) string {
	return c.dlmKeyFor("pms", projectId)
}

func (c *_cache) DlmKeyForTask(taskId Id) string {
	return c.dlmKeyFor("t", taskId)
}

func (c *_cache) DlmKeyForTasks(taskIds []Id) []string {
	strs := make([]string, 0, len(taskIds))
	for _, id := range taskIds {
		strs = append(strs, c.dlmKeyFor("t", id))
	}
	return strs
}

func (c *_cache) dlmKeyFor(typeKey string, id Id) string {
	return typeKey + ":" + id.String()
}

func (c *_cache) getDlm(dlmKeys []string) (int64, error) {
	panicIfRetrievedDlmsAreMissingEntries := false
	if len(c.retrievedDlms) > 0 {
		panicIfRetrievedDlmsAreMissingEntries = true
	}
	dlmsToFetch := make([]interface{}, 0, len(dlmKeys))
	latestDlm := int64(0)
	for _, dlmKey := range dlmKeys {
		dlm, exists := c.retrievedDlms[dlmKey]
		if !exists {
			if panicIfRetrievedDlmsAreMissingEntries {
				panic(&missingDlmErr{
					dlmKey:  dlmKey,
					reqPath: c.ctx.req.URL.Path,
				})
			}
			dlmsToFetch = append(dlmsToFetch, dlmKey)
		} else if exists && dlm > latestDlm {
			latestDlm = dlm
		}
	}
	if len(dlmsToFetch) > 0 {
		cnn := c.ctx.statics.redisPool.Get()
		defer cnn.Close()
		start := t.Now()
		dlms, err := redis.Int64s(cnn.Do("MGET", dlmsToFetch...))
		c.ctx.writeQueryInfo(&queryInfo{Query: "MGET", Args: dlmsToFetch, Duration: t.Now().Sub(start)})
		if err != nil {
			return 0, err
		}
		for _, dlm := range dlms {
			if dlm > latestDlm {
				latestDlm = dlm
			}
		}
	}
	return latestDlm, nil
}

func (c *_cache) setDlmsForUpdate(dlmKeys []string) {
	for _, key := range dlmKeys {
		c.dlmsToUpdate[key] = nil
	}
}

func (c *_cache) doCacheUpdate() {
	if len(c.dlmsToUpdate) == 0 && len(c.cacheItemsToUpdate) == 0 && len(c.cacheKeysToDelete) == 0 {
		return
	}
	setArgs := make([]interface{}, 0, (len(c.dlmsToUpdate)*2)+(len(c.cacheItemsToUpdate)*2))
	for k := range c.dlmsToUpdate {
		setArgs = append(setArgs, k, c.ctx.requestStartUnixMillis)
	}
	for k, v := range c.cacheItemsToUpdate {
		setArgs = append(setArgs, k, v)
	}
	delArgs := make([]interface{}, 0, len(c.cacheKeysToDelete))
	for k := range c.cacheKeysToDelete {
		delArgs = append(delArgs, k)
	}
	cnn := c.ctx.statics.redisPool.Get()
	defer cnn.Close()
	if len(setArgs) > 0 {
		start := t.Now()
		_, err := cnn.Do("MSET", setArgs...)
		c.ctx.writeQueryInfo(&queryInfo{Query: "MSET", Args: setArgs, Duration: t.Now().Sub(start)})
		if err != nil {
			c.ctx.Error().Log(err)
		}
	}
	if len(delArgs) > 0 {
		start := t.Now()
		_, err := cnn.Do("DEL", setArgs...)
		c.ctx.writeQueryInfo(&queryInfo{Query: "DEL", Args: setArgs, Duration: t.Now().Sub(start)})
		if err != nil {
			c.ctx.Error().Log(err)
		}
	}

}

type _validate struct {
	emailRegex            *regexp.Regexp //regexp.MustCompile(`.+@.+\..+`)
	nameRegexMatchers     []*regexp.Regexp
	pwdRegexMatchers      []*regexp.Regexp
	maxAvatarDim          uint
	maxProcessEntityCount int
	nameMinRuneCount      int
	nameMaxRuneCount      int
	pwdMinRuneCount       int
	pwdMaxRuneCount       int
}

func (v *_validate) Exists(exists bool) {
	if !exists {
		NoSuchEntityErr.Panic()
	}
}

func (v *_validate) Limit(limit int) int {
	if limit < 1 || limit > v.maxProcessEntityCount {
		limit = v.maxProcessEntityCount
	}
	return limit
}

func (v *_validate) EntityCount(count int) {
	if count < 1 || count > v.maxProcessEntityCount {
		InvalidEntityCountErr.Panic()
	}
}

func (v *_validate) Name(name string) {
	validateStringArg("name", name, v.nameMinRuneCount, v.nameMaxRuneCount, v.nameRegexMatchers)
}

func (v *_validate) Pwd(pwd string) {
	validateStringArg("pwd", pwd, v.pwdMinRuneCount, v.pwdMaxRuneCount, v.pwdRegexMatchers)
}

func (v *_validate) Email(email string) {
	validateStringArg("email", email, 6, 254, []*regexp.Regexp{v.emailRegex})
}

func (v *_validate) MemberHasAccountOwnerAccess(accountRole *AccountRole) {
	if accountRole == nil || *accountRole != AccountOwner {
		InsufficientPermissionErr.Panic()
	}
}

func (v *_validate) MemberHasAccountAdminAccess(accountRole *AccountRole) {
	if accountRole == nil || (*accountRole != AccountOwner && *accountRole != AccountAdmin) {
		InsufficientPermissionErr.Panic()
	}
}

func (v *_validate) MemberHasProjectAdminAccess(accountRole *AccountRole, projectRole *ProjectRole) {
	if accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || *projectRole != ProjectAdmin)) {
		InsufficientPermissionErr.Panic()
	}
}

func (v *_validate) MemberHasProjectWriteAccess(accountRole *AccountRole, projectRole *ProjectRole) {
	if accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter))) {
		InsufficientPermissionErr.Panic()
	}
}

func (v *_validate) MemberIsAProjectMemberWithWriteAccess(projectRole *ProjectRole) {
	if projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter) {
		InsufficientPermissionErr.Panic()
	}
}

func (v *_validate) MemberHasProjectReadAccess(accountRole *AccountRole, projectRole *ProjectRole, projectIsPublic *bool) {
	if projectIsPublic == nil || (!*projectIsPublic && (accountRole == nil || ((*accountRole != AccountOwner && *accountRole != AccountAdmin) && (projectRole == nil || (*projectRole != ProjectAdmin && *projectRole != ProjectWriter && *projectRole != ProjectReader))))) {
		InsufficientPermissionErr.Panic()
	}
}

type _crypt struct {
	err          *_err
	urlSafeRunes []rune //[]rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	cryptCodeLen int
	saltLen      int
	scryptN      int
	scryptR      int
	scryptP      int
	scryptKeyLen int
}

func (c *_crypt) CreatePwdSalt() []byte {
	k := make([]byte, c.saltLen)
	_, err := io.ReadFull(rand.Reader, k)
	c.err.PanicIf(err)
	return k
}

func (c *_crypt) CreateUrlSafeString() string {
	buf := make([]rune, c.cryptCodeLen)
	urlSafeRunesLength := big.NewInt(int64(len(c.urlSafeRunes)))
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, urlSafeRunesLength)
		c.err.PanicIf(err)
		buf[i] = c.urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf)
}

func (c *_crypt) ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte {
	key, err := scrypt.Key(password, salt, N, r, p, keyLen)
	c.err.PanicIf(err)
	return key
}

func (c *_crypt) ScryptN() int {
	return c.scryptN
}

func (c *_crypt) ScryptR() int {
	return c.scryptR
}

func (c *_crypt) ScryptP() int {
	return c.scryptP
}

func (c *_crypt) ScryptKeyLen() int {
	return c.scryptKeyLen
}

type _centralDb struct {
	ctx *_ctx
}

func (c *_centralDb) Account() isql.ReplicaSet {
	return &replicaSet{ctx: c.ctx, rs: c.ctx.statics.accountDb}
}

func (c *_centralDb) Pwd() isql.ReplicaSet {
	return &replicaSet{ctx: c.ctx, rs: c.ctx.statics.pwdDb}
}

type _regionalDb struct {
	ctx *_ctx
}

func (r *_regionalDb) TreeShardCount() int {
	return len(r.ctx.statics.treeShards)
}

func (r *_regionalDb) Tree(shard int) isql.ReplicaSet {
	return &replicaSet{ctx: r.ctx, rs: r.ctx.statics.treeShards[shard]}
}

func (r *_regionalDb) GetAccountRole(shard int, accountId, memberId Id) *AccountRole {
	row := r.Tree(shard).QueryRow(`SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?`, []byte(accountId), []byte(memberId))
	res := AccountRole(3)
	if r.ctx.statics.err.IsSqlErrNoRowsElsePanicIf(row.Scan(&res)) {
		return nil
	}
	return &res
}

func (r *_regionalDb) GetAccountAndProjectRoles(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole) {
	accountIdBytes := []byte(accountId)
	memberIdBytes := []byte(memberId)
	row := r.Tree(shard).QueryRow(`SELECT role accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM accountMembers WHERE account=? AND isActive=true AND id=?`, accountIdBytes, []byte(projectId), memberIdBytes, accountIdBytes, memberIdBytes)
	var accRole *AccountRole
	var projRole *ProjectRole
	if r.ctx.statics.err.IsSqlErrNoRowsElsePanicIf(row.Scan(&accRole, &projRole)) {
		return nil, nil
	}
	return accRole, projRole
}

func (r *_regionalDb) GetAccountAndProjectRolesAndProjectIsPublic(shard int, accountId, projectId, memberId Id) (*AccountRole, *ProjectRole, *bool) {
	accountIdBytes := []byte(accountId)
	projectIdBytes := []byte(projectId)
	memberIdBytes := []byte(memberId)
	row := r.Tree(shard).QueryRow(`SELECT isPublic, (SELECT role FROM accountMembers WHERE account=? AND isActive=true AND id=?) accountRole, (SELECT role FROM projectMembers WHERE account=? AND isActive=true AND project=? AND id=?) projectRole FROM projects WHERE account=? AND id=?`, accountIdBytes, memberIdBytes, accountIdBytes, projectIdBytes, memberIdBytes, accountIdBytes, projectIdBytes)
	isPublic := false
	var accRole *AccountRole
	var projRole *ProjectRole
	if r.ctx.statics.err.IsSqlErrNoRowsElsePanicIf(row.Scan(&isPublic, &accRole, &projRole)) {
		return nil, nil, nil
	}
	return accRole, projRole, &isPublic
}

func (r *_regionalDb) GetPublicProjectsEnabled(shard int, accountId Id) bool {
	row := r.Tree(shard).QueryRow(`SELECT publicProjectsEnabled FROM accounts WHERE id=?`, []byte(accountId))
	res := false
	r.ctx.statics.err.PanicIf(row.Scan(&res))
	return res
}

func (r *_regionalDb) MakeChangeHelper(shard int, sql string, args ...interface{}) {
	row := r.Tree(shard).QueryRow(sql, args...)
	changeMade := false
	r.ctx.statics.err.PanicIf(row.Scan(&changeMade))
	if !changeMade {
		noChangeMadeErr.Panic()
	}
}

type replicaSet struct {
	ctx *_ctx
	rs  isql.ReplicaSet
}

func (rs *replicaSet) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := t.Now()
	res, err := rs.rs.Exec(query, args...)
	rs.ctx.writeQueryInfo(&queryInfo{Query: query, Args: args, Duration: t.Now().Sub(start)})
	return res, err
}

func (rs *replicaSet) Query(query string, args ...interface{}) (isql.Rows, error) {
	start := t.Now()
	rows, err := rs.rs.Query(query, args...)
	rs.ctx.writeQueryInfo(&queryInfo{Query: query, Args: args, Duration: t.Now().Sub(start)})
	return rows, err
}

func (rs *replicaSet) QueryRow(query string, args ...interface{}) isql.Row {
	start := t.Now()
	row := rs.rs.QueryRow(query, args...)
	rs.ctx.writeQueryInfo(&queryInfo{Query: query, Args: args, Duration: t.Now().Sub(start)})
	return row
}

//helpers

func newCtx(myId *Id, req *http.Request, resp http.ResponseWriter, statics *statics) *_ctx {
	c := &_ctx{
		myId: myId,
		requestStartUnixMillis: statics.time.Now().UnixNano() / 1000,
		req:           req,
		resp:          resp,
		queryInfosMtx: &sync.RWMutex{},
		queryInfos:    make([]*queryInfo, 0, 10),
		statics:       statics,
	}
	c.cache = &_cache{
		ctx:                c,
		retrievedDlms:      map[string]int64{},
		dlmsToUpdate:       map[string]interface{}{},
		cacheItemsToUpdate: map[string]interface{}{},
		cacheKeysToDelete:  map[string]interface{}{},
	}
	c.centralDb = &_centralDb{
		ctx: c,
	}
	c.regionalDb = &_regionalDb{
		ctx: c,
	}
	return c
}

func newStatics(env, region, server, version, masterCacheKey string, redisPool iredis.Pool, privateRegionClient RegionalV1PrivateClient, mail mail, avatar avatar, errorLog errorLog, nameRegexMatchers, pwdRegexMatchers []*regexp.Regexp, maxAvatarDim uint, maxProcessEntityCount, nameMinRuneCount, nameMaxRuneCount, pwdMinRuneCount, pwdMaxRuneCount, cryptCodeLen, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int, accountDb, pwdDb isql.ReplicaSet, treeShards map[int]isql.ReplicaSet) *statics {
	err := &_err{log: errorLog}
	return &statics{
		env:                     env,
		region:                  region,
		server:                  server,
		version:                 version,
		masterCacheKey:          masterCacheKey,
		redisPool:               redisPool,
		regionalV1PrivateClient: privateRegionClient,
		mail:   mail,
		avatar: avatar,
		time:   &_time{},
		err:    err,
		validate: &_validate{
			emailRegex:            regexp.MustCompile(`.+@.+\..+`),
			nameRegexMatchers:     nameRegexMatchers,
			pwdRegexMatchers:      pwdRegexMatchers,
			maxAvatarDim:          maxAvatarDim,
			maxProcessEntityCount: maxProcessEntityCount,
			nameMinRuneCount:      nameMinRuneCount,
			nameMaxRuneCount:      nameMaxRuneCount,
			pwdMinRuneCount:       pwdMinRuneCount,
			pwdMaxRuneCount:       pwdMaxRuneCount,
		},
		crypt: &_crypt{
			err:          err,
			urlSafeRunes: []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
			cryptCodeLen: cryptCodeLen,
			saltLen:      saltLen,
			scryptN:      scryptN,
			scryptR:      scryptR,
			scryptP:      scryptP,
			scryptKeyLen: scryptKeyLen,
		},
		accountDb:  accountDb,
		pwdDb:      pwdDb,
		treeShards: treeShards,
	}
}

type valueCacheKey struct {
	MasterKey string      `json:"masterKey"`
	Key       string      `json:"key"`
	Args      interface{} `json:"args"`
}

type queryInfo struct {
	Query    string      `json:"query"`
	Args     interface{} `json:"args"`
	Duration t.Duration  `json:"duration"`
}
