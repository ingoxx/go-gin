package main

import (
	"github.com/ingoxx/go-gin/project/dao"
	"github.com/ingoxx/go-gin/project/model"
	"testing"
)

func TestAddUser(t *testing.T) {
	dao.DB.AutoMigrate(&model.User{})

}
