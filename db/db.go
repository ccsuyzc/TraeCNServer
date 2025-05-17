package config

import (
	 . "TraeCNServer/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "root:1234@tcp(127.0.0.1:3306)/myblog?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("链接数据库失败: %v", err))
	}

	err = AutoMigrate(DB)
	if err != nil {
		panic("迁移数据库失败")
	}
}
