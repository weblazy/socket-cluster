package node

import (
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/protocol"
)

type (
	// RedisNode the consistent hash redis node config
	RedisNode struct {
		RedisConf *redis.Options
		Position  uint32 //the position of hash ring
	}
	// NodeConf node config
	NodeConf struct {
		Host               string         // the ip or domain of the node
		ClientPath         string         // the client path
		TransPath          string         // the transport path
		RedisNodeList      []*RedisNode   // the slice of RedisNode
		RedisConf          *redis.Options // the redis config
		RedisMaxCount      uint32         // the hash ring node count
		Port               int64          // Node port
		Password           string         // Password for auth when connect to other node
		ClientPingInterval int64
		NodePingInterval   int64                  // Heartbeat interval
		onMsg              func(context *Context) // callback function when receive client message
		router             func(g *echo.Group)    // http router of echo
		echoObj            *echo.Echo             //Echo object
		discoveryHandler   discovery.ServiceDiscovery
		protocolHandler    protocol.Protocol
	}
	// Params of onMsg
	Context struct {
		Conn     *protocol.Connection
		ClientId string
		Msg      []byte
	}
)

// NewNodeConf creates a new NodeConf.
func NewNodeConf(host, clientPath, transPath string, redisConf *redis.Options, redisNodeList []*RedisNode, discoveryHandler discovery.ServiceDiscovery, onMsg func(context *Context)) *NodeConf {
	return &NodeConf{
		Host:               host,
		ClientPath:         clientPath,
		TransPath:          GetUUID(),
		Port:               defaultPort,
		RedisConf:          redisConf,
		Password:           defaultPassword,
		ClientPingInterval: defaultClientPingInterval,
		NodePingInterval:   defaultNodePingInterval,
		RedisNodeList:      redisNodeList,
		discoveryHandler:   discoveryHandler,
		onMsg:              onMsg,
		router:             func(g *echo.Group) {},
	}

}

// WithPassword sets the password for transport node
func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.Password = password
	return conf
}

// WithPort sets the port for websocket
func (conf *NodeConf) WithPort(port int64) *NodeConf {
	conf.Port = port
	return conf
}

// WithClientInterval sets the heartbeat interval
func (conf *NodeConf) WithClientInterval(pingInterval int64) *NodeConf {
	conf.ClientPingInterval = pingInterval
	return conf
}

// WithRouter sets the router
func (conf *NodeConf) WithRouter(router func(g *echo.Group)) *NodeConf {
	conf.router = router
	return conf
}

// WithEcho sets the echo
func (conf *NodeConf) WithEcho(echoObj *echo.Echo) *NodeConf {
	conf.echoObj = echoObj
	return conf
}
