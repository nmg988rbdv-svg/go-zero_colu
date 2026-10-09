package account

import (
	"context"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/account/rpc/client/accountservice"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/ctxdata"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/errs"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginOutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginOutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginOutLogic {
	return &LoginOutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginOutLogic) LoginOut() (resp *types.Empty, err error) {
	uid, err := ctxdata.GetUid(l.ctx)
	if err != nil {
		return nil, err
	}
	token := ctxdata.GetToken(l.ctx)
	if token == "" {
		return nil, errs.TokenValidateFailed
	}

	_, err = l.svcCtx.AccountRpc.LoginOut(l.ctx, &accountservice.LoginOutReq{
		Token: token,
		Uid:   uid,
	})
	if err != nil {
		l.Logger.Errorf("logout failed: %v", err)
		return nil, err
	}
	return &types.Empty{}, nil
}
