package service

import (
	"errors"
	"github.com/ingoxx/go-gin/project/config"

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

func (p *Paginate) GetPageData(page int, sql *gorm.DB) (*Paginate, error) {
	var total int64
	size := config.PageSize

	// Count total rows
	if err := sql.Count(&total).Error; err != nil {
		return nil, err
	}

	// 没有数据，直接返回
	if total == 0 {
		p.Total = 0
		p.PageSize = size
		p.Gd = sql
		return p, nil
	}

	// 计算总页数 + 校验页码范围
	totalPage := int((total + int64(size) - 1) / int64(size))
	if page <= 0 || page > totalPage {
		return nil, ErrorPageRangeSize
	}

	offset := (page - 1) * size

	// 只返回处理好的 DB 查询对象
	p.Total = int(total)
	p.PageSize = size
	p.Gd = sql.Order("id DESC").Offset(offset).Limit(size) // 不做 Order
	return p, nil
}
