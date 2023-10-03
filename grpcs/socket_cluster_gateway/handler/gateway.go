package handler

import (
	"context"

	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/logic/gateway_logic"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"
	"github.com/weblazy/socket-cluster/node"
	"go.uber.org/zap"

	"github.com/weblazy/easy/code_err"
	"github.com/weblazy/easy/econfig"
	"github.com/weblazy/easy/elog"
)

type GatewayService struct {
	gateway.UnimplementedGatewayServiceServer
	Node node.Node
}

func NewGatewayService(n node.Node) *GatewayService {
	return &GatewayService{
		Node: n,
	}
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
	if econfig.GlobalViper.GetBool("BaseConfig.Debug") {
		elog.InfoCtx(ctx, "SendToClientId", zap.Any("req", req))
	}
	resp := gateway.SendToClientIdResponse{}
	err := h.Node.SendToClientId(req.ClientId, req.Data)
	if err != nil {
		resp.Code = -1
		resp.Msg = err.Error()
	}
	return &resp, nil
}

func (h *GatewayService) SendToClientIds(ctx context.Context, req *gateway.SendToClientIdsRequest) (*gateway.SendToClientIdsResponse, error) {
	resp := gateway.SendToClientIdsResponse{}
	err := h.Node.SendToClientIds(req.ClientIds, req.Data)
	if err != nil {
		resp.Code = -1
		resp.Msg = err.Error()
	}
	return &resp, nil
}
