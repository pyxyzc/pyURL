package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pyShortURL/internal/svc"
	"pyShortURL/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShowLogic {
	return &ShowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 自己写缓存          surl -> lurl
// go-zero自带的缓存   surl -> 数据行

func (l *ShowLogic) Show(req *types.ShowRequest) (resp *types.ShowResponse, err error) {
	// 查看短链接 -> 重定向到真实的链接
	// 1. 根据短链接查询原始的长链接
	// 不存在短链接直接返回404，不需要后续处理
	// a. 基于内存版本，服务重启之后就没有了，每次重启都要重新加载短链接
	// b. 基于Redis版本
	exist, err := l.svcCtx.Filter.Exists([]byte(req.ShortUrl))
	if err != nil {
		logx.Errorw("Bloom Filter failed", logx.LogField{Value: err.Error(), Key: "err"})
	}
	// 不存在的短链接直接返回
	if !exist {
		return nil, errors.New("404")
	}
	fmt.Println("开始查询缓存和DB...")
	// 查询数据库之前可增加缓存层
	u, err := l.svcCtx.ShortURLModel.FindOneBySurl(l.ctx, sql.NullString{Valid: true, String: req.ShortUrl})
	if err != nil {
		if err == sql.ErrNoRows {
			logx.Errorw("短链接未找到", logx.LogField{Value: req.ShortUrl, Key: "shortUrl"})
			return nil, errors.New("404")
		}
		logx.Errorw("ShortUrlModel.FindOneBySurl failed", logx.LogField{Value: err.Error(), Key: "err"})
		return nil, err
	}
	// 2. 返回查询到的长链接，在调用 handler 层返回重定向响应
	return &types.ShowResponse{LongUrl: u.Lurl.String}, nil
}
