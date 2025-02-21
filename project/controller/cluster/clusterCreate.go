package cluster

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (ccs *CreateClusterStruct) Create(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(&ccs.CC); err != nil {
		return err
	}

	if err := mapstructure.Decode(ccs.CC, &ccs.SW); err != nil {
		return fmt.Errorf("集群: [%s] 创建失败, errMsg: %v", ccs.CC.Name, err.Error())
	}

	if err := ccs.SW.CreateCluster(); err != nil {
		return fmt.Errorf("集群: [%s] 创建失败, errMsg: %v", ccs.CC.Name, err.Error())
	}

	if err := mapstructure.Decode(&ccs.SW, &ccs.CM); err != nil {
		return fmt.Errorf("集群: [%s] 创建失败, errMsg: %v", ccs.CC.Name, err.Error())
	}

	if err := ccs.CM.Create(&ccs.CM); err != nil {
		return fmt.Errorf("集群: [%s] 创建失败, errMsg: %v", ccs.CC.Name, err.Error())
	}

	ccs.SW.ID = ccs.CM.ID
	if err := ccs.SW.UpdateServers(); err != nil {
		return fmt.Errorf("服务器更新失败, errMsg: %v", err.Error())
	}

	return nil
}
