package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/orbitspot/lib-metrics/pkg/log"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"
	"time"
	_ "github.com/joho/godotenv/autoload"
)

const (
	maxRedisConnections = 30
)

// CacheRepo cache repository structure
type redisCache struct {
	cache      *redis.Client
	host       string
	port       string
	database   int
	expiration int
	name       string
}

// Redis godoc
var (
	appName      = os.Getenv("APP_NAME")
	redisDefault = &redisCache{}
	R            = make(map[string]*redisCache)
)

// Init godoc
func Init() {

	// Init multiple Redis connections
	for i := 0; i < maxRedisConnections; i++ {
		redisConnEnv := os.Getenv(fmt.Sprintf("REDIS_CONNECTION_%v", i))
		if redisConnEnv != "" {
			initConnection(redisConnEnv, i)
		} else {
			break
		}
	}

	return
}

// Initiate a Redis connection
// Any fail must break (panic) application startup!
//
// Environment variable syntax: host, port, db, expiration (seconds), name (for connection)
//
// Example:
//	REDIS_CONNECTION_0=localhost,6379,15,  360,default  # Default connection, 10 minutes TTL
//	REDIS_CONNECTION_1=localhost,6379,05,    0,name1    # Infinite TTL
//	REDIS_CONNECTION_2=localhost,6379,12,   10,name2    # 10 seconds TTL
//	REDIS_CONNECTION_3=localhost,6379,07,86400,name3    # 1 day TTL
//
func initConnection(conn string, n int) {

	envMsg := fmt.Sprintf("[app: %s, connection: [number: %v, details: %s]", appName, n, conn)

	// Get default redis connection
	redisArgs := strings.Split(conn, ",")
	if len(redisArgs) < 5 {
		_ = log.ErrorNew("Redis WAS NOT STARTED because some environment variable is missing! %s", envMsg)
		panic("Redis WAS NOT STARTED!")
		return
	}

	// Init Redis default database connection
	var err error
	redisNewConn := redisCache{}
	redisNewConn.host = strings.TrimSpace(redisArgs[0])
	redisNewConn.port = strings.TrimSpace(redisArgs[1])
	redisDatabase    := strings.TrimSpace(redisArgs[2])
	redisExpiration  := strings.TrimSpace(redisArgs[3])
	redisNewConn.name = strings.TrimSpace(redisArgs[4])
	if redisNewConn.host == "" || redisNewConn.port == "" || redisNewConn.name == "" {
		log.Fatalf(err,"Redis WAS NOT STARTED because CONNECTION STRING is invalid! %s", envMsg)
		return
	}
	if redisNewConn.database, err = strconv.Atoi(redisDatabase); err != nil {
		log.Fatalf(err,"Redis WAS NOT STARTED because DATABASE is invalid! %s", envMsg)
		return
	}
	if redisNewConn.expiration, err = strconv.Atoi(redisExpiration); err != nil {
		log.Fatalf(err,"Redis WAS NOT STARTED because EXPIRATION is invalid! %s", envMsg)
		return
	}

	// Create a new cache instance
	redisNewConn.cache = redis.NewClient(&redis.Options{
		Addr: redisNewConn.host + ":" + redisNewConn.port,
		DB:   redisNewConn.database,
	})

	// Setup Default Redis Connection
	if n == 0 {
		*redisDefault = redisNewConn
	}

	// Check if Redis connection was started
	if err := redisNewConn.Ping(); err != nil {
		log.Fatalf(err,"Redis WAS NOT STARTED due an error! %s", envMsg)
	}

	R[redisNewConn.name] = &redisNewConn

	log.Info("Redis Connection '%s' STARTED! [%s]", redisNewConn.name, envMsg)
	return
}


//////////////////////////////////////
// DEFAULT IMPLEMENTATION FUNCTIONS //
//////////////////////////////////////

// Instance godoc
func (r *redisCache) Instance() *redis.Client {
	return r.cache
}

// Ping godoc
func (r *redisCache) Ping() error {
	if _, err := r.cache.Ping().Result(); err != nil {
		return err
	}
	return nil
}

// Set with specific TTL godoc
func (r *redisCache) SetT(key string, object interface{}, expiration int) error {
	jsonOject, err := json.Marshal(object)
	if err != nil {
		log.Debug("Error while trying to SET cache with object %v - error: %v", jsonOject, err.Error())
		return err
	}
	r.cache.Set(key, jsonOject, time.Duration(expiration) * time.Second)
	return nil
}

// Set with default TTL godoc
func (r *redisCache) Set(key string, object any) error {
	return r.SetT(key, object, r.expiration)
}

// Get godoc
func (r *redisCache) Get(key string, object interface{}) (error, bool) {
	cache := r.cache.Get(key).Val()
	if len(cache) > 0 {
		if err := json.Unmarshal([]byte(cache), &object); err != nil {
			log.Debug("Error while trying to GET cache value with key %s - error: %s", key, err.Error())
			return err, false
		}
		return nil, true
	}
	return nil, false
}

// Del godoc
func (r *redisCache) Del(key string) error{
	// Delete by Pattern
	if strings.Contains(key, "*") {
		var cursor uint64
		iter := r.cache.Scan(cursor, key, 0).Iterator()
		for iter.Next() {
			log.Debug("Deleting key [%s]", iter.Val())
			r.cache.Del(iter.Val())
		}
		if err := iter.Err(); err != nil {
			log.Debug("Error deleting pattern %s", key)
			return err
		}
	} else {
		// Delete by Fixed Key
		log.Debug("Deleting key [%s]", key)
		r.cache.Del(key)
	}
	return nil
}


///////////////////////////////////
// DEFAULT ABSTRACTION FUNCTIONS //
///////////////////////////////////

// Instance default godoc
func Instance() *redis.Client {
	return redisDefault.Instance()
}

// Ping default godoc
func Ping() error {
	return redisDefault.Ping()
}

// Set TTL default godoc
func Set(key string, object any) error {
	return redisDefault.SetT(key, object, redisDefault.expiration)
}

// Set default godoc
func SetT(key string, object interface{}, expiration int) error {
	return redisDefault.SetT(key, object, expiration)
}

// Get default godoc
func Get(key string, object interface{}) (error, bool) {
	return redisDefault.Get(key, object)
}

// Del default godoc
func Del(key string) error {
	return redisDefault.Del(key)
}


//////////////////////
// HELPER FUNCTIONS //
//////////////////////

// PreparerKey godoc
func PrepareKey(cacheName string, object interface{}, useMD5HashInObject bool) (string, error) {
	var key = ""
	if object != nil && len(cacheName) > 0  {
		jsonKey, err := json.Marshal(object)
		if err != nil {
			return key, err
		}
		if useMD5HashInObject {
			key = fmt.Sprintf("%s:%s:%s", appName, cacheName, GetMD5Hash(string(jsonKey)))
		} else {
			key = fmt.Sprintf("%s:%s:%s", appName, cacheName, string(jsonKey))
		}
		return key, err
	}
	if object != nil && len(cacheName) <= 0  {
		jsonKey, err := json.Marshal(object)
		if err != nil {
			return key, err
		}
		if useMD5HashInObject {
			key = fmt.Sprintf("%s:%s", appName, GetMD5Hash(string(jsonKey)))
		} else {
			key = fmt.Sprintf("%s:%s", appName, string(jsonKey))
		}
	}
	if object == nil && len(cacheName) > 0  {
		key = fmt.Sprintf("%s:%s", appName, cacheName)
	}
	return key, nil
}

// GetMD5Hash godoc
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
