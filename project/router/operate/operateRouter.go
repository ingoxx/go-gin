package operate

import (
	oc "github.com/Lxb921006/Gin-bms/project/controller/operatelog"

	"github.com/gin-gonic/gin"
)

func OperateRouter(r *gin.Engine) {
	operate := r.Group("/log")
	{
		operate.GET("/list", oc.OperateLogList)
	}
}
