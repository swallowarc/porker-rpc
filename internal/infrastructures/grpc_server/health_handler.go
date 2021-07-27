package grpc_server

import (
	"context"

	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type healthHandler struct {
}

func newHealthHandler() health.HealthServer {
	return &healthHandler{}
}

func (h *healthHandler) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthHandler) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented.")
}
