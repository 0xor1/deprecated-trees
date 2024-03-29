package static

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/0xor1/config"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"github.com/0xor1/panic"
	"github.com/0xor1/trees/server/util/avatar"
	"github.com/0xor1/trees/server/util/cnst"
	"github.com/0xor1/trees/server/util/id"
	"github.com/0xor1/trees/server/util/mail"
	"github.com/0xor1/trees/server/util/private"
	"github.com/0xor1/trees/server/util/queryinfo"
	"github.com/0xor1/trees/server/util/redis"
	t "github.com/0xor1/trees/server/util/time"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

// pass in empty strings for no config file
func Config(configFile string, newPrivateV1Client func(env cnst.Env, scheme, nakedHost string) private.V1Client) *Resources {
	config := config.New(configFile, "_")
	//defaults set up for onebox local environment configuration i.e everything running on one machine
	// must be one of "lcl", "dev", "stg", "prd"
	config.SetDefault("env", string(cnst.LclEnv))
	// bind address eg "127.0.0.1:80"
	config.SetDefault("bindAddress", "127.0.0.1:80")
	// naked host "project-trees"
	config.SetDefault("nakedHost", "project-trees.com")
	// must be one of "central", "use", "usw", "euw", "asp", "aus", only considered when env is not lcl or dev
	config.SetDefault("region", "")
	// commit sha
	config.SetDefault("version", string(cnst.LclEnv))
	// relative path from server executable to static file resource directory
	config.SetDefault("fileServerDir", "client")
	// api docs path
	config.SetDefault("apiDocsRoute", "/api/docs")
	// api mget path
	config.SetDefault("apiMDoRoute", "/api/mdo")
	// api logout path
	config.SetDefault("apiLogoutRoute", "/api/logout")
	// session cookie name
	config.SetDefault("sessionCookieName", "t")
	// session cookie store
	config.SetDefault("sessionAuthKey64s", []interface{}{
		"Va3ZMfhH4qSfolDHLU7oPal599DMcL93A80rV2KLM_om_HBFFUbodZKOHAGDYg4LCvjYKaicodNmwLXROKVgcA",
		"WK_2RgRx6vjfWVkpiwOCB1fvv1yklnltstBjYlQGfRsl6LyVV4mkt6UamUylmkwC8MEgb9bSGr1FYgM2Zk20Ug",
	})
	config.SetDefault("sessionEncrKey32s", []interface{}{
		"3ICuYRUelY-4Fhak0Iw0_5CW24bJvxFWM0jAA78IIp8",
		"u80sYkgbBav52fJXbENYhN3Iyof7WhuLHHMaS_rmUQw",
	})
	// is caching enabled
	config.SetDefault("cachingEnabled", true)
	// incremental base64 value
	config.SetDefault("masterCacheKey", "0")
	// regexes that account names must match to be valid during account creation or name setting
	config.SetDefault("nameRegexMatchers", []interface{}{
		`^[0-9a-zA-Z_]+$`,
	})
	// regexes that account displayNames must match to be valid during account creation or displayName setting
	config.SetDefault("displayNameRegexMatchers", []interface{}{
		`^[^%]+$`,
	})
	// regexes that account pwds must match to be valid during account creation or pwd setting
	config.SetDefault("pwdRegexMatchers", []interface{}{
		`[0-9]`,
		`[a-z]`,
		`[A-Z]`,
		`[\W]`,
	})
	// minimum number of runes required for a valid account name
	config.SetDefault("nameMinRuneCount", 3)
	// maximum number of runes required for a valid account name
	config.SetDefault("nameMaxRuneCount", 50)
	// minimum number of runes required for a valid account displayName
	config.SetDefault("displayNameMinRuneCount", 3)
	// maximum number of runes required for a valid account displayName
	config.SetDefault("displayNameMaxRuneCount", 100)
	// minimum number of runes required for a valid account pwd
	config.SetDefault("pwdMinRuneCount", 8)
	// maximum number of runes required for a valid account pwd
	config.SetDefault("pwdMaxRuneCount", 100)
	// max number of entities that can be processed at once, also used for max limit value on queries
	config.SetDefault("maxProcessEntityCount", 100)
	// length of cryptographic codes, used in email links for validating email addresses and resetting pwds
	config.SetDefault("cryptCodeLen", 100)
	// length of salts used for pwd hashing
	config.SetDefault("saltLen", 64)
	// scrypt N value
	config.SetDefault("scryptN", 32768)
	// scrypt R value
	config.SetDefault("scryptR", 8)
	// scrypt P value
	config.SetDefault("scryptP", 1)
	// scrypt key length
	config.SetDefault("scryptKeyLen", 32)
	// private client secret base64 encoded
	config.SetDefault("regionalV1PrivateClientSecret", "bwIwGNgOdTWxCifGdL5BW5XhoWoctcTQyN3LLeSTo1nuDNebpKmlda2XaF66jOh1jaV7cvFRHScJrdyn8gSnMQ")
	// max avatar dimension
	config.SetDefault("maxAvatarDim", 250)
	// local avatar storage directory, relative
	config.SetDefault("lclAvatarDir", "avatar")
	// api key for spark post client
	config.SetDefault("sparkPostApiKey", "")
	// account primary sql connection
	config.SetDefault("accountDbPrimary", "t_c_accounts:T@sk-@cc-0unt5@tcp(localhost:3306)/accounts?parseTime=true&loc=UTC&multiStatements=true")
	// account slave sql connections
	config.SetDefault("accountDbSlaves", []interface{}{})
	// pwd primary sql connection
	config.SetDefault("pwdDbPrimary", "t_c_pwds:T@sk-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	// account slave sql connections
	config.SetDefault("pwdDbSlaves", []interface{}{})
	// tree shard sql connections
	config.SetDefault("treeShards", map[string]interface{}{
		"0": []interface{}{"t_r_trees:T@sk-Tr335@tcp(localhost:3306)/trees?parseTime=true&loc=UTC&multiStatements=true"},
	})
	// redis pool for caching layer
	config.SetDefault("dlmAndDataRedisPool", "localhost:6379")
	// redis pool for private request keys to check for replay attacks
	config.SetDefault("privateKeyRedisPool", "localhost:6379")

	//validate env value
	env := cnst.Env(config.GetString("env"))
	env.Validate()
	//validate region value if this isnt a one box environment
	region := cnst.Region(config.GetString("region"))
	if env == cnst.StgEnv || env == cnst.ProEnv {
		region.Validate()
	}
	bindAddress := config.GetString("bindAddress")
	nakedHost := config.GetString("nakedHost")
	clientHost := fmt.Sprintf("%s.%s", env, nakedHost)
	allHosts := []string{
		clientHost,
		fmt.Sprintf("%s-api.%s", env, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.CentralRegion, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.USWRegion, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.USERegion, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.EUWRegion, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.ASPRegion, nakedHost),
		fmt.Sprintf("%s-%s-api.%s", env, cnst.AUSRegion, nakedHost),
	}

	scheme := "https://"
	if env == cnst.LclEnv {
		scheme = "http://"
	}

	authKey64s := config.GetStringSlice("sessionAuthKey64s")
	encrKey32s := config.GetStringSlice("sessionEncrKey32s")
	sessionAuthEncrKeyPairs := make([][]byte, 0, len(authKey64s)*2)
	for i := range authKey64s {
		authBytes, e := base64.RawURLEncoding.DecodeString(authKey64s[i])
		panic.IfNotNil(e)
		panic.If(len(authBytes) != 64, "sessionAuthBytes length is not 64")
		encrBytes, e := base64.RawURLEncoding.DecodeString(encrKey32s[i])
		panic.IfNotNil(e)
		panic.If(len(encrBytes) != 32, "sessionEncrBytes length is not 32")
		sessionAuthEncrKeyPairs = append(sessionAuthEncrKeyPairs, authBytes, encrBytes)
	}
	sessionStore := sessions.NewCookieStore(sessionAuthEncrKeyPairs...)
	sessionStore.Options.MaxAge = 0
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = env != cnst.LclEnv
	sessionStore.Options.Domain = clientHost
	gob.Register(id.New()) //register Id type for sessionCookie

	var logError func(error)
	var logStats func(status int, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo)
	var avatarClient avatar.Client
	var mailClient mail.Client

	if env == cnst.LclEnv {
		//setup local environment interfaces
		logError = func(err error) {
			fmt.Println(fmt.Sprintf("%s\n%s", err, string(debug.Stack())))
		}
		logStats = func(status int, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo) {
			fmt.Println(status, fmt.Sprintf("%dms", t.NowUnixMillis()-reqStartUnixMillis), path)
			//often too much info when running locally, makes too much noise, but feel free to uncomment when necessary
			//queryInfosBytes, _ := json.Marshal(queryInfos)
			//fmt.Println(string(queryInfosBytes))
		}
		avatarClient = avatar.NewLocalClient(config.GetString("lclAvatarDir"), uint(config.GetInt("maxAvatarDim")))
		mailClient = mail.NewLocalClient()
	} else {
		//setup aws environment interfaces
		logError = func(err error) {
			fmt.Println(fmt.Sprintf("%s\n%s", err, string(debug.Stack())))
		}
		logStats = func(status int, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo) {
			fmt.Println(status, fmt.Sprintf("%dms", t.NowUnixMillis()-reqStartUnixMillis), path)
			//often too much info when running locally, makes too much noise, but feel free to uncomment when necessary
			//queryInfosBytes, _ := json.Marshal(queryInfos)
			//fmt.Println(string(queryInfosBytes))
		}
		avatarClient = avatar.NewLocalClient(config.GetString("lclAvatarDir"), uint(config.GetInt("maxAvatarDim")))
		mailClient = mail.NewSparkPostClient("noreply@"+clientHost, config.GetString("sparkPostApiKey"))
		//TODO setup aws s3 avatarStore storage
		//TODO setup datadog stats and error logging
	}

	nameRegexMatchers := make([]*regexp.Regexp, 0, len(config.GetStringSlice("nameRegexMatchers")))
	for _, str := range config.GetStringSlice("nameRegexMatchers") {
		nameRegexMatchers = append(nameRegexMatchers, regexp.MustCompile(str))
	}

	displayNameRegexMatchers := make([]*regexp.Regexp, 0, len(config.GetStringSlice("displayNameRegexMatchers")))
	for _, str := range config.GetStringSlice("displayNameRegexMatchers") {
		displayNameRegexMatchers = append(displayNameRegexMatchers, regexp.MustCompile(str))
	}
	pwdRegexMatchers := make([]*regexp.Regexp, 0, len(config.GetStringSlice("pwdRegexMatchers")))
	for _, str := range config.GetStringSlice("pwdRegexMatchers") {
		pwdRegexMatchers = append(pwdRegexMatchers, regexp.MustCompile(str))
	}

	var accountDb isql.ReplicaSet
	if config.GetString("accountDbPrimary") != "" {
		accountDb = isql.MustNewReplicaSet("mysql", config.GetString("accountDbPrimary"), config.GetStringSlice("accountDbSlaves")...)
	}

	var pwdDb isql.ReplicaSet
	if config.GetString("pwdDbPrimary") != "" {
		pwdDb = isql.MustNewReplicaSet("mysql", config.GetString("pwdDbPrimary"), config.GetStringSlice("pwdDbSlaves")...)
	}

	treeShardDbs := map[int]isql.ReplicaSet{}
	treeShards := config.GetMap("treeShards")
	if treeShards != nil {
		for k := range treeShards {
			shardId, e := strconv.ParseInt(k, 10, 32)
			panic.IfNotNil(e)
			v := config.GetStringSlice(fmt.Sprintf("treeShards.%s", k))
			treeShardDbs[int(shardId)] = isql.MustNewReplicaSet("mysql", v[0], v[1:]...)
		}
	}

	dlmAndDataRedisPool := redis.CreatePool(config.GetString("dlmAndDataRedisPool"))
	privateKeyRedisPool := redis.CreatePool(config.GetString("privateKeyRedisPool"))

	regionalV1PrivateClientSecret, e := base64.RawURLEncoding.DecodeString(config.GetString("regionalV1PrivateClientSecret"))
	panic.IfNotNil(e)

	var regionalV1PrivateClient private.V1Client
	if newPrivateV1Client != nil {
		regionalV1PrivateClient = newPrivateV1Client(env, scheme, nakedHost)
	}

	return &Resources{
		ServerCreatedOn:               t.NowUnixMillis(),
		BindAddress:                   bindAddress,
		Env:                           env,
		NakedHost:                     nakedHost,
		ClientHost:                    clientHost,
		AllHosts:                      allHosts,
		ClientScheme:                  scheme,
		Region:                        region,
		Version:                       config.GetString("version"),
		FileServerDir:                 config.GetString("fileServerDir"),
		ApiDocsRoute:                  strings.ToLower(config.GetString("apiDocsRoute")),
		ApiMDoRoute:                   strings.ToLower(config.GetString("apiMDoRoute")),
		ApiLogoutRoute:                strings.ToLower(config.GetString("apiLogoutRoute")),
		SessionCookieName:             config.GetString("sessionCookieName"),
		SessionStore:                  sessionStore,
		CachingEnabled:                config.GetBool("cachingEnabled"),
		MasterCacheKey:                config.GetString("masterCacheKey"),
		NameRegexMatchers:             nameRegexMatchers,
		DisplayNameRegexMatchers:      displayNameRegexMatchers,
		PwdRegexMatchers:              pwdRegexMatchers,
		NameMinRuneCount:              config.GetInt("nameMinRuneCount"),
		NameMaxRuneCount:              config.GetInt("nameMaxRuneCount"),
		DisplayNameMinRuneCount:       config.GetInt("displayNameMinRuneCount"),
		DisplayNameMaxRuneCount:       config.GetInt("displayNameMaxRuneCount"),
		PwdMinRuneCount:               config.GetInt("pwdMinRuneCount"),
		PwdMaxRuneCount:               config.GetInt("pwdMaxRuneCount"),
		MaxProcessEntityCount:         config.GetInt("maxProcessEntityCount"),
		CryptCodeLen:                  config.GetInt("cryptCodeLen"),
		SaltLen:                       config.GetInt("saltLen"),
		ScryptN:                       config.GetInt("scryptN"),
		ScryptR:                       config.GetInt("scryptR"),
		ScryptP:                       config.GetInt("scryptP"),
		ScryptKeyLen:                  config.GetInt("scryptKeyLen"),
		RegionalV1PrivateClientSecret: regionalV1PrivateClientSecret,
		RegionalV1PrivateClient:       regionalV1PrivateClient,
		MailClient:                    mailClient,
		AvatarClient:                  avatarClient,
		LogError:                      logError,
		LogStats:                      logStats,
		AccountDb:                     accountDb,
		PwdDb:                         pwdDb,
		TreeShards:                    treeShardDbs,
		DlmAndDataRedisPool:           dlmAndDataRedisPool,
		PrivateKeyRedisPool:           privateKeyRedisPool,
	}
}

// Collection of application static resources, in lcl and dev "onebox" environments all values must be set
// but in stg and prd environments central and regional endpoints are physically separated and so not all values are valid
// e.g. account and pwd dbs are only initialised on central service, whilst redis pool and tree shards are only initialised
// for regional endpoints.
type Resources struct {
	// server created on unix millisecs
	ServerCreatedOn int64
	// server address eg "127.0.0.1:80"
	BindAddress string
	// must be one of "lcl", "dev", "stg", "pro"
	Env cnst.Env
	// must be "project-trees.com"
	NakedHost string
	// must be one of "lcl.project-trees.com", "dev.project-trees.com", "stg.project-trees.com", "pro.project-trees.com"
	ClientHost string
	// must be one of collection of all API hosts e.g ["pro-api.project-trees.com", "pro-usw-api.project-trees.com", "pro-use-api.project-trees.com", "pro-euw-api.project-trees.com", "pro-asp-api.project-trees.com", "pro-asp-api.project-trees.com"]
	AllHosts []string
	// must be one of "https://", "http://"
	ClientScheme string
	// must be one of "central", "use", "usw", "euw", only considered in non lcl and dev envs
	Region cnst.Region
	// commit sha
	Version string
	// relative path from server executable to static file resource directory
	FileServerDir string
	// api docs path
	ApiDocsRoute string
	// api mget path
	ApiMDoRoute string
	// api logout path
	ApiLogoutRoute string
	// session cookie name
	SessionCookieName string
	// session cookie store
	SessionStore *sessions.CookieStore
	// indented json api docs
	ApiDocs []byte
	// is caching enabled
	CachingEnabled bool
	// incremental base64 value
	MasterCacheKey string
	// regexes that account names must match to be valid during account creation or name setting
	NameRegexMatchers []*regexp.Regexp
	// regexes that account names must match to be valid during account creation or name setting
	DisplayNameRegexMatchers []*regexp.Regexp
	// regexes that account pwds must match to be valid during account creation or pwd setting
	PwdRegexMatchers []*regexp.Regexp
	// minimum number of runes required for a valid account name
	NameMinRuneCount int
	// maximum number of runes required for a valid account name
	NameMaxRuneCount int
	// minimum number of runes required for a valid account displayName
	DisplayNameMinRuneCount int
	// maximum number of runes required for a valid account displayName
	DisplayNameMaxRuneCount int
	// minimum number of runes required for a valid account pwd
	PwdMinRuneCount int
	// maximum number of runes required for a valid account pwd
	PwdMaxRuneCount int
	// max number of entities that can be processed at once, also used for max limit value on queries
	MaxProcessEntityCount int
	// length of cryptographic codes, used in email links for validating email addresses and resetting pwds
	CryptCodeLen int
	// length of salts used for pwd hashing
	SaltLen int
	// scrypt N value
	ScryptN int
	// scrypt R value
	ScryptR int
	// scrypt P value
	ScryptP int
	// scrypt key length
	ScryptKeyLen int
	// regional v1 private client secret
	RegionalV1PrivateClientSecret []byte
	// regional v1 private client used by central endpoints
	RegionalV1PrivateClient private.V1Client
	// mail client for sending emails
	MailClient mail.Client
	// avatar client for storing avatar images
	AvatarClient avatar.Client
	// error logging function
	LogError func(error)
	// stats logging function
	LogStats func(status int, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo)
	// account sql connection
	AccountDb isql.ReplicaSet
	// pwd sql connection
	PwdDb isql.ReplicaSet
	// tree shard sql connections
	TreeShards map[int]isql.ReplicaSet
	// redis pool for caching layer
	DlmAndDataRedisPool iredis.Pool
	// redis pool for private request keys to check for replay attacks
	PrivateKeyRedisPool iredis.Pool
}
