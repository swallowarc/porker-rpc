package env

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/swallowarc/porker-rpc/internal/infrastructures/redis"
)

var (
	Server Config
	Redis  redis.Config
)

type (
	Config struct {
		IsDevelopment bool   `envconfig:"is_development" default:"true"`
		PORT          string `envconfig:"grpc_port" default:"50051"`
	}
)

func init() {
	setup()
}

func setup() {
	check(envconfig.Process("", &Server))
	check(envconfig.Process("redis", &Redis))
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
