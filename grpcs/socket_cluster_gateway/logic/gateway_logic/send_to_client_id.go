package gateway_logic

import (
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"

	"github.com/weblazy/easy/code_err"
)

type SendToClientIdCtx struct {
	*code_err.Log
	Req *gateway.SendToClientIdRequest
	Res *gateway.SendToClientIdResponse
}

func SendToClientId(ctx *SendToClientIdCtx) *code_err.CodeErr {
	return nil
}
