package grpc_server

import (
	"context"
	"fmt"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type (
	healthServer struct {
		memDBCli gateways.MemDBClient
	}

	healthRegister struct {
		healthServer health.HealthServer
	}
)

func NewHealthRegister(memDBCli gateways.MemDBClient) ControllerRegister {
	return &healthRegister{
		healthServer: newHealthServer(memDBCli),
	}
}

func (hr *healthRegister) Register(grpcServer grpc.ServiceRegistrar) {
	health.RegisterHealthServer(grpcServer, hr.healthServer)
}

func newHealthServer(memDBCli gateways.MemDBClient) health.HealthServer {
	return &healthServer{
		memDBCli: memDBCli,
	}
}

func (h *healthServer) Check(ctx context.Context, _ *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	if err := h.memDBCli.Ping(ctx); err != nil {
		panic(fmt.Errorf("failed to ping redis in health check: %v", err))
	}

	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthServer) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented.")
}
