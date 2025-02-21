package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/command/server/redis"
	"github.com/Lxb921006/Gin-bms/project/command/server/script"
	"github.com/Lxb921006/Gin-bms/project/tools/dockerSwarmApi"
	"github.com/Lxb921006/Gin-bms/project/tools/dockerSwarmStatusCheck"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	savePath = "/opt"
)

type server struct {
	pb.UnimplementedStreamUpdateProgramServiceServer
	pb.UnimplementedFileTransferServiceServer
	pb.UnimplementedStreamCheckSystemLogServiceServer
	pb.UnimplementedClusterOperateServiceServer
}

type runScriptData struct {
	req        *pb.StreamRequest
	systemLog  *pb.StreamSystemLogRequest
	stream     pb.StreamUpdateProgramService_DockerUpdateServer
	program    string
	programLog string
}

func (s *server) DockerUpdate(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_DockerUpdateServer) (err error) {
	log.Println("received DockerUpdate")

	data := runScriptData{
		req:        req,
		stream:     stream,
		program:    script.DockerUpdateScript,
		programLog: script.DockerUpdateLog,
	}

	if err = s.scriptOutPut(data); err != nil {
		if err = data.stream.Send(&pb.StreamReply{Message: fmt.Sprintf("fail to run DockerUpdate, errMsg: %s\n", err.Error())}); err != nil {
			log.Printf("fail to run send msg, errMsg: %s\n", err.Error())
		}
		return
	}

	return
}

func (s *server) DockerUpdateLog(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_DockerUpdateLogServer) (err error) {
	log.Println("received DockerUpdateLog")

	cmd := exec.Command("more", script.DockerUpdateLog, req.GetUuid())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err = stream.Send(&pb.StreamReply{Message: scanner.Text()}); err != nil {
			return
		}
	}

	if err = cmd.Wait(); err != nil {
		return
	}

	return
}

func (s *server) JavaUpdateLog(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_JavaUpdateLogServer) (err error) {
	log.Println("received JavaUpdateLog")

	cmd := exec.Command("more", script.JavaUpdateLog, req.GetUuid())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err = stream.Send(&pb.StreamReply{Message: scanner.Text()}); err != nil {
			return
		}
	}

	if err = cmd.Wait(); err != nil {
		return
	}

	return
}

func (s *server) JavaUpdate(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_JavaUpdateServer) (err error) {
	log.Println("received JavaUpdate")

	data := runScriptData{
		req:        req,
		stream:     stream,
		program:    script.JavaUpdateScript,
		programLog: script.JavaUpdateLog,
	}

	if err = s.scriptOutPut(data); err != nil {
		if err = data.stream.Send(&pb.StreamReply{Message: fmt.Sprintf("fail to run JavaUpdate, errMsg: %s\n", err.Error())}); err != nil {
			log.Printf("fail to run send msg, errMsg: %s\n", err.Error())
		}
		return
	}

	return
}

func (s *server) scriptOutPut(data runScriptData) (err error) {
	if _, err = os.Open(data.program); err != nil {
		return
	}

	var makeCmd string
	if data.req != nil {
		if data.req.GetUuid() != "" {
			makeCmd = fmt.Sprintf("bash %s %s | tee %s", data.program, data.req.GetUuid(), data.programLog)
		} else if data.req.GetCmd() != "" {
			makeCmd = fmt.Sprintf("bash %s \"%s\" | tee %s", data.program, data.req.GetCmd(), data.programLog)
		}
	} else if data.systemLog != nil {
		makeCmd = fmt.Sprintf("bash %s %s %s %s %s| tee %s",
			data.program,
			data.systemLog.GetLogName(),
			data.systemLog.GetStart(),
			data.systemLog.GetEnd(),
			data.systemLog.GetField(),
			data.programLog)
	}

	cmd := exec.Command("bash", "-c", makeCmd)
	// 标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	// 标准错误
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if err = data.stream.Send(&pb.StreamReply{Message: scanner.Text()}); err != nil {
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if err = data.stream.Send(&pb.StreamReply{Message: scanner.Text()}); err != nil {
				return
			}
		}
	}()

	if err = cmd.Wait(); err != nil {
		return
	}

	return
}

func (s *server) DockerReload(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_DockerReloadServer) (err error) {
	return
}

func (s *server) JavaReload(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_JavaReloadServer) (err error) {
	return
}

func (s *server) RunLinuxCmd(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_DockerUpdateServer) (err error) {
	log.Println("received RunLinuxCmd")

	data := runScriptData{
		req:        req,
		stream:     stream,
		program:    script.RunLinuxCmdScript,
		programLog: script.RunLinuxCmdLog,
	}

	if err = s.scriptOutPut(data); err != nil {
		if err = data.stream.Send(&pb.StreamReply{Message: fmt.Sprintf("%s fail to run %s, errMsg: %s\n", req.GetIp(), data.program, err.Error())}); err != nil {
			log.Printf("fail to run send msg, errMsg: %s\n", err.Error())
		}
		return
	}

	return
}

func (s *server) CheckSystemLog(req *pb.StreamSystemLogRequest, stream pb.StreamCheckSystemLogService_CheckSystemLogServer) (err error) {
	log.Println("received CheckSystemLog")

	data := runScriptData{
		systemLog:  req,
		stream:     stream,
		program:    script.CheckSystemLogScript,
		programLog: script.CheckSystemLog,
	}

	if err = s.scriptOutPut(data); err != nil {
		if err = data.stream.Send(&pb.StreamReply{Message: err.Error()}); err != nil {
			log.Printf("fail to run send msg, errMsg: %s\n", err.Error())
		}
		return
	}

	return
}

func (s *server) ClusterInit(req *pb.StreamClusterOperateReq, stream pb.ClusterOperateService_ClusterInitServer) (err error) {
	log.Println("received ClusterInit")

	cid, wToken, mToken, err := dockerSwarmApi.NewDockerSwarmOp(req.GetMasterIp(), "", "", "", context.Background()).CreateSwarm()
	if err != nil {
		log.Printf("fail to init swarm, errMsg: %s\n", err.Error())
		if err = stream.Send(&pb.StreamClusterOperateResp{Message: fmt.Sprintf("fail to init swarm, errMsg: %s\n", err.Error()), Code: 10001}); err != nil {
			log.Printf("ClusterInit, fail to send data, errMsg: %s\n", err.Error())
		}

		return err
	}

	if err = stream.Send(&pb.StreamClusterOperateResp{Message: "ok", WToken: wToken, MToken: mToken, ClusterID: cid, Ip: req.GetMasterIp(), Code: 10000}); err != nil {
		log.Printf("ClusterInit, fail to send data,  errMsg: %s\n", err.Error())
		return err
	}

	// 启动集群的健康监测脚本
	go dockerSwarmStatusCheck.Check(cid)

	return
}

func (s *server) ClusterJoinWork(req *pb.StreamClusterOperateReq, stream pb.ClusterOperateService_ClusterJoinWorkServer) (err error) {
	log.Println("received ClusterJoinWork")

	if err := dockerSwarmApi.NewDockerSwarmOp(req.GetMasterIp(), req.GetNodeIp(), req.GetWToken(), "", context.Background()).JoinWorkSwarm(); err != nil {
		log.Println(fmt.Sprintf("faied to join work swarm, errMsg: %s\n", err.Error()))
		if err = stream.Send(&pb.StreamClusterOperateResp{Message: fmt.Sprintf("faied to join work swarm, errMsg: %s\n", err.Error()), Code: 10001}); err != nil {
			log.Printf("ClusterJoinWork, fail to send data, errMsg: %s\n", err.Error())
		}
		return err
	}

	if err = stream.Send(&pb.StreamClusterOperateResp{Message: "ok", Ip: req.GetNodeIp(), Code: 10000}); err != nil {
		log.Printf("ClusterJoinWork, fail to send data,  errMsg: %s\n", err.Error())
		return err
	}

	return
}

func (s *server) ClusterJoinMaster(req *pb.StreamClusterOperateReq, stream pb.ClusterOperateService_ClusterJoinMasterServer) (err error) {
	log.Println("received ClusterJoinMaster")

	if err := dockerSwarmApi.NewDockerSwarmOp(req.GetMasterIp(), req.GetNodeIp(), "", req.GetMToken(), context.Background()).JoinMasterSwarm(); err != nil {
		log.Println(fmt.Sprintf("faied to join master swarm, errMsg: %s\n", err.Error()))
		if err = stream.Send(&pb.StreamClusterOperateResp{Message: fmt.Sprintf("faied to join master swarm, errMsg: %s\n", err.Error()), Code: 10001}); err != nil {
			log.Printf("ClusterJoinMaster, fail to send data, errMsg: %s\n", err.Error())
		}
		return err
	}

	if err = stream.Send(&pb.StreamClusterOperateResp{Message: "ok", Ip: req.GetNodeIp(), Code: 10000}); err != nil {
		log.Printf("ClusterJoinMaster, fail to send data,  errMsg: %s\n", err.Error())
		return err
	}

	return
}

func (s *server) ClusterLeaveSwarm(req *pb.StreamClusterOperateReq, stream pb.ClusterOperateService_ClusterLeaveSwarmServer) (err error) {
	log.Println("received ClusterLeaveSwarm")

	if err := dockerSwarmApi.NewDockerSwarmOp("", "", "", "", context.Background()).LeaveSwarm(); err != nil {
		log.Println(fmt.Sprintf("faied to leave swarm, errMsg: %s\n", err.Error()))
		if err = stream.Send(&pb.StreamClusterOperateResp{Message: fmt.Sprintf("faied to leave swarm, errMsg: %s\n", err.Error()), Code: 10001}); err != nil {
			log.Printf("ClusterLeaveSwarm, fail to send data, errMsg: %s\n", err.Error())
		}

		return err
	}

	if err = stream.Send(&pb.StreamClusterOperateResp{Message: "ok", Code: 10000}); err != nil {
		log.Printf("ClusterLeaveSwarm, fail to send data,  errMsg: %s\n", err.Error())
		return err
	}

	return
}

// SendFile 接收文件并返回文件md5
func (s *server) SendFile(stream pb.FileTransferService_SendFileServer) (err error) {
	if err = s.ProcessMsg(stream); err != nil {
		log.Println(err)
	}

	return
}

func (s *server) ProcessMsg(stream pb.FileTransferService_SendFileServer) (err error) {
	var file string
	var msg string
	var ip string
	var chunks [][]byte
	var fileTmp = filepath.Join(savePath, file+".tmp")

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if file == "" {
			file = filepath.Join(savePath, resp.GetName())
		}

		if ip == "" {
			ip = resp.GetIp()
		}
		chunks = append(chunks, resp.Byte)
	}

	log.Println("received file: ", filepath.Base(file))

	fw, err := os.Create(fileTmp)
	if err != nil {
		s.send(err.Error(), stream)
		return
	}

	defer fw.Close()

	nw := bufio.NewWriter(fw)
	for _, chunk := range chunks {
		if _, err = nw.Write(chunk); err != nil {
			s.send(err.Error(), stream)
			return
		}
	}
	nw.Flush()

	if s.comparison(file, fileTmp) {
		msg = fmt.Sprintf("%s|%s, same md5, no need to update", filepath.Base(file), s.fileMd5(file))
		s.send(msg, stream)
		return
	}

	if err = os.Rename(fileTmp, file); err != nil {
		s.send(err.Error(), stream)
		return
	}

	msg = fmt.Sprintf("%s|%s", filepath.Base(file), s.fileMd5(file))
	s.send(msg, stream)

	return
}

func (s *server) send(msg string, stream pb.FileTransferService_SendFileServer) {
	if err := stream.Send(&pb.FileMessage{Byte: []byte("md5"), Name: msg}); err != nil {
		log.Println(err.Error())
	}
}

func (s *server) comparison(src, dst string) bool {
	return s.fileMd5(src) == s.fileMd5(dst)
}

func (s *server) fileMd5(file string) (m5 string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}

	defer f.Close()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	m5 = hex.EncodeToString(h.Sum(nil))

	return
}

// failList 如果文件重命名覆盖失败就记录，当前端执行更新操作时就可以忽略掉当前的服务器
func (s *server) failList(ip string) {
}

// grpc请求验证
func (s *server) verify(req *pb.StreamRequest, stream pb.StreamUpdateProgramService_JavaReloadServer) (err error) {
	if err = redis.NewRdbOp().ReqVerify("lxb", "ttt"); err != nil {
		if err = stream.Send(&pb.StreamReply{Message: err.Error()}); err != nil {
			return
		}
		return
	}
	return
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 12306))
	if err != nil {
		log.Fatalln(fmt.Sprintf("failed to listen 12306, errMsg: %v", err))
	}

	if err = redis.InitPoolRdb(); err != nil {
		log.Fatalln(fmt.Sprintf("failed to connect redis, errMsg: %s", err.Error()))
	}

	s := grpc.NewServer()
	pb.RegisterStreamUpdateProgramServiceServer(s, &server{})
	pb.RegisterFileTransferServiceServer(s, &server{})
	pb.RegisterStreamCheckSystemLogServiceServer(s, &server{})
	pb.RegisterClusterOperateServiceServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
