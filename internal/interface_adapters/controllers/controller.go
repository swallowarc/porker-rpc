package controllers

import (
	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/usecases/interactors"
	"go.uber.org/zap"
)

type (
	porkerController struct {
		logger *zap.Logger
		porker.UnimplementedPorkerServiceServer

		loginInteractor interactors.LoginInteractor
		pokerInteractor interactors.PokerInteractor
	}
)

func NewPorkerController(logger *zap.Logger, iFactory interactors.Factory) porker.PorkerServiceServer {
	return &porkerController{
		logger:          logger,
		loginInteractor: iFactory.LoginInteractor(),
		pokerInteractor: iFactory.PokerInteractor(),
	}
}
