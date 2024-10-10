package dao

import "errors"

var (
	ErrorMysqlConnectFailed = errors.New("数据库连接失败")
	ErrorRedisConnectFailed = errors.New("redis连接失败")
)
