package service

import (
	"encoding/json"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/command/client"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Ws struct {
	Conn        *websocket.Conn `json:"-"`
	Ip          string          `json:"ip"`
	ProcessName string          `json:"name"`
	Uuid        string          `json:"uuid"`
}

func NewWs(conn *websocket.Conn) *Ws {
	return &Ws{
		Conn: conn,
	}
}

func (ws *Ws) Run() (err error) {
	_, message, err := ws.Conn.ReadMessage()

	if err != nil {
		return
	}

	if err = ParseJsonToStruct(message, ws); err != nil {
		return
	}

	if err = ws.Send(); err != nil {
		return err
	}

	return
}

func (ws *Ws) Send() (err error) {
	server := fmt.Sprintf("%s:12306", ws.Ip)
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	defer conn.Close()

	cn := client.NewRpcClient(ws.ProcessName, ws.Uuid, ws.Conn, conn)
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

func ParseJsonToStruct(data []byte, ws interface{}) (err error) {
	if err = json.Unmarshal(data, &ws); err != nil {
		return
	}

	return
}
