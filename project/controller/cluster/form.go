package cluster

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type AddClusterForm struct {
	Name       string `form:"name" json:"name" binding:"required"`
	Region     string `form:"region" json:"region" binding:"required"`
	ClusterCid string `form:"cluster_cid" json:"cluster_cid" binding:"required"`
}

type CheckClusterQuery struct {
	Page       int    `form:"page" validate:"min=1" binding:"required" json:"page"`
	Name       string `form:"name" json:"name"`
	Region     string `form:"region" json:"region"`
	ClusterCid string `form:"cluster_cid" json:"cluster_cid"`
}
