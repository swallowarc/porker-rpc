package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/swallowarc/porker-proto/pkg/porker"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
	"github.com/swallowarc/porker-rpc/internal/usecases/ports"
	"golang.org/x/xerrors"
)

const (
	loginKeyPrefix = "porker_login"
	loginTimeout   = time.Hour
)

type (
	loginRepository struct {
		memDBCli gateways.MemDBClient
	}
)

func NewLoginRepository(gwFactory gateways.Factory) ports.LoginRepository {
	return &loginRepository{
		memDBCli: gwFactory.MemDBClient(),
	}
}

func (r *loginRepository) FindByID(ctx context.Context, loginID string) (*porker.Login, error) {
	sessionID, err := r.memDBCli.Get(ctx, loginKey(loginID))
	if err != nil {
		return nil, xerrors.Errorf("failed to memdb get: %w", err)
	}
	return &porker.Login{
		LoginId:   loginID,
		SessionId: sessionID,
	}, nil
}

func (r *loginRepository) NewLogin(ctx context.Context, loginID string) (*porker.Login, error) {
	sessionID := uuid.New()
	if err := r.memDBCli.SetNX(ctx, loginKey(loginID), sessionID, loginTimeout); err != nil {
		return nil, xerrors.Errorf("failed to SetNX: %w", err)
	}
	return &porker.Login{
		LoginId:   loginID,
		SessionId: sessionID.String(),
	}, nil
}

func (r *loginRepository) ReLogin(ctx context.Context, login *porker.Login) error {
	if err := r.memDBCli.SetNX(ctx, loginKey(login.LoginId), login.SessionId, loginTimeout); err != nil {
		return xerrors.Errorf("failed to SetNX: %w", err)
	}
	return nil
}

func (r *loginRepository) Logout(ctx context.Context, loginID string) error {
	err := r.memDBCli.Del(ctx, loginKey(loginID))
	if errs.IsNotFoundError(err) {
		return nil
	}
	if err != nil {
		return xerrors.Errorf("failed to Del: %w", err)
	}
	return nil
}

func loginKey(loginID string) string {
	return fmt.Sprintf("%s:%s", loginKeyPrefix, loginID)
}
