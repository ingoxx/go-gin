package role

import (
	rc "github.com/ingoxx/go-gin/project/controller/role"

	"github.com/gin-gonic/gin"
)

func Router(r *gin.Engine) {
	role := r.Group("/role")
	{
		role.GET("/rolesname", rc.GetRolesInfo) //角色详情
		role.GET("/list", rc.GetRolesList)
		role.GET("/userperms", rc.GetUserPerms)  //用户权限查询
		role.GET("/roleperms", rc.GetRolePerms)  //角色权限查询
		role.GET("/pmenu", rc.GetAllFormatPerms) //权限菜单
		role.POST("/create", rc.CreateRole)
		role.POST("/delete", rc.DeleteRoles)
		role.POST("/allotperms", rc.AllotPermsToRole) //分配权限
		role.POST("/rmperms", rc.RemoveRolePerms)     //移除权限
	}
}
