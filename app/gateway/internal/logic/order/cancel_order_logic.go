package order

import (
	"context"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/client/orderservice"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/ctxdata"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelOrderLogic) CancelOrder(req *types.CancelOrderReq) (resp *types.Empty, err error) {
	uid, err := ctxdata.GetUid(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.OrderRpc.CancelOrder(l.ctx, &orderservice.CancelOrderReq{
		OrderId:    req.ID,
		Uid:        uid,
		SymbolName: req.SymbolName,
	})
	if err != nil {
		l.Logger.Errorf("cancel order failed: %v", err)
		return nil, err
	}
	return &types.Empty{}, nil
}
