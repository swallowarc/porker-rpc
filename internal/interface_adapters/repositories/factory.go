package repositories

import (
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
)

type (
	factory struct {
		loginRepository ports.LoginRepository
		pokerRepository ports.PokerRepository
	}
)

func NewFactory(gwFactory gateways.Factory) ports.RepositoriesFactory {
	return &factory{
		loginRepository: NewLoginRepository(gwFactory),
		pokerRepository: NewPokerRepository(gwFactory),
	}
}

func (f *factory) LoginRepository() ports.LoginRepository {
	return f.loginRepository
}

func (f *factory) PokerRepository() ports.PokerRepository {
	return f.pokerRepository
}
