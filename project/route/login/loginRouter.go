package login

import (
	lc "github.com/Lxb921006/Gin-bms/project/controller/login"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	r.POST("/login", lc.Login)
	r.POST("/galogin", lc.MFAVerify)
	r.POST("/logout", lc.Logout)
}
