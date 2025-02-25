package operate

import (
	oc "github.com/ingoxx/go-gin/project/controller/operatelog"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	operate := r.Group("/log")
	{
		operate.GET("/list", oc.LogListController)
		operate.GET("/get-login-num", oc.GetLoginNumDataController)
		operate.GET("/get-run-linux-cmd-num", oc.GetLinuxCmdDataController)
		operate.GET("/get-user-login-num", oc.GetUserLoginNumController)
	}
}
