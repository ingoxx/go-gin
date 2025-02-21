package cluster

import (
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

var cm model.ClusterModel

func CheckClusterListController(ctx *gin.Context) {
	var cm model.ClusterModel
	var am model.AssetsModel
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
			"code":    10002,
		})
		return
	}

	servers, err := am.GetAllServersIp()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10003,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     data.ModelSlice,
		"total":    data.Total,
		"pageSize": data.PageSize,
		"servers":  servers,
		"code":     10000,
		"message":  "ok",
	})

	return
}

func CreateClusterController(ctx *gin.Context) {
	var ccs CreateClusterStruct
	if err := ccs.Create(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("集群: [%s], 添加成功", ccs.CC.Name),
		"code":    10000,
	})
	return
}

func JoinMasterClusterController(ctx *gin.Context) {
	var jj JoinJson
	if err := jj.JoinMaster(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("加入【%s】集群成功", jj.sw.Name),
		"code":    10000,
	})

	return
}

func JoinWorkClusterController(ctx *gin.Context) {
	var jj JoinJson
	if err := jj.JoinWork(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("加入【%s】集群成功", jj.sw.Name),
		"code":    10000,
	})

	return
}

func LeaveClusterController(ctx *gin.Context) {
	var js JoinJson
	if err := js.LeaveSwarm(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "移除成功",
		"code":    10000,
	})
	return
}

func DeleteClusterController(ctx *gin.Context) {
	var dc GenericClusterJson
	if err := dc.Delete(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("集群: %v, 删除成功", dc.deleteClusterName),
		"code":    10000,
	})
	return
}

func UpdateClusterController(ctx *gin.Context) {
	var uc UpdateClusterForm
	if err := uc.Update(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("集群: %s, 更新成功", uc.Name),
		"code":    10000,
	})
}

func HealthCheckController(ctx *gin.Context) {
	var dc GenericClusterJson
	if err := dc.StartHealthCheck(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("集群: %v, 监控启动成功", dc.deleteClusterName),
		"code":    10000,
	})
	return
}
