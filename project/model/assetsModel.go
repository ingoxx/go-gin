package model

import (
	"fmt"
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/service"
	"gorm.io/gorm"
	"time"
)

// AssetsModel 服务器列表
type AssetsModel struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Ip          string       `json:"ip" gorm:"not null;unique"`
	NodeType    uint         `json:"node_type" gorm:"default:3;comment:1-master节点, 2-work节点, 3-未知节点类型"`
	Project     string       `json:"project" gorm:"not null"`
	Status      uint         `json:"status" gorm:"default:200;comment:100-服务器异常,200-服务器正常,300-未知状态"`
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
	NodeStatus  uint         `json:"node_status" gorm:"default:300;comment:100-节点异常,200-节点正常,300-未知状态"`
	LeaveType   uint         `json:"leave_type" gorm:"default:3;comment:1-手动离开集群,2-被动离开集群,3-未知原因"`
}

type AssetsWithCluster struct {
	AssetsModel
	ClusterID     uint   `json:"id"`
	ClusterName   string `json:"name"`
	ClusterStatus uint   `json:"status"`
	ClusterCid    string `json:"cluster_cid"`
}

func (o *AssetsModel) List2(page int, pageSize int, filter AssetsModel) ([]AssetsModel, int64, error) {
	var total int64
	var result []AssetsModel

	db := dao.DB.Model(&AssetsModel{}).Where(&filter)

	// 1. 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []AssetsModel{}, 0, nil
	}

	// 2. 分页获取符合条件的 ID（只查 ID，速度非常快）
	var ids []uint
	offset := (page - 1) * pageSize
	if err := db.
		Select("id").
		Order("id desc").
		Limit(pageSize).
		Offset(offset).
		Pluck("id", &ids).Error; err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []AssetsModel{}, total, nil
	}

	// 3. 使用 ID 列表再次查询，携带关联字段
	if err := dao.DB.
		Model(&AssetsModel{}).
		Preload("Cluster", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "cluster_cid", "name", "status")
		}).
		Where("id IN ?", ids).
		Order("id desc").
		Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (o *AssetsModel) List(page int, am AssetsModel) (data *service.Paginate, err error) {
	var os []AssetsModel
	sql := dao.DB.Model(o).Where(&am).Preload("Cluster", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "status", "cluster_cid")
	})
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
	var ips []string
	err := dao.DB.Model(&AssetsModel{}).Pluck("ip", &ips).Error
	return ips, err
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

func (o *AssetsModel) Delete(id []uint) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var servers []AssetsModel
	// 查出这些服务器的信息，只取ID和ClusterID
	if err := dao.DB.Select("id", "cluster_id", "ip").Where("id IN ?", id).Find(&servers).Error; err != nil {
		return err
	}

	// 判断是否有服务器还绑定了集群
	for _, s := range servers {
		if s.ClusterID != nil {
			return fmt.Errorf("服务器'%s' 已绑定集群，无法删除，请先移除集群绑定", s.Ip)
		}
	}

	if err = tx.Where("id IN ?", id).Unscoped().Delete(o).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (o *AssetsModel) Update(am AssetsModel) (err error) {
	tx := dao.DB.Begin()
	if err = tx.Model(o).Where("id = ?", am.ID).Updates(am).Error; err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit().Error
}

func (o *AssetsModel) AfterUpdate(tx *gorm.DB) (err error) {
	o.Start = time.Now()
	return
}
