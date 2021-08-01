package grpc_server

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	fakeRegister struct{}
)

func (f fakeRegister) Register(grpc.ServiceRegistrar) {}

func TestGrpcServer_RunGRPCServer(t *testing.T) {
	ctx := context.Background()

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to new zap logger: %v", err)
	}
	srv := NewGRPCServer(zapLogger, "18080", true, ControllerRegisters{fakeRegister{}}, func() {}, func() {})

	ctx2, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.RunGRPCServer(ctx2)
	}()
	wg.Wait()
}
