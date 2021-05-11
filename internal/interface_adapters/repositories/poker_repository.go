package repositories

import (
	"context"
	"encoding/json"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const (
	situationMessageKey = "poker_situation_message_key"
)

type (
	PokerRepository struct {
		memDBCli gateways.MemDBClient
	}
)

func NewPokerRepository(gwFactory gateways.Factory) ports.PokerRepository {
	return &PokerRepository{
		memDBCli: gwFactory.MemDBClient(),
	}
}

func (r *PokerRepository) Create(ctx context.Context, loginID string) (room.ID, error) {
	var roomID room.ID
	for {
		roomID = room.NewID()
		_, err := r.memDBCli.Get(ctx, roomID.IDKey())
		if err == nil {
			continue
		}
		if errs.IsNotFoundError(err) {
			break
		}
		return "", xerrors.Errorf("failed to memDBCli.Get: %w", err)
	}

	if err := r.memDBCli.SetNX(ctx, roomID.IDKey(), "", room.TimeoutDuration); err != nil {
		return "", xerrors.Errorf("failed to SetNX: %w", err)
	}

	situation := &porker.PokerSituation{
		RoomId:        roomID.String(),
		MasterLoginId: loginID,
		State:         porker.RoomState_ROOM_STATE_TURN_DOWN,
		Ballots:       []*porker.Ballot{},
	}
	if err := r.Update(ctx, situation); err != nil { // UpdateでもStreamがなければ新規作成される
		return "", err
	}

	return roomID, nil
}

func (r *PokerRepository) refreshRoomDuration(ctx context.Context, roomID room.ID) error {
	if _, err := r.memDBCli.Get(ctx, roomID.IDKey()); err != nil {
		return xerrors.Errorf("failed to Get room_id from memdb: %w", err)
	}

	eg := errgroup.Group{}

	eg.Go(func() error {
		if err := r.memDBCli.Expire(ctx, roomID.IDKey(), room.TimeoutDuration); err != nil {
			return xerrors.Errorf("failed to Expire room: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := r.memDBCli.Expire(ctx, roomID.MemberKey(), room.TimeoutDuration); err != nil {
			return xerrors.Errorf("failed to Expire member: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := r.memDBCli.Expire(ctx, roomID.StreamKey(), room.TimeoutDuration); err != nil {
			return xerrors.Errorf("failed to Expire stream: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (r *PokerRepository) Update(ctx context.Context, ps *porker.PokerSituation) error {
	roomID := room.ID(ps.RoomId)
	if err := r.refreshRoomDuration(ctx, roomID); err != nil {
		return xerrors.Errorf("failed to refreshRoomDuration: %w", err)
	}

	jm, err := json.Marshal(ps)
	if err != nil {
		return xerrors.Errorf("failed to json.Marshal: %w", err)
	}

	if err := r.memDBCli.PublishStream(ctx, roomID.StreamKey(), map[string]interface{}{
		situationMessageKey: jm,
	}); err != nil {
		return xerrors.Errorf("failed to PublishStream: %w", err)
	}

	return nil
}

func (r *PokerRepository) Enter(ctx context.Context, roomID room.ID, loginID string) error {
	if err := r.refreshRoomDuration(ctx, roomID); err != nil {
		return xerrors.Errorf("failed to refreshRoomDuration: %w", err)
	}

	if err := r.memDBCli.SAdd(ctx, roomID.MemberKey(), loginID); err != nil {
		return xerrors.Errorf("failed to SAdd room member from memdb: %w", err)
	}

	return nil
}

func (r *PokerRepository) Leave(ctx context.Context, roomID room.ID, loginID string) error {
	if err := r.refreshRoomDuration(ctx, roomID); err != nil {
		return xerrors.Errorf("failed to refreshRoomDuration: %w", err)
	}

	if err := r.memDBCli.SRem(ctx, roomID.MemberKey(), loginID); err != nil {
		return xerrors.Errorf("failed to SRem from memdb: %w", err)
	}

	return nil
}

func (r *PokerRepository) ReadStreamLatest(ctx context.Context, roomID room.ID) (string, *porker.PokerSituation, error) {
	msgID, msg, err := r.memDBCli.ReadStreamLatest(ctx, roomID.StreamKey(), situationMessageKey)
	if err != nil {
		return "", nil, xerrors.Errorf("failed to ReadStreamLatest: %w", err)
	}

	result, err := unmarshal(msg)
	return msgID, result, err
}

func (r *PokerRepository) ReadStream(ctx context.Context, roomID room.ID, messageID string) (string, *porker.PokerSituation, error) {
	msgID, msg, err := r.memDBCli.ReadStream(ctx, roomID.StreamKey(), situationMessageKey, messageID)
	if err != nil {
		return "", nil, xerrors.Errorf("failed to ReadStream: %w", err)
	}

	result, err := unmarshal(msg)
	return msgID, result, err
}

func unmarshal(message string) (*porker.PokerSituation, error) {
	var result porker.PokerSituation
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		return nil, xerrors.Errorf("failed to json unmarshal. err: %w, msg: %s", err, message)
	}
	return &result, nil
}

func (r *PokerRepository) ListMembers(ctx context.Context, roomID room.ID) ([]string, error) {
	members, err := r.memDBCli.SMembers(ctx, roomID.MemberKey())
	if err != nil {
		return nil, xerrors.Errorf("failed to SMembers from memdb: %w", err)
	}

	return members, nil
}

func (r *PokerRepository) IsExistsInRoom(ctx context.Context, roomID room.ID, loginID string) (bool, error) {
	memberIDs, err := r.ListMembers(ctx, roomID)
	if err != nil {
		return false, xerrors.Errorf("failed to ListMembers: %w", err)
	}

	for _, id := range memberIDs {
		if loginID == id {
			return true, nil
		}
	}

	return false, nil
}

func (r *PokerRepository) Delete(ctx context.Context, roomID room.ID) error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return r.memDBCli.Del(ctx, roomID.IDKey())
	})
	eg.Go(func() error {
		return r.memDBCli.Del(ctx, roomID.MemberKey())
	})
	eg.Go(func() error {
		return r.memDBCli.Del(ctx, roomID.StreamKey())
	})
	if err := eg.Wait(); err != nil {
		return xerrors.Errorf("failed to Del from memdb: %w", err)
	}

	return nil
}
