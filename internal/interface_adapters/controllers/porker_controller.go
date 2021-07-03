package controllers

import (
	"context"
	"time"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/commons/loggers"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
	"github.com/swallowarc/porker-rpc/internal/usecases/listener"
	"golang.org/x/xerrors"
)

func (c *porkerController) Login(ctx context.Context, request *porker.LoginRequest) (*porker.LoginResponse, error) {
	login := request.Login
	loggers.With(ctx, loggers.Map{
		"login_id":   login.LoginId,
		"session_id": login.SessionId,
	})

	newLogin, err := c.loginInteractor.Login(ctx, login)
	if err != nil {
		return nil, xerrors.Errorf("failed to login: %w", err)
	}

	return &porker.LoginResponse{
		Login: newLogin,
	}, nil
}

func (c *porkerController) Logout(ctx context.Context, request *porker.LogoutRequest) (*porker.NoBody, error) {
	login := request.Login
	loggers.With(ctx, loggers.Map{
		"login_id":   login.LoginId,
		"session_id": login.SessionId,
	})
	if err := c.loginInteractor.Logout(ctx, login); err != nil {
		return nil, xerrors.Errorf("failed to logout: %w", err)
	}
	return &porker.NoBody{}, nil
}

func (c *porkerController) CreateRoom(ctx context.Context, req *porker.CreateRoomRequest) (*porker.CreateRoomResponse, error) {
	roomID, err := c.pokerInteractor.Create(ctx, req.LoginId)
	if err != nil {
		return nil, xerrors.Errorf("failed to Create: %w", err)
	}

	return &porker.CreateRoomResponse{
		RoomId: roomID.String(),
	}, nil
}

func (c *porkerController) CanEnterRoom(ctx context.Context, req *porker.CanEnterRoomRequest) (*porker.CanEnterRoomResponse, error) {
	can, err := c.pokerInteractor.CanEnter(ctx, room.ID(req.RoomId))
	if err != nil {
		return nil, xerrors.Errorf("failed to CanEnter: %w", err)
	}

	return &porker.CanEnterRoomResponse{CanEnterRoom: can}, nil
}

func (c *porkerController) EnterRoom(request *porker.EnterRoomRequest, stream porker.PorkerService_EnterRoomServer) error {
	ctx := loggers.LoggerToContext(stream.Context(), c.logger)
	lsnr, err := c.pokerInteractor.Enter(ctx, room.ID(request.RoomId), request.LoginId)
	if err != nil {
		return xerrors.Errorf("failed to Enter: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			bt, err := lsnr.Listen(ctx)
			if err != nil {
				if errs.IsNotFoundError(err) {
					time.Sleep(time.Second)
					continue
				}
				if xerrors.As(err, &listener.LeftError) {
					loggers.Logger(ctx).Info("already left the room")
					return nil
				}
				if xerrors.As(err, &context.Canceled) {
					loggers.Logger(ctx).Info("context canceled")
					return nil
				}

				return xerrors.Errorf("failed to Listen: %w", err)
			}
			if err := stream.Send(bt); err != nil {
				if xerrors.As(err, &context.Canceled) {
					loggers.Logger(ctx).Debug("client context canceled")
					return nil
				}

				return xerrors.Errorf("failed to Send: %w", err)
			}
		}
	}
}

func (c *porkerController) LeaveRoom(ctx context.Context, req *porker.LeaveRoomRequest) (*porker.NoBody, error) {
	if err := c.pokerInteractor.Leave(ctx, room.ID(req.RoomId), req.LoginId); err != nil {
		return nil, xerrors.Errorf("failed to Leave: %w", err)
	}

	return &porker.NoBody{}, nil
}

func (c *porkerController) Voting(ctx context.Context, req *porker.VotingRequest) (*porker.NoBody, error) {
	if err := c.pokerInteractor.Voting(ctx, room.ID(req.RoomId), req.Ballot.LoginId, req.Ballot.Point); err != nil {
		return nil, xerrors.Errorf("failed to Voting: %w", err)
	}

	return &porker.NoBody{}, nil
}

func (c *porkerController) VoteCounting(ctx context.Context, req *porker.VoteCountingRequest) (*porker.NoBody, error) {
	if err := c.pokerInteractor.VoteCounting(ctx, room.ID(req.RoomId), req.LoginId); err != nil {
		return nil, xerrors.Errorf("failed to Pick: %w", err)
	}

	return &porker.NoBody{}, nil
}

func (c *porkerController) ResetRoom(ctx context.Context, req *porker.ResetRoomRequest) (*porker.NoBody, error) {
	if err := c.pokerInteractor.Reset(ctx, room.ID(req.RoomId)); err != nil {
		return nil, xerrors.Errorf("failed to Reset: %w", err)
	}

	return &porker.NoBody{}, nil
}
