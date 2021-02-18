package main

import (
	"github.com/weblazy/socket-cluster/examples/config"
	"github.com/weblazy/socket-cluster/examples/master"
	"github.com/weblazy/socket-cluster/examples/model"

	"github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

func main() {
	//初始化配置中心
	config.InitNacos(utils.GetRunTime())
	// 初始化数据库
	gorm.NewDB("dbDefault")
	model.CreateTable()
	// go websocket_cluster.StartMaster(websocket_cluster.NewMasterConf())
	master.Node()
	select {}
}
