package user

import (
	uc "github.com/Lxb921006/Gin-bms/project/controller/user"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.Engine) {
	user := r.Group("/user")
	{
		user.POST("/add", uc.AddUser)
		user.POST("/del", uc.DeleteUser)
		user.POST("/update", uc.UpdateUser)
		user.GET("/getinfobyname", uc.GetUserByName)
		user.GET("/list", uc.GetUsersByPaginate)
	}
}
