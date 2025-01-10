package utils

import (
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/api"
	"github.com/Lxb921006/Gin-bms/project/command/client"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/Lxb921006/Gin-bms/project/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProgramAsyncRunCelery struct {
	Works chan api.CeleryInterface
}

func NewProgramAsyncRunCelery() *ProgramAsyncRunCelery {
	var aprm model.AssetsProgramUpdateRecordModel
	var dataModel = make(map[string]interface{})

	c := &ProgramAsyncRunCelery{
		Works: make(chan api.CeleryInterface),
	}

	go func() {
		for w := range c.Works {
			data, err := w.Data()
			if err != nil {
				logger.Error("获取grpc参数失败, errMsg: ", err)
				return
			}

			dataModel["uuid"] = data["uuid"].(string)
			dataModel["status"] = 400

			conn, err := grpc.NewClient(fmt.Sprintf("%s:12306", data["ip"].(string)), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				if err = aprm.Update(dataModel); err != nil {
					logger.Error("更新失败, errMsg: ", err)
				}
				logger.Error("连接grpc失败, errMsg: ", err)
				return
			}

			cn := client.NewGrpcClient(data["update_name"].(string), data["uuid"].(string), nil, conn)
			go func() {
				if err = cn.Send(); err != nil {
					if err = aprm.Update(dataModel); err != nil {
						logger.Error("更新失败, errMsg: ", err)
					}
					logger.Error("grpc发送数据失败, errMsg: ", err)
				}
			}()
		}

	}()

	return c
}

func (c *ProgramAsyncRunCelery) Task(task api.CeleryInterface) {
	c.Works <- task
}

func (c *ProgramAsyncRunCelery) Close() {
	close(c.Works)
}
