package business_client

import (
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/node"
	"github.com/weblazy/socket-cluster/session_storage"
)

type (
	// NodeConf node config
	BusinessClientConf struct {
		password              string                         // Password for auth when connect to other node
		discoveryHandler      discovery.ServiceDiscovery     // Discover service
		sessionStorageHandler session_storage.SessionStorage // On-line state storage components
	}
)

// NewNodeConf creates a new NodeConf.
func NewBusinessClientConf(discoveryHandler discovery.ServiceDiscovery, sessionStorageHandler session_storage.SessionStorage) *BusinessClientConf {
	return &BusinessClientConf{
		password:              node.DefaultPassword,
		discoveryHandler:      discoveryHandler,
		sessionStorageHandler: sessionStorageHandler,
	}

}

// WithPassword sets the password for transport node
func (conf *BusinessClientConf) WithPassword(password string) *BusinessClientConf {
	conf.password = password
	return conf
}
