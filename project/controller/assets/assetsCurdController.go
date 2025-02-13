package assets

import (
	"errors"
	"github.com/Lxb921006/Gin-bms/project/command/rpcConfig"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/Lxb921006/Gin-bms/project/utils/encryption"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
)

// 服务器的增删改查

type ListForm struct {
	Ip        string `form:"ip,omitempty" json:"ip"`
	Project   string `form:"project,omitempty" json:"project"`
	ClusterID *uint  `form:"cluster_id,omitempty" json:"cluster_id"`
	Page      int    `form:"page" json:"page" validate:"min=1" binding:"required"`
}

func (a *ListForm) List(ctx *gin.Context) (data *service.Paginate, err error) {
	var al model.AssetsModel
	if err = ctx.ShouldBind(a); err != nil {
		return
	}

	//validate := validator.New()
	vd := NewValidateData(validate)
	if err = vd.ValidateStruct(a); err != nil {
		return
	}

	if err = mapstructure.Decode(a, &al); err != nil {
		return
	}

	data, err = al.List(a.Page, al)
	if err != nil {
		return
	}

	return
}

func (a *ListForm) GetAllClusterData() ([]*model.ClusterModel, error) {
	var cm *model.ClusterModel
	allData, err := cm.GetAllClusterData()
	if err != nil {
		return nil, err
	}

	return allData, nil
}

type DelForm struct {
	Ips []string `form:"ips" json:"ips" binding:"required"`
}

func (a *DelForm) Del(ctx *gin.Context) (err error) {
	var am model.AssetsModel
	if err = ctx.BindJSON(a); err != nil {
		return
	}

	if err = am.Delete(a.Ips); err != nil {
		return
	}

	return
}

type CreateUpdateAssetsForm struct {
	ID          uint   `json:"id" form:"id"`
	NodeType    uint   `json:"node_type" form:"node_type"`
	Project     string `json:"project" form:"project" binding:"required"`
	Ip          string `json:"ip" form:"ip" binding:"required"`
	User        string `json:"user" form:"user" binding:"required"`
	Port        uint   `json:"port" form:"port" binding:"required"`
	Password    string `json:"password"  form:"password"`
	Key         string `json:"key"`
	Operator    string `json:"operator" form:"operator" binding:"required"`
	ConnectType uint   `json:"connect_type" form:"connect_type" binding:"required"`
	ctx         *gin.Context
	am          model.AssetsModel
	ClusterID   *uint `json:"cluster_id" form:"cluster_id"`
	//Cluster     model.ClusterModel `json:"cluster" gorm:"constraint:OnDelete:SET NULL;"`
}

func (caf *CreateUpdateAssetsForm) VerifyFrom() (err error) {
	if err := caf.ctx.ShouldBind(&caf); err != nil {
		return err
	}
	return
}

func (caf *CreateUpdateAssetsForm) assetsData() (map[string]interface{}, error) {
	var data = make(map[string]interface{})
	var keyFile string
	formData, err := caf.ctx.MultipartForm()
	if err != nil {
		return data, err
	}

	if caf.ConnectType == 1 {
		ks, err := encryption.NewKeyPwdEncryption(caf.Password, 1).Encryption()
		if err != nil {
			return nil, err
		}
		caf.Password = ks
	} else if caf.ConnectType == 2 {
		fileHandles := formData.File["file"]
		for _, file := range fileHandles {
			keyFile = filepath.Join(rpcConfig.UploadPath, file.Filename)
			if err = caf.ctx.SaveUploadedFile(file, keyFile); err != nil {
				return nil, err
			}
		}
		ks, err := encryption.NewKeyPwdEncryption(keyFile, 2).Encryption()
		if err != nil {
			return data, err
		}
		caf.Key = ks
	} else {
		return nil, errors.New("未知登陆类型")
	}

	if err := mapstructure.Decode(caf, &caf.am); err != nil {
		return nil, err
	}

	if caf.Key != "" {
		caf.am.Key = caf.Key
	}

	return data, nil
}

func (caf *CreateUpdateAssetsForm) Create() (err error) {
	_, err = caf.assetsData()
	if err != nil {
		return err
	}

	if err := caf.am.Create(caf.am); err != nil {
		return err
	}

	return
}

func (caf *CreateUpdateAssetsForm) Update() (err error) {
	_, err = caf.assetsData()
	if err != nil {
		return err
	}

	if err := caf.am.Update(caf.am); err != nil {
		return err
	}

	return
}

func NewCreateUpdateAssetsForm(ctx *gin.Context) *CreateUpdateAssetsForm {
	return &CreateUpdateAssetsForm{
		ctx: ctx,
	}
}
