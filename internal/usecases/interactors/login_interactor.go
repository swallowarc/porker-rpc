package interactors

import (
	"context"

	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
	"golang.org/x/xerrors"
)

type (
	loginInteractor struct {
		loginRepo ports.LoginRepository
	}
)

func NewLoginInteractor(rFactory ports.RepositoriesFactory) LoginInteractor {
	return &loginInteractor{
		loginRepo: rFactory.LoginRepository(),
	}
}

func (li *loginInteractor) Login(ctx context.Context, login *porker.Login) (*porker.Login, error) {
	registeredLogin, err := li.loginRepo.FindByID(ctx, login.LoginId)

	switch {
	case errs.IsNotFoundError(err):
		newLogin, err := li.loginRepo.NewLogin(ctx, login.LoginId)
		if err != nil {
			return nil, xerrors.Errorf("failed to NewLogin: %w", err)
		}
		return newLogin, nil

	case err != nil:
		return nil, xerrors.Errorf("failed to FindByID: %w", err)

	case login.SessionId != "" && login.SessionId != registeredLogin.SessionId:
		return nil, errs.NewPreConditionError("session id does not match")
	}

	if err := li.loginRepo.ReLogin(ctx, login); err != nil {
		return nil, xerrors.Errorf("failed to ReLogin: %w", err)
	}

	return registeredLogin, nil
}

func (li *loginInteractor) Logout(ctx context.Context, login *porker.Login) error {
	if err := li.loginRepo.Logout(ctx, login.LoginId); err != nil {
		return xerrors.Errorf("failed to Logout: %w", err)
	}
	return nil
}
