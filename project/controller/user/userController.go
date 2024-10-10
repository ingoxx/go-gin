package user

import (
	"fmt"
	"net/http"

	"github.com/Lxb921006/Gin-bms/project/logic/user"
	"github.com/Lxb921006/Gin-bms/project/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CreateUserForm struct {
	Name       string `form:"name" binding:"required"`
	Email      string `form:"email" binding:"required"`
	RoleId     uint   `form:"roleId" binding:"required"`
	Tel        int    `form:"tel" binding:"required" validate:"min=1500000"`
	Isopenga   uint   `form:"isopenga"`
	Isopenqr   uint   `form:"isopenqr"`
	Password   string `form:"password" binding:"required"`
	RePassword string `form:"rePassword" binding:"required" validate:"eqfield=Password"`
}

type UpdateUserForm struct {
	Name       string `form:"name"`
	Uid        uint   `form:"uid" binding:"required"`
	Password   string `form:"password"`
	RePassword string `form:"rePassword"`
	Rid        uint   `form:"rid" binding:"required"`
	Isopenga   uint   `form:"isopenga"`
	Isopenqr   uint   `form:"isopenqr"`
}

type DelUserByIdJson struct {
	Uid []uint `form:"uid" json:"uid" binding:"required" validate:"contains=1"` //我是super user, 防止把自己给误删
}

type GetUserByNameQuery struct {
	Name string `form:"name" binding:"required"`
}

func AddUser(ctx *gin.Context) {
	var u model.User
	var ud CreateUserForm

	if err := ctx.ShouldBind(&ud); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    60001,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := vd.ValidateStruct(ud); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg=%v", ud.Name, err.Error()),
			"code":    60002,
		})
		return
	}

	u.Name = ud.Name
	u.Email = ud.Email
	u.Tel = ud.Tel
	u.Isopenga = ud.Isopenga
	u.Isopenqr = ud.Isopenqr
	u.Password = ud.Password

	if err := u.AddUser(u, ud.RoleId); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg=%v", ud.Name, err.Error()),
			"code":    60003,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("添加%v成功", ud.Name),
		"code":    10000,
	})

}

func DeleteUser(ctx *gin.Context) {
	var u model.User
	var ud DelUserByIdJson
	if err := ctx.ShouldBindJSON(&ud); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    60004,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := vd.ValidateStruct(ud); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "该用户不能删除, 他是大哥",
			"code":    60002,
		})
		return
	}

	if err := u.DeleteUser(ud.Uid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    60005,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v删除成功", ud.Uid),
		"code":    10000,
	})

}

func UpdateUser(ctx *gin.Context) {

	var u model.User
	var ud UpdateUserForm
	if e := ctx.ShouldBind(&ud); e != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": e.Error(),
			"code":    60006,
		})
		return
	}

	if ud.Password != ud.RePassword {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "密码不一致",
			"code":    60007,
		})
		return
	}

	u.Password = ud.RePassword
	u.Isopenga = ud.Isopenga
	u.Isopenqr = ud.Isopenqr
	u.ID = ud.Uid

	if err := u.UpdateUser(u, ud.Rid); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("更新%v失败, errMsg=%v", ud.Name, err.Error()),
			"code":    60008,
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
			"code":    60009,
		})
		return
	}

	r, e := u.GetUserByName(guf.Name)
	if e != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": e.Error(),
			"code":    60010,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"data": r,
			"code": 10000,
		})
	}

}

func GetUsersByPaginate(ctx *gin.Context) {
	var pd user.UserListQuery
	var u model.User
	if err := ctx.ShouldBindQuery(&pd); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    60011,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := pd.PaginateLogic(u, vd); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    60012,
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
