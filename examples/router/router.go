package router

import (
	"github.com/weblazy/socket-cluster/examples/api"
	"github.com/weblazy/socket-cluster/protocol/ws_protocol"

	"github.com/labstack/echo/v4"
)

// Router
func Router(g *echo.Group) {
	g.POST("/login", api.Login)
	g.POST("/register", api.Register)
	g.POST("/sendSmsCode", api.SendSmsCode)
	g.POST("/chatInit", api.ChatInit, ws_protocol.OriginMiddlewareFunc)
	g.POST("/getGroupMembers", api.GetGroupMembers)
	g.POST("/search", api.Search)
	g.POST("/createGroup", api.CreateGroup)
	g.POST("/addFriend", api.AddFriend)
	g.POST("/manageSystemMsg", api.ManageSystemMsg)
	g.POST("/manageAddFriend", api.ManageAddFriend)
	g.POST("/joinGroup", api.JoinGroup)
	g.POST("/manageJoinGroup", api.ManageJoinGroup)
	g.POST("/getSystemMsg", api.GetSystemMsg)

	// Solve cross-domain problems
	g.OPTIONS("/login", ws_protocol.OptionHandler)
	g.OPTIONS("/register", ws_protocol.OptionHandler)
	g.OPTIONS("/sendSmsCode", ws_protocol.OptionHandler)
	g.OPTIONS("/chatInit", ws_protocol.OptionHandler, ws_protocol.OriginMiddlewareFunc)
	g.OPTIONS("/getGroupMembers", ws_protocol.OptionHandler)
	g.OPTIONS("/search", ws_protocol.OptionHandler)
	g.OPTIONS("/createGroup", ws_protocol.OptionHandler)
	g.OPTIONS("/addFriend", ws_protocol.OptionHandler)
	g.OPTIONS("/manageSystemMsg", ws_protocol.OptionHandler)
	g.OPTIONS("/manageAddFriend", ws_protocol.OptionHandler)
	g.OPTIONS("/joinGroup", ws_protocol.OptionHandler)
	g.OPTIONS("/manageJoinGroup", ws_protocol.OptionHandler)
	g.OPTIONS("/getSystemMsg", ws_protocol.OptionHandler)
}
