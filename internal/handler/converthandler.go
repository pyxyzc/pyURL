package handler

import (
	"net/http"

	"pyShortURL/internal/logic"
	"pyShortURL/internal/svc"
	"pyShortURL/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ConvertHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析请求参数
		var req types.ConvertRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 参数规则校验
		err := validator.New().StructCtx(r.Context(), &req)
		if err != nil {
			logx.Errorw("validator check failed", logx.LogField{Key: "err", Value: err})
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 执行业务逻辑
		l := logic.NewConvertLogic(r.Context(), svcCtx)
		resp, err := l.Convert(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
