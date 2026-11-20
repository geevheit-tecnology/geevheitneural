package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type GetSystemStateUseCase struct{ uow UnitOfWork }

func NewGetSystemStateUseCase(uow UnitOfWork) *GetSystemStateUseCase {
	return &GetSystemStateUseCase{uow: uow}
}

func (uc *GetSystemStateUseCase) Execute(ctx context.Context, owner system.OwnerPrincipal) (bool, error) {
	if owner.ID == "" || owner.AAL != "aal2" {
		return false, system.ErrOwnerRequired
	}
	var paused bool
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.SystemState == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		state, err := tx.SystemState.Get(ctx)
		if err != nil {
			return err
		}
		if state == nil {
			return fmt.Errorf("estado do sistema indisponível")
		}
		paused = state.Paused
		return nil
	})
	return paused, err
}
