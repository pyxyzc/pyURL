package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pyShortURL/internal/svc"
	"pyShortURL/internal/types"
	"pyShortURL/model"
	"pyShortURL/pkg/base62"
	"pyShortURL/pkg/connect"
	"pyShortURL/pkg/md5"
	"pyShortURL/pkg/urltool"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ConvertLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConvertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConvertLogic {
	return &ConvertLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Convert 转链：长链接 -> 短链接
func (l *ConvertLogic) Convert(req *types.ConvertRequest) (resp *types.ConvertResponse, err error) {
	// 1. 校验长链接
	// 1.1 数据不能为空
	// 使用validator包校验参数

	// 1.2 能请求通过的链接
	ok := connect.Get(req.LongUrl)
	if !ok {
		return nil, errors.New("链接无法请求通")
	}

	// 1.3 数据库中是否已经存在
	// 不能直接查库，并没有给直接的长链接建立索引
	// 生成md5
	md5Value := md5.Sum([]byte(req.LongUrl))
	u, err := l.svcCtx.ShortURLModel.FindOneByMd5(l.ctx, sql.NullString{String: md5Value, Valid: true})
	// 不用管查到的是什么值，只看查没查到即可
	if err != sqlx.ErrNotFound {
		if err == nil {
			return nil, fmt.Errorf("该链接已被转为短链接%s", u.Surl.String)
		}
		logx.Errorw("ShortUrlModel.FindOneByMd5 failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	// 1.4 避免循环转链接（即输入的即是短链接）
	// 对输入的进行解析，并且查一查
	basePath, err := urltool.GetBasePath(req.LongUrl)
	if err != nil {
		logx.Errorw("urltool.GetBasePath failed", logx.LogField{Key: "lurl", Value: req.LongUrl}, logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	u, err = l.svcCtx.ShortURLModel.FindOneBySurl(l.ctx, sql.NullString{String: basePath, Valid: true})
	if err != sqlx.ErrNotFound {
		if err == nil {
			return nil, fmt.Errorf("该链接已是短链接%s", u.Surl.String)
		}
		logx.Errorw("ShortUrlModel.FindOneBySurl failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	// 2. 取号（发号器）
	// 每来一个转链请求，就使用REPLACE INTO语句往seq表插入一条数据，并使用取出主键id作为号码
	var short string
	for {
		// 2. 取号 基于MySQL实现的发号器
		// 每来一个转链请求，我们就使用 REPLACE INTO语句往 sequence 表插入一条数据，并且取出主键id作为号码
		seq, err := l.svcCtx.Sequence.Next()
		if err != nil {
			logx.Errorw("Sequence.Next() failed", logx.LogField{Key: "err", Value: err.Error()})
			return nil, err
		}
		// 3. 号码转短链
		// 3.1 安全性  1En = 6347
		short = base62.Int2String(seq)
		// 3.2 短域名黑名单避免某些特殊词比如 api、health、fuck等等
		if _, ok := l.svcCtx.ShortUrlBlackList[short]; !ok {
			break // 生成不在黑名单里的短链接就跳出for循环
		}
	}
	// 4. 存储长短链接映射关系
	if _, err := l.svcCtx.ShortURLModel.Insert(
		l.ctx,
		&model.ShortUrlMap{
			Lurl: sql.NullString{String: req.LongUrl, Valid: true},
			Md5:  sql.NullString{String: md5Value, Valid: true},
			Surl: sql.NullString{String: short, Valid: true},
		},
	); err != nil {
		logx.Errorw("ShortUrlModel.Insert() failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	// 将生成的短链接加到布隆过滤器中
	if err := l.svcCtx.Filter.Add([]byte(short)); err != nil {
		logx.Errorw("BloomFilter.Add() failed", logx.LogField{Key: "err", Value: err.Error()})
	}
	// 5. 返回响应
	// 返回短域名+短链接
	shortUrl := l.svcCtx.Config.ShortDomain + "/" + short
	return &types.ConvertResponse{ShortUrl: shortUrl}, nil
}
