package router

import (
	"github.com/weblazy/socket-cluster/examples/api"
	"github.com/weblazy/socket-cluster/protocol/websocket_protocol"

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
	g.OPTIONS("/login", websocket_protocol.OptionHandler)
	g.OPTIONS("/register", websocket_protocol.OptionHandler)
	g.OPTIONS("/sendSmsCode", websocket_protocol.OptionHandler)
	g.OPTIONS("/chatInit", websocket_protocol.OptionHandler)
	g.OPTIONS("/getGroupMembers", websocket_protocol.OptionHandler)
	g.OPTIONS("/search", websocket_protocol.OptionHandler)
	g.OPTIONS("/createGroup", websocket_protocol.OptionHandler)
	g.OPTIONS("/addFriend", websocket_protocol.OptionHandler)
	g.OPTIONS("/manageSystemMsg", websocket_protocol.OptionHandler)
	g.OPTIONS("/manageAddFriend", websocket_protocol.OptionHandler)
	g.OPTIONS("/joinGroup", websocket_protocol.OptionHandler)
	g.OPTIONS("/manageJoinGroup", websocket_protocol.OptionHandler)
	g.OPTIONS("/getSystemMsg", websocket_protocol.OptionHandler)
}
