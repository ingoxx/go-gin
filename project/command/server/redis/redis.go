package redis

import (
	"github.com/Lxb921006/Gin-bms/project/config"
	"github.com/go-redis/redis"
	"time"
)

var (
	Rdb *redis.Client
)

func InitPoolRdb() (err error) {
	Rdb = redis.NewClient(&redis.Options{
		Addr:         config.RedisConAddre,
		DB:           config.RedisUserDb,
		Password:     config.RedisPwd,
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

func (r *RdbOp) Check() {}
