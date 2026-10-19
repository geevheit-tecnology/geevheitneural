package app

import (
	"context"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type ClientLGPDUseCases struct {
	uow UnitOfWork
}

func NewClientLGPDUseCases(uow UnitOfWork) *ClientLGPDUseCases {
	return &ClientLGPDUseCases{uow: uow}
}

func (uc *ClientLGPDUseCases) Export(ctx context.Context, principal system.OwnerPrincipal, clientID string) (*ClientExportDTO, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client_id é obrigatório")
	}

	var export *ClientExportDTO
	now := time.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		data, err := tx.Clients.Export(ctx, clientID)
		if err != nil {
			return err
		}
		export = data

		// Registra no audit_log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    principal.ID,
			Action:     "lgpd.client_exported",
			ActorType:  "owner",
			TargetType: "client",
			TargetID:   &clientID,
			Meta: map[string]any{
				"client_name": data.Client.Name,
				"cards_count": len(data.Cards),
			},
			At: now,
		})
	})

	if err != nil {
		return nil, err
	}
	return export, nil
}

func (uc *ClientLGPDUseCases) Purge(ctx context.Context, principal system.OwnerPrincipal, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client_id é obrigatório")
	}

	now := time.Now()

	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		client, err := tx.Clients.GetByID(ctx, clientID)
		if err != nil {
			return fmt.Errorf("cliente não encontrado para expurgo: %w", err)
		}

		if err := tx.Clients.Delete(ctx, clientID); err != nil {
			return fmt.Errorf("falha ao expurgar cliente: %w", err)
		}

		// Registra auditoria com identificador imutável
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    principal.ID,
			Action:     "lgpd.client_purged",
			ActorType:  "owner",
			TargetType: "client",
			TargetID:   &clientID,
			Meta: map[string]any{
				"purged_client_id": clientID,
				"client_name_hash": fmt.Sprintf("%x", client.Name),
			},
			At: now,
		})
	})
}
