package model

import (
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
