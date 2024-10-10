package user

import (
	"strconv"

	"github.com/Lxb921006/Gin-bms/project/api"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
)

// 这里是存放user这个app的复杂逻辑代码
type UserListQuery struct {
	CurPage  int               `form:"page"  validate:"min=1" binding:"required"`
	Name     string            `form:"name"`
	Email    string            `form:"email"`
	Tel      string            `form:"tel"`
	Isopenga uint              `form:"isopenga"`
	Isopenqr uint              `form:"isopenqr"`
	RoleName string            `form:"rolename"`
	PageData *service.Paginate `form:"-"`
}

func (ul *UserListQuery) PaginateLogic(u model.User, api api.Api) (err error) {
	//验证器
	if err = api.ValidateStruct(ul); err != nil {
		return
	}

	u.Name = ul.Name
	u.Email = ul.Email
	u.Tel, _ = strconv.Atoi(ul.Tel)
	u.Isopenga = ul.Isopenga
	u.Isopenqr = ul.Isopenqr

	if ul.RoleName == "" {
		ul.PageData, err = u.GetUserByPaginate(ul.CurPage, u)
		if err != nil {
			return
		}
	} else {
		ul.PageData, err = u.GetUserByMmPaginate(ul.CurPage, ul.RoleName, u)
		if err != nil {
			return
		}
	}

	return
}
