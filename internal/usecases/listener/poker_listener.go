package listener

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
	"golang.org/x/xerrors"
)

type (
	pokerListener struct {
		roomID        room.ID
		loginID       string
		pokerRepo     ports.PokerRepository
		lastMessageID string
	}
)

var LeftError = xerrors.New("already left the room")

func NewPokerListener(roomID room.ID, loginID string, pokerRepo ports.PokerRepository) ports.PokerListener {
	return &pokerListener{
		roomID:    roomID,
		loginID:   loginID,
		pokerRepo: pokerRepo,
	}
}

func (l *pokerListener) Listen(ctx context.Context) (*porker.PokerSituation, error) {
	isExists, err := l.pokerRepo.IsExistsInRoom(ctx, l.roomID, l.loginID)
	if err != nil {
		return nil, xerrors.Errorf("failed to IsExistsInRoom: %w", err)
	}
	if !isExists {
		return nil, LeftError
	}

	var ps *porker.PokerSituation
	if l.lastMessageID == "" {
		l.lastMessageID, ps, err = l.pokerRepo.ReadStreamLatest(ctx, l.roomID)
		if err != nil {
			return nil, xerrors.Errorf("failed to ReadStreamLatest: %w", err)
		}
	} else {
		l.lastMessageID, ps, err = l.pokerRepo.ReadStream(ctx, l.roomID, l.lastMessageID)
		if err != nil {
			return nil, xerrors.Errorf("failed to ReadStream: %w", err)
		}
	}

	return ps, nil
}
