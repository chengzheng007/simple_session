package store

import (
	"encoding/json"
	"errors"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"sync"
)

type Store struct {
	mutex sync.RWMutex
	sid   string
	// value    map[interface{}]interface{}
	value    map[string]interface{}
	lifeTime int64
}

var (
	maxPoolSize     int   = 1
	sessionLifeTime int64 = 86400
	pool            *redis.Pool
)

// initilize redis pool
// ip:port,maxIdleNum,pwd,redisDbNum
func InitPool(maxLifeTime int64, cfgStr string) error {
	sessionLifeTime = maxLifeTime
	connCfg := strings.Split(cfgStr, ",")
	if len(connCfg) < 1 {
		return errors.New("Invalid pool config")
	}
	var (
		poolSize int
		pwd      string
		dbNum    int
		err      error
	)
	if len(connCfg) >= 2 {
		poolSize, err = strconv.Atoi(connCfg[1])
		if poolSize <= 0 || err != nil {
			poolSize = maxPoolSize
		}
	} else {
		poolSize = maxPoolSize
	}

	if len(connCfg) >= 3 {
		pwd = connCfg[2]
	}

	if len(connCfg) >= 4 {
		dbNum, err = strconv.Atoi(connCfg[3])
		if err != nil {
			dbNum = 0
		}
	} else {
		dbNum = 0
	}

	pool = redis.NewPool(
		func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", connCfg[0])
			if err != nil {
				return nil, err
			}

			if pwd != "" {
				if _, err = conn.Do("AUTH", pwd); err != nil {
					conn.Close()
					return nil, err
				}
			}

			if _, err = conn.Do("SELECT", dbNum); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		}, poolSize)

	return pool.Get().Err()
}

func (this *Store) Set(key string, val interface{}) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.value[key] = val
}

func (this *Store) Get(key string) interface{} {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if v, ok := this.value[key]; ok {
		return v
	}
	return nil
}

func (this *Store) GetAll() map[string]interface{} {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.value
}

func (this *Store) Del(key string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.value, key)
}

// 持久化：重新设定过期时间
func (this *Store) Persistence() error {
	data, err := json.Marshal(this.value)
	if err != nil {
		return err
	}
	conn := pool.Get()
	defer conn.Close()
	_, err = conn.Do("SETEX", this.sid, this.lifeTime, string(data))
	return err
}

func (this *Store) GC() {

}

// 读取得到存储对象
func SessionRead(sid string) (*Store, error) {
	conn := pool.Get()
	defer conn.Close()
	var err error
	var kv map[string]interface{}
	val, _ := redis.String(conn.Do("GET", sid))

	if len(val) == 0 {
		kv = make(map[string]interface{})
	} else {
		// kv, err = serialize.DecodeGob([]byte(val))
		err = json.Unmarshal([]byte(val), &kv)
		if err != nil {
			kv = make(map[string]interface{})
		}
	}

	StoreObj := &Store{sid: sid, value: kv, lifeTime: sessionLifeTime}
	return StoreObj, err
}

func SessionExist(sid string) bool {
	conn := pool.Get()
	defer conn.Close()

	existed, err := redis.Int(conn.Do("EXISTS", sid))
	if err != nil || existed <= 0 {
		return false
	}

	return true
}

func SessionDestroy(sid string) error {
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", sid)
	return err
}
