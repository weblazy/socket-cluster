package model

import (
	"fmt"
)

func CreateTable() {
	fmt.Println("开始初始化数据库")
	//自动建表，数据迁移
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='消息表' AUTO_INCREMENT=1;").AutoMigrate(&Message{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='认证表' AUTO_INCREMENT=1;").AutoMigrate(&Auth{})
	fmt.Println("数据库初始化完成")
}
