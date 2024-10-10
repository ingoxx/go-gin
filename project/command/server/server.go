package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/utils"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

type server struct {
	pb.UnimplementedStreamUpdateProcessServiceServer
	pb.UnimplementedFileTransferServiceServer
}

func (s *server) DockerUpdate(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_DockerUpdateServer) (err error) {
	log.Println("rev run DockerUpdate")

	file := fmt.Sprintf("sh /root/shellscript/DockerUpdate.sh %s | tee /root/shellscript/DockerUpdate.log", req.GetUuid())
	cmd := exec.Command("sh", "-c", file)
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

func (s *server) DockerUpdateLog(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_DockerUpdateLogServer) (err error) {
	log.Println("rev run DockerUpdateLog")

	cmd := exec.Command("more", "/root/shellscript/DockerUpdate.log", req.GetUuid())
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

func (s *server) JavaUpdateLog(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_JavaUpdateLogServer) (err error) {
	log.Println("rev run JavaUpdateLog")

	cmd := exec.Command("more", "/root/shellscript/JavaUpdate.log", req.GetUuid())
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

func (s *server) JavaUpdate(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_JavaUpdateServer) (err error) {
	log.Println("rev run JavaUpdate")

	file := fmt.Sprintf("sh /root/shellscript/JavaUpdate.sh %s | tee /root/shellscript/JavaUpdate.log", req.GetUuid())
	cmd := exec.Command("sh", "-c", file)
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

func (s *server) DockerReload(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_DockerReloadServer) (err error) {
	return
}

func (s *server) JavaReload(req *pb.StreamRequest, stream pb.StreamUpdateProcessService_JavaReloadServer) (err error) {
	return
}

func (s *server) SendFile(stream pb.FileTransferService_SendFileServer) (err error) {
	if err = s.ProcessMsg(stream); err != nil {
		log.Println(err)
	}

	return
}

func (s *server) ProcessMsg(stream pb.FileTransferService_SendFileServer) (err error) {
	var file string
	var chunks [][]byte
	path := "/opt"

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if file == "" {
			log.Println("rec file >>>", resp.GetName())
			file = filepath.Join(path, resp.GetName())
		}

		chunks = append(chunks, resp.Byte)
	}

	_, err = os.Stat(file)
	if err != nil {
		fw, err := os.Create(file)
		if err != nil {
			utils.Error(err.Error())
		}
		defer fw.Close()

		nw := bufio.NewWriter(fw)

		for _, chunk := range chunks {
			_, err := nw.Write(chunk)
			if err != nil {
				utils.Error(err.Error())
			}
		}

		nw.Flush()
	}

	m, _ := s.FileMd5(file)

	if err = stream.Send(&pb.FileMessage{Byte: []byte("md5"), Name: filepath.Base(file) + "|" + m}); err != nil {
		utils.Error(err.Error())
	}

	log.Printf("%s %s send ok", file, m)

	return
}

func (s *server) FileMd5(file string) (m5 string, err error) {
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

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 12306))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	utils.SetLogFile("./rpc_file.log")
	utils.SetLogLevel(utils.ErrorLevel)

	s := grpc.NewServer()

	pb.RegisterStreamUpdateProcessServiceServer(s, &server{})
	pb.RegisterFileTransferServiceServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
