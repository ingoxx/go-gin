package cluster

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
)

func (cl *SwarmOperate) initGrpc(ip string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.NewClient(fmt.Sprintf("%s:12306", ip), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}

	return
}

func (cl *SwarmOperate) clearStatus() {
	if cl.Code > 0 {
		cl.Code = 0
	}

	if cl.Message != "" {
		cl.Message = ""
	}
}

func (cl *SwarmOperate) StartHealthCheck() error {
	conn, err := cl.initGrpc(cl.MasterIp)
	if err != nil {
		return err
	}

	defer conn.Close()

	ds := pb.NewClusterOperateServiceClient(conn)
	stream, err := ds.StartClusterMonitor(context.Background(), &pb.StreamClusterOperateReq{ClusterID: cl.ClusterCid})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		cl.Code = resp.GetCode()
	}

	if cl.Code != 10000 {
		return errors.New(cl.Message)
	}

	return nil
}

func (cl *SwarmOperate) CreateCluster() error {
	conn, err := cl.initGrpc(cl.MasterIp)
	if err != nil {
		return err
	}

	defer conn.Close()

	ds := pb.NewClusterOperateServiceClient(conn)
	stream, err := ds.ClusterInit(context.Background(), &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		cl.WorkToken = resp.WToken
		cl.MasterToken = resp.MToken
		cl.ClusterCid = resp.ClusterID
		cl.Code = resp.GetCode()
	}

	if cl.Code != 10000 {
		return errors.New(cl.Message)
	}

	// 调用 JoinCluster 并处理返回的错误
	if errs := cl.JoinWork(); len(errs) > 0 {
		return fmt.Errorf("加入到集群是发生错误, errMsg: %v", errs)
	}

	return nil
}

func (cl *SwarmOperate) JoinMaster() []error {
	var errs = make([]error, 0)
	for _, v := range cl.ServersInfo {
		if v.NodeType == 1 {
			if err := cl.joinMaster(v.Ip); err != nil {
				me := fmt.Errorf("%s|%s", err.Error(), cl.Message)
				errs = append(errs, me)
				cl.clearStatus()
			}
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (cl *SwarmOperate) joinMaster(ip string) error {
	conn, err := cl.initGrpc(ip)
	if err != nil {
		return err
	}

	defer conn.Close()

	ds := pb.NewClusterOperateServiceClient(conn)

	stream, err := ds.ClusterJoinMaster(context.Background(), &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp, NodeIp: ip, MToken: cl.MasterToken})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cl.Code = resp.GetCode()
		cl.Message = resp.GetMessage()
	}

	if cl.Code != 10000 {
		return errors.New(cl.Message)
	}

	return nil
}

func (cl *SwarmOperate) JoinWork() []error {
	var errs = make([]error, 0)
	for _, v := range cl.ServersInfo {
		if v.NodeType == 2 {
			if err := cl.joinWork(v.Ip); err != nil {
				me := fmt.Errorf("%s|%s", err.Error(), cl.Message)
				errs = append(errs, me)
				cl.clearStatus()
			}
		}
	}

	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (cl *SwarmOperate) joinWork(ip string) error {
	conn, err := cl.initGrpc(ip)
	if err != nil {
		return nil
	}

	defer conn.Close()

	ds := pb.NewClusterOperateServiceClient(conn)

	stream, err := ds.ClusterJoinWork(context.Background(), &pb.StreamClusterOperateReq{MasterIp: cl.MasterIp, NodeIp: ip, WToken: cl.WorkToken})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cl.Code = resp.GetCode()
		cl.Message = resp.GetMessage()
	}

	if cl.Code != 10000 {
		return errors.New(cl.Message)
	}

	return nil
}

func (cl *SwarmOperate) LeaveCluster() []error {
	var errs = make([]error, 0)
	for _, v := range cl.ServersInfo {
		if err := cl.leaveCluster(v.Ip); err != nil {
			me := fmt.Errorf("%s|%s", err.Error(), cl.Message)
			errs = append(errs, me)
			cl.clearStatus()
		}
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (cl *SwarmOperate) leaveCluster(ip string) error {
	conn, err := cl.initGrpc(ip)
	if err != nil {
		return nil

	}
	defer conn.Close()

	ds := pb.NewClusterOperateServiceClient(conn)
	stream, err := ds.ClusterLeaveSwarm(context.Background(), &pb.StreamClusterOperateReq{MasterIp: ip, NodeIp: ip, WToken: cl.WorkToken})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		cl.Code = resp.GetCode()
		cl.Message = resp.GetMessage()
	}

	if cl.Code != 10000 {
		return errors.New(cl.Message)
	}

	cl.ID = 0

	return nil
}

func (cl *SwarmOperate) UpdateServers() error {
	for _, v := range cl.ServersInfo {
		if cl.ID != 0 {
			if err := dao.DB.Model(&model.AssetsModel{}).Where("ip = ?", v.Ip).Updates(map[string]interface{}{
				"cluster_id": cl.ID,
				//"node_type":   v.NodeType,
				//"node_status": 3,
			}).Error; err != nil {
				return err
			}
		} else {
			if err := dao.DB.Model(&model.AssetsModel{}).Where("ip = ?", v.Ip).Updates(map[string]interface{}{
				"cluster_id": nil,
				//"node_type":   3,
				//"node_status": 300,
				"leave_type": 1,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
