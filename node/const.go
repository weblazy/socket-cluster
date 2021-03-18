package node

import "time"

const (
	clientPrefix              = "client#"
	defaultPassword           = "password"
	defaultClientPingInterval = 120
	defaultNodePingInterval   = 10
	defaultInternalPort       = 9527
	defaultPort               = 9528
	NodeAddress               = "node_address"
	authTime                  = 10 * time.Second
)

const (
	AuthMsgType   int32 = 1
	ClientMsgType int32 = 2
	PingMsgType   int32 = 3
)
