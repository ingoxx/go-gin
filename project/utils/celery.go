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
	c := &ProgramAsyncRunCelery{
		Works: make(chan api.CeleryInterface),
	}

	go func() {
		var apr model.AssetsProgramUpdateRecordModel
		var dataModel = make(map[string]interface{})
		for w := range c.Works {
			data, err := w.Data()
			if err != nil {
				logger.Error(fmt.Sprintf("获取grpc参数失败, errMsg: %s", err.Error()))
				continue
			}

			dataModel["uuid"] = data["uuid"].(string)
			dataModel["status"] = 400
			dataModel["ip"] = data["ip"].(string)

			conn, err := grpc.NewClient(fmt.Sprintf("%s:12306", data["ip"].(string)), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				dataModel["status"] = 300
				if err := apr.Update(dataModel); err != nil {
					logger.Error("uuid: %v, 更新失败, errMsg: %s, 1", dataModel["uuid"], err.Error())
				}
				logger.Error("ip: %v, grpc连接失败, errMsg: %s", dataModel["ip"], err.Error())
				continue
			}

			cn := client.NewGrpcClient(data["update_name"].(string), data["uuid"].(string), "", dataModel["ip"].(string), nil, conn)
			go func() {
				if err := cn.CallSendProgramCmdMth(); err != nil {
					dataModel["status"] = 300
					if err := apr.Update(dataModel); err != nil {
						logger.Error("uuid: %v, 更新失败, errMsg: %s, 2", dataModel["uuid"], err.Error())
					}
					logger.Error("uuid: %v, ip: %v, grpc连接失败, errMsg: %s, 2", dataModel["uuid"], dataModel["ip"], err.Error())
					return
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
