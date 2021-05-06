package interactors

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
	"github.com/swallowarc/porker-rpc/internal/usecases/listener"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
	"golang.org/x/xerrors"
)

type (
	pokerInteractor struct {
		pokerRepo ports.PokerRepository
	}
)

func NewPokerInteractor(rFactory ports.RepositoriesFactory) PokerInteractor {
	return &pokerInteractor{
		pokerRepo: rFactory.PokerRepository(),
	}
}

func (bi *pokerInteractor) Create(ctx context.Context, loginID string) (room.ID, error) {
	roomID, err := bi.pokerRepo.Create(ctx, loginID)
	if err != nil {
		return "", xerrors.Errorf("failed to Create: %w", err)
	}

	return roomID, nil
}

func (bi *pokerInteractor) CanEnter(ctx context.Context, roomID room.ID) (bool, error) {
	_, _, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		if errs.IsNotFoundError(err) {
			return false, nil
		}
		return false, xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	return true, nil
}

func (bi *pokerInteractor) Enter(ctx context.Context, roomID room.ID, loginID string) (ports.PokerListener, error) {
	if err := bi.pokerRepo.Enter(ctx, roomID, loginID); err != nil {
		return nil, xerrors.Errorf("failed to Enter: %w", err)
	}

	_, ps, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		return nil, xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	ps.Ballots = append(ps.Ballots, &porker.Ballot{
		LoginId: loginID,
		Point:   porker.Point_POINT_UNKNOWN,
	})

	if err := bi.pokerRepo.Update(ctx, ps); err != nil {
		return nil, xerrors.Errorf("failed to Update: %w", err)
	}

	return listener.NewPokerListener(roomID, loginID, bi.pokerRepo), nil
}

func (bi *pokerInteractor) Leave(ctx context.Context, roomID room.ID, loginID string) error {
	if err := bi.pokerRepo.Leave(ctx, roomID, loginID); err != nil {
		return xerrors.Errorf("failed to Leave: %w", err)
	}

	// 最終退出者だった場合はroomを削除する
	members, err := bi.pokerRepo.ListMembers(ctx, roomID)
	if err != nil {
		return xerrors.Errorf("failed to ListMembers: %w", err)
	}
	if len(members) != 0 {
		return nil
	}
	if err := bi.pokerRepo.Delete(ctx, roomID); err != nil {
		return xerrors.Errorf("failed to Delete: %w", err)
	}
	return nil
}

func (bi *pokerInteractor) Voting(ctx context.Context, roomID room.ID, loginID string, point porker.Point) error {
	_, ps, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		return xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	for i, ballot := range ps.Ballots {
		if ballot.LoginId == loginID {
			ps.Ballots[i].Point = point
			if err := bi.pokerRepo.Update(ctx, ps); err != nil {
				return xerrors.Errorf("failed to bt Update: %w", err)
			}

			return nil
		}
	}

	return xerrors.Errorf("login_id: %s is not found in room_id: %s", loginID, roomID)
}

func (bi *pokerInteractor) VoteCounting(ctx context.Context, roomID room.ID, loginID string) error {
	_, ps, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		return xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	ps.State = porker.RoomState_ROOM_STATE_OPEN
	if err := bi.pokerRepo.Update(ctx, ps); err != nil {
		return xerrors.Errorf("failed to bt Update: %w", err)
	}

	return nil
}

func (bi *pokerInteractor) Reset(ctx context.Context, roomID room.ID) error {
	_, ps, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		return xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	ps.State = porker.RoomState_ROOM_STATE_TURN_DOWN
	for i, ballot := range ps.Ballots {
		if ballot.Point != porker.Point_NOT_VOTE {
			ps.Ballots[i].Point = porker.Point_POINT_UNKNOWN
		}
	}

	if err := bi.pokerRepo.Update(ctx, ps); err != nil {
		return xerrors.Errorf("failed to Update: %w", err)
	}

	return nil
}
