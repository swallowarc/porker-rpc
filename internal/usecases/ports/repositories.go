//go:generate mockgen -source=$GOFILE -destination=../../tests/mocks/$GOPACKAGE/mock_$GOFILE -package=mock_$GOPACKAGE
package ports

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/domains/room"
)

type (
	LoginRepository interface {
		FindByID(ctx context.Context, loginID string) (*porker.Login, error)
		NewLogin(ctx context.Context, loginID string) (*porker.Login, error)
		ReLogin(ctx context.Context, login *porker.Login) error
		Logout(ctx context.Context, loginID string) error
	}

	PokerRepository interface {
		Create(ctx context.Context, loginID string) (room.ID, error)
		Update(ctx context.Context, ps *porker.PokerSituation) error
		Enter(ctx context.Context, roomID room.ID, loginID string) error
		Leave(ctx context.Context, roomID room.ID, loginID string) error
		ReadStreamLatest(ctx context.Context, roomID room.ID) (string, *porker.PokerSituation, error)
		ReadStream(ctx context.Context, roomID room.ID, messageID string) (string, *porker.PokerSituation, error)
		ListMembers(ctx context.Context, roomID room.ID) ([]string, error)
		IsExistsInRoom(ctx context.Context, roomID room.ID, loginID string) (bool, error)
		Delete(ctx context.Context, roomID room.ID) error
	}
)
