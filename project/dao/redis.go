package dao

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/ingoxx/go-gin/project/config"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// 把关于redis的处理都放这里了,想不到好的位置放
var (
	RdPool *redis.Client
	Rds    *RedisDb
)

// 初始化redis连接池

func InitPoolRds() (err error) {
	RdPool = redis.NewClient(&redis.Options{
		Addr:         config.RedisConAddr,
		DB:           config.RedisUserDb,
		Password:     config.RedisPwd,
		MinIdleConns: 5,
		PoolSize:     30,
		PoolTimeout:  30 * time.Second,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	return RdPool.Ping().Err()
}

type Md struct {
	Count uint
	Rtime uint64
	Wait  uint64
}

type RedisDb struct {
	pool *redis.Client
	md   map[string]Md
	lock *sync.Mutex
}

func NewRedisDb(pool *redis.Client, md map[string]Md) *RedisDb {
	return &RedisDb{
		pool: pool,
		md:   md,
		lock: &sync.Mutex{},
	}
}

// ForbiddenLogin 60s内只要累计超过三次登陆失败就限制登陆
func (r *RedisDb) ForbiddenLogin(key string) (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	val, err := r.HGetAllKey(key)
	if err != nil {
		return
	}

	if len(val) > 0 {
		date, errN := strconv.Atoi(val["date"])
		if errN != nil {
			return errN
		}

		num, errN := strconv.Atoi(val["num"])
		if err != nil {
			return errN
		}

		now := time.Now().Unix()
		s := int(now) - date

		if s > 60 {
			if err = r.HMSetKey(key, map[string]interface{}{"date": 1, "num": 1}); err != nil {
				return
			}

			return
		}

		if (s <= 60) && num >= 3 {
			return errors.New("你已被限制登陆, 请1分钟后再试")
		}
	}

	return
}
func (r *RedisDb) SetData(key string, val []byte) (err error) {
	_, err = r.pool.Set(key, val, 0).Result()
	if err != nil {
		return
	}

	return

}

func (r *RedisDb) GetData(key string) (data []byte, err error) {
	d := r.pool.Get(key).Val()
	if d == "" {
		return
	}

	data = []byte(d)

	return
}

func (r *RedisDb) HMSetKey(key string, setData map[string]interface{}) (err error) {
	if _, err = r.pool.HMSet(key, setData).Result(); err != nil {
		return
	}

	return
}

func (r *RedisDb) HGetAllKey(key string) (data map[string]string, err error) {
	res := r.pool.HGetAll(key)
	if res.Err() != nil {
		return
	}

	return res.Val(), nil
}

func (r *RedisDb) RecordLoginFailedNum(key string) (err error) {
	val, err := r.HGetAllKey(key)
	if err != nil {
		return
	}

	if len(val) == 0 {
		if err = r.HMSetKey(key, map[string]interface{}{"date": time.Now().Unix(), "num": 1}); err != nil {
			return
		}

		return
	}

	num, err := strconv.Atoi(val["num"])
	if err != nil {
		return
	}

	num += 1

	date, err := strconv.Atoi(val["date"])
	if err != nil {
		return
	}

	if date == 1 {
		date = int(time.Now().Unix())
	}

	if err = r.HMSetKey(key, map[string]interface{}{"date": date, "num": num}); err != nil {
		return
	}

	return
}

func (r *RedisDb) RequestVerify(user, token string) (err error) {
	getToken, err := r.pool.Get(user).Result()
	if err != nil {
		return
	}

	if getToken != token {
		err = errors.New("token已过期, 请重新登录")
		return
	}

	return
}

func (r *RedisDb) RegisterUserInfo(user string) (t string, err error) {
	token := r.HashToken(user)
	_, err = r.pool.Set(user, token, time.Second*259200).Result()
	if err != nil {
		return
	}
	t = token
	return
}

func (r *RedisDb) SaveGaKey(user, key string) (err error) {
	_, err = r.pool.Set(user+"-ga", key, time.Second*1576800000).Result()
	if err != nil {
		return
	}
	return
}

func (r *RedisDb) GetGaKey(user string) (key string, err error) {
	key, err = r.pool.Get(user + "-ga").Result()
	if err != nil {
		return
	}
	return
}

func (r *RedisDb) ClearToken(user string) (err error) {
	_, err = r.pool.Del(user).Result()
	if err != nil {
		return
	}
	return
}

// ReqFrequencyLimit 简单的限流功能，每秒只能接收5次访问，超过5次返回502并需要等待10秒后才能访问
func (r *RedisDb) ReqFrequencyLimit(host string) (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	mdd := Md{}
	ut := uint64(time.Now().Unix())

	vd, ok := r.md["visit_"+host]
	if !ok {
		mdd.Count = 1
		mdd.Rtime = ut + 1
		r.md["visit_"+host] = mdd
		vd = r.md["visit_"+host]
	} else if vd.Count <= 1 {
		vd.Rtime = ut + 1
		r.md["visit_"+host] = vd
	}

	if vd.Wait > ut {
		err = errors.New("频繁访问已被限制")
		return
	}

	if vd.Rtime >= ut && vd.Count > uint(config.Frequency) {
		vd.Count = 1
		vd.Wait = ut + 10
		r.md["visit_"+host] = vd
		err = errors.New("频繁访问已被限制")
		return
	}

	vd.Count += 1

	if ut > vd.Rtime {
		vd.Count = 1
		vd.Wait = 0
	}

	r.md["visit_"+host] = vd

	return
}

func (r *RedisDb) HashToken(user string) string {
	dateString := strconv.Itoa(int(time.Now().UnixNano()))
	hash := sha256.New()
	hash.Write([]byte(user + config.Secret + dateString))
	nnd := hex.EncodeToString(hash.Sum(nil))
	return nnd
}

func (r *RedisDb) GetProcessStatus() (sm map[string]string, err error) {
	var data = make(map[string]string)

	running, err := r.pool.HGet("prcessstatus", "running").Result()
	if err != nil {
		r.pool.HGet("prcessstatus", "running")
		return
	}

	finished, err := r.pool.HGet("prcessstatus", "finished").Result()
	if err != nil {
		return
	}

	failed, err := r.pool.HGet("prcessstatus", "failed").Result()
	if err != nil {
		return
	}

	data["running"] = running
	data["finished"] = finished
	data["failed"] = failed

	sm = data

	return
}

func (r *RedisDb) GetServerCpuLoadData(key string) ([]string, error) {
	values, err := r.pool.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	return values, nil
}
