package model

import (
	"bytes"
	"io"
	"time"

	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperateLog struct {
	gorm.Model
	Url      string    `json:"url" gorm:"type:text;not null"`
	Operator string    `json:"operator" gorm:"not null"`
	Ip       string    `json:"ip" gorm:"not null"`
	Start    time.Time `json:"start" gorm:"-"`
	End      time.Time `json:"end" gorm:"-"`
}

func (o *OperateLog) OperateLogList(page int, op OperateLog) (data *service.Paginate, err error) {
	var os []OperateLog
	sql := dao.DB.Model(o).Where(op)
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

func (o *OperateLog) OperateLogListByDate(page int, op OperateLog) (data *service.Paginate, err error) {
	var os []OperateLog
	sql := dao.DB.Model(o).Or(op).Where("created_at between ? and ?", op.Start, op.End)
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

func (o *OperateLog) AddOperateLog(ctx *gin.Context) (err error) {
	// 没办法, 上传的不给这样操作
	if ctx.Request.URL.Path == "/assets/upload" {
		return
	}
	b, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return
	}

	// 将其放入到缓冲区缓存
	buf := bytes.NewBuffer(b)

	// 保存在resp变量
	resp := buf.Bytes()

	buf2 := bytes.NewBuffer(resp)
	rb := io.NopCloser(buf2)

	// 读取了ctx.Request.Body需要再放回去给后续的函数用
	ctx.Request.Body = rb

	// 重置缓冲区
	buf.Reset()

	o.Url = ctx.Request.URL.Path + ", " + string(resp)
	o.Operator = ctx.Query("user")
	o.Ip = ctx.RemoteIP()

	if err = dao.DB.Create(o).Error; err != nil {
		return
	}
	return
}

func (o *OperateLog) AloneAddOperateLog(data map[string]string) error {
	o.Url = data["url"]
	o.Operator = data["user"]
	o.Ip = data["ip"]

	if err := dao.DB.Create(o).Error; err != nil {
		return err
	}

	return nil
}
