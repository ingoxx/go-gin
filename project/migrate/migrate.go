package migrate

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
)

func InitTable() (err error) {
	err = dao.DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.OperateLog{},
		&model.AssetsModel{},
		&model.AssetsProcessUpdateRecordModel{},
	)

	return
}
