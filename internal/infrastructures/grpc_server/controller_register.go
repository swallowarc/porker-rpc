package grpc_server

import (
	"github.com/swallowarc/porker-proto/pkg/porker"
	"google.golang.org/grpc"
)

type (
	controllerRegister struct {
		porkerController porker.PorkerServiceServer
	}
)

func NewControllerRegister(controller porker.PorkerServiceServer) ControllerRegister {
	return &controllerRegister{
		porkerController: controller,
	}
}

func (cr *controllerRegister) Register(grpcServer grpc.ServiceRegistrar) {
	porker.RegisterPorkerServiceServer(grpcServer, cr.porkerController)
}
