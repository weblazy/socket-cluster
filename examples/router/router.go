package router

import (
	websocket_cluster "websocket-cluster"
	"websocket-cluster/examples/api"

	"github.com/labstack/echo/v4"
)

// @desc
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Router(g *echo.Group) {
	g.POST("/login", api.Login)
	g.POST("/register", api.Register)
	g.POST("/sendSmsCode", api.SendSmsCode)
	g.POST("/chatInit", api.ChatInit)
	g.POST("/getGroupMembers", api.GetGroupMembers)
	g.POST("/search", api.Search)
	g.POST("/createGroup", api.CreateGroup)

	g.OPTIONS("/login", websocket_cluster.OptionHandler)
	g.OPTIONS("/register", websocket_cluster.OptionHandler)
	g.OPTIONS("/sendSmsCode", websocket_cluster.OptionHandler)
	g.OPTIONS("/chatInit", websocket_cluster.OptionHandler)
	g.OPTIONS("/getGroupMembers", websocket_cluster.OptionHandler)
	g.OPTIONS("/search", websocket_cluster.OptionHandler)
	g.OPTIONS("/createGroup", websocket_cluster.OptionHandler)
}