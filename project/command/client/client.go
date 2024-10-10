package client

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/command/rpcConfig"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type RpcClient struct {
	Name    string
	Uuid    string
	File    string
	RpcConn *grpc.ClientConn
	WsConn  *websocket.Conn
	ctx     context.Context
}

func NewRpcClient(name, uuid string, ws *websocket.Conn, rc *grpc.ClientConn) *RpcClient {
	return &RpcClient{
		Name:    name,
		Uuid:    uuid,
		WsConn:  ws,
		RpcConn: rc,
	}
}

func (rc *RpcClient) Send() (err error) {
	switch rc.Name {
	case "dockerUpdate":
		if err = rc.DockerUpdate(); err != nil {
			return err
		}
	case "javaUpdate":
		if err = rc.JavaUpdate(); err != nil {
			return err
		}
	case "dockerUpdateLog":
		if err = rc.DockerUpdateLog(); err != nil {
			return err
		}
	case "javaUpdateLog":
		if err = rc.JavaUpdateLog(); err != nil {
			return err
		}
	default:
		return errors.New("无效操作")
	}

	return
}

func (rc *RpcClient) DockerUpdate() (err error) {
	c := pb.NewStreamUpdateProcessServiceClient(rc.RpcConn)
	stream, err := c.DockerUpdate(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
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

func (rc *RpcClient) DockerUpdateLog() (err error) {
	c := pb.NewStreamUpdateProcessServiceClient(rc.RpcConn)
	stream, err := c.DockerUpdateLog(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
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

func (rc *RpcClient) JavaUpdate() (err error) {
	c := pb.NewStreamUpdateProcessServiceClient(rc.RpcConn)
	stream, err := c.JavaUpdate(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

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

func (rc *RpcClient) JavaUpdateLog() (err error) {
	c := pb.NewStreamUpdateProcessServiceClient(rc.RpcConn)
	stream, err := c.JavaUpdateLog(context.Background(), &pb.StreamRequest{Uuid: rc.Uuid})
	if err != nil {
		return
	}

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

// 分发文件
type SyncFileRpcClient struct {
	Ip      []string
	File    []string
	RpcConn *grpc.ClientConn
	WsConn  *websocket.Conn
	ctx     context.Context
	wg      sync.WaitGroup
	resChan chan string
}

func NewSyncFileRpcClient(ip, file []string, ws *websocket.Conn) *SyncFileRpcClient {
	return &SyncFileRpcClient{
		Ip:      ip,
		File:    file,
		WsConn:  ws,
		resChan: make(chan string),
	}
}

func (sfrc *SyncFileRpcClient) Run() (err error) {
	for _, file := range sfrc.File {
		for _, ip := range sfrc.Ip {
			sfrc.wg.Add(1)
			go func(ip, file string) {
				file = filepath.Join(rpcConfig.UploadPath, file)
				if err = sfrc.Send(ip, file); err != nil {
					if err = sfrc.ReturnWsData(fmt.Sprintf("%s\n", err.Error())); err != nil {
						return
					}
				}
			}(ip, file)
		}
	}

	go func() {
		sfrc.wg.Wait()
		close(sfrc.resChan)
	}()

	for data := range sfrc.resChan {
		if err = sfrc.ReturnWsData(fmt.Sprintf("%s\n", data)); err != nil {
			return
		}

	}

	return
}

func (sfrc *SyncFileRpcClient) ReturnWsData(data string) (err error) {
	if err = sfrc.WsConn.WriteMessage(1, []byte(data)); err != nil {
		return
	}
	return
}

func (sfrc *SyncFileRpcClient) Send(ip, file string) (err error) {
	defer sfrc.wg.Done()
	server := fmt.Sprintf("%s:12306", ip)

	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

		if err = stream.Send(&pb.FileMessage{Byte: buffer[:b], Name: filepath.Base(file)}); err != nil {
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

		sfrc.resChan <- ip + "-" + resp.GetName()
	}

	return
}
