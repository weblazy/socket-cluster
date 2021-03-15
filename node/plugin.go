package node

import "github.com/weblazy/socket-cluster/protocol"

type (
	Plugin interface {
		OnConnect(conn protocol.Connection)
		OnClose(conn protocol.Connection)
	}
	DefaultPlugin struct {
	}
)

var defaultPlugin = &DefaultPlugin{}

func (this *DefaultPlugin) OnConnect(conn protocol.Connection) {

}
func (this *DefaultPlugin) OnClose(conn protocol.Connection) {

}
