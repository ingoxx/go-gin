package middleware

import (
	"net/http"

	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/gin-gonic/gin"
)

// 中间件
// 允许跨域访问
func AllowCos() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if method := ctx.Request.Method; method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}

		ctx.Next()
	}
}

func TokenVerify() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if service.W.WhileList(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		token := ctx.Query("token")
		user := ctx.Query("user")
		if token == "" || user == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "非法请求, 参数缺失",
			})
			ctx.Abort()
		} else {
			if err := dao.Rds.RquestVerify(user, token); err == nil {
				ctx.Next()
			} else {
				ctx.JSON(http.StatusBadGateway, gin.H{
					"message": err.Error(),
				})
				ctx.Abort()
			}
		}
	}
}

func PermsVerify() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if service.W.WhileList(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		var p []model.Permission
		var whileUrl []string
		pass := false
		user := ctx.Query("user")

		err := dao.DB.Model(&model.Permission{}).
			Joins("inner join role_permissions on role_permissions.permission_id = permissions.id").
			Joins("inner join role_users on role_users.role_id = role_permissions.role_id").
			Joins("inner join users on role_users.user_id = users.id and users.name = ?", user).
			Select("permissions.path").Find(&p).Error

		if err != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": err.Error(),
			})
			ctx.Abort()
			return
		}

		for _, v := range p {
			whileUrl = append(whileUrl, v.Path)
		}

		url := ctx.Request.URL.Path
		for i := 0; i < len(whileUrl); i++ {
			if url == whileUrl[i] {
				pass = true
				break
			}
		}

		if pass {
			ctx.Next()
		} else {
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "您没有权限操作!",
			})
			ctx.Abort()
		}
	}
}

func Visitlimit() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := dao.Rds.Visitlimit(ctx.Request.Host); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			ctx.Abort()
			return
		}
	}
}

func OperateRecord() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		url := ctx.Request.URL.Path
		op := model.OperateLog{}
		if service.W.OperateWhileList(url) {
			ctx.Next()
			return
		}

		if err := op.AddOperateLog(ctx); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			ctx.Abort()
		}
	}
}
