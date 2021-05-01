package business_client

import (
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/node"
	"github.com/weblazy/socket-cluster/protocol"
	"github.com/weblazy/socket-cluster/protocol/tcp_protocol"
	"github.com/weblazy/socket-cluster/session_storage"
)

type (
	// NodeConf node config
	BusinessClientConf struct {
		hostList                []string                       // A list of IP or domain names for DNS resolution
		password                string                         // Password for auth when connect to other node
		nodePingInterval        int64                          // Node with node Heartbeat interval
		onMsg                   func(context *node.Context)    // Callback function when receive client message
		discoveryHandler        discovery.ServiceDiscovery     // Discover service
		internalProtocolHandler protocol.Protocol              // Direct protocol between node and node
		sessionStorageHandler   session_storage.SessionStorage // On-line state storage components
	}
)

// NewNodeConf creates a new NodeConf.
func NewBusinessClientConf(hostList []string, discoveryHandler discovery.ServiceDiscovery, onMsg func(context *node.Context)) *BusinessClientConf {
	return &BusinessClientConf{
		hostList:                hostList,
		password:                node.DefaultPassword,
		nodePingInterval:        node.DefaultNodePingInterval,
		internalProtocolHandler: &tcp_protocol.TcpProtocol{},
		discoveryHandler:        discoveryHandler,
		onMsg:                   onMsg,
	}

}

// WithPassword sets the password for transport node
func (conf *BusinessClientConf) WithPassword(password string) *BusinessClientConf {
	conf.password = password
	return conf
}

// WithInternalProtocolHandler sets the internal protocol for node
func (conf *BusinessClientConf) WithInternalProtocolHandler(internalProtocolHandler protocol.Protocol) *BusinessClientConf {
	conf.internalProtocolHandler = internalProtocolHandler
	return conf
}
