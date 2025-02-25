package login

import (
	lc "github.com/ingoxx/go-gin/project/controller/login"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	r.POST("/login", lc.Login)
	r.POST("/galogin", lc.MFAVerify)
	r.POST("/logout", lc.Logout)
}
