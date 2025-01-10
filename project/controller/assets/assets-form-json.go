package assets

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
