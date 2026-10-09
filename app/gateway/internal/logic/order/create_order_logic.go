package order

import (
	"context"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/client/orderservice"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/ctxdata"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/proto/enum"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.Empty, err error) {
	uid, err := ctxdata.GetUid(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.OrderRpc.CreateOrder(l.ctx, &orderservice.CreateOrderReq{
		UserId:      uid,
		SymbolName:  req.SymbolName,
		BaseAmount:  req.BaseAmount,
		Price:       req.Price,
		QuoteAmount: req.QuoteAmount,
		Side:        enum.Side(req.Side),
		OrderType:   enum.OrderType(req.OrderType),
	})
	if err != nil {
		l.Logger.Errorf("create order failed: %v", err)
		return nil, err
	}
	return &types.Empty{}, nil
}
