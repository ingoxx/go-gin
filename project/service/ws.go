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
)

type Ws struct {
	Conn        *websocket.Conn `json:"-"`
	mc          api.ModelCurd
	Ip          string `json:"ip"`
	ProcessName string `json:"name"`
	Uuid        string `json:"uuid"`
}

func NewWs(conn *websocket.Conn, mc api.ModelCurd) *Ws {
	return &Ws{
		Conn: conn,
		mc:   mc,
	}
}

func (ws *Ws) Error(err error) {
	var data = make(map[string]interface{})
	if Err := ws.Conn.WriteMessage(1, []byte(fmt.Sprintf("%s", err.Error()))); Err != nil {
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

	if errS := ws.Send(); errS != nil {
		ws.Error(errS)
		return
	}

	return
}

func (ws *Ws) Send() (err error) {
	server := fmt.Sprintf("%s:12306", ws.Ip)
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	cn := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Conn, conn)
	if err = cn.Send(); err != nil {
		return err
	}

	return
}

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

	cn := client.NewSyncFileRpcClient(sfw.Ip, sfw.File, sfw.Conn)
	if err = cn.Run(); err != nil {
		return
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
