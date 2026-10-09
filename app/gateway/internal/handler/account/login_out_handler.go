// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"net/http"

	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/logic/account"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/response"
)

// 登出（从 Authorization 读取 token）
func LoginOutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := account.NewLoginOutLogic(r.Context(), svcCtx)
		resp, err := l.LoginOut()
		response.Response(w, r, resp, err)
	}
}
