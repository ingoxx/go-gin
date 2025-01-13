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
				logger.Error(fmt.Sprintf("获取grpc参数失败, errMsg: %s", err.Error()))
				continue
			}

			dataModel["uuid"] = data["uuid"].(string)
			dataModel["status"] = 400

			conn, err := grpc.NewClient(fmt.Sprintf("%s:12306", data["ip"].(string)), grpc.WithTransportCredentials(insecure.NewCredentials()))
			defer conn.Close()
			if err != nil {
				dataModel["status"] = 300
				if err = aprm.Update(dataModel); err != nil {
					logger.Error(fmt.Sprintf("uuid: %s, 更新失败, errMsg: %s, 1", data["uuid"].(string), err.Error()))
				}
				logger.Error(fmt.Sprintf("ip: %s, grpc连接失败, errMsg: %s", data["ip"].(string), err.Error()))
				continue
			}

			cn := client.NewGrpcClient(data["update_name"].(string), data["uuid"].(string), nil, conn)
			go func() {
				if err = cn.Send(); err != nil {
					dataModel["status"] = 300
					if err = aprm.Update(dataModel); err != nil {
						logger.Error(fmt.Sprintf("uuid: %s, 更新失败, errMsg: %s, 2", data["uuid"].(string), err.Error()))
					}
					logger.Error(fmt.Sprintf("ip: %s, grpc发送数据失败, errMsg: %s", data["ip"].(string), err.Error()))
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
