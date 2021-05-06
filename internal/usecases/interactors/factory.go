package interactors

import (
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
)

type (
	Factory interface {
		LoginInteractor() LoginInteractor
		PokerInteractor() PokerInteractor
	}

	factory struct {
		loginInteractor LoginInteractor
		pokerInteractor PokerInteractor
	}
)

func NewFactory(rFactory ports.RepositoriesFactory) Factory {
	return &factory{
		loginInteractor: NewLoginInteractor(rFactory),
		pokerInteractor: NewPokerInteractor(rFactory),
	}
}

func (f factory) LoginInteractor() LoginInteractor {
	return f.loginInteractor
}

func (f factory) PokerInteractor() PokerInteractor {
	return f.pokerInteractor
}
