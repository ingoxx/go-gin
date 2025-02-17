package cluster

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type AddClusterForm struct {
	Name        string   `form:"name" json:"name" binding:"required"`
	Region      string   `form:"region" json:"region" binding:"required"`
	MasterIp    string   `form:"master_ip" json:"master_ip" binding:"required"`
	WorkIp      []string `form:"work_ip" json:"work_ip" binding:"required"`
	ClusterType string   `form:"cluster_type" json:"cluster_type" binding:"required"`
}

type CheckClusterQuery struct {
	Page        int    `form:"page" validate:"min=1" binding:"required" json:"page"`
	Name        string `form:"name" json:"name"`
	Region      string `form:"region" json:"region"`
	ClusterCid  string `form:"cluster_cid" json:"cluster_cid"`
	ClusterType string `form:"cluster_type" json:"cluster_type"`
}
