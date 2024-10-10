package assets

import (
	ac "github.com/Lxb921006/Gin-bms/project/controller/assets"
	"github.com/gin-gonic/gin"
)

func AssetsRouter(r *gin.Engine) {
	assets := r.Group("/assets")
	{
		// 执行后端程序的接口
		assets.GET("/ws", ac.RunProgramWsController)
		assets.GET("/file/ws", ac.SyncFilePassWsController)
		assets.GET("/process/status", ac.GetMissionStatusController)
		assets.GET("/process/update/list", ac.ProgramUpdateListController)
		assets.GET("/list", ac.AssetsListController)
		assets.POST("/process/update/create", ac.CreateUpdateProgramRecordController)
		assets.POST("/api", ac.RunProgramApiController)
		// 服务器的增删改查接口
		assets.POST("/upload", ac.UploadController)
		assets.POST("/add", ac.AssetsCreateController)
		assets.POST("/del", ac.AssetsDeleteController)
		assets.POST("/update", ac.AssetsModifyController)
	}
}
