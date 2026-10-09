// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/logic/account"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/types"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/errs"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/response"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// 登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(w, r, nil, errs.WarpMessage(errs.ParamValidateFailed, err.Error()))
			return
		}
		l := account.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		response.Response(w, r, resp, err)

	}
}
