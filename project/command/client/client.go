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
	Name    string
	Uuid    string
	File    string
	RpcConn *grpc.ClientConn
	WsConn  *websocket.Conn
	ctx     context.Context
	lock    *sync.Mutex
}

func NewGrpcClient(name, uuid string, ws *websocket.Conn, rc *grpc.ClientConn) *GrpcClient {
	return &GrpcClient{
		Name:    name,
		Uuid:    uuid,
		WsConn:  ws,
		RpcConn: rc,
	}
}

func (rc *GrpcClient) Send() (err error) {
	c := pb.NewStreamUpdateProgramServiceClient(rc.RpcConn)
	switch rc.Name {
	case "dockerUpdate":
		if err = rc.DockerUpdate(c); err != nil {
			return err
		}
	case "javaUpdate":
		if err = rc.JavaUpdate(c); err != nil {
			return err
		}
	case "dockerUpdateLog":
		if err = rc.DockerUpdateLog(c); err != nil {
			return err
		}
	case "javaUpdateLog":
		if err = rc.JavaUpdateLog(c); err != nil {
			return err
		}
	default:
		if err = rc.sendErr(); err != nil {
			logger.Error(fmt.Sprintf("sendErr errMsg: %s", err.Error()))
		}
	}

	return
}

func (rc *GrpcClient) sendErr() (err error) {
	if rc.WsConn == nil {
		return errors.New("websocket已经关闭")
	}

	if err = rc.WsConn.WriteMessage(1, []byte(fmt.Sprintf("%s\n", err.Error()))); err != nil {
		return err
	}

	return
}

func (rc *GrpcClient) DockerUpdate(c pb.StreamUpdateProgramServiceClient) (err error) {
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

func (rc *GrpcClient) DockerUpdateLog(c pb.StreamUpdateProgramServiceClient) (err error) {
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

func (rc *GrpcClient) JavaUpdate(c pb.StreamUpdateProgramServiceClient) (err error) {
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

func (rc *GrpcClient) JavaUpdateLog(c pb.StreamUpdateProgramServiceClient) (err error) {
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
				file = filepath.Join(rpcConfig.UploadPath, file)
				if err = sfc.Send(ip, file); err != nil {
					if err1 := sfc.ReturnWsData(fmt.Sprintf("%s\n", err.Error())); err1 != nil {
						return
					}
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
		return
	}
	return
}

func (sfc *SyncFileClient) Send(ip, file string) (err error) {
	defer sfc.wg.Done()

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
