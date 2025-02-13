package cluster

import (
	cc "github.com/Lxb921006/Gin-bms/project/controller/cluster"
	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	assets := r.Group("/cluster")
	{
		assets.GET("/list", cc.CheckClusterListController)
		assets.POST("/add", cc.AddClusterController)
	}
}
