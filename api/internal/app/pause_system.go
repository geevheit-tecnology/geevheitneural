package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type PauseSystemUseCase struct {
	uow       UnitOfWork
	publisher EventPublisher
	clock     clock.Clock
}

func NewPauseSystemUseCase(uow UnitOfWork, publisher EventPublisher, clk clock.Clock) *PauseSystemUseCase {
	return &PauseSystemUseCase{
		uow:       uow,
		publisher: publisher,
		clock:     clk,
	}
}

// Execute pausa imediatamente todas as execuções de agentes e conexões (Botão Parar).
func (uc *PauseSystemUseCase) Execute(ctx context.Context, owner system.OwnerPrincipal, reason string) error {
	now := uc.clock.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		st, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		st.Pause(owner, reason, now)

		if err := tx.SystemState.Save(ctx, st); err != nil {
			return fmt.Errorf("erro ao salvar estado do sistema: %w", err)
		}

		audit := AuditLogDTO{
			ID:         uuid.NewString(),
			OwnerID:    owner.ID,
			Action:     "system.paused",
			ActorType:  "owner",
			ActorID:    &owner.ID,
			TargetType: "system",
			Meta: map[string]any{
				"reason": reason,
			},
			At: now,
		}
		if err := tx.Audit.Record(ctx, audit); err != nil {
			return fmt.Errorf("erro ao registrar auditoria: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	_ = uc.publisher.Publish(ctx, TopicSystemStateChanged, map[string]any{
		"paused": true,
		"reason": reason,
	})

	return nil
}
