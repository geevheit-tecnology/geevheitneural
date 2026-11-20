package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type GetDurableWorkspaceUseCase struct{ uow UnitOfWork }

func NewGetDurableWorkspaceUseCase(uow UnitOfWork) *GetDurableWorkspaceUseCase {
	return &GetDurableWorkspaceUseCase{uow: uow}
}

func (uc *GetDurableWorkspaceUseCase) Execute(ctx context.Context, owner system.OwnerPrincipal, cardID string) ([]ProjectFile, error) {
	if owner.ID == "" || owner.AAL != "aal2" {
		return nil, system.ErrOwnerRequired
	}
	if cardID == "" {
		return nil, fmt.Errorf("cartão obrigatório")
	}
	var files []ProjectFile
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if tx.DurableFactory == nil {
			return fmt.Errorf("armazenamento durável indisponível")
		}
		var err error
		files, err = tx.DurableFactory.ListFiles(ctx, cardID)
		return err
	})
	return files, err
}
