package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type ResumeSystemUseCase struct {
	uow       UnitOfWork
	publisher EventPublisher
	clock     clock.Clock
}

func NewResumeSystemUseCase(uow UnitOfWork, publisher EventPublisher, clk clock.Clock) *ResumeSystemUseCase {
	return &ResumeSystemUseCase{
		uow:       uow,
		publisher: publisher,
		clock:     clk,
	}
}

// Execute retoma a operação do sistema após validação do dono.
func (uc *ResumeSystemUseCase) Execute(ctx context.Context, owner system.OwnerPrincipal) error {
	now := uc.clock.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		st, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		st.Resume(owner, now)

		if err := tx.SystemState.Save(ctx, st); err != nil {
			return fmt.Errorf("erro ao salvar estado do sistema: %w", err)
		}

		audit := AuditLogDTO{
			ID:         uuid.NewString(),
			OwnerID:    owner.ID,
			Action:     "system.resumed",
			ActorType:  "owner",
			ActorID:    &owner.ID,
			TargetType: "system",
			Meta:       map[string]any{},
			At:         now,
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
		"paused": false,
	})

	return nil
}
