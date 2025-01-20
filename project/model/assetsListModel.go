package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"gorm.io/gorm"
	"time"
)

type AssetsModel struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Ip        string    `json:"ip" gorm:"not null;unique"`
	Project   string    `json:"project" gorm:"not null"`
	Status    int32     `json:"status" gorm:"default:200;comment:100-服务器异常,200-服务器正常"`
	Operator  string    `json:"operator" gorm:"default:lxb"`
	RamUsage  int32     `json:"ram_usage" gorm:"default:1"`
	DiskUsage int32     `json:"disk_usage" gorm:"default:1"`
	CpuUsage  int32     `json:"cpu_usage"  gorm:"default:1"`
	Start     time.Time `json:"start" gorm:"default:CURRENT_TIMESTAMP;nullable"`
}

func (o *AssetsModel) List(page int, am AssetsModel) (data *service.Paginate, err error) {
	var os []AssetsModel
	sql := dao.DB.Model(o).Where(&am)
	pg := service.NewPaginate()
	data, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = data.Gd.Find(&os).Error; err != nil {
		return
	}

	data.ModelSlice = os

	return
}

func (o *AssetsModel) Create(am []*AssetsModel) (err error) {
	if err = dao.DB.Create(am).Error; err != nil {
		return
	}

	return
}

func (o *AssetsModel) Delete(ip []string) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("ip IN ?", ip).Unscoped().Delete(o).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (o *AssetsModel) Update(data map[string]interface{}) (err error) {
	if err = dao.DB.Model(o).Where("id = ?", data["id"].(int64)).Updates(data).Error; err != nil {
		return
	}
	dao.DB.Save(o)

	return
}

func (o *AssetsModel) AfterUpdate(tx *gorm.DB) (err error) {
	o.Start = time.Now()
	return
}
