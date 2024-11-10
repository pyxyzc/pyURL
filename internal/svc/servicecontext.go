package svc

import (
	"pyShortURL/internal/config"
	"pyShortURL/model"
	"pyShortURL/sequence"

	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	ShortURLModel model.ShortUrlMapModel

	Sequence sequence.Sequence

	ShortUrlBlackList map[string]struct{}

	Filter *bloom.Filter
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.ShortURLDB.DSN)
	// 配置文件中的黑名单加载到ShortUrlBlackList中
	m := make(map[string]struct{}, len(c.ShortUrlBlackList))
	for _, v := range c.ShortUrlBlackList {
		m[v] = struct{}{}
	}
	// 初始化布隆过滤器
	store := redis.New(c.BloomRedis[0].Host, func(r *redis.Redis) {
		r.Type = redis.NodeType
	})
	// 声明一个 bitSet
	filter := bloom.New(store, "bloom_filter", 20*(1<<20))
	// 加载已经有的短链接数据
	return &ServiceContext{
		Config:            c,
		ShortURLModel:     model.NewShortUrlMapModel(conn, c.CacheRedis),
		Sequence:          sequence.NewMySQL(c.Sequence.DSN),
		ShortUrlBlackList: m,
		Filter:            filter,
	}
}
