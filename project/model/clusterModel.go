package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"time"
)

type ClusterModel struct {
	ID         uint          `json:"id" gorm:"primaryKey"`
	ClusterCid string        `json:"cluster_cid" gorm:"default:n21q22l9bxkf0hhi7d971hh9o;comment:docker info可以查询"`
	Name       string        `json:"name" gorm:"unique"`
	Region     string        `json:"region" gorm:"default:cn-sz"`
	Date       time.Time     `json:"date" gorm:"default:CURRENT_TIMESTAMP;nullable"`
	Status     uint          `json:"status" gorm:"default:100;comment:100-集群异常,200-集群正常"`
	Servers    []AssetsModel `json:"servers" gorm:"foreignKey:ClusterID"`
}

func (cm *ClusterModel) List(page int, c ClusterModel) (data *service.Paginate, err error) {
	var cs []ClusterModel
	sql := dao.DB.Model(cm).Where(c)
	pg := service.NewPaginate()
	data, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = data.Gd.Find(&cs).Error; err != nil {
		return
	}

	data.ModelSlice = cs

	return
}

func (cm *ClusterModel) GetAllClusterData() (data []*ClusterModel, err error) {
	if err := dao.DB.Model(cm).Find(&data).Error; err != nil {
		return data, err
	}

	return
}

func (cm *ClusterModel) Update() {}

func (cm *ClusterModel) Delete(id uint) error {
	if err := dao.DB.Where("id = ?", id).Delete(cm).Error; err != nil {
		return err
	}

	return nil
}

func (cm *ClusterModel) Add(c ClusterModel) error {
	if err := dao.DB.Create(&c).Error; err != nil {
		return err
	}

	return nil
}
