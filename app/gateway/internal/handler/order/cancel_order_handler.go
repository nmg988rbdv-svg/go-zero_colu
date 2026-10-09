// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/logic/order"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/errs"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 取消订单
func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(w, r, nil, errs.WarpMessage(errs.ParamValidateFailed, err.Error()))
			return
		}
		l := order.NewCancelOrderLogic(r.Context(), svcCtx)
		resp, err := l.CancelOrder(&req)
		response.Response(w, r, resp, err)

	}
}
