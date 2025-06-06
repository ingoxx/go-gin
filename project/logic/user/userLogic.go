package user

import (
	"github.com/ingoxx/go-gin/project/api"
	"github.com/ingoxx/go-gin/project/model"
	"github.com/ingoxx/go-gin/project/service"
	"github.com/mitchellh/mapstructure"
)

// UserListQuery 这里是存放user这个app的复杂逻辑代码
type UserListQuery struct {
	ID       uint              `json:"id"`
	CurPage  int               `json:"page" form:"page"  validate:"min=1" binding:"required"`
	Name     string            `json:"name" form:"name"`
	Email    string            `json:"email" form:"email"`
	Tel      int               `json:"tel" form:"tel"`
	Isopenga uint              `json:"isopenga" form:"isopenga"`
	Isopenqr uint              `json:"isopenqr" form:"isopenqr"`
	RoleName string            `json:"rolename" form:"rolename"`
	PageData *service.Paginate `form:"-"`
}

func (ul *UserListQuery) PaginateLogic(u model.User, api api.Api) (err error) {
	//验证器
	if err = api.ValidateStruct(ul); err != nil {
		return
	}

	if err = mapstructure.Decode(ul, &u); err != nil {
		return
	}

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
