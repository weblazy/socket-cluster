package gateway_logic

import (
	"socket_cluster/grpcs/socket_cluster_gateway/proto/gateway"

	"github.com/weblazy/easy/code_err"
)

type SendToClientIdsCtx struct {
	*code_err.Log
	Req *gateway.SendToClientIdsRequest
	Res *gateway.SendToClientIdsResponse
}

func SendToClientIds(ctx *SendToClientIdsCtx) *code_err.CodeErr {
	return nil
}
