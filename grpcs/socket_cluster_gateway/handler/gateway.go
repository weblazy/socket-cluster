package handler

import (
	"context"

	"socket_cluster/grpcs/socket_cluster_gateway/logic/gateway_logic"
	"socket_cluster/grpcs/socket_cluster_gateway/proto/gateway"

	"github.com/weblazy/easy/code_err"
)

type GatewayService struct {
	gateway.UnimplementedGatewayServiceServer
}

func NewGatewayService() *GatewayService {
	return &GatewayService{}
}

func (h *GatewayService) IsOnline(ctx context.Context, req *gateway.IsOnlineRequest) (*gateway.IsOnlineResponse, error) {
	svcCtx := &gateway_logic.IsOnlineCtx{
		Log: code_err.NewLog(ctx),
		Req: req,
		Res: new(gateway.IsOnlineResponse),
	}
	err := gateway_logic.IsOnline(svcCtx)
	if err != nil {
		svcCtx.Res.Code = err.Code
		svcCtx.Res.Msg = err.Msg
	}
	return svcCtx.Res, nil
}

func (h *GatewayService) SendToClientId(ctx context.Context, req *gateway.SendToClientIdRequest) (*gateway.SendToClientIdResponse, error) {
	svcCtx := &gateway_logic.SendToClientIdCtx{
		Log: code_err.NewLog(ctx),
		Req: req,
		Res: new(gateway.SendToClientIdResponse),
	}
	err := gateway_logic.SendToClientId(svcCtx)
	if err != nil {
		svcCtx.Res.Code = err.Code
		svcCtx.Res.Msg = err.Msg
	}
	return svcCtx.Res, nil
}

func (h *GatewayService) SendToClientIds(ctx context.Context, req *gateway.SendToClientIdsRequest) (*gateway.SendToClientIdsResponse, error) {
	svcCtx := &gateway_logic.SendToClientIdsCtx{
		Log: code_err.NewLog(ctx),
		Req: req,
		Res: new(gateway.SendToClientIdsResponse),
	}
	err := gateway_logic.SendToClientIds(svcCtx)
	if err != nil {
		svcCtx.Res.Code = err.Code
		svcCtx.Res.Msg = err.Msg
	}
	return svcCtx.Res, nil
}
