package node

import (
	"github.com/go-redis/redis/v8"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/protocol"
	"github.com/weblazy/socket-cluster/protocol/internal_protocol"
	"github.com/weblazy/socket-cluster/session_storage"
)

type (
	// RedisNode the consistent hash redis node config
	RedisNode struct {
		RedisConf *redis.Options
		Position  uint32 //the position of hash ring
	}
	// NodeConf node config
	NodeConf struct {
		HostList                []string // A list of IP or domain names for DNS resolution
		Host                    string   // the ip or domain of the node
		NodeId                  string   // Node unique identification
		Port                    int64    // Node port
		InternalPort            int64    // Node port
		Password                string   // Password for auth when connect to other node
		ClientPingInterval      int64
		NodePingInterval        int64                  // Heartbeat interval
		onMsg                   func(context *Context) // callback function when receive client message
		discoveryHandler        discovery.ServiceDiscovery
		protocolHandler         protocol.Protocol
		internalProtocolHandler protocol.Protocol
		sessionStorageHandler   session_storage.SessionStorage
		plugin                  Plugin
	}
	// Params of onMsg
	Context struct {
		Conn     protocol.Connection
		ClientId string
		Msg      []byte
	}
)

// NewNodeConf creates a new NodeConf.
func NewNodeConf(host string, protocolHandler protocol.Protocol, sessionStorageHandler session_storage.SessionStorage, discoveryHandler discovery.ServiceDiscovery, onMsg func(context *Context)) *NodeConf {
	return &NodeConf{
		Host:                    host,
		HostList:                []string{host},
		NodeId:                  GetUUID(),
		Port:                    defaultPort,
		Password:                defaultPassword,
		ClientPingInterval:      defaultClientPingInterval,
		NodePingInterval:        defaultNodePingInterval,
		protocolHandler:         protocolHandler,
		internalProtocolHandler: &internal_protocol.TcpProtocol{},
		InternalPort:            defaultInternalPort,
		sessionStorageHandler:   sessionStorageHandler,
		discoveryHandler:        discoveryHandler,
		onMsg:                   onMsg,
		plugin:                  defaultPlugin,
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

// WithInternalPort sets the port for internal protocol
func (conf *NodeConf) WithInternalPort(port int64) *NodeConf {
	conf.InternalPort = port
	return conf
}

// WithInternalProtocolHandler sets the internal protocol for node
func (conf *NodeConf) WithInternalProtocolHandler(internalProtocolHandler protocol.Protocol) *NodeConf {
	conf.internalProtocolHandler = internalProtocolHandler
	return conf
}

// WithClientInterval sets the heartbeat interval
func (conf *NodeConf) WithClientInterval(pingInterval int64) *NodeConf {
	conf.ClientPingInterval = pingInterval
	return conf
}

// WithHostList sets the host list for node cluster
func (conf *NodeConf) WithHostList(hostList []string) *NodeConf {
	conf.HostList = hostList
	return conf
}

// WithPlugin sets the plugin
func (conf *NodeConf) WithPlugin(plugin Plugin) *NodeConf {
	conf.plugin = plugin
	return conf
}
