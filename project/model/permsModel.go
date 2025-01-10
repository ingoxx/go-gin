package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/errors"
	"github.com/Lxb921006/Gin-bms/project/service"

	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model
	Path     string `json:"path" gorm:"not null"`
	Title    string `json:"title" gorm:"not null"`
	Hidden   uint   `json:"hidden" gorm:"default:1;comment:'1是可见,2是隐藏'"`
	ParentId uint   `json:"parentId" gorm:"default 0"`
	Level    uint   `json:"level" gorm:"not null"`
	Roles    []Role `json:"roles" gorm:"many2many:role_permissions"`
}

func (p *Permission) CreatePerms(pd Permission) (err error) {
	if err = dao.DB.Create(&pd).Error; err != nil {
		return
	}
	return
}

func (p *Permission) DeletePerms(pid []uint) (err error) {
	tx := dao.DB.Begin()
	var perm Permission
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, v := range pid {
		if err = dao.DB.Where("id = ?", v).Find(&perm).Error; err != nil {
			return
		}

		if err = tx.Model(&perm).Association("Roles").Clear(); err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Unscoped().Delete(&perm, v).Error; err != nil {
			tx.Rollback()
			return
		}

	}

	return tx.Commit().Error
}

func (p *Permission) GetAllPerms() ([]Permission, error) {
	var perms []Permission
	if err := dao.DB.Where("id > 0").Find(&perms).Error; err != nil {
		return perms, err
	}

	if len(perms) == 0 {
		return perms, errors.ErrEmptyPermList
	}

	return perms, nil
}

func (p *Permission) GetPermsList(page int) (data *service.Paginate, err error) {
	var ps []Permission
	sql := dao.DB.Model(p)
	pg := service.NewPaginate()
	data, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = data.Gd.Find(&ps).Error; err != nil {
		return
	}

	data.ModelSlice = ps

	return
}
