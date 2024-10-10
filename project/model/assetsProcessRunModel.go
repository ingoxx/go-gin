package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"gorm.io/gorm"
	"time"
)

type AssetsProcessUpdateRecordModel struct {
	ID         int64     `form:"id,omitempty" json:"id,omitempty" gorm:"primaryKey"`
	Ip         string    `form:"ip" json:"ip,omitempty" gorm:"not null"`
	Uuid       string    `form:"uuid" json:"uuid,omitempty" gorm:"not null;unique"`
	UpdateName string    `form:"update_name" json:"update_name,omitempty" gorm:"not null"`
	Project    string    `form:"project" json:"project,omitempty" gorm:"not null"`
	Operator   string    `form:"operator" json:"operator,omitempty" gorm:"not null"`
	Progress   int32     `form:"progress,omitempty" json:"progress,omitempty" gorm:"default:0;nullable"`
	Status     int32     `form:"status,omitempty" json:"status,omitempty" gorm:"default:400;comment:200-success,300-failed,400-running;nullable"`
	CostTime   int32     `form:"cost_time,omitempty" json:"cost_time,omitempty" gorm:"default:0;nullable"`
	Start      time.Time `form:"start,omitempty" json:"start,omitempty" gorm:"default:CURRENT_TIMESTAMP;nullable"`
	End        time.Time `form:"end,omitempty" json:"end,omitempty" gorm:"default:CURRENT_TIMESTAMP;nullable"`
}

func (pur *AssetsProcessUpdateRecordModel) List(page int, am AssetsProcessUpdateRecordModel) (data *service.Paginate, err error) {
	var os []AssetsProcessUpdateRecordModel
	sql := dao.DB.Model(pur).Where(&am)
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

func (pur *AssetsProcessUpdateRecordModel) Create(data []AssetsProcessUpdateRecordModel) (err error) {
	if err = dao.DB.Create(&data).Error; err != nil {
		return
	}
	return
}

func (pur *AssetsProcessUpdateRecordModel) Update(data map[string]interface{}) (err error) {
	if err = dao.DB.Model(pur).Where("uuid = ?", data["uuid"].(string)).Updates(data).Error; err != nil {
		return
	}

	return
}

func (pur *AssetsProcessUpdateRecordModel) Del(pid []int32) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("id IN ?", pid).Delete(pur).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error

}

func (pur *AssetsProcessUpdateRecordModel) BeforeSave(tx *gorm.DB) (err error) {
	if pur.Start.IsZero() {
		pur.Start = time.Now()
	}

	if pur.End.IsZero() {
		pur.End = time.Now()
	}

	return
}
