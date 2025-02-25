package operate

import (
	"time"

	"github.com/ingoxx/go-gin/project/api"
	"github.com/ingoxx/go-gin/project/model"
	"github.com/ingoxx/go-gin/project/service"
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

func NewOperateLogListQuery() *OperateLogListQuery {
	return &OperateLogListQuery{}
}

func (op *OperateLogListQuery) PaginateLogic(opm model.OperateLogModel, api api.Api) (err error) {
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
