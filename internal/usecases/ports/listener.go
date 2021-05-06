package ports

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
)

type (
	PokerListener interface {
		Listen(ctx context.Context) (*porker.PokerSituation, error)
	}
)
