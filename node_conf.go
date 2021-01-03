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
		Host               string
		ClientPath         string
		TransPath          string
		RedisNodeList      []*RedisNode
		RedisConf          RedisConf
		RedisMaxCount      uint32
		Port               int64
		Password           string //Password for auth when connect to master
		ClientPingInterval int64
		NodePingInterval   int64 //Heartbeat interval
		onMsg              func(context *Context)
		router             func(g *echo.Group)
		echoObj            *echo.Echo
	}

	Context struct {
		Conn     *Connection
		ClientId string
		Msg      []byte
	}
)

// NewPeer creates a new peer.
func NewNodeConf(host, clientPath, transPath string, redisConf RedisConf, redisNodeList []*RedisNode, onMsg func(context *Context)) *NodeConf {
	return &NodeConf{
		Host:               host,
		ClientPath:         clientPath,
		TransPath:          transPath,
		Port:               9528,
		RedisConf:          redisConf,
		Password:           defaultPassword,
		ClientPingInterval: defaultClientPingInterval,
		NodePingInterval:   defaultNodePingInterval,
		RedisNodeList:      redisNodeList,
		onMsg:              onMsg,
		router:             func(g *echo.Group) {},
	}

}

// set the password for transport node
func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.Password = password
	return conf
}

// set the port for websocket
func (conf *NodeConf) WithPort(port int64) *NodeConf {
	conf.Port = port
	return conf
}

// set the heartbeat interval
func (conf *NodeConf) WithClientInterval(pingInterval int64) *NodeConf {
	conf.ClientPingInterval = pingInterval
	return conf
}

// set the router
func (conf *NodeConf) WithRouter(router func(g *echo.Group)) *NodeConf {
	conf.router = router
	return conf
}

//set the echo
func (conf *NodeConf) WithEcho(echoObj *echo.Echo) *NodeConf {
	conf.echoObj = echoObj
	return conf
}
