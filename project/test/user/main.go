package main

import (
	"github.com/Lxb921006/Gin-bms/project/dao"
	"github.com/Lxb921006/Gin-bms/project/model"
	"testing"
)

func TestAddUser(t *testing.T) {
	dao.DB.AutoMigrate(&model.User{})

}
