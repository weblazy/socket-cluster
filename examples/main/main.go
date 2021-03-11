package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/weblazy/socket-cluster/examples/config"
	"github.com/weblazy/socket-cluster/examples/master"
	"github.com/weblazy/socket-cluster/examples/model"
	"github.com/weblazy/socket-cluster/examples/router"
	"github.com/weblazy/socket-cluster/protocol/ws_protocol"

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
	e := echo.New()
	router.Router(e.Group("/p1/web", ws_protocol.OriginMiddlewareFunc))
	go func() {
		err := e.Start(fmt.Sprintf(":%d", 80))
		if err != nil {
			panic(err)
		}
	}()
	master.Node()
	select {}
}
