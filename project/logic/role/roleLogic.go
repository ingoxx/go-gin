package role

import (
	"github.com/ingoxx/go-gin/project/model"
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
	var role model.Role

	if err = role.AllotPerms(rid, newPermsId); err != nil {
		return
	}

	return
}

func mergeNotPidPerms(p []model.Permission) (m []Menu) {

	return
}
