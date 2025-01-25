package root

import (
	"net/http"
	"time"

	"github.com/Lxb921006/Gin-bms/project/middleware"
	"github.com/Lxb921006/Gin-bms/project/route/assets"
	"github.com/Lxb921006/Gin-bms/project/route/login"
	"github.com/Lxb921006/Gin-bms/project/route/operate"
	"github.com/Lxb921006/Gin-bms/project/route/perms"
	"github.com/Lxb921006/Gin-bms/project/route/role"
	"github.com/Lxb921006/Gin-bms/project/route/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *http.Server {
	// gin.SetMode(gin.ReleaseMode) 正式生产环境需切换到Release模式，测试是debug模式
	router := gin.Default()
	router.Static("/static", "../static")
	// route.LoadHTMLGlob("../../templates")
	router.Use(middleware.AllowCos(), middleware.TokenVerify(), middleware.PermsVerify(), middleware.ReqFrequencyLimit(), middleware.OperateRecord())

	//加载路由配置
	user.Router(router)
	role.Router(router)
	perms.Router(router)
	login.Router(router)
	operate.Router(router)
	assets.Router(router)

	t := &http.Server{
		Addr:           ":9293",
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second, //响应时间
		MaxHeaderBytes: 8 << 20,          //body大小8M
	}

	return t
}
