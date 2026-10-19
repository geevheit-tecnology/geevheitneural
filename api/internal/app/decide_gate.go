package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/approval"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type DecideGateUseCase struct {
	uow       UnitOfWork
	publisher EventPublisher
	clock     clock.Clock
}

func NewDecideGateUseCase(uow UnitOfWork, publisher EventPublisher, clk clock.Clock) *DecideGateUseCase {
	return &DecideGateUseCase{
		uow:       uow,
		publisher: publisher,
		clock:     clk,
	}
}

type DecideGateInput struct {
	Owner    system.OwnerPrincipal
	CardID   string
	Decision card.Decision
	Note     string
}

// Execute executa a decisão em um portão do dono de forma atômica e imutável.
func (uc *DecideGateUseCase) Execute(ctx context.Context, input DecideGateInput) error {
	now := uc.clock.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Carrega o cartão
		c, err := tx.Cards.GetByID(ctx, input.CardID)
		if err != nil {
			return fmt.Errorf("erro ao buscar cartão: %w", err)
		}

		// 2. Carrega o estado de pausa do sistema
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao carregar estado do sistema: %w", err)
		}

		gateOrigin := c.Stage
		// 3. Aplica a máquina de estados do domínio
		nextStage, err := c.DecideGate(input.Decision, sysState.Paused, now)
		if err != nil {
			return err
		}

		// 4. Cria o registro imutável de aprovação
		appr, err := approval.New(
			uuid.NewString(),
			input.Owner.ID,
			c.ID,
			gateOrigin,
			input.Decision,
			input.Owner.ID,
			input.Note,
			now,
		)
		if err != nil {
			return fmt.Errorf("erro ao gerar aprovação: %w", err)
		}

		if err := tx.Approvals.Create(ctx, appr); err != nil {
			return fmt.Errorf("erro ao salvar aprovação: %w", err)
		}

		// 5. Atualiza o cartão
		if err := tx.Cards.Update(ctx, c); err != nil {
			return fmt.Errorf("erro ao atualizar cartão: %w", err)
		}

		// 6. Registra o evento imutável do histórico do cartão
		event := CardEventDTO{
			ID:        uuid.NewString(),
			OwnerID:   input.Owner.ID,
			CardID:    c.ID,
			Kind      : fmt.Sprintf("gate_%s", input.Decision),
			Payload: map[string]any{
				"gate":       gateOrigin,
				"decision":   input.Decision,
				"next_stage": nextStage,
				"note":       input.Note,
			},
			ActorType: "owner",
			ActorID:   &input.Owner.ID,
			At:        now,
		}
		if err := tx.CardEvents.Create(ctx, event); err != nil {
			return fmt.Errorf("erro ao registrar evento do cartão: %w", err)
		}

		// 7. Registra auditoria
		audit := AuditLogDTO{
			ID:         uuid.NewString(),
			OwnerID:    input.Owner.ID,
			Action:     "card.gate_decided",
			ActorType:  "owner",
			ActorID:    &input.Owner.ID,
			TargetType: "card",
			TargetID:   &c.ID,
			Meta: map[string]any{
				"gate":     gateOrigin,
				"decision": input.Decision,
			},
			At: now,
		}
		if err := tx.Audit.Record(ctx, audit); err != nil {
			return fmt.Errorf("erro ao gravar auditoria: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 8. Publica notificação para atualização de valor parado em tempo real
	_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, map[string]any{
		"card_id": input.CardID,
	})

	return nil
}
