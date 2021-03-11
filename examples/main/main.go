package main

import (
	"github.com/weblazy/socket-cluster/examples/config"
	"github.com/weblazy/socket-cluster/examples/master"
	"github.com/weblazy/socket-cluster/examples/model"

	"github.com/sunmi-OS/gocore/gorm"
	"github.com/sunmi-OS/gocore/utils"
)

func main() {
	// Initialize the configuration center
	config.InitNacos(utils.GetRunTime())
	// Initialize the database
	gorm.NewDB("DbLocal")
	model.CreateTable()
	// go websocket_cluster.StartMaster(websocket_cluster.NewMasterConf())
	master.Node()
	select {}
}
