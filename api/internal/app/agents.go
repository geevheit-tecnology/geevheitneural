package app

import (
	"context"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type AgentUseCases struct {
	uow       UnitOfWork
	guard     system.ExecutionGuard
	publisher EventPublisher
}

func NewAgentUseCases(uow UnitOfWork, publisher EventPublisher) *AgentUseCases {
	return &AgentUseCases{
		uow:       uow,
		guard:     system.ExecutionGuard{},
		publisher: publisher,
	}
}

// ListAvailableCards retorna cartões em execução para o agente, sujeito ao ExecutionGuard.
func (uc *AgentUseCases) ListAvailableCards(ctx context.Context, principal system.AgentPrincipal) ([]CardListItemDTO, error) {
	var cards []CardListItemDTO

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Checa estado global do sistema
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		// 2. ExecutionGuard: Se sistema pausado ou agente inativo => 423 Locked
		if !uc.guard.CanExecute(principal, *sysState) {
			return ErrExecutionLocked
		}

		// 3. Agente só pode visualizar cartões que estão ativamente em execução
		allCards, err := tx.Cards.ListEnriched(ctx)
		if err != nil {
			return fmt.Errorf("erro ao listar cartões: %w", err)
		}

		for _, c := range allCards {
			if c.Stage == string(card.StageEmExecucao) {
				cards = append(cards, c)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return cards, nil
}

// GetCardDetails retorna detalhes técnicos de um cartão em execução para o agente.
func (uc *AgentUseCases) GetCardDetails(ctx context.Context, principal system.AgentPrincipal, cardID string) (*CardDetailsDTO, error) {
	var details *CardDetailsDTO

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		if !uc.guard.CanExecute(principal, *sysState) {
			return ErrExecutionLocked
		}

		c, err := tx.Cards.GetEnrichedByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}

		if c.Stage != string(card.StageEmExecucao) {
			return fmt.Errorf("cartão não está em execução disponível para agentes")
		}

		events, err := tx.Cards.ListEvents(ctx, cardID)
		if err != nil {
			return fmt.Errorf("erro ao buscar histórico: %w", err)
		}

		details = &CardDetailsDTO{
			Card:   *c,
			Events: events,
			IsGate: false,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return details, nil
}

// RecordProgress registra log de progresso técnico do agente.
func (uc *AgentUseCases) RecordProgress(ctx context.Context, principal system.AgentPrincipal, cardID string, cmd AgentProgressCommand) error {
	now := time.Now()

	return uc.uow.RunInTx(ctx, func(tx Repositories) error {
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		if !uc.guard.CanExecute(principal, *sysState) {
			return ErrExecutionLocked
		}

		// Valida limite de gasto (SpendLimit)
		if cmd.Cost > 0 && principal.SpendLimitCents > 0 {
			if err := tx.Agents.AddSpend(ctx, principal.ID, cmd.Cost); err != nil {
				return fmt.Errorf("falha ao debitar gasto do agente: %w", err)
			}
		}

		// Registra evento no cartão
		payload := map[string]any{
			"message":    cmd.Message,
			"cost_cents": cmd.Cost,
		}
		if cmd.Meta != nil {
			payload["meta"] = cmd.Meta
		}

		actorID := principal.ID
		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cardID,
			Kind:      "agent_progress",
			ActorType: "agent",
			ActorID:   &actorID,
			Payload:   payload,
			At:        now,
		}); err != nil {
			return fmt.Errorf("erro ao registrar evento de progresso: %w", err)
		}

		// Audit log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    "00000000-0000-0000-0000-000000000001",
			Action:     "agent.progress_recorded",
			ActorType:  "agent",
			ActorID:    &actorID,
			TargetType: "card",
			TargetID:   &cardID,
			Meta: map[string]any{
				"agent_role": principal.Role,
			},
			At: now,
		})
	})
}

// SubmitDelivery submete a entrega técnica para aprovação do dono (Portão 3).
// INVARIANTE: O agente NUNCA aprova nem entrega o cartão. O cartão transita para entrega_aprovar.
func (uc *AgentUseCases) SubmitDelivery(ctx context.Context, principal system.AgentPrincipal, cardID string, cmd AgentSubmitDeliveryCommand) error {
	now := time.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		sysState, err := tx.SystemState.Get(ctx)
		if err != nil {
			return fmt.Errorf("erro ao obter estado do sistema: %w", err)
		}

		if !uc.guard.CanExecute(principal, *sysState) {
			return ErrExecutionLocked
		}

		c, err := tx.Cards.GetByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}

		if c.Stage != card.StageEmExecucao {
			return fmt.Errorf("apenas cartões em execução podem ser submetidos para entrega")
		}

		// Move para o Portão 3: Entrega para Aprovar
		c.Stage = card.StageEntregaAprovar
		c.UpdatedAt = now
		if err := tx.Cards.Update(ctx, c); err != nil {
			return fmt.Errorf("erro ao atualizar estágio para entrega_aprovar: %w", err)
		}

		// Registra evento
		actorID := principal.ID
		payload := map[string]any{
			"summary": cmd.Summary,
			"pr_url":  cmd.PRURL,
		}
		if cmd.Meta != nil {
			payload["meta"] = cmd.Meta
		}

		if err := tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.New().String(),
			CardID:    cardID,
			Kind:      "delivery_submitted_by_agent",
			ActorType: "agent",
			ActorID:   &actorID,
			Payload:   payload,
			At:        now,
		}); err != nil {
			return fmt.Errorf("erro ao registrar evento: %w", err)
		}

		// Audit log
		return tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.New().String(),
			OwnerID:    "00000000-0000-0000-0000-000000000001",
			Action:     "agent.delivery_submitted",
			ActorType:  "agent",
			ActorID:    &actorID,
			TargetType: "card",
			TargetID:   &cardID,
			Meta: map[string]any{
				"agent_role": principal.Role,
				"to_stage":   string(card.StageEntregaAprovar),
			},
			At: now,
		})
	})

	if err != nil {
		return err
	}

	_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": cardID})
	_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)

	return nil
}
