package root

import (
	"github.com/ingoxx/go-gin/project/route/cluster"
	"net/http"
	"time"

	"github.com/ingoxx/go-gin/project/middleware"
	"github.com/ingoxx/go-gin/project/route/assets"
	"github.com/ingoxx/go-gin/project/route/login"
	"github.com/ingoxx/go-gin/project/route/operate"
	"github.com/ingoxx/go-gin/project/route/perms"
	"github.com/ingoxx/go-gin/project/route/role"
	"github.com/ingoxx/go-gin/project/route/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *http.Server {
	// gin.SetMode(gin.ReleaseMode) 正式生产环境需切换到Release模式，测试是debug模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Static("/static", "../static")
	// route.LoadHTMLGlob("../../templates")
	router.Use(middleware.AllowCos(),
		middleware.TokenVerify(),
		middleware.PermsVerify(),
		//middleware.ReqFrequencyLimit(),
		//middleware.OperateRecord(),
	)

	//加载路由配置
	user.Router(router)
	role.Router(router)
	perms.Router(router)
	login.Router(router)
	operate.Router(router)
	assets.Router(router)
	cluster.Router(router)

	t := &http.Server{
		Addr:           ":9293",
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second, //响应时间
		MaxHeaderBytes: 8 << 20,          //body大小8M
	}

	return t
}
