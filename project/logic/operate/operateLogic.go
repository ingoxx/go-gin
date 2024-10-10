package operate

import (
	"time"

	"github.com/Lxb921006/Gin-bms/project/api"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
)

type OperateLogListQuery struct {
	Page      int               `form:"page" validate:"min=1" binding:"required" json:"page"`
	Url       string            `form:"url" json:"url"`
	Ip        string            `form:"ip" json:"ip"`
	Operator  string            `form:"operator" json:"operator"`
	StartTime string            `form:"starttime" json:"starttime"`
	EndTime   string            `form:"endtime" json:"endtime"`
	PageData  *service.Paginate `form:"-"`
}

func (op *OperateLogListQuery) PaginateLogic(opm model.OperateLog, api api.Api) (err error) {
	//验证器
	if err = api.ValidateStruct(op); err != nil {
		return
	}

	opm.Url = op.Url
	opm.Ip = op.Ip
	opm.Operator = op.Operator

	if op.StartTime != "" || op.EndTime != "" {
		opm.Start, _ = time.Parse("2006-01-02", op.StartTime)
		opm.End, _ = time.Parse("2006-01-02", op.EndTime)
		opm.End = opm.End.Add(time.Hour * 24)
		op.PageData, err = opm.OperateLogListByDate(op.Page, opm)
		if err != nil {
			return
		}
	} else {
		op.PageData, err = opm.OperateLogList(op.Page, opm)
		if err != nil {
			return
		}
	}

	return
}
