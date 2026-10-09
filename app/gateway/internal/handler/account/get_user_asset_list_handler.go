// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package account

import (
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/logic/account"
	"github.com/nmg988rbdv-svg/go-zero_columbina/app/gateway/internal/svc"
	"github.com/nmg988rbdv-svg/go-zero_columbina/common/pkg/response"
	"net/http"
)

// 获取用户所有资产
func GetUserAssetListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := account.NewGetUserAssetListLogic(r.Context(), svcCtx)
		resp, err := l.GetUserAssetList()
		response.Response(w, r, resp, err)

	}
}
