package role

import (
	"fmt"
	"net/http"

	"github.com/Lxb921006/Gin-bms/project/model"

	"github.com/Lxb921006/Gin-bms/project/logic/role"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type CreateRoleForm struct {
	RoleName string `form:"rolename" binding:"required"`
}

type DeleteRoleJson struct {
	Rid []uint `form:"rid" binding:"required"`
}

type RoleListQuery struct {
	RoleName string `form:"rolename"`
	Page     int    `form:"page" validate:"min=1" binding:"required"`
}

type OperatePermsJson struct {
	Rid      uint   `form:"rid" binding:"required"`
	Pid      []uint `form:"pid" binding:"required"`
	RoleName string `form:"rolename" binding:"required"`
}

type UserPermsQuery struct {
	Uid uint `form:"uid" binding:"required"`
}

type RolePermsQuery struct {
	Rid uint `form:"rid" binding:"required"`
}

func CreateRole(ctx *gin.Context) {
	var r model.Role
	var cr CreateRoleForm

	if err := ctx.ShouldBindWith(&cr, binding.Form); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50001,
		})
		return
	}

	r.RoleName = cr.RoleName

	if err := r.CreateRole(r); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("[%s] 创建失败, errMsg: %s", cr.RoleName, err.Error()),
			"code":    50002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("[%s] 创建成功", cr.RoleName),
		"code":    10000,
	})
}

func DeleteRoles(ctx *gin.Context) {
	var r model.Role
	var dr DeleteRoleJson

	if err := ctx.ShouldBindJSON(&dr); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50003,
		})
		return
	}

	if err := r.DeleteRole(dr.Rid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50004,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v删除成功", r.GetRoleNames(dr.Rid)),
		"code":    10000,
	})
}

func GetRolesInfo(ctx *gin.Context) {
	var r model.Role
	data, err := r.GetAllRoles()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50005,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    data,
		"code":    10000,
		"message": "ok",
	})
}

func AllotPermsToRole(ctx *gin.Context) {
	var ap OperatePermsJson
	var r model.Role

	if err := ctx.ShouldBindJSON(&ap); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50006,
		})
		return
	}

	if err := role.UpdateUserPerms(ap.Pid, ap.Rid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50007,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("角色: %s, 权限已更新: %v", ap.RoleName, r.GetPermNames(ap.Pid)),
		"code":    10000,
	})
}

func RemoveRolePerms(ctx *gin.Context) {
	var r model.Role
	var rp OperatePermsJson

	if err := ctx.ShouldBindJSON(&rp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50008,
		})
		return
	}

	data, err := r.RemovePerms(rp.Rid, rp.Pid)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50009,
		})
		return
	}

	mp := r.FormatUserPerms(data, 0)

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("角色: %s, 权限已移除: %v", rp.RoleName, r.GetPermNames(rp.Pid)),
		"data":    mp,
		"code":    10000,
	})
}

func GetRolesList(ctx *gin.Context) {
	var r model.Role
	var rp RoleListQuery
	if err := ctx.ShouldBindQuery(&rp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50010,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := vd.ValidateStruct(rp); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50011,
		})
		return
	}

	r.RoleName = rp.RoleName

	data, err := r.GetRolesList(rp.Page, r)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50012,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     data.ModelSlice,
		"total":    data.Total,
		"pageSize": data.PageSize,
		"code":     10000,
		"message":  "ok",
	})
}

func GetUserPerms(ctx *gin.Context) {
	var r model.Role
	var up UserPermsQuery

	if err := ctx.ShouldBindQuery(&up); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10012,
		})
		return
	}

	data, err := r.GetUserPerms(up.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10013,
		})
		return
	}

	menuData := r.FormatUserPerms(data, 0)

	if len(menuData) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "当前用户还未授权, 请联系管理员",
			"code":    10013,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    menuData,
		"code":    10000,
		"message": "ok",
	})
}

func GetRolePerms(ctx *gin.Context) {
	var r model.Role
	var up RolePermsQuery
	var pidList []uint

	if err := ctx.ShouldBindQuery(&up); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    50014,
		})
		return
	}

	data, err := r.GetRolePerms(up.Rid)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50015,
		})
		return
	}

	//var fData []role.Menu
	//if len(data) != 0 {
	//	fData = role.FormatUserPerms(data, 0)
	//	for _, v := range data {
	//		pidList = append(pidList, v.ID)
	//	}
	//} else {
	//	fData = []role.Menu{}
	//}

	fData := r.FormatUserPerms(data, 0)
	if len(fData) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "当前用户还未授权, 请联系管理员",
			"code":    10013,
		})
		return
	}

	for _, v := range data {
		pidList = append(pidList, v.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    fData,
		"pidList": pidList,
		"code":    10000,
		"message": "ok",
	})
}

func GetAllFormatPerms(ctx *gin.Context) {
	var r model.Role
	data, err := r.GetAllFormatPerms()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    50016,
		})
		return
	}

	//var fdata []role.Menu
	//
	//if len(data) != 0 {
	//	fdata = role.FormatUserPerms(data, 0)
	//
	//} else {
	//	fdata = []role.Menu{}
	//}

	fData := role.FormatUserPerms(data, 0)
	if len(fData) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "当前用户还未授权, 请联系管理员",
			"code":    10013,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    fData,
		"code":    10000,
		"message": "ok",
	})
}
