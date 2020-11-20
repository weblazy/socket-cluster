package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	g "github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

func Orm() *gorm.DB {
	db := g.Gorm().GetORM("dbDefault")
	if utils.GetRunTime() == "local" {
		db = db.Debug()
	}
	return db
}

func CreateTable() {
	fmt.Println("开始初始化数据库")
	//自动建表，数据迁移
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='消息表' AUTO_INCREMENT=1;").AutoMigrate(&Message{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='认证表' AUTO_INCREMENT=1;").AutoMigrate(&Auth{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='好友表' AUTO_INCREMENT=1;").AutoMigrate(&Friend{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='分组表' AUTO_INCREMENT=1;").AutoMigrate(&Group{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='用户分组绑定表' AUTO_INCREMENT=1;").AutoMigrate(&UserGroup{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='分组消息表' AUTO_INCREMENT=1;").AutoMigrate(&GroupMessage{})
	fmt.Println("数据库初始化完成")
}
