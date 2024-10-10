package utils

import (
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/api"
	"github.com/Lxb921006/Gin-bms/project/command/client"
	"github.com/Lxb921006/Gin-bms/project/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProgramAsyncRunCelery struct {
	Works chan api.CeleryInterface
}

func NewProgramAsyncRunCelery() *ProgramAsyncRunCelery {
	var aprm model.AssetsProcessUpdateRecordModel
	var dataModel = make(map[string]interface{})

	c := &ProgramAsyncRunCelery{
		Works: make(chan api.CeleryInterface),
	}

	SetLogFile(CeleryLogPath)
	SetLogLevel(ErrorLevel)

	go func() {
		for w := range c.Works {
			data, err := w.Data()
			if err != nil {
				Error("获取grpc参数失败: ", err)
				return
			}

			dataModel["uuid"] = data["uuid"].(string)
			dataModel["status"] = 400

			conn, err := grpc.Dial(fmt.Sprintf("%s:12306", data["ip"].(string)), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				if err = aprm.Update(dataModel); err != nil {
					Error("更新失败-connect: ", err)
				}
				Error("连接grpc失败: ", err)
				return
			}

			cn := client.NewRpcClient(data["update_name"].(string), data["uuid"].(string), nil, conn)
			go func() {
				if err = cn.Send(); err != nil {
					if err = aprm.Update(dataModel); err != nil {
						Error("更新失败-send: ", err)
					}
					Error("grpc发送数据失败: ", err)
				}
			}()
		}

	}()

	return c
}

func (c *ProgramAsyncRunCelery) Task(task api.CeleryInterface) {
	c.Works <- task
}
