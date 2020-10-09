package websocket_cluster

const (
	userPrefix           = "user#"
	groupPrefix          = "group#"
	defaultMasterAddress = "127.0.0.1:9527"
	defaultPassword      = "password"
	defaultPingInterval  = 1
	redisInterval        = 10
	redisZsortKey        = "tpcluster_node"
)
