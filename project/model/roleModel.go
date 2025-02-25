package model

import (
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/errors"
	"github.com/ingoxx/go-gin/project/service"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	RoleName    string `json:"rolename" gorm:"unique;not null"`
	Description string `json:"description" gorm:"-"`
	//不同步更新permission表
	//Permissions  []Permissions `json:"permission" gorm:"many2many:role_permissions;association_autoupdate:false;association_autocreate:false"`
	//同步更新permission表
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions"`
	Mm          []*Menu      `json:"mm" gorm:"-"`
	Users       []User       `json:"users" gorm:"many2many:role_users"`
}

type Menu struct {
	Permission
	Children []*Menu `json:"children"`
}

type RolePermission struct {
	RoleID       uint `json:"roleId"`
	PermissionID uint `json:"permissionId"`
}

type RoleUser struct {
	RoleID uint `json:"roleId"`
	UserID uint `json:"userId"`
}

func (rl *Role) CreateRole(rd Role) (err error) {
	if err = dao.DB.Create(&rd).Error; err != nil {
		return
	}
	return
}

func (rl *Role) GetPermNames(pid []uint) []string {
	var perms []Permission
	var pns []string
	if err := dao.DB.Where("id IN ?", pid).Find(&perms).Error; err != nil {
		return pns
	}

	for _, v := range perms {
		pns = append(pns, v.Title)
	}

	return pns
}

func (rl *Role) GetRoleNames(rid []uint) []string {
	var role []Role
	var pns []string
	if err := dao.DB.Where("id IN ?", rid).Find(&role).Error; err != nil {
		return pns
	}

	for _, v := range role {
		pns = append(pns, v.RoleName)
	}

	return pns
}

// AllotPerms 分配权限
func (rl *Role) AllotPerms(rid uint, pid []uint) (err error) {
	var perms []Permission
	var role Role
	tx := dao.DB.Begin()

	if err = dao.DB.Where("id IN ?", pid).Find(&perms).Error; err != nil {
		return
	}

	if len(perms) == 0 {
		return errors.EmptyPermListError
	}

	if err = dao.DB.Where("id = ?", rid).Find(&role).Error; err != nil {
		return
	}

	if err = tx.Model(&role).Association("Permissions").Replace(&perms); err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit().Error
}

// RemovePerms 移除权限
func (rl *Role) RemovePerms(rid uint, pid []uint) (perms []Permission, err error) {
	tx := dao.DB.Begin() // 开启事务

	// 1. 通过原生 SQL 直接删除 role_permissions 中间表的关联数据
	if err = tx.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id IN (?)", rid, pid).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Preload("Permissions").First(&rl, rid).Error; err != nil {
		tx.Rollback()
		return
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	perms = rl.Permissions
	return
}

func (rl *Role) DeleteRole(rid []uint) (err error) {
	var role Role

	tx := dao.DB.Begin()

	for _, v := range rid {
		if err = dao.DB.First(&role, v).Error; err != nil {
			return
		}

		if role.RoleName == "管理员" {
			continue
		}

		if err = tx.Model(&role).Association("Permissions").Clear(); err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Model(&role).Association("Users").Clear(); err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Unscoped().Delete(&role, v).Error; err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit().Error; err != nil {
			return
		}
	}

	return
}

func (rl *Role) GetAllRoles() (ul []Role, err error) {
	err = dao.DB.Find(&ul).Error
	if err != nil {
		return
	}
	return
}

func (rl *Role) GetRolesList(page int, rolename Role) (data *service.Paginate, err error) {
	var rs []Role
	var frs []Role
	sql := dao.DB.Model(rl).Where(rolename)
	pg := service.NewPaginate()
	data, err = pg.GetPageData(page, sql)
	if err != nil {
		return
	}

	if err = data.Gd.Find(&rs).Error; err != nil {
		return
	}

	for i := 0; i < len(rs); i++ {
		p, _ := rl.GetRolePerms(rs[i].ID)
		m := rl.FormatUserPerms(p, 0)
		rs[i].Mm = m
		frs = append(frs, rs[i])
		dao.DB.Save(&rs[i])
	}

	data.ModelSlice = frs

	return
}

func (rl *Role) GetUserPerms(uid uint) (p []Permission, err error) {
	var user User

	if err = dao.DB.Preload("Roles.Permissions").First(&user, uid).Error; err != nil {
		return
	}

	permissionsMap := make(map[uint]Permission)
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissionsMap[permission.ID] = permission
		}
	}

	p = make([]Permission, 0, len(permissionsMap))
	for _, permission := range permissionsMap {
		p = append(p, permission)
	}

	return
}

func (rl *Role) GetRolePerms(rid uint) (p []Permission, err error) {
	var role Role
	if err = dao.DB.Preload("Permissions").Find(&role, rid).Error; err != nil {
		return
	}

	return role.Permissions, nil
}

func (rl *Role) GetAllFormatPerms() (p []Permission, err error) {
	err = dao.DB.Find(&p).Error
	if err != nil {
		return
	}
	return
}

func (rl *Role) FormatUserPerms(perms []Permission, pid uint) []*Menu {
	menuMap := make(map[uint]*Menu)
	for _, perm := range perms {
		menuMap[perm.ID] = &Menu{Permission: perm, Children: []*Menu{}}
	}

	var roots []*Menu
	// 第二次遍历，把节点正确地添加到其父节点的 Children 里
	for _, menu := range menuMap {
		if menu.ParentId == pid {
			// 父 ID 为 0，说明是根节点，加入 roots
			roots = append(roots, menu)
		} else if parent, exists := menuMap[menu.ParentId]; exists {
			// 找到父节点，加入其 Children
			parent.Children = append(parent.Children, menu)
		} else {
			// 父 ID 不为 0，但在 map 里找不到对应的父级，说明数据不完整
			// 这里选择直接当成根节点，避免数据丢失
			roots = append(roots, menu)
		}
	}

	return roots
}
