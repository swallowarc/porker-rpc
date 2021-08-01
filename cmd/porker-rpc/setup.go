package main

import (
	"context"

	"github.com/swallowarc/porker-rpc/internal/commons/loggers"
	"github.com/swallowarc/porker-rpc/internal/infrastructures"
	"github.com/swallowarc/porker-rpc/internal/infrastructures/env"
	"github.com/swallowarc/porker-rpc/internal/infrastructures/grpc_server"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/controllers"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/repositories"
	"github.com/swallowarc/porker-rpc/internal/usecases/interactors"
	"go.uber.org/zap"
)

func setup() grpc_server.GRPCServer {
	zapLogger := loggers.NewZapLogger(env.Server.IsDevelopment)

	// factories
	gwFactory := infrastructures.NewFactory()
	repoFactory := repositories.NewFactory(gwFactory)
	iFactory := interactors.NewFactory(repoFactory)

	// interface_adapters
	controller := controllers.NewPorkerController(zapLogger, iFactory)
	// grpc_controller_register
	grpcControllerRegisters := grpc_server.ControllerRegisters{
		grpc_server.NewControllerRegister(controller),
		grpc_server.NewHealthRegister(gwFactory.MemDBClient()),
	}

	// initializer & closer
	init := func() {
		if err := gwFactory.MemDBClient().Ping(context.Background()); err != nil {
			zapLogger.Panic("failed to ping to redis", zap.Error(err))
		}
		zapLogger.Info("ping to redis was successful")
	}
	closer := func() {}

	grpcServer := grpc_server.NewGRPCServer(
		zapLogger,
		env.Server.PORT,
		env.Server.IsDevelopment,
		grpcControllerRegisters,
		init,
		closer,
	)

	return grpcServer
}
