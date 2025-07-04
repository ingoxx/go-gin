package assets

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/go-gin/project/command/rpcConfig"
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/model"
	"github.com/ingoxx/go-gin/project/service"
	"github.com/ingoxx/go-gin/project/utils"
	"path/filepath"
	"strings"
)

var (
	validate = validator.New()
)

type ProgramUpdateRecordDelForm struct {
	ID []int64 `json:"id" form:"id" binding:"required"`
}

type RunProgramApiForm struct {
	Ip         string `form:"ip" json:"ip" gorm:"not null" binding:"required"`
	UpdateName string `form:"update_name" json:"update_name" gorm:"not null" binding:"required"`
	Uuid       string `form:"uuid" json:"uuid" gorm:"not null;unique" binding:"required"`
}

func (apf *RunProgramApiForm) Data() (data map[string]interface{}, err error) {
	b, err := json.Marshal(apf)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &data); err != nil {
		return
	}

	return

}

func (apf *RunProgramApiForm) Run(ctx *gin.Context) (err error) {
	if err = ctx.ShouldBind(apf); err != nil {
		return
	}

	cy := utils.NewProgramAsyncRunCelery()
	cy.Task(apf)
	cy.Close()

	return
}

type ProgramUpdateListForm struct {
	Ip         string `form:"ip,omitempty" json:"ip"`
	Uuid       string `form:"uuid,omitempty" json:"uuid"`
	UpdateName string `form:"update_name,omitempty" json:"update_name"`
	Project    string `form:"project,omitempty" json:"project"`
	Operator   string `form:"operator,omitempty" json:"operator"`
	Progress   int32  `form:"progress,omitempty" json:"progress"`
	Status     int32  `form:"status,omitempty" json:"status"`
	Page       int    `form:"page" json:"page" validate:"min=1" binding:"required"`
}

func (apul *ProgramUpdateListForm) List(ctx *gin.Context) (data *service.Paginate, err error) {
	var lm model.AssetsProgramUpdateRecordModel
	if err = ctx.ShouldBind(apul); err != nil {
		return
	}

	vd := NewValidateData(validate)
	if err = vd.ValidateStruct(apul); err != nil {
		return
	}

	if err = utils.CopyStruct(apul, &lm); err != nil {
		return
	}

	data, err = lm.List(apul.Page, lm)
	if err != nil {
		return
	}

	return
}

type CreateUpdateProgramRecordForm struct {
	DataList []model.AssetsProgramUpdateRecordModel `form:"data_list" json:"data_list" binding:"required"`
}

func (c *CreateUpdateProgramRecordForm) Create(ctx *gin.Context) (err error) {
	var cm model.AssetsProgramUpdateRecordModel
	if err = ctx.ShouldBindJSON(c); err != nil {
		return
	}

	if err = cm.Create(c.DataList); err != nil {
		return
	}

	return
}

type GetMissionStatusForm struct {
	Result string `form:"result" binding:"required"`
}

func (ps *GetMissionStatusForm) GetProgress(ctx *gin.Context) (data map[string]string, err error) {
	if err = ctx.ShouldBind(ps); err != nil {
		return
	}

	data, err = dao.Rds.GetProcessStatus()
	if err != nil {
		return
	}

	return
}

type UploadForm struct {
	File    []string `form:"file" json:"file" binding:"required"`
	resChan chan string
}

func NewUploadForm() *UploadForm {
	return &UploadForm{
		resChan: make(chan string),
	}
}

func (u *UploadForm) UploadFiles(ctx *gin.Context) (md5 map[string]string, err error) {
	var addLog model.OperateLogModel
	var record = make(map[string]string)
	var fileList = make([]string, 0)
	form, err := ctx.MultipartForm()
	if err != nil {
		return
	}

	files := form.File["file"]

	if len(files) == 0 {
		return md5, errors.New("上传失败")
	}

	for _, file := range files {
		fullFile := filepath.Join(rpcConfig.UploadPath, file.Filename)
		if err = ctx.SaveUploadedFile(file, fullFile); err != nil {
			return
		}
		fileList = append(fileList, file.Filename)
	}

	// 这个上传文件的日志记录有点特殊，只能在这里走日志记录，middleware不好记录
	record["user"] = ctx.Query("user")
	record["url"] = ctx.Request.URL.Path + ", 文件名：" + strings.Join(fileList, ",")
	record["ip"] = ctx.RemoteIP()

	if err = addLog.AloneAddOperateLog(record); err != nil {
		return
	}

	return
}
