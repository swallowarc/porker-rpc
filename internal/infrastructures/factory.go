package infrastructures

import (
	"github.com/swallowarc/porker-rpc/internal/infrastructures/env"
	"github.com/swallowarc/porker-rpc/internal/infrastructures/redis"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
)

type (
	factory struct {
		redisClient gateways.MemDBClient
	}
)

func NewFactory() gateways.Factory {
	return &factory{
		redisClient: redis.NewRedisClient(env.Redis),
	}
}

func (f factory) MemDBClient() gateways.MemDBClient {
	return f.redisClient
}
