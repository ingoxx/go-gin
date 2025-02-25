package user

import (
	uc "github.com/ingoxx/go-gin/project/controller/user"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	user := r.Group("/user")
	{
		user.POST("/add", uc.AddUser)
		user.POST("/del", uc.DeleteUser)
		user.POST("/update", uc.UpdateUser)
		user.POST("/update-pwd", uc.UpdateUserPwd)
		user.GET("/getinfobyname", uc.GetUserByName)
		user.GET("/list", uc.GetUsersByPaginate)
	}
}
