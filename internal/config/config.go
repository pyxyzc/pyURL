package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type ShortURLDB struct {
	DSN string
}

type Sequence struct {
	DSN string
}

type Config struct {
	rest.RestConf
	ShortURLDB ShortURLDB
	Sequence   Sequence

	BaseString        string
	ShortUrlBlackList []string

	ShortDomain string

	CacheRedis cache.CacheConf
	BloomRedis cache.CacheConf
}
