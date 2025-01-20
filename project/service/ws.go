package service

import (
	"encoding/json"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/api"
	"github.com/Lxb921006/Gin-bms/project/command/client"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
)

type Ws struct {
	Conn        *websocket.Conn `json:"-"`
	mc          api.ModelCurd
	Ip          []string `json:"ip"`
	ProcessName string   `json:"name"`
	Uuid        string   `json:"uuid"`
	Cmd         string   `json:"cmd"`
	wg          *sync.WaitGroup
	limit       chan struct{}
	output      chan map[string][]string
}

func NewWs(conn *websocket.Conn, mc api.ModelCurd) *Ws {
	return &Ws{
		Conn:   conn,
		mc:     mc,
		wg:     new(sync.WaitGroup),
		output: make(chan map[string][]string),
		limit:  make(chan struct{}, 20),
	}
}

func (ws *Ws) Error(err error) {
	var data = make(map[string]interface{})
	if Err := ws.Conn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", err.Error()))); Err != nil {
		data["uuid"] = ws.Uuid
		data["status"] = 300
		if err := ws.mc.Update(data); err != nil {
			logger.Error(fmt.Sprintf("fail to update AssetsProgramUpdateRecordModel, errMsg: %s", err.Error()))
			return
		}
		logger.Error(fmt.Sprintf("Ws writeMessage errMsg: %s", err.Error()))
	}
}

func (ws *Ws) Run() (err error) {
	_, message, err := ws.Conn.ReadMessage()
	if err != nil {
		ws.Error(err)
		return
	}

	if err = ParseJsonToStruct(message, ws); err != nil {
		return
	}

	if ws.ProcessName == "runLinuxCmd" {
		if errS := ws.AcpLinuxCmd(); errS != nil {
			return
		}
	} else {
		if errS := ws.AcpProgramCmd(); errS != nil {
			ws.Error(errS)
			return
		}
	}

	return
}

func (ws *Ws) AcpLinuxCmd() (err error) {
	for _, ip := range ws.Ip {
		ws.wg.Add(1)
		ws.limit <- struct{}{}
		go func(ip string) {
			server := fmt.Sprintf("%s:12306", ip)
			conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				ws.Error(err)
				return
			}

			defer conn.Close()

			if err := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Cmd, ip, ws.Conn, conn).SendLinuxCmd(ws.wg, ws.limit, ws.output); err != nil {
				ws.Error(err)
				return
			}

		}(ip)
	}

	go func() {
		ws.wg.Wait()
		close(ws.output)
	}()

	for data := range ws.output {
		for _, v1 := range data {
			for _, v2 := range v1 {
				if err = ws.Conn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", v2))); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (ws *Ws) AcpProgramCmd() (err error) {
	for _, ip := range ws.Ip {
		server := fmt.Sprintf("%s:12306", ip)
		conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		if err := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Cmd, ip, ws.Conn, conn).SendProgramCmd(); err != nil {
			return err
		}
	}

	return
}

// SendFileWs 文件分发
type SendFileWs struct {
	Conn *websocket.Conn `json:"-"`
	Ip   []string        `json:"ip"`
	File []string        `json:"file"`
}

func NewSendFileWs(conn *websocket.Conn) *SendFileWs {
	return &SendFileWs{
		Conn: conn,
	}
}

func (sfw *SendFileWs) Send() (err error) {
	_, message, err := sfw.Conn.ReadMessage()
	if err != nil {
		return
	}

	if err = ParseJsonToStruct(message, sfw); err != nil {
		return
	}

	if err := client.NewSyncFileRpcClient(sfw.Ip, sfw.File, sfw.Conn).Run(); err != nil {
		return err
	}

	return
}

func (sfw *SendFileWs) Error(err error) {
	if Err := sfw.Conn.WriteMessage(1, []byte(fmt.Sprintf("%s", err.Error()))); Err != nil {
		log.Println(fmt.Sprintf("SendFileWs writeMessage errMsg: %s", Err.Error()))
	}
}

func ParseJsonToStruct(data []byte, ws interface{}) (err error) {
	if err = json.Unmarshal(data, &ws); err != nil {
		return
	}

	return
}
