package cluster

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (uc *UpdateClusterForm) Update(ctx *gin.Context) error {
	if err := ctx.ShouldBind(uc); err != nil {
		return err
	}

	if err := mapstructure.Decode(uc, &cms); err != nil {
		return fmt.Errorf("集群: [%s] 更新失败, errMsg: %v", uc.Name, err.Error())
	}

	if err := cms.Update(uc.ID, cms); err != nil {
		return fmt.Errorf("集群: [%s] 更新失败, errMsg: %v", uc.Name, err.Error())
	}

	return nil
}
