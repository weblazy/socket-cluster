package gateway_logic

import (
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"

	"github.com/weblazy/easy/code_err"
)

type IsOnlineCtx struct {
	*code_err.Log
	Req *gateway.IsOnlineRequest
	Res *gateway.IsOnlineResponse
}

func IsOnline(ctx *IsOnlineCtx) *code_err.CodeErr {
	return nil
}
