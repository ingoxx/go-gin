package model

import (
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/service"
	"time"
)

type ClusterModel struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	ClusterCid  string        `json:"cluster_cid" gorm:"default:n21q22l9bxkf0hhi7d971hh9o;comment:docker info可以查询"`
	Name        string        `json:"name" gorm:"unique"`
	Region      string        `json:"region" gorm:"default:cn-sz"`
	WorkToken   string        `json:"-" gorm:"null;comment:work节点的token"`
	MasterToken string        `json:"-" gorm:"null;comment:master节点token"`
	MasterIp    string        `json:"master_ip"  gorm:"default:1.1.1.1"`
	Date        time.Time     `json:"date" gorm:"default:CURRENT_TIMESTAMP;nullable"`
	Status      uint          `json:"status" gorm:"default:300;comment:100-集群异常,200-集群正常,300-正在初始化"`
	Servers     []AssetsModel `json:"servers" gorm:"foreignKey:ClusterID"`
	ClusterType string        `json:"cluster_type" gorm:"default:1"`
	HealthPort  int32         `json:"health_port" gorm:"default:12306;comment:管理节点的健康检测端口"`
}

func (cm *ClusterModel) List(page int, c ClusterModel) (data *service.Paginate, err error) {
	var cs []ClusterModel
	sql := dao.DB.Model(cm).Where(c).Preload("Servers")
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

func (cm *ClusterModel) Update(id uint, data ClusterModel) error {
	if err := dao.DB.Model(cm).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}

	return nil
}

func (cm *ClusterModel) Delete(id uint) error {
	if err := dao.DB.Where("id = ?", id).Delete(cm).Error; err != nil {
		return err
	}

	return nil
}

func (cm *ClusterModel) Create(c *ClusterModel) error {
	if err := dao.DB.Create(c).Error; err != nil {
		return err
	}

	return nil
}

func (cm *ClusterModel) GetCluster(id uint) (ClusterModel, error) {
	var cms ClusterModel
	if err := dao.DB.Model(cm).Where("id = ?", id).Preload("Servers").Find(&cms).Error; err != nil {
		return cms, err
	}

	return cms, nil
}
