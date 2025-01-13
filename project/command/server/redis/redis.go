package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"time"
)

var (
	Rdb *redis.Client
)

func InitPoolRdb() (err error) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:         RedisConAddre,
		DB:           RedisUserDb,
		Password:     RedisPwd,
		MinIdleConns: 5,
		PoolSize:     30,
		PoolTimeout:  30 * time.Second,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	return Rdb.Ping().Err()
}

type RdbOp struct {
}

func NewRdbOp() *RdbOp {
	return &RdbOp{}
}

func (r *RdbOp) Set() {}

func (r *RdbOp) Get() {}

func (r *RdbOp) Del() {}

func (r *RdbOp) Check() {
}

func (r *RdbOp) ReqVerify(user, token string) (err error) {
	getToken, err := Rdb.Get(user).Result()
	if err != nil {
		return
	}

	if getToken != token {
		err = errors.New("token验证失败")
		return
	}

	return
}
