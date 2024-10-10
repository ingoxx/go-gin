package role

import (
	rc "github.com/Lxb921006/Gin-bms/project/controller/role"

	"github.com/gin-gonic/gin"
)

func RoleRouter(r *gin.Engine) {
	role := r.Group("/role")
	{
		role.GET("/rolesname", rc.GetRolesInfo) //角色详情
		role.GET("/list", rc.GetRolesList)
		role.GET("/userperms", rc.GetUserPerms)  //用户权限
		role.GET("/roleperms", rc.GetRolePerms)  //角色权限
		role.GET("/pmenu", rc.GetAllFormatPerms) //权限菜单(格式化后的权限列表)
		role.POST("/create", rc.CreateRole)
		role.POST("/delete", rc.DeleteRoles)
		role.POST("/allotperms", rc.AllotPermsToRole) //分配权限
		role.POST("/rmperms", rc.RemoveRolePerms)     //移除权限
	}
}
