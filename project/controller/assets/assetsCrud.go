package assets

import (
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/Lxb921006/Gin-bms/project/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

//资产列表的crud-1

type AssetsListForm struct {
	Ip      string `form:"ip,omitempty" json:"ip"`
	Project string `form:"project,omitempty" json:"project"`
	Page    int    `form:"page" json:"page" validate:"min=1" binding:"required"`
}

func (a *AssetsListForm) List(ctx *gin.Context) (data *service.Paginate, err error) {
	var al model.AssetsModel
	if err = ctx.ShouldBind(a); err != nil {
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err = vd.ValidateStruct(a); err != nil {
		return
	}

	if err = utils.CopyStruct(a, &al); err != nil {
		return
	}

	data, err = al.List(a.Page, al)
	if err != nil {
		return
	}

	return
}

type AssetsCreateForm struct {
	Ip      []string `form:"ip" json:"ip"`
	Project string   `form:"project" json:"project"`
}

func (a *AssetsCreateForm) Create(ctx *gin.Context) (err error) {
	var am model.AssetsModel
	var aml []*model.AssetsModel
	if err = ctx.ShouldBindJSON(a); err != nil {
		return
	}

	for _, ip := range a.Ip {
		data := &model.AssetsModel{
			Project: a.Project,
			Ip:      ip,
		}

		aml = append(aml, data)
	}

	if err = am.Create(aml); err != nil {
		return
	}

	return
}

type AssetsDelForm struct {
	Ips []string `form:"ips" json:"ips" binding:"required"`
}

func (a *AssetsDelForm) Del(ctx *gin.Context) (err error) {
	var am model.AssetsModel
	if err = ctx.BindJSON(a); err != nil {
		return
	}

	if err = am.Del(a.Ips); err != nil {
		return
	}

	return
}

type AssetsModifyForm struct {
	Id      int64  `form:"id" json:"id"`
	Ip      string `form:"ip" json:"ip"`
	Project string `form:"project" json:"project"`
}

func (aef *AssetsModifyForm) Modify(ctx *gin.Context) (err error) {
	var am model.AssetsModel
	var data = make(map[string]interface{})
	if err = ctx.ShouldBind(aef); err != nil {
		return
	}

	data["id"] = aef.Id
	data["ip"] = aef.Ip
	data["project"] = aef.Project

	if err = am.Modify(data); err != nil {
		return
	}

	return
}
