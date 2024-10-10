package model

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/service"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	RoleName    string `json:"rolename" gorm:"unique;not null"`
	Description string `json:"description" gorm:"-"`
	//不同步更新permission表
	//Permission  []Permission `json:"permission" gorm:"many2many:role_permissions;association_autoupdate:false;association_autocreate:false"`
	//同步更新permission表
	Permission []Permission `json:"permission" gorm:"many2many:role_permissions"`
	Mm         []Menu       `json:"mm" gorm:"-"`
}

type Menu struct {
	Permission
	Children []Menu `json:"children"`
}

type RolePermission struct {
	RoleID       uint `json:"roleId"`
	PermissionID uint `json:"permissionId"`
}

type RoleUser struct {
	RoleID uint `json:"roleId"`
	UserID uint `json:"userId"`
}

func (u *Role) CreateRole(rd Role) (err error) {
	if err = dao.DB.Create(&rd).Error; err != nil {
		return
	}
	return
}

// 分配权限
func (u *Role) AllotPerms(rid uint, pid []uint) (err error) {
	var p []Permission
	if err = dao.DB.Where("id IN ?", pid).Find(&p).Error; err != nil {
		return

	}
	if err = dao.DB.Where("id = ?", rid).Find(u).Error; err != nil {
		return
	}

	u.Permission = p

	if err = dao.DB.Save(&u).Error; err != nil {
		return
	}

	return
}

// 移除权限
func (u *Role) RemovePerms(rid uint, pid []uint) (p []Permission, err error) {
	var rp RolePermission
	tx := dao.DB.Begin()

	if err = tx.Where("role_id = ? AND permission_id IN ?", rid, pid).Delete(&rp).Error; err != nil {
		return
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	p, _ = u.GetRolePerms(rid)

	return
}

func (u *Role) DeleteRole(rid []uint) (err error) {
	tx := dao.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Where("role_id IN ?", rid).Delete(&RoleUser{}).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Where("id IN ?", rid).Unscoped().Delete(u).Error; err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

func (u *Role) GetAllRoles() (ul []Role, err error) {
	err = dao.DB.Find(&ul).Error
	if err != nil {
		return
	}
	return
}

func (u *Role) GetRolesList(page int, rolename Role) (data *service.Paginate, err error) {
	var rs []Role
	var frs []Role
	sql := dao.DB.Model(u).Where(rolename)
	pg := service.NewPaginate()
	data, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = data.Gd.Find(&rs).Error; err != nil {
		return
	}

	for i := 0; i < len(rs); i++ {
		p, _ := u.GetRolePerms(rs[i].ID)
		m := u.FormatUserPerms(p, 0)
		rs[i].Mm = m
		frs = append(frs, rs[i])
		dao.DB.Save(&rs[i])
	}

	data.ModelSlice = frs

	return
}

func (u *Role) GetUserPerms(uid uint) (p []Permission, err error) {
	var ud User
	// err = dao.DB.Model(&Permission{}).
	// 	Joins("inner join role_permissions on role_permissions.permission_id = permissions.id").
	// 	Joins("inner join role_users on role_users.role_id = role_permissions.role_id and role_users.user_id = ?", uid).
	// 	Select("permissions.id", "permissions.path", "permissions.title", "permissions.hidden", "permissions.parent_id", "permissions.level").Find(&p).Error
	// if err != nil {
	// 	return
	// }

	if err = dao.DB.Where("id = ?", uid).Find(&ud).Error; err != nil {
		return
	}

	if err = dao.DB.Model(&ud).Association("Role").Find(u); err != nil {
		return
	}

	p, err = u.GetRolePerms(u.ID)
	if err != nil {
		return
	}

	return
}

func (u *Role) GetRolePerms(rid uint) (p []Permission, err error) {
	err = dao.DB.Model(&Permission{}).
		Joins("inner join role_permissions on role_permissions.permission_id = permissions.id").
		Joins("inner join roles on roles.id = role_permissions.role_id and roles.id = ?", rid).
		Select("permissions.id", "permissions.path", "permissions.title", "permissions.hidden", "permissions.parent_id", "permissions.level").Find(&p).Error
	if err != nil {
		return
	}

	return
}

func (u *Role) GetAllFormatPerms() (p []Permission, err error) {
	err = dao.DB.Find(&p).Error
	if err != nil {
		return
	}
	return
}

func (u *Role) FormatUserPerms(p []Permission, pid uint) (m []Menu) {
	m = []Menu{}
	m1 := Menu{}
	for i := 0; i < len(p); i++ {
		if p[i].ParentId == pid {
			m1.Children = u.FormatUserPerms(p, p[i].ID)
			m1.ID = p[i].ID
			m1.ParentId = p[i].ParentId
			m1.Title = p[i].Title
			m1.Path = p[i].Path
			m1.Level = p[i].Level
			m = append(m, m1)
		}
	}
	return
}
