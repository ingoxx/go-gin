package cluster

import (
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

func CheckClusterListController(ctx *gin.Context) {
	var cm model.ClusterModel
	var query CheckClusterQuery
	if err := ctx.ShouldBind(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	vd := NewValidateData(validate)
	if err := vd.ValidateStruct(query); err != nil {
		return
	}

	if err := mapstructure.Decode(query, &cm); err != nil {
		return
	}

	data, err := cm.List(query.Page, cm)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	fmt.Println("data >>> ", data)

	ctx.JSON(http.StatusOK, gin.H{
		"data":     data.ModelSlice,
		"total":    data.Total,
		"pageSize": data.PageSize,
		"code":     10000,
		"message":  "ok",
	})

	return
}

func AddClusterController(ctx *gin.Context) {
	var add AddClusterForm
	var cm model.ClusterModel
	if err := ctx.ShouldBind(&add); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	if err := mapstructure.Decode(add, &cm); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("集群: [%s] 创建失败, errMsg: %v", add.Name, err.Error()),
			"code":    10002,
		})
		return
	}

	if err := cm.Add(cm); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("集群: [%s] 创建失败, errMsg: %v", add.Name, err.Error()),
			"code":    10003,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("集群: [%s], 添加成功", add.Name),
		"code":    10000,
	})
	return

}
