package user

import (
	"errors"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/go-playground/validator/v10"
)

type ve error

var (
	notAllowDelUser = []string{"admin", "二毛"}
	validateErr     ve
)

type ValidateData struct {
	validate *validator.Validate
}

func (v *ValidateData) ValidateStruct(s interface{}) (err error) {
	if err = v.validate.Struct(s); err != nil {
		return validateErr
	}
	return
}

func (v *ValidateData) ValidateForAdminUid(fl validator.FieldLevel) bool {
	var user model.User
	userList, ok := fl.Field().Interface().([]uint)
	if !ok {
		return false
	}

	for _, u := range userList {
		if err := dao.DB.First(&user, u).Error; err != nil {
			return false
		}
		for _, admin := range notAllowDelUser {
			if admin == user.Name {
				validateErr = errors.New(fmt.Sprintf("【%s】超级管理员不能删除", admin))
				return false
			}
		}
	}

	return true
}

func (v *ValidateData) ValidateNumber(fl validator.FieldLevel) bool {
	num := fl.Field().Interface()
	switch num := num.(type) {
	case int:
		if num <= 0 {
			return false
		} else {
			return true
		}
	case uint:
		if num <= 0 {
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func (v *ValidateData) RegisterValidation() (err error) {
	if err = v.validate.RegisterValidation("containsAdminUid", v.ValidateForAdminUid); err != nil {
		return err
	}

	return nil
}

func NewValidateData(v *validator.Validate) *ValidateData {
	var vd = &ValidateData{
		validate: v,
	}

	return vd
}
