package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"gorm.io/gorm"
	"time"
)

type AssetsModel struct {
	ID       int64     `json:"id" gorm:"primaryKey"`
	Ip       string    `json:"ip" gorm:"not null;unique"`
	Project  string    `json:"project" gorm:"not null"`
	Status   string    `json:"status" gorm:"default:100"`
	Operator string    `json:"operator" gorm:"default:lxb"`
	Start    time.Time `json:"start" gorm:"default:CURRENT_TIMESTAMP;nullable"`
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

func (o *AssetsModel) Del(ip []string) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("ip IN ?", ip).Delete(o).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (o *AssetsModel) Modify(data map[string]interface{}) (err error) {
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
