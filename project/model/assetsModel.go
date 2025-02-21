package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"gorm.io/gorm"
	"time"
)

// AssetsModel 服务器列表
type AssetsModel struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Ip          string       `json:"ip" gorm:"not null;unique"`
	NodeType    uint         `json:"node_type" gorm:"default:3;comment:1-master节点, 2-work节点, 3-未知节点类型"`
	Project     string       `json:"project" gorm:"not null"`
	Status      uint         `json:"status" gorm:"default:200;comment:100-服务器异常,200-服务器正常"`
	Operator    string       `json:"operator" gorm:"default:lxb"`
	RamUsage    uint         `json:"ram_usage" gorm:"default:1"`
	DiskUsage   uint         `json:"disk_usage" gorm:"default:1"`
	CpuUsage    uint         `json:"cpu_usage"  gorm:"default:1"`
	Start       time.Time    `json:"start" gorm:"default:CURRENT_TIMESTAMP;nullable"`
	User        string       `json:"user" gorm:"default:root"`
	Password    string       `json:"-" gorm:"not null"`
	Key         string       `json:"-" gorm:"type:TEXT"`
	Port        uint         `json:"port" gorm:"default:22"`
	OSType      uint         `json:"os_type" gorm:"default:1;comment:1-ubuntu,2-centos,3-debian"`
	ConnectType uint         `json:"connect_type" gorm:"default:1;comment:1-密码登陆, 2-秘钥登陆"`
	ClusterID   *uint        `json:"cluster_id" gorm:"index;onDelete:SET NULL;default:NULL"`
	Cluster     ClusterModel `json:"cluster" gorm:"constraint:OnDelete:SET NULL;"`
	NodeStatus  uint         `json:"node_status" gorm:"default:300;comment:100-节点异常,200-节点正常,100-未知状态"`
}

func (o *AssetsModel) List(page int, am AssetsModel) (data *service.Paginate, err error) {
	var os []AssetsModel
	sql := dao.DB.Model(o).Where(&am).Preload("Cluster")
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

func (o *AssetsModel) GetTerminalIp(id uint) (ip string, err error) {
	var am AssetsModel
	if err = dao.DB.Where("id = ?", id).Find(&am).Error; err != nil {
		return
	}

	ip = am.Ip

	return
}

func (o *AssetsModel) GetAllServersIp() ([]string, error) {
	var ams []AssetsModel
	var ip = make([]string, 0)
	if err := dao.DB.Find(&ams).Error; err != nil {
		return ip, err
	}

	for _, v := range ams {
		ip = append(ip, v.Ip)
	}

	return ip, nil
}

func (o *AssetsModel) GetServer(ip string) (am AssetsModel, err error) {
	if err = dao.DB.Where("ip = ?", ip).Find(&am).Error; err != nil {
		return
	}

	return
}

func (o *AssetsModel) Create(am AssetsModel) (err error) {
	if err = dao.DB.Create(&am).Error; err != nil {
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

func (o *AssetsModel) Update(am AssetsModel) (err error) {
	tx := dao.DB.Begin()
	if err = tx.Model(o).Where("ip = ?", am.Ip).Updates(am).Error; err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit().Error
}

func (o *AssetsModel) AfterUpdate(tx *gorm.DB) (err error) {
	o.Start = time.Now()
	return
}
