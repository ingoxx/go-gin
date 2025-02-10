package assets

import (
	"github.com/go-playground/validator/v10"
)

type ValidateData struct {
	validate *validator.Validate
}

func (v *ValidateData) ValidateStruct(s interface{}) (err error) {
	if err = v.validate.Struct(s); err != nil {
		return
	}
	return
}

func (v *ValidateData) ValidatePwdKey(fl validator.FieldLevel) (err error) {
	return
}

func (v *ValidateData) RegisterValidation() (err error) {

	return nil
}

func NewValidateData(v *validator.Validate) *ValidateData {
	return &ValidateData{
		validate: v,
	}
}
