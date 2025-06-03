package cluster

import (
	"github.com/gin-gonic/gin"
	cc "github.com/ingoxx/go-gin/project/controller/cluster"
)

func Router(r *gin.Engine) {
	assets := r.Group("/cluster")
	{
		assets.GET("/list", cc.CheckClusterListController)
		assets.POST("/add", cc.CreateClusterController)
		assets.POST("/del", cc.DeleteClusterController)
		assets.POST("/update", cc.UpdateClusterController)
		assets.POST("/join-work", cc.JoinWorkClusterController)
		assets.POST("/join-master", cc.JoinMasterClusterController)
		assets.POST("/leave-cluster", cc.LeaveClusterController)
		//assets.POST("/health-check", cc.HealthCheckController)
	}
}
