//go:generate mockgen -source=$GOFILE -destination=../../tests/mocks/$GOPACKAGE/mock_$GOFILE -package=mock_$GOPACKAGE
package interactors

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
)

type (
	LoginInteractor interface {
		Login(ctx context.Context, login *porker.Login) (*porker.Login, error)
		Logout(ctx context.Context, login *porker.Login) error
	}

	PokerInteractor interface {
		Create(ctx context.Context, loginID string) (room.ID, error)
		CanEnter(ctx context.Context, roomID room.ID) (bool, error)
		Enter(ctx context.Context, roomID room.ID, loginID string) (ports.PokerListener, error)
		Leave(ctx context.Context, roomID room.ID, loginID string) error
		Voting(ctx context.Context, roomID room.ID, loginID string, point porker.Point) error
		VoteCounting(ctx context.Context, roomID room.ID, loginID string) error
		Reset(ctx context.Context, roomID room.ID) error
	}
)
