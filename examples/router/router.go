package router

import (
	"github.com/weblazy/socket-cluster/examples/api"
	"github.com/weblazy/socket-cluster/node"

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
	g.OPTIONS("/login", node.OptionHandler)
	g.OPTIONS("/register", node.OptionHandler)
	g.OPTIONS("/sendSmsCode", node.OptionHandler)
	g.OPTIONS("/chatInit", node.OptionHandler)
	g.OPTIONS("/getGroupMembers", node.OptionHandler)
	g.OPTIONS("/search", node.OptionHandler)
	g.OPTIONS("/createGroup", node.OptionHandler)
	g.OPTIONS("/addFriend", node.OptionHandler)
	g.OPTIONS("/manageSystemMsg", node.OptionHandler)
	g.OPTIONS("/manageAddFriend", node.OptionHandler)
	g.OPTIONS("/joinGroup", node.OptionHandler)
	g.OPTIONS("/manageJoinGroup", node.OptionHandler)
	g.OPTIONS("/getSystemMsg", node.OptionHandler)
}
