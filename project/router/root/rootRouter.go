package root

import (
	"net/http"
	"time"

	"github.com/Lxb921006/Gin-bms/project/middleware"
	"github.com/Lxb921006/Gin-bms/project/router/assets"
	"github.com/Lxb921006/Gin-bms/project/router/login"
	"github.com/Lxb921006/Gin-bms/project/router/operate"
	"github.com/Lxb921006/Gin-bms/project/router/perms"
	"github.com/Lxb921006/Gin-bms/project/router/role"
	"github.com/Lxb921006/Gin-bms/project/router/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *http.Server {
	// gin.SetMode(gin.ReleaseMode) 正式生产环境需切换到Release模式，测试是debug模式
	router := gin.Default()
	router.Static("/static", "../static")
	// router.LoadHTMLGlob("../../templates")
	router.Use(middleware.AllowCos(), middleware.TokenVerify(), middleware.PermsVerify(), middleware.Visitlimit(), middleware.OperateRecord())

	//加载路由配置
	user.UserRouter(router)
	role.RoleRouter(router)
	perms.PermsRouter(router)
	login.LoginRouter(router)
	operate.OperateRouter(router)
	assets.AssetsRouter(router)

	t := &http.Server{
		Addr:           ":9293",
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second, //响应时间
		MaxHeaderBytes: 8 << 20,          //body大小8M
	}

	return t
}
