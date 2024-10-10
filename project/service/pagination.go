package service

import (
	"errors"
	"github.com/Lxb921006/Gin-bms/project/config"

	"gorm.io/gorm"
)

var (
	ErrorPageRangeSize = errors.New("页码不在范围内")
)

type Paginate struct {
	Gd         *gorm.DB
	ModelSlice interface{}
	Total      int
	PageSize   int
}

func NewPaginate() *Paginate {
	return &Paginate{}
}

func (p *Paginate) GetPageData(page int, sql *gorm.DB) (pg *Paginate, err error) {
	var total int64
	//这里是写死了每页最多展示的数据
	var size = config.PageSize
	if err = sql.Count(&total).Error; err != nil {
		return
	}

	if total == 0 {
		p.Total = int(total)
		p.PageSize = size
		p.Gd = sql
		pg = p
		return
	}

	if page < 0 {
		err = ErrorPageRangeSize
		return
	}

	totalPage := int(total) / size

	if mod := int(total) % size; mod != 0 {
		totalPage += 1
	}

	if totalPage < page {
		err = ErrorPageRangeSize
		return
	}

	offset := (page - 1) * size

	p.Gd = sql.Limit(size).Offset(offset).Order("id desc")

	if p.Gd.Error != nil {
		err = p.Gd.Error
		return
	}

	p.Total = int(total)
	p.PageSize = size
	pg = p

	return
}
