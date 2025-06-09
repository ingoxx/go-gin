package service

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ingoxx/go-gin/project/api"
	"github.com/ingoxx/go-gin/project/command/client"
	"github.com/ingoxx/go-gin/project/logger"
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
	LogName     string   `json:"log_name"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Field       string   `json:"field"`
	wg          *sync.WaitGroup
	gCtx        *gin.Context
	limit       chan struct{}
	output      chan map[string][]string
	record      api.RecordWebsocketLog
	lock        *sync.Mutex
}

func NewWs(conn *websocket.Conn, mc api.ModelCurd, gCtx *gin.Context, record api.RecordWebsocketLog) *Ws {
	return &Ws{
		Conn:   conn,
		mc:     mc,
		wg:     new(sync.WaitGroup),
		gCtx:   gCtx,
		output: make(chan map[string][]string),
		limit:  make(chan struct{}, 20),
		record: record,
		lock:   new(sync.Mutex),
	}

}

func (ws *Ws) Error(err error) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

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

	// 解析websocket数据
	if err = ParseJsonToStruct(message, ws); err != nil {
		return
	}

	go ws.recordLog()

	if ws.ProcessName == "runLinuxCmd" {
		if err := ws.AcpLinuxCmd(); err != nil {
			return err
		}
	} else if ws.ProcessName == "checkSystemLog" {
		if err := ws.AcpSystemLog(); err != nil {
			return err
		}
	} else {
		if err := ws.AcpProgramCmd(); err != nil {
			ws.Error(err)
			return err
		}
	}

	return
}

// AcpLinuxCmd 批量命令执行
func (ws *Ws) AcpLinuxCmd() (err error) {
	for _, ip := range ws.Ip {
		ws.wg.Add(1)
		ws.limit <- struct{}{}
		go func(ip string) {
			server := fmt.Sprintf("%s:12306", ip)
			conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				ws.inputErr(ip, err)
				return
			}

			defer conn.Close()

			if err := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Cmd, ip, ws.Conn, conn).CallSendLinuxCmdMth(ws.wg, ws.limit, ws.output); err != nil {
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
				if err := ws.wsSendData(v2); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (ws *Ws) inputErr(ip string, err error) {
	var res = make(map[string][]string)
	res[ip] = append(res[ip], err.Error())
	ws.output <- res
}

func (ws *Ws) wsSendData(data string) error {
	ws.lock.Lock()
	defer ws.lock.Unlock()
	if err := ws.Conn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", data))); err != nil {
		logger.Error(fmt.Sprintf("wsSendData fail to send data, errMsg: %s", err.Error()))
		return err
	}

	return nil
}

func (ws *Ws) recordLog() {
	var data = make(map[string]interface{})
	if ws.gCtx.Request.URL.Path == "/assets/ws" {
		data["url"] = fmt.Sprintf("%s, 更新操作程序: %s, 操作服务器: %v", ws.gCtx.Request.URL.Path, ws.ProcessName, ws.Ip)
		data["operator"] = ws.gCtx.Query("user")
		data["ip"] = ws.gCtx.RemoteIP()
	}

	if ws.gCtx.Request.URL.Path == "/assets/run-linux-cmd" {
		data["url"] = fmt.Sprintf("%s, 批量执行命令: %s, 操作服务器: %v", ws.gCtx.Request.URL.Path, ws.Cmd, ws.Ip)
		data["operator"] = ws.gCtx.Query("user")
		data["ip"] = ws.gCtx.RemoteIP()
	}

	if ws.gCtx.Request.URL.Path == "/assets/view-system-log" {
		data["url"] = fmt.Sprintf("%s, 查询日志: %s, 查询字段: %s, 查询日期: %s-%s, 操作服务器: %v", ws.gCtx.Request.URL.Path, ws.LogName, ws.Field, ws.Start, ws.End, ws.Ip)
		data["operator"] = ws.gCtx.RemoteIP()
		data["ip"] = ws.gCtx.Query("user")
	}

	if err := ws.RecordLog(data); err != nil {
		logger.Error(fmt.Sprintf("操作记录失败, errMsg: %s", err.Error()))
	}
}

func (ws *Ws) RecordLog(data map[string]interface{}) error {
	if err := ws.record.RecordLog(data); err != nil {
		return err
	}
	if err := ws.record.DataCount(ws.gCtx.Request.URL.Path); err != nil {
		return err
	}

	return nil
}

func (ws *Ws) AcpProgramCmd() (err error) {
	for _, ip := range ws.Ip {
		server := fmt.Sprintf("%s:12306", ip)
		conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		if err := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Cmd, ip, ws.Conn, conn).CallSendProgramCmdMth(); err != nil {
			return err
		}
	}

	return
}

// AcpSystemLog 服务器系统日志查询
func (ws *Ws) AcpSystemLog() error {
	for _, ip := range ws.Ip {
		server := fmt.Sprintf("%s:12306", ip)
		conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		if err := client.NewGrpcClient(ws.ProcessName, ws.Uuid, ws.Cmd, ip, ws.Conn, conn).CallSystemLogMth(ws.LogName, ws.Start, ws.End, ws.Field); err != nil {
			if err := ws.wsSendData(err.Error()); err != nil {
				return err
			}
			return err
		}
	}

	return nil
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
