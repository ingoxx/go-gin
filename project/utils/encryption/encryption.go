package encryption

import (
	"github.com/Lxb921006/Gin-bms/project/config"
	ginErr "github.com/Lxb921006/Gin-bms/project/errors"
	"golang.org/x/crypto/bcrypt"
)

type DataEncryption struct {
	password string
	user     string
}

func (e *DataEncryption) EncryptionPwd() (string, error) {
	password := e.user + e.password + config.Sign
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil || len(hashedPassword) == 0 {
		return "", ginErr.NewEncryptionDataError(err.Error())
	}

	return string(hashedPassword), nil
}

func (e *DataEncryption) DecryptionPwd(hashPwd string) (err error) {
	password := e.user + e.password + config.Sign
	err = bcrypt.CompareHashAndPassword([]byte(hashPwd), []byte(password))
	if err != nil {
		return
	}

	return
}

func NewDataEncryption(u, p string) *DataEncryption {
	return &DataEncryption{
		user:     u,
		password: p,
	}
}
