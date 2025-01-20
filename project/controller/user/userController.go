package user

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/http"

	"github.com/Lxb921006/Gin-bms/project/logic/user"
	"github.com/Lxb921006/Gin-bms/project/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
)

type CreateUserForm struct {
	Name       string `json:"name" form:"name" binding:"required"`
	Email      string `json:"email" form:"email" binding:"required"`
	RoleId     uint   `json:"roleId" form:"roleId" binding:"required"`
	Tel        int    `json:"tel" form:"tel" binding:"required" validate:"min=1500000"`
	Isopenga   uint   `json:"isopenga" form:"isopenga"`
	Isopenqr   uint   `json:"isopenqr" form:"isopenqr"`
	MfaApp     uint   `json:"mfa_app" form:"mfa_app"`
	Password   string `json:"password" form:"password" binding:"required"`
	RePassword string `json:"rePassword" form:"rePassword" binding:"required" validate:"eqfield=Password"`
}

type UpdateUserForm struct {
	Name       string `json:"name" form:"name"`
	Uid        uint   `json:"uid" form:"uid" binding:"required"`
	Password   string `json:"password" form:"password"`
	RePassword string `json:"rePassword" form:"rePassword"  validate:"eqfield=Password"`
	Rid        uint   `json:"rid" form:"rid" binding:"required"`
	Isopenga   uint   `json:"isopenga" form:"isopenga"`
	Isopenqr   uint   `json:"isopenqr" form:"isopenqr"`
	MfaApp     uint   `json:"mfa_app" form:"mfa_app"`
}

type DelUserByIdJson struct {
	Uid []uint `json:"uid" form:"uid" binding:"required" validate:"containsAdminUid"`
}

type GetUserByNameQuery struct {
	Name string `form:"name" binding:"required"`
}

func AddUser(ctx *gin.Context) {
	var u model.User
	var addUser CreateUserForm

	if err := ctx.ShouldBind(&addUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	if err := NewValidateData(validate).ValidateStruct(addUser); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg: %v", addUser.Name, err.Error()),
			"code":    10002,
		})
		return
	}

	if err := mapstructure.Decode(addUser, &u); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg: %v", addUser.Name, err.Error()),
			"code":    10003,
		})
		return
	}

	if err := u.AddUser(u, addUser.RoleId); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg: %v", addUser.Name, err.Error()),
			"code":    10004,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("添加%v成功", addUser.Name),
		"code":    10000,
	})

}

func DeleteUser(ctx *gin.Context) {
	var u model.User
	var delUser DelUserByIdJson
	if err := ctx.ShouldBindJSON(&delUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	var vd = NewValidateData(validate)
	if err := vd.RegisterValidation(); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	users, err := u.GetUserNameById(delUser.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10003,
		})
		return
	}

	if err := vd.ValidateStruct(delUser); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10004,
		})
		return
	}

	if err := u.DeleteUser(delUser.Uid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10005,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v删除成功", users),
		"code":    10000,
	})

}

func UpdateUser(ctx *gin.Context) {
	var u model.User
	var ud UpdateUserForm
	if e := ctx.ShouldBind(&ud); e != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": e.Error(),
			"code":    10001,
		})
		return
	}

	if err := NewValidateData(validate).ValidateStruct(ud); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("更新%v失败, errMsg: %v", ud.Name, err.Error()),
			"code":    10002,
		})
		return
	}

	if err := mapstructure.Decode(ud, &u); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("更新%v失败, errMsg: %v", ud.Name, err.Error()),
			"code":    10003,
		})
		return
	}

	if err := u.UpdateUser(u, ud.Rid, ud.Uid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("更新%v失败, errMsg: %s", ud.Name, err.Error()),
			"code":    10004,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("更新%v成功", ud.Name),
		"code":    10000,
	})

}

func GetUserByName(ctx *gin.Context) {
	var u model.User
	var guf GetUserByNameQuery
	if e := ctx.ShouldBindQuery(&guf); e != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": e.Error(),
			"code":    10001,
		})
		return
	}

	r, e := u.GetUserByName(guf.Name)
	if e != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": e.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": r,
		"code": 10000,
	})

}

func GetUsersByPaginate(ctx *gin.Context) {
	var pd user.UserListQuery
	var u model.User
	if err := ctx.ShouldBindQuery(&pd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	var vd = NewValidateData(validate)
	if err := pd.PaginateLogic(u, vd); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     pd.PageData.ModelSlice,
		"total":    pd.PageData.Total,
		"pageSize": pd.PageData.PageSize,
		"code":     10000,
	})

}
