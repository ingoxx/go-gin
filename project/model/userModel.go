package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"unique;not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Hobby    string `json:"-" gorm:"default:'basketball'"`
	Tel      int    `json:"tel" gorm:"default:168888"`
	Password string `json:"-" gorm:"not null"`
	Role     []Role `json:"role" gorm:"many2many:role_users"`
	Isopenga uint   `json:"isopenga" gorm:"default:1"`
	Isopenqr uint   `json:"isopenqr" gorm:"default:1"`
}

func (u *User) AddUser(d User, rid uint) (err error) {
	if err = dao.DB.Create(&d).Error; err != nil {
		return
	}

	if err = u.AssignRoles(d.Name, rid); err != nil {
		return
	}

	return
}

// 分配角色
func (u *User) AssignRoles(name string, rid uint) (err error) {
	var role Role

	if err = dao.DB.Where("name = ?", name).Find(u).Error; err != nil {
		return
	}

	if err = dao.DB.Where("id = ?", rid).Find(&role).Error; err != nil {
		return
	}

	u.Role = append(u.Role, role)
	dao.DB.Save(u)
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

func (u *User) UpdateUser(ud User, rid uint) (err error) {
	var ur RoleUser
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Model(u).Where("id = ?", ud.ID).Updates(ud).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Model(&ur).Where("user_id = ?", ud.ID).Update("role_id", rid).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (u *User) GetUserByName(name string) (ud User, err error) {
	if err = dao.DB.Model(u).Where("name = ?", name).Preload("Role").Find(&ud).Error; err != nil {
		return
	}

	return
}

// 单表中过滤出row
func (u *User) GetUserByPaginate(page int, d User) (ul *service.Paginate, err error) {
	var us []User
	sql := dao.DB.Model(u).Where(d).Preload("Role")
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

// 通过m2m关系表中过滤出row
func (u *User) GetUserByMmPaginate(page int, rolename string, user User) (ul *service.Paginate, err error) {
	var us []User
	var uid []uint
	if err = dao.DB.Preload("Role", Role{RoleName: rolename}).Where(&user).Find(&us).Error; err != nil {
		return
	}

	for _, v := range u.FormatData(us) {
		uid = append(uid, v.ID)
	}

	sql := dao.DB.Model(u).Where("id IN ?", uid).Preload("Role")
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
		if len(v.Role) != 0 {
			usd = append(usd, v)
		}
	}
	return
}
