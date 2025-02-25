package main

import (
	"github.com/ingoxx/go-gin/project/config"
	"github.com/ingoxx/go-gin/project/logger"
	"github.com/ingoxx/go-gin/project/migrate"
	"log"

	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/route/root"
)

func main() {
	logger.SetLogFile(config.LoggerFile)
	logger.SetLogLevel(logger.ErrorLevel)

	//初始化mysql
	err := dao.InitPoolMysql()
	if err != nil {
		log.Fatalf(err.Error())
	}

	//初始化数据库表
	err = migrate.InitTable()
	if err != nil {
		log.Fatalf(err.Error())
	}

	//初始化redis连接池
	err = dao.InitPoolRds()
	if err != nil {
		log.Fatalf(err.Error())
	}
	if dao.RdPool == nil {
		log.Fatalf(dao.ErrorRedisConnectFailed.Error())
	}
	dao.Rds = dao.NewRedisDb(dao.RdPool, map[string]dao.Md{})

	//初始化gin并启动
	t := root.SetupRouter()
	err = t.ListenAndServe()
	if err != nil {
		log.Fatalf(err.Error())
	}
}
