package migrate

import (
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/model"
)

func InitTable() (err error) {
	err = dao.DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.OperateLogModel{},
		&model.AssetsModel{},
		&model.AssetsProgramUpdateRecordModel{},
		&model.AssetsProgramModel{},
		&model.ClusterModel{},
	)

	return
}
