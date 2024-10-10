package role

import (
	"github.com/Lxb921006/Gin-bms/project/model"
)

type Menu struct {
	model.Permission
	Children []Menu `json:"children"`
}

func FormatUserPerms(p []model.Permission, pid uint) (m []Menu) {
	m = []Menu{}
	m1 := Menu{}
	for i := 0; i < len(p); i++ {
		if p[i].ParentId == pid {
			m1.Children = FormatUserPerms(p, p[i].ID)
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

func UpdateUserPerms(newPermsId []uint, rid uint) (err error) {
	var rm model.Role
	var oldPermsId []uint
	var delOldPermsId []uint

	oldPerms, err := rm.GetRolePerms(rid)
	if err != nil {
		return
	}

	if err = rm.AllotPerms(rid, newPermsId); err != nil {
		return
	}

	for _, v := range oldPerms {
		oldPermsId = append(oldPermsId, v.ID)
	}

	for _, v2 := range oldPermsId {
		flag := 0
		for _, v1 := range newPermsId {
			if v2 == v1 {
				flag = 1
				break
			}
		}
		if flag == 0 {
			delOldPermsId = append(delOldPermsId, v2)
		}
	}

	if len(delOldPermsId) >= 0 {
		_, err = rm.RemovePerms(rid, delOldPermsId)
		if err != nil {
			return
		}
	}

	return
}
