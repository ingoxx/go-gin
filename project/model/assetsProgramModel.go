package model

import "github.com/Lxb921006/Gin-bms/project/dao"

// AssetsProgramModel 程序操作列表, 可在这里添加新的功能
type AssetsProgramModel struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	CnName string `json:"cnname" gorm:"not null;unique"`
	EnName string `json:"enname" gorm:"not null;unique"`
	Path   string `json:"path" gorm:"not null"`
	IsLoad bool   `json:"load" gorm:"default:false;nullable"`
}

func (apm *AssetsProgramModel) List() (data []AssetsProgramModel, err error) {
	if err = dao.DB.Find(&data).Error; err != nil {
		return
	}
	return
}

func (apm *AssetsProgramModel) Create(data AssetsProgramModel) (err error) {
	if err = dao.DB.Create(&data).Error; err != nil {
		return
	}
	return
}

func (apm *AssetsProgramModel) Delete() {}

func (apm *AssetsProgramModel) Update() {}
