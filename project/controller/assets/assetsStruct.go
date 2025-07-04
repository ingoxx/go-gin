package assets

import "github.com/ingoxx/go-gin/project/model"

var am model.AssetsProgramUpdateRecordModel
var om model.OperateLogModel

type ProgramAddForm struct {
	CnName string `json:"cnname" form:"cnname"  binding:"required"`
	EnName string `json:"enname" form:"enname"  binding:"required"`
	Path   string `json:"path" form:"path"  binding:"required"`
}

type ProgramListForm struct {
	CnName string `json:"cnname" form:"cname"`
	EnName string `json:"enname" form:"ename"`
	Path   string `json:"path" form:"path"`
}

type WebTerminalQuery struct {
	ID uint `json:"server_id" form:"server_id" binding:"required"`
}

type GetServerStatusQuery struct {
	Ip   string `json:"ip" form:"ip" binding:"required"`
	Days uint   `json:"days" form:"days" binding:"omitempty" validate:"min=1"`
}
