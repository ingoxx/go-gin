package perms

import (
	pc "github.com/ingoxx/go-gin/project/controller/perms"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	perms := r.Group("/perms")
	{
		perms.GET("/list", pc.GetPermsList)
		perms.POST("/create", pc.CreatePermMenu)
		perms.POST("/delete", pc.DeletePermsMenu)
	}
}
