package websocket_cluster

import "github.com/weblazy/core/database/redis"

type (
	NodeConf struct {
		Host          string
		Path          string
		RedisNodeList []*RedisNode
		RedisConf     redis.RedisConf
		RedisMaxCount uint32
		Port          int64
		MasterAddress string //Master address
		Password      string //Password for auth when connect to master
		PingInterval  int64  //Heartbeat interval
		onMessage     func(nodeInfo *NodeInfo, context *Context)
	}

	Context struct {
		Conn    *Connection
		Uid     string
		Message []byte
	}
)

// NewPeer creates a new peer.
func NewNodeConf(host, path, masterAddress string, redisConf redis.RedisConf, redisNodeList []*RedisNode, onMessage func(nodeInfo *NodeInfo, context *Context)) *NodeConf {
	return &NodeConf{
		Host:          host,
		Path:          path,
		Port:          9528,
		RedisConf:     redisConf,
		MasterAddress: masterAddress,
		Password:      defaultPassword,
		PingInterval:  defaultPingInterval,
		RedisNodeList: redisNodeList,
		onMessage:     onMessage,
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
