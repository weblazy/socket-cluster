package websocket_cluster

import (
	"github.com/labstack/echo/v4"
)

type (
	RedisConf struct {
		Addr     string
		Password string
		DB       int64
	}
	NodeConf struct {
		Host          string
		Path          string
		RedisNodeList []*RedisNode
		RedisConf     RedisConf
		RedisMaxCount uint32
		Port          int64
		Password      string //Password for auth when connect to master
		PingInterval  int64  //Heartbeat interval
		onMsg         func(nodeInfo *NodeInfo, context *Context)
		router        func(g *echo.Group)
	}

	Context struct {
		Conn *Connection
		Uid  string
		Msg  []byte
	}
)

// NewPeer creates a new peer.
func NewNodeConf(host, path string, redisConf RedisConf, redisNodeList []*RedisNode, onMsg func(nodeInfo *NodeInfo, context *Context)) *NodeConf {
	return &NodeConf{
		Host:          host,
		Path:          path,
		Port:          9528,
		RedisConf:     redisConf,
		Password:      defaultPassword,
		PingInterval:  defaultPingInterval,
		RedisNodeList: redisNodeList,
		onMsg:         onMsg,
		router:        func(g *echo.Group) {},
	}

}

func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.Password = password
	return conf
}

func (conf *NodeConf) WithPort(port int64) *NodeConf {
	conf.Port = port
	return conf
}

func (conf *NodeConf) WithPing(pingInterval int64) *NodeConf {
	conf.PingInterval = pingInterval
	return conf
}

func (conf *NodeConf) WithRouter(router func(g *echo.Group)) *NodeConf {
	conf.router = router
	return conf
}
