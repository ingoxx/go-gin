package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
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
}

func (p *Permission) CreatePerms(pd Permission) (err error) {
	if err = dao.DB.Create(&pd).Error; err != nil {
		return
	}
	return
}

func (p *Permission) DeletePerms(pid []uint) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("permission_id IN ?", pid).Delete(&RolePermission{}).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Where("id IN ?", pid).Unscoped().Delete(p).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
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
