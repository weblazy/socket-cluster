package main

import (
	"websocket-cluster/examples/config"
	"websocket-cluster/examples/master"
	"websocket-cluster/examples/model"

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
	master.Node1()
	master.Node2()
	select {}
}
