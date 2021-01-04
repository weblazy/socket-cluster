package model

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	g "github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

const TableNum int64 = 10

func Orm() *gorm.DB {
	db := g.GetORM("DbLocal")
	if utils.GetRunTime() == "local" {
		db = db.Debug()
	}
	return db
}

func CreateTable() {
	fmt.Println("开始初始化数据库")
	//自动建表，数据迁移
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='认证表' AUTO_INCREMENT=1;").AutoMigrate(&Auth{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='好友表' AUTO_INCREMENT=1;").AutoMigrate(&Friend{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='分组表' AUTO_INCREMENT=1;").AutoMigrate(&Group{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='用户分组绑定表' AUTO_INCREMENT=1;").AutoMigrate(&UserGroup{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='验证码表' AUTO_INCREMENT=1;").AutoMigrate(&SmsCode{})
	Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='系统消息表' AUTO_INCREMENT=1;").AutoMigrate(&SystemMsg{})

	for i := 0; i < int(TableNum); i++ {
		index := strconv.FormatInt(int64(i), 10)
		userMsgMap[int64(i)] = &UserMsg{
			table: "user_msg_" + index,
		}
		groupMsgMap[int64(i)] = &GroupMsg{
			table: "group_msg_" + index,
		}
		Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='用户消息表' AUTO_INCREMENT=1;").AutoMigrate(userMsgMap[int64(i)])
		Orm().Set("gorm:table_options", "CHARSET=utf8mb4 comment='分组消息表' AUTO_INCREMENT=1;").AutoMigrate(groupMsgMap[int64(i)])
	}
	fmt.Println("数据库初始化完成")
}
