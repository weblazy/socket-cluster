package router

import (
	socket_cluster "socket-cluster"

	"github.com/weblazy/socket-cluster/examples/api"

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
	g.POST("/addFriend", api.AddFriend)
	g.POST("/manageSystemMsg", api.ManageSystemMsg)
	g.POST("/manageAddFriend", api.ManageAddFriend)
	g.POST("/joinGroup", api.JoinGroup)
	g.POST("/manageJoinGroup", api.ManageJoinGroup)
	g.POST("/getSystemMsg", api.GetSystemMsg)

	//解决跨域问题
	g.OPTIONS("/login", socket_cluster.OptionHandler)
	g.OPTIONS("/register", socket_cluster.OptionHandler)
	g.OPTIONS("/sendSmsCode", socket_cluster.OptionHandler)
	g.OPTIONS("/chatInit", socket_cluster.OptionHandler)
	g.OPTIONS("/getGroupMembers", socket_cluster.OptionHandler)
	g.OPTIONS("/search", socket_cluster.OptionHandler)
	g.OPTIONS("/createGroup", socket_cluster.OptionHandler)
	g.OPTIONS("/addFriend", socket_cluster.OptionHandler)
	g.OPTIONS("/manageSystemMsg", socket_cluster.OptionHandler)
	g.OPTIONS("/manageAddFriend", socket_cluster.OptionHandler)
	g.OPTIONS("/joinGroup", socket_cluster.OptionHandler)
	g.OPTIONS("/manageJoinGroup", socket_cluster.OptionHandler)
	g.OPTIONS("/getSystemMsg", socket_cluster.OptionHandler)
}
