package errors

import "errors"

var (
	EmptyPermListError = errors.New("权限列表为空")
)

type ForForbiddenError struct {
	errMsg string
}

func (e ForForbiddenError) Error() string {
	return e.errMsg
}

func IsForForbiddenError(e error) bool {
	var err ForForbiddenError
	return errors.As(e, &err)
}

func NewForForbiddenError(errMsg string) error {
	return ForForbiddenError{
		errMsg: errMsg,
	}
}

type EncryptionDataError struct {
	errMsg string
}

func (e EncryptionDataError) Error() string {
	return e.errMsg
}

func IsEncryptionDataError(e error) bool {
	var err EncryptionDataError
	return errors.As(e, &err)
}

func NewEncryptionDataError(errMsg string) error {
	return EncryptionDataError{
		errMsg: errMsg,
	}
}
