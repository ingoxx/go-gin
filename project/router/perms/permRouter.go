package perms

import (
	pc "github.com/Lxb921006/Gin-bms/project/controller/perms"

	"github.com/gin-gonic/gin"
)

func PermsRouter(r *gin.Engine) {
	perms := r.Group("/perms")
	{
		perms.GET("/list", pc.GetPermsList)
		perms.POST("/create", pc.CreatePermMenu)
		perms.POST("/delete", pc.DeletePermsMenu)
	}
}
