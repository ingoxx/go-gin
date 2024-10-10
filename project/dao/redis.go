package dao

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/Lxb921006/Gin-bms/project/config"
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
		Addr:         config.RedisConAddre,
		DB:           config.RedisUserDb,
		Password:     config.RedisPwd,
		MinIdleConns: 5,
		PoolSize:     30,
		PoolTimeout:  30 * time.Second,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	_, err = RdPool.Ping().Result()
	return
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

func (r *RedisDb) RquestVerify(user, token string) (err error) {
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

// 很简单很简单的限流功能，每秒只能接收20次访问，超过5次返回502并需要等待10秒后才能访问
func (r *RedisDb) Visitlimit(host string) (err error) {
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

	if vd.Rtime >= ut && vd.Count > config.Frequency {
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
	secret := "K0ka03kadk0kdko4951kadKMBNQPZLAJGHWQ"
	date := time.Now().UnixNano()
	dateString := strconv.Itoa(int(date))
	data := user + secret + dateString
	hash := sha256.New()
	hash.Write([]byte(data))
	nd := hash.Sum(nil)
	nnd := hex.EncodeToString(nd)
	return nnd
}

func (r *RedisDb) GetProcessStatus() (sm map[string]string, err error) {
	var data = make(map[string]string, 0)

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
