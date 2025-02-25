package model

import (
	"errors"
	"fmt"
	ginErr "github.com/ingoxx/go-gin/project/errors"
	"github.com/ingoxx/go-gin/project/utils/encryption"

	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/service"
)

type Login struct {
	ID       uint   `json:"uid"`
	Name     string `json:"name"`
	Isopenga uint   `json:"isopenga"`
	Isopenqr uint   `json:"isopenqr"`
	QrUrl    string `json:"qrurl"`
	Password string `json:"-"`
	Token    string `json:"token"`
	MfaApp   uint   `json:"mfa_app"`
}

func (l *Login) GaLogin(code, user string) (ui *Login, err error) {
	if err = dao.Rds.ForbiddenLogin(user + "-lm-ga"); err != nil {
		return nil, ginErr.NewForForbiddenError(err.Error())
	}

	key, err := dao.Rds.GetGaKey(user)
	if err != nil {
		return
	}

	gas := service.NewGoogleAuthenticator(key)
	gaCode, err := gas.GaCode()
	if err != nil {
		return
	}

	if gaCode != code {
		if err = dao.Rds.RecordLoginFailedNum(user + "-lm-ga"); err != nil {
			return
		}
		err = errors.New("验证失败")
		return
	}

	//扫描User表填充到Login表
	if err = l.FillData(user); err != nil {
		return
	}

	//用户扫完码就关闭qr
	if l.Isopenqr == 1 || l.MfaApp == 1 {
		if err = l.CloseGoogleAuthQr(user); err != nil {
			return
		}
	}

	token, err := dao.Rds.RegisterUserInfo(user)
	if err != nil {
		return
	}

	l.Token = token
	ui = l

	return
}

func (l *Login) UserLogin(u, p string) (ui *Login, err error) {
	if err = l.Authenticate(u, p); err != nil {
		return
	}

	isGa, err := l.IsOpenGoogleAuth(u)
	if err != nil {
		return
	}

	if isGa {
		ui = l
		return
	}

	token, err := dao.Rds.RegisterUserInfo(u)
	if err != nil {
		return
	}

	l.Token = token
	ui = l

	return
}

func (l *Login) UserLogout(u string) (err error) {
	if err = dao.Rds.ClearToken(u); err != nil {
		return
	}
	return
}

func (l *Login) Authenticate(u, p string) (err error) {
	if err = l.FillData(u); err != nil {
		return
	}

	if err = dao.Rds.ForbiddenLogin(u + "-lm"); err != nil {
		return ginErr.NewForForbiddenError(err.Error())
	}

	if l.Name != u {
		if err = dao.Rds.RecordLoginFailedNum(u + "-lm"); err != nil {
			return
		}
		return fmt.Errorf("用户%s不存在", u)
	}

	if err := encryption.NewDataEncryption(u, p).DecryptionPwd(l.Password); err != nil {
		if err := dao.Rds.RecordLoginFailedNum(u + "-lm"); err != nil {
			return err
		}

		return errors.New("密码错误")
	}

	return
}

func (l *Login) IsOpenGoogleAuth(u string) (b bool, err error) {
	var ui User
	gas := service.NewGoogleAuthenticator("")
	if err = dao.DB.Where("name = ?", u).Find(&ui).Error; err != nil {
		return
	}

	if ui.Isopenga == 1 {
		b = true
	}

	if ui.Isopenqr == 1 {
		l.QrUrl = gas.QrUrl("golang-cmdb", u)
		err = dao.Rds.SaveGaKey(u, gas.Secret)
		if err != nil {
			return
		}
	}

	l.Isopenga = ui.Isopenga
	l.Isopenqr = ui.Isopenqr

	return
}

func (l *Login) CloseGoogleAuthQr(u string) (err error) {
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Model(&User{}).Where("name = ?", u).Update("isopenqr", 2).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Model(&User{}).Where("name = ?", u).Update("mfa_app", 2).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (l *Login) FillData(user string) (err error) {
	if err = dao.DB.Model(&User{}).Where("name = ?", user).Scan(l).Error; err != nil {
		return
	}
	return
}
