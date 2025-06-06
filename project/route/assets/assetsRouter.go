package assets

import (
	"github.com/gin-gonic/gin"
	ac "github.com/ingoxx/go-gin/project/controller/assets"
)

func Router(r *gin.Engine) {
	assets := r.Group("/assets")
	{
		// 执行后端程序的接口
		assets.GET("/run-linux-cmd", ac.RunProgramController)
		assets.GET("/view-system-log", ac.RunProgramController)
		assets.GET("/ws", ac.RunProgramController)
		assets.GET("/file/ws", ac.SyncFileController)
		//assets.GET("/program/status", ac.GetMissionStatusController)
		assets.GET("/program/update/list", ac.ProgramUpdateListController)
		//assets.GET("/list", ac.ListController)
		assets.GET("/list", ac.ListController2)
		assets.POST("/program/update/create", ac.CreateUpdateProgramRecordController)
		assets.POST("/api", ac.RunProgramApiController)
		assets.POST("/program/add", ac.AddProgramOperateController)
		assets.POST("/program/del", ac.DelProgramUpdateRecordController)
		assets.GET("/program/list", ac.ProgramListController)
		//assets.Any("/terminal/*path", ac.WebTerminalController)
		assets.GET("/terminal", ac.WebTerminalController)
		// 服务器的增删改查接口
		assets.POST("/upload", ac.UploadController)
		assets.POST("/add", ac.CreateController)
		assets.POST("/del", ac.DeleteController)
		assets.POST("/update", ac.UpdateController)
	}
}
