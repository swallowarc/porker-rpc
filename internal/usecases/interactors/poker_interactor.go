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

	var isExists bool
	for _, ballot := range ps.Ballots {
		if ballot.LoginId == loginID {
			isExists = true
			break
		}
	}

	if !isExists {
		ps.Ballots = append(ps.Ballots, &porker.Ballot{
			LoginId: loginID,
			Point:   porker.Point_POINT_UNKNOWN,
		})
	}

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

	if ps.State != porker.RoomState_ROOM_STATE_TURN_DOWN {
		return xerrors.Errorf(
			"cannot vote in any state other than TURN_DOWN. room_id: %s, state: %s", roomID, ps.State)
	}

	var isExists bool
	var numOfVoted int
	for _, ballot := range ps.Ballots {
		if ballot.LoginId == loginID {
			ballot.Point = point
			isExists = true
		}
		if ballot.Point != porker.Point_POINT_UNKNOWN {
			numOfVoted++
		}
	}

	if !isExists {
		return xerrors.Errorf("login_id: %s is not found in room. room_id: %s", loginID, roomID)
	}

	if numOfVoted == len(ps.Ballots) {
		ps.State = porker.RoomState_ROOM_STATE_OPEN
	}

	if err := bi.pokerRepo.Update(ctx, ps); err != nil {
		return xerrors.Errorf("failed to bt Update: %w", err)
	}

	return nil
}

func (bi *pokerInteractor) VoteCounting(ctx context.Context, roomID room.ID, loginID string) error {
	_, ps, err := bi.pokerRepo.ReadStreamLatest(ctx, roomID)
	if err != nil {
		return xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	if ps.State != porker.RoomState_ROOM_STATE_TURN_DOWN {
		return xerrors.Errorf(
			"cannot vote counting in any state other than TURN_DOWN. room_id: %s, state: %s", roomID, ps.State)
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
