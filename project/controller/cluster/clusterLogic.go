package cluster

import (
	"context"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"sync"
)

type SwarmOperate struct {
	MasterIp   string   `json:"master_ip"`
	Token      string   `json:"token"`
	ClusterCid string   `json:"cluster_cid"`
	WorkIp     []string `json:"work_ip"`
	Message    string   `json:"message"`
	ds         pb.ClusterOperateServiceClient
	ctx        context.Context
	once       sync.Once
}

func (cl *SwarmOperate) initGrpc(ip string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.NewClient(fmt.Sprintf("%s:12306", ip), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	cl.ds = pb.NewClusterOperateServiceClient(conn)

	return
}

func (cl *SwarmOperate) CreateCluster() (interface{}, error) {
	conn, err := cl.initGrpc(cl.MasterIp)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	stream, err := cl.ds.ClusterInit(cl.ctx, &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp})
	if err != nil {
		return nil, err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		cl.Token = resp.Token
		cl.ClusterCid = resp.ClusterID
	}

	// 调用 JoinCluster 并处理返回的错误
	data, errs := cl.JoinCluster()
	if len(errs) > 0 {
		return nil, fmt.Errorf("加入到集群是发生错误, errMsg: %v", errs)
	}

	return data, nil
}

func (cl *SwarmOperate) JoinCluster() ([]map[string]string, []error) {
	var respData = make([]map[string]string, 0)
	var errs = make([]error, 0)
	for _, workIp := range cl.WorkIp {
		conn, err := cl.initGrpc(cl.MasterIp)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		data := cl.joinCluster(workIp, conn)
		respData = append(respData, data)
	}

	if len(errs) != 0 {
		return respData, errs
	}

	return respData, nil
}

func (cl *SwarmOperate) joinCluster(ip string, conn *grpc.ClientConn) map[string]string {
	defer conn.Close()
	var data = make(map[string]string)
	stream, err := cl.ds.ClusterAddNode(cl.ctx, &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp, NodeIp: ip, Token: cl.Token})
	if err != nil {
		data["errMsg"] = err.Error()
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		data["ip"] = resp.Message
		data["message"] = resp.Ip
	}

	return nil
}

func (cl *SwarmOperate) LeaveCluster() ([]map[string]string, []error) {
	var respData = make([]map[string]string, 0)
	var errs = make([]error, 0)
	var data = make(map[string]string)
	for _, workIp := range cl.WorkIp {
		conn, err := cl.initGrpc(cl.MasterIp)
		if err != nil {
			errs = append(errs, err)
		}
		cl.leaveCluster(workIp, conn)
		respData = append(respData, data)
	}

	return respData, errs
}

func (cl *SwarmOperate) leaveCluster(ip string, conn *grpc.ClientConn) map[string]string {
	defer conn.Close()
	var data = make(map[string]string)
	stream, err := cl.ds.ClusterLeaveSwarm(cl.ctx, &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp, NodeIp: ip, Token: cl.Token})
	if err != nil {
		data["errMsg"] = err.Error()
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		data["ip"] = resp.Ip
		data["message"] = resp.Message
	}

	return nil
}
