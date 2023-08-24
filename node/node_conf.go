package node

import (
	"github.com/go-redis/redis/v8"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/protocol"
	"github.com/weblazy/socket-cluster/session_storage"
)

type (
	// RedisNode the hash redis node config
	RedisNode struct {
		RedisConf *redis.Options
		Position  uint32 // The position of hash
	}
	// NodeConf node config
	NodeConf struct {
		addr               string                     // The ip or domain of the node
		port               int64                      // Node clientPort
		internalPort       int64                      // Node internalPort
		password           string                     // Password for auth when connect to other node
		clientPingInterval int64                      // Node with client heartbeat interval
		nodePingInterval   int64                      // Node with node Heartbeat interval
		onMsg              func(context *Context)     // Callback function when receive client message
		discoveryHandler   discovery.ServiceDiscovery // Discover service
		protocolHandler    protocol.Protocol          // Direct protocol between node and client
		// internalProtocolHandler protocol.Protocol              // Direct protocol between node and node
		sessionStorageHandler session_storage.SessionStorage // On-line state storage components
		plugin                Plugin                         // The interface that the client connects to or closes
	}
	// Params of onMsg
	Context struct {
		Conn     protocol.Connection
		ClientId string
		Msg      []byte
	}
)

// NewNodeConf creates a new NodeConf.
func NewNodeConf(addr string, protocolHandler protocol.Protocol, sessionStorageHandler session_storage.SessionStorage, discoveryHandler discovery.ServiceDiscovery, onMsg func(context *Context)) *NodeConf {
	return &NodeConf{
		addr:               defaultAddr,
		port:               defaultPort,
		password:           DefaultPassword,
		clientPingInterval: defaultClientPingInterval,
		nodePingInterval:   DefaultNodePingInterval,
		protocolHandler:    protocolHandler,
		// internalProtocolHandler: &tcp_protocol.TcpProtocol{},
		internalPort:          defaultInternalPort,
		sessionStorageHandler: sessionStorageHandler,
		discoveryHandler:      discoveryHandler,
		onMsg:                 onMsg,
		plugin:                defaultPlugin,
	}

}

// WithPassword sets the password for transport node
func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.password = password
	return conf
}

// WithPort sets the port
func (conf *NodeConf) WithPort(port int64) *NodeConf {
	conf.port = port
	return conf
}

// WithInternalPort sets the port for internal protocol
func (conf *NodeConf) WithInternalPort(port int64) *NodeConf {
	conf.internalPort = port
	return conf
}

// WithInternalProtocolHandler sets the internal protocol for node
// func (conf *NodeConf) WithInternalProtocolHandler(internalProtocolHandler protocol.Protocol) *NodeConf {
// 	conf.internalProtocolHandler = internalProtocolHandler
// 	return conf
// }

// WithClientInterval sets the heartbeat interval
func (conf *NodeConf) WithClientInterval(pingInterval int64) *NodeConf {
	conf.clientPingInterval = pingInterval
	return conf
}

// WithAddr sets the addr for node cluster
func (conf *NodeConf) WithAddr(addr string) *NodeConf {
	conf.addr = addr
	return conf
}

// WithPlugin sets the plugin
func (conf *NodeConf) WithPlugin(plugin Plugin) *NodeConf {
	conf.plugin = plugin
	return conf
}
