package account

import (
	"context"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/client/accountservice"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/ctxdata"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddUserAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserAssetLogic {
	return &AddUserAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserAssetLogic) AddUserAsset(req *types.AddUserAssetReq) (resp *types.Empty, err error) {
	uid, err := ctxdata.GetUid(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.AccountRpc.AddUserAsset(l.ctx, &accountservice.AddUserAssetReq{
		Uid:      uid,
		CoinName: req.CoinName,
		Amount:   req.Qty,
	})
	if err != nil {
		l.Logger.Errorf("add user asset failed: %v", err)
		return nil, err
	}
	return &types.Empty{}, nil
}
