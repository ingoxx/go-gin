package model

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OperateLogModel struct {
	gorm.Model
	Url      string    `json:"url" gorm:"type:text;not null"`
	Operator string    `json:"operator" gorm:"not null"`
	Ip       string    `json:"ip" gorm:"not null"`
	Start    time.Time `json:"start" gorm:"-"`
	End      time.Time `json:"end" gorm:"-"`
}

func (o *OperateLogModel) OperateLogList(page int, op OperateLogModel) (data *service.Paginate, err error) {
	var os []OperateLogModel
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

func (o *OperateLogModel) OperateLogListByDate(page int, op OperateLogModel) (data *service.Paginate, err error) {
	var os []OperateLogModel
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

func (o *OperateLogModel) AddOperateLog(ctx *gin.Context) (err error) {
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

	if ctx.Request.URL.Path == "/assets/terminal" {
		o.Url = ctx.Request.URL.Path + ", connect server: " + ctx.Query("ip")
	} else {
		o.Url = ctx.Request.URL.Path + ", " + string(resp)
	}

	o.Operator = ctx.Query("user")
	o.Ip = ctx.RemoteIP()

	if err = dao.DB.Create(o).Error; err != nil {
		return
	}

	if err := o.dataCount(ctx.Request.URL.Path); err != nil {

		return err
	}

	return
}

func (o *OperateLogModel) dataCount(url string) (err error) {
	if url == "/login" {
		if err := o.recordHighFrequencyData(dao.LoginNum); err != nil {
			return err
		}
		if err := o.recordHighFrequencyData(dao.UserLoginNum); err != nil {
			return err
		}
	}

	if url == "/assets/run-linux-cmd" {
		if err := o.recordHighFrequencyData(dao.RunLinuxCmdNum); err != nil {
			return err
		}
	}

	return
}

func (o *OperateLogModel) recordHighFrequencyData(key string) (err error) {
	var data []map[string]interface{}
	if key == dao.LoginNum {
		data, err = o.findSevenDaysLoginNum()
		if err != nil {

			return
		}
	} else if key == dao.RunLinuxCmdNum {
		data, err = o.findSevenDaysRunLinuxCmdNum()
		if err != nil {
			return
		}
	} else if key == dao.UserLoginNum {
		data, err = o.findUserLoginNum()
		if err != nil {
			return
		}
	}

	b, err := json.Marshal(&data)
	if err != nil {

		return
	}

	if err := dao.Rds.SetData(key, b); err != nil {
		return err
	}

	return
}

func (o *OperateLogModel) GetLoginNum() (data interface{}, err error) {
	var md = make([]map[string]interface{}, 0)
	b, err := dao.Rds.GetData(dao.LoginNum)
	if err != nil {
		return
	}

	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}

	var rd = make(map[string]interface{})
	rd["columns"] = []string{"日期", "平台登录次数"}
	rd["rows"] = md

	data = rd

	return
}

func (o *OperateLogModel) GetRunLinuxCmdNum() (data interface{}, err error) {
	var md = make([]map[string]interface{}, 0)
	var rd = make(map[string]interface{})

	b, err := dao.Rds.GetData(dao.RunLinuxCmdNum)
	if err != nil {
		return
	}

	if len(b) == 0 {
		return rd, nil
	}

	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}

	rd["columns"] = []string{"日期", "linux命令执行次数"}
	rd["rows"] = md

	data = rd

	return
}

func (o *OperateLogModel) GetUserLoginNum() (data interface{}, err error) {
	var md = make([]map[string]interface{}, 0)
	var rd = make(map[string]interface{})

	b, err := dao.Rds.GetData(dao.UserLoginNum)
	if err != nil {
		return
	}

	if len(b) == 0 {
		return rd, nil
	}

	if err := json.Unmarshal(b, &md); err != nil {
		return nil, err
	}

	rd["columns"] = []string{"用户名", "总的平台登陆次数"}
	rd["rows"] = md

	data = rd

	return
}

func (o *OperateLogModel) AloneAddOperateLog(data map[string]string) error {
	o.Url = data["url"]
	o.Operator = data["user"]
	o.Ip = data["ip"]

	if err := dao.DB.Create(o).Error; err != nil {
		return err
	}

	return nil
}

func (o *OperateLogModel) findSevenDaysLoginNum() ([]map[string]interface{}, error) {
	var dataList = make([]map[string]interface{}, 0)
	rows, err := dao.DB.Raw(`
		SELECT DATE(created_at) as date, count(1) as login_num FROM operate_log_models 
		where DATE(created_at) > NOW() - INTERVAL 7 DAY and url like '%/login%'
		GROUP BY DATE(created_at);
	`).Rows()

	if err != nil {
		return dataList, err
	}

	for rows.Next() {
		var date string
		var loginNum int
		var data = make(map[string]interface{})

		if err := rows.Scan(&date, &loginNum); err != nil {
			return dataList, err
		}

		parsedTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return dataList, err
		}

		data["日期"] = parsedTime.Format("2006-01-02")
		data["平台登录次数"] = loginNum
		dataList = append(dataList, data)
	}

	return dataList, nil
}

func (o *OperateLogModel) findSevenDaysRunLinuxCmdNum() ([]map[string]interface{}, error) {
	var dataList = make([]map[string]interface{}, 0)
	rows, err := dao.DB.Raw(`
		SELECT DATE(created_at) as date, count(1) as run_num FROM operate_log_models 
		where DATE(created_at) > NOW() - INTERVAL 7 DAY and url like '%/assets/run-linux-cmd%' 
		GROUP BY DATE(created_at);
	`).Rows()

	if err != nil {
		return dataList, err
	}

	for rows.Next() {
		var date string
		var runLinuxCmdNum int
		var data = make(map[string]interface{})

		if err := rows.Scan(&date, &runLinuxCmdNum); err != nil {
			return dataList, err
		}

		parsedTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return dataList, err
		}

		data["日期"] = parsedTime.Format("2006-01-02")
		data["linux命令执行次数"] = runLinuxCmdNum
		dataList = append(dataList, data)
	}

	return dataList, nil
}

func (o *OperateLogModel) findUserLoginNum() ([]map[string]interface{}, error) {
	var dataList = make([]map[string]interface{}, 0)
	rows, err := dao.DB.Raw(`
		select operator, count(1) as user_login_num from operate_log_models 
		where url like '%/login%' 
		GROUP BY operator ORDER BY user_login_num LIMIT 5;
	`).Rows()

	if err != nil {
		return dataList, err
	}

	for rows.Next() {
		var user string
		var loginNum int
		var data = make(map[string]interface{})

		if err := rows.Scan(&user, &loginNum); err != nil {
			return dataList, err
		}

		data["用户名"] = user
		data["总的平台登陆次数"] = loginNum
		dataList = append(dataList, data)
	}

	return dataList, nil
}

func (o *OperateLogModel) serializeData() (err error) {
	return
}
