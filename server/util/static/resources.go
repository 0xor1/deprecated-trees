package static

import (
	"bitbucket.org/0xor1/task/server/util/avatar"
	"bitbucket.org/0xor1/task/server/util/err"
	"bitbucket.org/0xor1/task/server/util/id"
	"bitbucket.org/0xor1/task/server/util/mail"
	"bitbucket.org/0xor1/task/server/util/private"
	"bitbucket.org/0xor1/task/server/util/queryinfo"
	"bitbucket.org/0xor1/task/server/util/redis"
	"bitbucket.org/0xor1/task/server/util/time"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

// pass in empty strings for no config file
func Config(configFile, configPath string, createPrivateV1Client func(map[string]string) private.V1Client) *Resources {
	//defaults set up for onebox local environment configuration i.e everything running on one machine
	// server address eg "127.0.0.1:8787"
	viper.SetDefault("serverAddress", "127.0.0.1:8787")
	// must be one of "lcl", "dev", "stg", "prd"
	viper.SetDefault("env", "lcl")
	// must be one of "lcl", "dev", "central", "use", "usw", "euw"
	viper.SetDefault("region", "lcl")
	// commit sha
	viper.SetDefault("version", "lcl")
	// api docs path
	viper.SetDefault("apiDocsRoute", "/api/docs")
	// session cookie name
	viper.SetDefault("sessionCookieName", "t")
	// cookie session domain
	viper.SetDefault("sessionDomain", "127.0.0.1")
	// session cookie store
	viper.SetDefault("sessionAuthKey64s", []string{
		"Va3ZMfhH4qSfolDHLU7oPal599DMcL93A80rV2KLM_om_HBFFUbodZKOHAGDYg4LCvjYKaicodNmwLXROKVgcA",
		"WK_2RgRx6vjfWVkpiwOCB1fvv1yklnltstBjYlQGfRsl6LyVV4mkt6UamUylmkwC8MEgb9bSGr1FYgM2Zk20Ug",
	})
	viper.SetDefault("sessionEncrKey32s", []string{
		"3ICuYRUelY-4Fhak0Iw0_5CW24bJvxFWM0jAA78IIp8",
		"u80sYkgbBav52fJXbENYhN3Iyof7WhuLHHMaS_rmUQw",
	})
	// incremental base64 value
	viper.SetDefault("masterCacheKey", "0")
	// regexes that account names must match to be valid during account creation or name setting
	viper.SetDefault("nameRegexMatchers", []string{})
	// regexes that account pwds must match to be valid during account creation or pwd setting
	viper.SetDefault("pwdRegexMatchers", []string{})
	// minimum number of runes required for a valid account name
	viper.SetDefault("nameMinRuneCount", 3)
	// maximum number of runes required for a valid account name
	viper.SetDefault("nameMaxRuneCount", 50)
	// minimum number of runes required for a valid account pwd
	viper.SetDefault("pwdMinRuneCount", 8)
	// maximum number of runes required for a valid account pwd
	viper.SetDefault("pwdMaxRuneCount", 200)
	// max number of entities that can be processed at once, also used for max limit value on queries
	viper.SetDefault("maxProcessEntityCount", 100)
	// length of cryptographic codes, used in email links for validating email addresses and resetting pwds
	viper.SetDefault("cryptCodeLen", 100)
	// length of salts used for pwd hashing
	viper.SetDefault("saltLen", 64)
	// scrypt N value
	viper.SetDefault("scryptN", 32768)
	// scrypt R value
	viper.SetDefault("scryptR", 8)
	// scrypt P value
	viper.SetDefault("scryptP", 1)
	// scrypt key length
	viper.SetDefault("scryptKeyLen", 32)
	// private client secret base64 encoded
	viper.SetDefault("regionalV1PrivateClientSecret", "bwIwGNgOdTWxCifGdL5BW5XhoWoctcTQyN3LLeSTo1nuDNebpKmlda2XaF66jOh1jaV7cvFRHScJrdyn8gSnMQ")
	// private client config
	viper.SetDefault("regionalV1PrivateClientConfig", map[string]string{
		"lcl": "http//127.0.0.1:8787",
	})
	// max avatar dimension
	viper.SetDefault("maxAvatarDim", 250)
	// local avatar storage directory, relative
	viper.SetDefault("lclAvatarDir", "avatar")
	// account primary sql connection
	viper.SetDefault("accountDbPrimary", "t_c_accounts:T@sk-@cc-0unt5@tcp(127.0.0.1:3307)/accounts?parseTime=true&loc=UTC&multiStatements=true")
	// account slave sql connections
	viper.SetDefault("accountDbSlaves", []string{})
	// pwd primary sql connection
	viper.SetDefault("pwdDbPrimary", "t_c_pwds:T@sk-Pwd5@tcp(127.0.0.1:3307)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	// account slave sql connections
	viper.SetDefault("pwdDbSlaves", []string{})
	// tree shard sql connections
	viper.SetDefault("treeShards", map[string][]string{
		"0": {"t_r_trees:T@sk-Tr335@tcp(127.0.0.1:3307)/trees?parseTime=true&loc=UTC&multiStatements=true"},
	})
	// redis pool for caching layer
	viper.SetDefault("dlmAndDataRedisPool", "127.0.0.1:6379")
	// redis pool for private request keys to check for replay attacks
	viper.SetDefault("privateKeyRedisPool", "127.0.0.1:6379")
	if configFile != "" && configPath != "" {
		viper.SetConfigName(configFile)
		viper.AddConfigPath(configPath)
		err.PanicIf(viper.ReadInConfig())
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	authKey64s := viper.GetStringSlice("sessionAuthKey64s")
	encrKey32s := viper.GetStringSlice("sessionEncrKey32s")
	sessionAuthEncrKeyPairs := make([][]byte, 0, len(authKey64s)*2)
	for i := range authKey64s {
		authBytes, e := base64.RawURLEncoding.DecodeString(authKey64s[i])
		err.PanicIf(e)
		if len(authBytes) != 64 {
			err.FmtPanic("sessionAuthBytes length is not 64")
		}
		encrBytes, e := base64.RawURLEncoding.DecodeString(encrKey32s[i])
		err.PanicIf(e)
		if len(encrBytes) != 32 {
			err.FmtPanic("sessionEncrBytes length is not 32")
		}
		sessionAuthEncrKeyPairs = append(sessionAuthEncrKeyPairs, authBytes, encrBytes)
	}
	sessionStore := sessions.NewCookieStore(sessionAuthEncrKeyPairs...)
	sessionStore.Options.MaxAge = 0
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = true
	sessionStore.Options.Domain = viper.GetString("sessionDomain")
	gob.Register(id.New()) //register Id type for sessionCookie

	var logError func(error)
	var logStats func(status int, method, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo)
	var avatarClient avatar.Client
	var mailClient mail.Client

	if viper.GetString("env") == "lcl" {
		//setup local environment interfaces
		logError = func(err error) {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
		logStats = func(status int, method, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo) {
			fmt.Println(status, fmt.Sprintf("%dms", time.NowUnixMillis()-reqStartUnixMillis), method, path)
			//often too much info when running locally, makes too much noise, but feel free to uncomment when necessary
			//queryInfosBytes, _ := json.Marshal(queryInfos)
			//fmt.Println(string(queryInfosBytes))
		}
		avatarClient = avatar.NewLocalClient(viper.GetString("lclAvatarDir"), uint(viper.GetInt("maxAvatarDim")))
		mailClient = mail.NewLocalClient()
	} else {
		//setup deployed environment interfaces
		//TODO setup aws s3 avatarStore storage
		//TODO setup sparkpost/mailgun/somthing mailClient client
		//TODO setup datadog stats and error logging
		panic(err.NotImplemented)
	}

	nameRegexMatchers := make([]*regexp.Regexp, 0, len(viper.GetStringSlice("nameRegexMatchers")))
	for _, str := range viper.GetStringSlice("nameRegexMatchers") {
		nameRegexMatchers = append(nameRegexMatchers, regexp.MustCompile(str))
	}
	pwdRegexMatchers := make([]*regexp.Regexp, 0, len(viper.GetStringSlice("pwdRegexMatchers")))
	for _, str := range viper.GetStringSlice("pwdRegexMatchers") {
		nameRegexMatchers = append(pwdRegexMatchers, regexp.MustCompile(str))
	}

	var accountDb isql.ReplicaSet
	if viper.GetString("accountDbPrimary") != "" {
		accountDb = isql.NewReplicaSet("mysql", viper.GetString("accountDbPrimary"), viper.GetStringSlice("accountDbSlaves"))
	}

	var pwdDb isql.ReplicaSet
	if viper.GetString("pwdDbPrimary") != "" {
		pwdDb = isql.NewReplicaSet("mysql", viper.GetString("pwdDbPrimary"), viper.GetStringSlice("pwdDbSlaves"))
	}

	treeShardDbs := map[int]isql.ReplicaSet{}
	treeShards := viper.GetStringMapStringSlice("treeShards")
	if treeShards != nil {
		for k, v := range treeShards {
			shardId, e := strconv.ParseInt(k, 10, 0)
			err.PanicIf(e)
			treeShardDbs[int(shardId)] = isql.NewReplicaSet("mysql", v[0], v[1:])
		}
	}

	var dlmAndDataRedisPool iredis.Pool
	if viper.GetString("dlmAndDataRedisPool") != "" {
		dlmAndDataRedisPool = redis.CreatePool(viper.GetString("dlmAndDataRedisPool"), logError)
	}

	var privateKeyRedisPool iredis.Pool
	if viper.GetString("privateKeyRedisPool") != "" {
		privateKeyRedisPool = redis.CreatePool(viper.GetString("privateKeyRedisPool"), logError)
	}

	return &Resources{
		ServerAddress:           viper.GetString("serverAddress"),
		Env:                     viper.GetString("env"),
		Region:                  viper.GetString("region"),
		Version:                 viper.GetString("version"),
		ApiDocsRoute:            strings.ToLower(viper.GetString("apiDocsRoute")),
		SessionCookieName:       viper.GetString("sessionCookieName"),
		SessionStore:            sessionStore,
		MasterCacheKey:          viper.GetString("masterCacheKey"),
		NameRegexMatchers:       nameRegexMatchers,
		PwdRegexMatchers:        pwdRegexMatchers,
		NameMinRuneCount:        viper.GetInt("nameMinRuneCount"),
		NameMaxRuneCount:        viper.GetInt("nameMaxRuneCount"),
		PwdMinRuneCount:         viper.GetInt("pwdMinRuneCount"),
		PwdMaxRuneCount:         viper.GetInt("pwdMaxRuneCount"),
		MaxProcessEntityCount:   viper.GetInt("maxProcessEntityCount"),
		CryptCodeLen:            viper.GetInt("cryptCodeLen"),
		SaltLen:                 viper.GetInt("saltLen"),
		ScryptN:                 viper.GetInt("scryptN"),
		ScryptR:                 viper.GetInt("scryptR"),
		ScryptP:                 viper.GetInt("scryptP"),
		ScryptKeyLen:            viper.GetInt("scryptKeyLen"),
		RegionalV1PrivateClient: createPrivateV1Client(viper.GetStringMapString("regionalV1PrivateClientConfig")),
		MailClient:              mailClient,
		AvatarClient:            avatarClient,
		LogError:                logError,
		LogStats:                logStats,
		AccountDb:               accountDb,
		PwdDb:                   pwdDb,
		TreeShards:              treeShardDbs,
		DlmAndDataRedisPool:     dlmAndDataRedisPool,
		PrivateKeyRedisPool:     privateKeyRedisPool,
	}
}

// Collection of application static resources, in lcl and dev "onebox" environments all values must be set
// but in stg and prd environments central and regional endpoints are physically separated and so not all values are valid
// e.g. account and pwd dbs are only initialised on central service, whilst redis pool and tree shards are only initialised
// for regional endpoints.
type Resources struct {
	// server address eg "127.0.0.1:8787"
	ServerAddress string
	// must be one of "lcl", "dev", "stg", "prd"
	Env string
	// must be one of "lcl", "dev", "central", "use", "usw", "euw"
	Region string
	// commit sha
	Version string
	// api docs path
	ApiDocsRoute string
	// session cookie name
	SessionCookieName string
	// session cookie store
	SessionStore *sessions.CookieStore
	// indented json api docs
	ApiDocs []byte
	// incremental base64 value
	MasterCacheKey string
	// regexes that account names must match to be valid during account creation or name setting
	NameRegexMatchers []*regexp.Regexp
	// regexes that account pwds must match to be valid during account creation or pwd setting
	PwdRegexMatchers []*regexp.Regexp
	// minimum number of runes required for a valid account name
	NameMinRuneCount int
	// maximum number of runes required for a valid account name
	NameMaxRuneCount int
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
	LogStats func(status int, method, path string, reqStartUnixMillis int64, queryInfos []*queryinfo.QueryInfo)
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
