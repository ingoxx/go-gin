package client

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/command/rpcConfig"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type GrpcClient struct {
	Name    string // 操作函数名
	Uuid    string
	File    string
	Cmd     string
	Ip      string
	RpcConn *grpc.ClientConn
	WsConn  *websocket.Conn
	ctx     context.Context
	lock    *sync.Mutex
	wg      sync.WaitGroup
	sc      pb.StreamUpdateProgramServiceClient
	sl      pb.StreamCheckSystemLogServiceClient
}

func NewGrpcClient(name, uuid, cmd, ip string, ws *websocket.Conn, rc *grpc.ClientConn) *GrpcClient {
	return &GrpcClient{
		Name:    name,
		Uuid:    uuid,
		Cmd:     cmd,
		Ip:      ip,
		WsConn:  ws,
		RpcConn: rc,
		lock:    new(sync.Mutex),
		sc:      pb.NewStreamUpdateProgramServiceClient(rc),
		sl:      pb.NewStreamCheckSystemLogServiceClient(rc),
	}
}

func (rc *GrpcClient) CallSendLinuxCmdMth(wg *sync.WaitGroup, limit chan struct{}, output chan map[string][]string) (err error) {
	if err := rc.RunLinuxCmd(wg, limit, output); err != nil {
		return err
	}
	return
}

func (rc *GrpcClient) CallSystemLogMth(log, start, end, field string) (err error) {
	data := &pb.StreamSystemLogRequest{LogName: log, Start: start, End: end, Field: field}
	stream, err := rc.sl.CheckSystemLog(context.Background(), data)
	if err != nil {
		return
	}

	defer rc.RpcConn.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", resp.Message))); err != nil {
				return err
			}
		}
	}

	return
}

func (rc *GrpcClient) RunLinuxCmd(wg *sync.WaitGroup, limit chan struct{}, output chan map[string][]string) (err error) {
	defer wg.Done()
	stream, err := rc.sc.RunLinuxCmd(context.Background(), &pb.StreamRequest{Cmd: rc.Cmd, Ip: rc.Ip})
	if err != nil {
		return
	}

	var res = make(map[string][]string)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			res[rc.Ip] = append(res[rc.Ip], resp.Message)
		}
	}

	output <- res
	<-limit

	return
}

func (rc *GrpcClient) CallSendProgramCmdMth() (err error) {
	switch rc.Name {
	case "dockerUpdate":
		if err := rc.DockerUpdate(); err != nil {
			return err
		}
	case "javaUpdate":
		if err := rc.JavaUpdate(); err != nil {
			return err
		}
	case "dockerUpdateLog":
		if err := rc.DockerUpdateLog(); err != nil {
			return err
		}
	case "javaUpdateLog":
		if err := rc.JavaUpdateLog(); err != nil {
			return err
		}
	default:
		err = errors.New(fmt.Sprintf("method not found, errMsg: %s", err.Error()))
		logger.Error(fmt.Sprintf("method not found, errMsg: %s", err.Error()))
	}

	return
}

func (rc *GrpcClient) DockerUpdate() (err error) {
	stream, err := rc.sc.DockerUpdate(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

	defer rc.RpcConn.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", resp.Message))); err != nil {
				return err
			}
		}
	}

	return
}

func (rc *GrpcClient) DockerUpdateLog() (err error) {
	stream, err := rc.sc.DockerUpdateLog(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

	defer rc.RpcConn.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", resp.Message))); err != nil {
				return err
			}
		}
	}

	return
}

func (rc *GrpcClient) JavaUpdate() (err error) {
	stream, err := rc.sc.JavaUpdate(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

	defer rc.RpcConn.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", resp.Message))); err != nil {
				return err
			}
		}

	}

	return
}

func (rc *GrpcClient) JavaUpdateLog() (err error) {
	stream, err := rc.sc.JavaUpdateLog(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

	defer rc.RpcConn.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if rc.WsConn != nil {
			if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", resp.Message))); err != nil {
				return err
			}
		}

	}

	return
}

type SyncFileClient struct {
	Ip      []string
	File    []string
	RpcConn *grpc.ClientConn
	WsConn  *websocket.Conn
	ctx     context.Context
	wg      sync.WaitGroup
	resChan chan string
	lock    *sync.Mutex
}

func NewSyncFileRpcClient(ip, file []string, ws *websocket.Conn) *SyncFileClient {
	return &SyncFileClient{
		Ip:      ip,
		File:    file,
		WsConn:  ws,
		resChan: make(chan string),
	}
}

func (sfc *SyncFileClient) Run() (err error) {
	for _, file := range sfc.File {
		for _, ip := range sfc.Ip {
			sfc.wg.Add(1)
			go func(ip, file string) {
				defer sfc.wg.Done()
				file = filepath.Join(rpcConfig.UploadPath, file)
				if err = sfc.Send(ip, file); err != nil {
					sfc.resChan <- err.Error()
					return
				}
			}(ip, file)
		}
	}

	go func() {
		sfc.wg.Wait()
		close(sfc.resChan)
	}()

	for data := range sfc.resChan {
		if err = sfc.ReturnWsData(fmt.Sprintf("%s\n", data)); err != nil {
			return
		}
	}

	return
}

func (sfc *SyncFileClient) ReturnWsData(data string) (err error) {
	if err = sfc.WsConn.WriteMessage(1, []byte(data)); err != nil {
		logger.Error(fmt.Sprintf("websocket err, errMsg: %s\n", err.Error()))
		return
	}
	return
}

func (sfc *SyncFileClient) Send(ip, file string) (err error) {
	server := fmt.Sprintf("%s:12306", ip)
	conn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	defer conn.Close()

	c := pb.NewFileTransferServiceClient(conn)

	stream, err := c.SendFile(context.Background())

	if err != nil {
		return
	}

	buffer := make([]byte, 8092)

	f, err := os.Open(file)
	if err != nil {
		return
	}

	defer f.Close()

	for {
		b, err := f.Read(buffer)
		if err == io.EOF {
			break
		}

		if b == 0 {
			break
		}

		if err = stream.Send(&pb.FileMessage{Byte: buffer[:b], Name: filepath.Base(file), Ip: ip}); err != nil {
			return err
		}
	}

	stream.CloseSend()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		sfc.resChan <- ip + "-" + resp.GetName()
	}

	return
}
