package model

import (
	"encoding/json"
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/service"
	"github.com/ingoxx/go-gin/project/utils/encryption"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	//gorm.Model
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Name      string         `json:"name" gorm:"unique;not null"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Hobby     string         `json:"-" gorm:"default:'basketball'"`
	Tel       int            `json:"tel" gorm:"default:168888"`
	Password  string         `json:"password" gorm:"not null"`
	Roles     []Role         `json:"roles" gorm:"many2many:role_users"`
	Isopenga  uint           `json:"isopenga" gorm:"default:1;comment:1-打开MFA,2-关闭MFA"`
	Isopenqr  uint           `json:"isopenqr" gorm:"default:1;comment:1-打开重置MFA,2-关闭重置MFA"`
	MfaApp    uint           `json:"mfa_app" gorm:"default:1;comment:1-打开MFA应用下载,2-关闭MFA应用下载"`
}

func (u *User) AddUser(au User, rid uint) (err error) {
	var roles []Role

	role, err := u.AssignRoles(rid)
	if err != nil {
		return err
	}

	roles = append(roles, role)
	au.Roles = roles

	if err = dao.DB.Create(&au).Error; err != nil {
		return
	}

	b, err := json.Marshal(au)
	if err != nil {
		return err
	}

	if err := dao.Rds.SetData(au.Name+"-rc", b); err != nil {
		return err
	}

	return
}

// AssignRoles 分配角色
func (u *User) AssignRoles(rid uint) (role Role, err error) {
	if err = dao.DB.Where("id = ?", rid).First(&role).Error; err != nil {
		return
	}

	return
}

func (u *User) DeleteUser(uid []uint) (err error) {
	var us []User
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("id IN ?", uid).Find(&us).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Select(clause.Associations).Unscoped().Delete(&us).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (u *User) UpdateUser(ud User, rid uint, uid uint) (err error) {
	var user User
	var role Role
	var roles []Role

	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if ud.Password != "" {
		enData, err := encryption.NewDataEncryption(ud.Name, ud.Password).EncryptionPwd()
		if err != nil {
			return err
		}

		ud.Password = enData
	}

	if err = dao.DB.Where("id = ?", uid).Find(&user).Error; err != nil {
		return
	}

	// 目前设计就是一个用户只能在一个角色组
	if err = dao.DB.Where("id = ?", rid).First(&role).Error; err != nil {
		return err
	}

	roles = append(roles, role)

	if err = tx.Model(&user).Association("Roles").Replace(roles); err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Model(&user).Updates(&ud).Error; err != nil {
		tx.Rollback()
		return
	}

	b, err := json.Marshal(ud)
	if err != nil {
		return err
	}

	if err := dao.Rds.SetData(ud.Name+"-rc", b); err != nil {
		return err
	}

	return tx.Commit().Error
}

func (u *User) GetUserNameById(uid []uint) (us []string, err error) {
	var users []User
	if err = dao.DB.Where("id IN ?", uid).Find(&users).Error; err != nil {
		return
	}

	if len(users) > 0 {
		for _, un := range users {
			us = append(us, un.Name)
		}
	}

	return
}

func (u *User) GetUserByName(name string) (ud User, err error) {
	if err = dao.DB.Model(u).Where("name = ?", name).Preload("Roles").Find(&ud).Error; err != nil {
		return
	}

	return
}

// GetUserByPaginate 单表中过滤出row
func (u *User) GetUserByPaginate(page int, user User) (ul *service.Paginate, err error) {
	var us []User
	sql := dao.DB.Omit("Password").Model(u).Where(&user).Preload("Roles")
	pg := service.NewPaginate()
	ul, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = ul.Gd.Find(&us).Error; err != nil {
		return
	}

	ul.ModelSlice = us

	return
}

// GetUserByMmPaginate 通过m2m关系表中过滤出row
func (u *User) GetUserByMmPaginate(page int, rolename string, user User) (ul *service.Paginate, err error) {
	var us []User
	var uid []uint
	if err = dao.DB.Omit("Password").Preload("Roles", Role{RoleName: rolename}).Where(&user).Find(&us).Error; err != nil {
		return
	}

	for _, v := range u.FormatData(us) {
		uid = append(uid, v.ID)
	}

	sql := dao.DB.Omit("Password").Model(u).Where("id IN ?", uid).Preload("Roles")
	pg := service.NewPaginate()
	ul, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = ul.Gd.Find(&us).Error; err != nil {
		return
	}

	ul.ModelSlice = us

	return
}

func (u *User) FormatData(ud []User) (usd []User) {
	for _, v := range ud {
		if len(v.Roles) != 0 {
			usd = append(usd, v)
		}
	}
	return
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	enData, err := encryption.NewDataEncryption(u.Name, u.Password).EncryptionPwd()
	if err != nil {
		return
	}

	u.Password = enData

	return
}
