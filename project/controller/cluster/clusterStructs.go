package cluster

import (
	pb "github.com/Lxb921006/Gin-bms/project/command/command"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()
var cms model.ClusterModel

type CreateClusterStruct struct {
	CC CreateClusterJson
	CM model.ClusterModel
	SW SwarmOperate
}

type CreateClusterJson struct {
	Name        string            `form:"name" json:"name" binding:"required"`
	Region      string            `form:"region" json:"region" binding:"required"`
	MasterIp    string            `form:"master_ip" json:"master_ip"  binding:"required"`
	ServersInfo []ServerNodeInput `form:"servers" json:"servers" binding:"required"`
	ClusterType string            `form:"cluster_type" json:"cluster_type" binding:"required"`
}

type ServerNodeInput struct {
	Ip       string `json:"ip"`
	NodeType uint   `json:"node_type"`
}

type LeaveClusterJson struct {
	ServersInfo []ServerNodeInput `form:"servers" json:"servers" binding:"required"`
}

type SwarmOperate struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MasterIp    string `json:"master_ip"`
	ClusterType string `json:"cluster_type"`
	Region      string `json:"region"`
	WorkToken   string `json:"work_token"`
	MasterToken string `json:"master_token"`
	ClusterCid  string `json:"cluster_cid"`
	Message     string `json:"message"`
	Code        int32  `json:"code"`
	ds          pb.ClusterOperateServiceClient
	ServersInfo []ServerNodeInput `form:"servers" json:"servers"`
}

type GeneralClusterOpStruct struct {
	sw SwarmOperate
	cm model.ClusterModel
}

type JoinJson struct {
	ID          uint              `form:"id" json:"id"  binding:"required"`
	MasterIp    string            `form:"master_ip" json:"master_ip"  binding:"required"`
	ServersInfo []ServerNodeInput `form:"servers" json:"servers" binding:"required"`
	GeneralClusterOpStruct
	ctx *gin.Context
}

type DeleteSwarmJson struct {
	ID                []uint `form:"id" json:"id"  binding:"required"`
	serversInfo       []ServerNodeInput
	gs                GeneralClusterOpStruct
	deleteClusterName []string
}

type GeneralClusterFieldStruct struct {
	Name        string `form:"name" json:"name"`
	Region      string `form:"region" json:"region"`
	ClusterCid  string `form:"cluster_cid" json:"cluster_cid"`
	ClusterType string `form:"cluster_type" json:"cluster_type"`
}

type CheckClusterQuery struct {
	Page int `form:"page" validate:"min=1" binding:"required" json:"page"`
	//GeneralClusterFieldStruct
	Name        string `form:"name" json:"name"`
	Region      string `form:"region" json:"region"`
	ClusterCid  string `form:"cluster_cid" json:"cluster_cid"`
	ClusterType string `form:"cluster_type" json:"cluster_type"`
}

type UpdateClusterForm struct {
	ID          uint   `json:"id" form:"id" binding:"required"`
	Name        string `form:"name" json:"name"`
	Region      string `form:"region" json:"region"`
	ClusterCid  string `form:"cluster_cid" json:"cluster_cid"`
	ClusterType string `form:"cluster_type" json:"cluster_type"`
}
