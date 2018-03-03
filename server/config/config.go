package config

import (
	centralAccount "bitbucket.org/0xor1/task/server/central/api/v1/account"
	"bitbucket.org/0xor1/task/server/regional/api/v1/account"
	"bitbucket.org/0xor1/task/server/regional/api/v1/private"
	"bitbucket.org/0xor1/task/server/regional/api/v1/project"
	"bitbucket.org/0xor1/task/server/regional/api/v1/task"
	. "bitbucket.org/0xor1/task/server/util"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0xor1/iredis"
	"github.com/0xor1/isql"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// pass in empty strings for no config file
func Config(configFile, configPath string) *StaticResources {
	sr := &StaticResources{}
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
	// cookie session name
	viper.SetDefault("sessionName", "t")
	// cookie session domain
	viper.SetDefault("sessionDomain", "127.0.0.1")
	// session cookie store
	viper.SetDefault("sessionAuthKey64s", []string{
		"2Kt39ndjI2praE8+rTBDb9OwEgR/JXXf8hO1Tl/mjx4SBpvugYSYRYkx3zBn+ofsRYFFr42Ap07NavrBjv1tzQ==",
		"co1AGV8VPSMOzjcokXEaxtKmKsi/BUXsidpaNHRyGyx0PwbnwNOMs0FzUWxhg+SO0AcqRATVtMWlhvGyLjtOIg==",
	})
	viper.SetDefault("sessionEncrKey32s", []string{
		"HQSpwNjT8Ra4RQkaZwYPPXMnF+sHHvJbZ0H7O+hQa/A=",
		"1V0QdFGKMGIFb4N5ss9f4q8c5QW0lZ5+yvA93aoJYmY=",
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
	viper.SetDefault("regionalV1PrivateClientSecret", "d8RBy1UvIT/u3bhEZlrZAHrKusB6c3UNFqz2tl3pT56ENN3f+kFjGxY2BvqTNM8pngMiuwBBMeuQRjl1Gcazlg==")
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
	viper.SetDefault("redisPool", "127.0.0.1:6379")
	if configFile != "" && configPath != "" {
		viper.SetConfigName(configFile)
		viper.AddConfigPath(configPath)
		PanicIf(viper.ReadInConfig())
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	authKey64s := viper.GetStringSlice("sessionAuthKey64s")
	encrKey32s := viper.GetStringSlice("sessionEncrKey32s")
	sessionAuthEncrKeyPairs := make([][]byte, 0, len(authKey64s)*2)
	for i := range authKey64s {
		authBytes, err := base64.StdEncoding.DecodeString(authKey64s[i])
		PanicIf(err)
		if len(authBytes) != 64 {
			FmtPanic("sessionAuthBytes length is not 64")
		}
		encrBytes, err := base64.StdEncoding.DecodeString(encrKey32s[i])
		PanicIf(err)
		if len(encrBytes) != 32 {
			FmtPanic("sessionEncrBytes length is not 32")
		}
		sessionAuthEncrKeyPairs = append(sessionAuthEncrKeyPairs, authBytes, encrBytes)
	}
	sessionStore := sessions.NewCookieStore(sessionAuthEncrKeyPairs...)
	sessionStore.Options.MaxAge = 0
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = true
	sessionStore.Options.Domain = viper.GetString("sessionDomain")
	gob.Register(NewId()) //register Id type for sessionCookie

	routes := map[string]*Endpoint{}

	var log func(error)
	var avatarClient AvatarClient
	var mailClient MailClient
	endpoints := make([]*Endpoint, 0, len(centralAccount.Endpoints)+len(private.Endpoints)+len(account.Endpoints)+len(project.Endpoints)+len(task.Endpoints))

	if viper.GetString("env") == "lcl" {
		//setup all routes on lcl environment
		endpoints = append(endpoints, centralAccount.Endpoints...)
		endpoints = append(endpoints, private.Endpoints...)
		endpoints = append(endpoints, account.Endpoints...)
		endpoints = append(endpoints, project.Endpoints...)
		endpoints = append(endpoints, task.Endpoints...)

		//setup local environment interfaces
		log = func(err error) {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
		avatarClient = NewLocalAvatarStore(viper.GetString("lclAvatarDir"), uint(viper.GetInt("maxAvatarDim")))
		mailClient = NewLocalMailClient()
	} else {
		//setup deployed environment interfaces
		//TODO setup aws s3 avatarStore storage
		//TODO setup sparkpost/mailgun/somthing mailClient client
		//TODO setup datadog stats and error logging
		NotImplementedErr.Panic()
	}

	for _, ep := range endpoints {
		ep.ValidateEndpoint()
		lowerPath := strings.ToLower(ep.Path)
		if _, exists := routes[lowerPath]; exists {
			FmtPanic("duplicate endpoint path %q", lowerPath)
		}
		routes[lowerPath] = ep
		ep.StaticResources = sr
	}
	routeDocs := make([]interface{}, 0, len(routes))
	for _, ep := range routes {
		routeDocs = append(routeDocs, ep.GetEndpointDocumentation())
	}
	apiDocs, err := json.MarshalIndent(routeDocs, "", "    ")
	PanicIf(err)

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
			shardId, err := strconv.ParseInt(k, 10, 0)
			PanicIf(err)
			treeShardDbs[int(shardId)] = isql.NewReplicaSet("mysql", v[0], v[1:])
		}
	}

	var redisPool iredis.Pool
	if viper.GetString("redisPool") != "" {
		redisPool = &redis.Pool{
			MaxIdle:     300,
			IdleTimeout: time.Minute,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial("tcp", viper.GetString("redisPool"), redis.DialDatabase(0), redis.DialConnectTimeout(1*time.Second), redis.DialReadTimeout(2*time.Second), redis.DialWriteTimeout(2*time.Second))
				// Log any Redis connection error on stdout
				if err != nil {
					log(err)
				}

				return conn, err
			},
			TestOnBorrow: func(c redis.Conn, ti time.Time) error {
				if time.Since(ti) < time.Minute {
					return nil
				}
				return errors.New("Redis connection timed out")
			},
		}
	}

	sr.ServerAddress = viper.GetString("serverAddress")
	sr.Env = viper.GetString("env")
	sr.Region = viper.GetString("region")
	sr.Version = viper.GetString("version")
	sr.ApiDocsRoute = strings.ToLower(viper.GetString("apiDocsRoute"))
	sr.SessionName = viper.GetString("sessionName")
	sr.SessionStore = sessionStore
	sr.Routes = routes
	sr.ApiDocs = apiDocs
	sr.MasterCacheKey = viper.GetString("masterCacheKey")
	sr.NameRegexMatchers = nameRegexMatchers
	sr.PwdRegexMatchers = pwdRegexMatchers
	sr.NameMinRuneCount = viper.GetInt("nameMinRuneCount")
	sr.NameMaxRuneCount = viper.GetInt("nameMaxRuneCount")
	sr.PwdMinRuneCount = viper.GetInt("pwdMinRuneCount")
	sr.PwdMaxRuneCount = viper.GetInt("pwdMaxRuneCount")
	sr.MaxProcessEntityCount = viper.GetInt("maxProcessEntityCount")
	sr.CryptCodeLen = viper.GetInt("cryptCodeLen")
	sr.SaltLen = viper.GetInt("saltLen")
	sr.ScryptR = viper.GetInt("scryptR")
	sr.ScryptP = viper.GetInt("scryptP")
	sr.ScryptKeyLen = viper.GetInt("scryptKeyLen")
	sr.RegionalV1PrivateClient = private.NewClient(viper.GetStringMapString("regionalV1PrivateClientConfig"))
	sr.MailClient = mailClient
	sr.AvatarClient = avatarClient
	sr.Log = log
	sr.AccountDb = accountDb
	sr.PwdDb = pwdDb
	sr.TreeShards = treeShardDbs
	sr.RedisPool = redisPool
	return sr
}
