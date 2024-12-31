package init

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
)

func AdminPerms() (err error) {
	var p []model.Permission
	if err := dao.DB.Where("id > ?", 0).Find(&p).Error; err != nil {
		return
	}

	var r model.Role
	if err := dao.DB.Where("rolename = ?", "管理员").Find(&r).Error; err != nil {
		return
	}

	return nil
}
