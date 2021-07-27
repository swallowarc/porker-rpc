package grpc_server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type (
	GRPCServer interface {
		RunGRPCServer(ctx context.Context)
	}

	ControllerRegister interface {
		Register(grpcServer grpc.ServiceRegistrar)
	}

	InitFunc   func()
	CloserFunc func()

	grpcServer struct {
		logger             *zap.Logger
		port               string
		isDevelop          bool
		controllerRegister ControllerRegister
		initFunction       InitFunc
		closerFunction     CloserFunc
	}
)

var (
	catchSignals = []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	}
)

func NewGRPCServer(
	logger *zap.Logger,
	port string,
	isDevelop bool,
	controllerRegister ControllerRegister,
	initFunction InitFunc,
	closerFunction CloserFunc,
) GRPCServer {
	return &grpcServer{
		logger:             logger,
		port:               port,
		isDevelop:          isDevelop,
		controllerRegister: controllerRegister,
		initFunction:       initFunction,
		closerFunction:     closerFunction,
	}
}

func (s *grpcServer) RunGRPCServer(ctx context.Context) {
	s.logger.Info("Initialize gRPC Server ...")
	s.initFunction()

	s.logger.Info("Startup gRPC Server ...")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := s.newServer()

	s.logger.Info(fmt.Sprintf("Startup using port : %s", s.port))
	go func() {
		if err := server.Serve(lis); err != nil {
			s.logger.Panic("failed to gRPC Server running", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, catchSignals...)

	select {
	case v := <-ctx.Done():
		s.logger.Info("!! Receive context cancel !!", zap.Reflect("message", v))
	case sig := <-quit:
		s.logger.Info("!! Receive signal !!", zap.String("signal", sig.String()))
	}

	wait := time.Duration(5)
	if s.isDevelop {
		wait = 1
	}
	sdCtx, sdCancel := context.WithTimeout(ctx, wait*time.Second)
	defer sdCancel()

	<-sdCtx.Done()

	s.logger.Info("Closing gRPC Server ...")
	server.GracefulStop()
	s.closerFunction()

	s.logger.Info("Shutdown gRPC Server")
}

func (s *grpcServer) newServer() *grpc.Server {
	var (
		zapOpts = []grpc_zap.Option{
			grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
				return zap.Int64("grpc.time_ns", duration.Nanoseconds())
			}),
		}

		kasp = keepalive.ServerParameters{
			MaxConnectionIdle: 600 * time.Second,
		}
	)

	grpc_zap.ReplaceGrpcLoggerV2(s.logger)
	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(s.logger, zapOpts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.StreamServerInterceptor(s.logger, zapOpts...),
		),
		grpc.KeepaliveParams(kasp),
	)

	s.controllerRegister.Register(server)

	health.RegisterHealthServer(server, newHealthHandler())

	if s.isDevelop {
		reflection.Register(server)
	}
	return server
}
