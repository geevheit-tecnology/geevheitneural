package app

import (
	"context"
	"fmt"
	"time"
)

type PublicScopeViewDTO struct {
	CardTitle          string           `json:"card_title"`
	UserStories        []map[string]any `json:"user_stories"`
	TestsPassed        bool             `json:"tests_passed"`
	ClientApproved     bool             `json:"client_approved"`
	PaymentSignalCents int64            `json:"payment_signal_cents"`
	PaymentSignalPaid  bool             `json:"payment_signal_paid"`
}

type GetPublicScopeUseCase struct {
	uow UnitOfWork
}

func NewGetPublicScopeUseCase(uow UnitOfWork) *GetPublicScopeUseCase {
	return &GetPublicScopeUseCase{uow: uow}
}

func (uc *GetPublicScopeUseCase) Execute(ctx context.Context, rawToken string) (*PublicScopeViewDTO, error) {
	hash := hashToken(rawToken)

	var view PublicScopeViewDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		spec, err := tx.Specifications.GetByTokenHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("especificação não encontrada para o token informado: %w", err)
		}

		c, err := tx.Cards.GetByID(ctx, spec.CardID)
		if err != nil {
			return fmt.Errorf("cartão da especificação não encontrado: %w", err)
		}

		view = PublicScopeViewDTO{
			CardTitle:          c.Title,
			UserStories:        spec.UserStories,
			TestsPassed:        spec.TestsPassed,
			ClientApproved:     spec.ClientApproved,
			PaymentSignalCents: spec.PaymentSignalCents,
			PaymentSignalPaid:  spec.PaymentSignalPaid,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &view, nil
}

type ApprovePublicScopeUseCase struct {
	uow    UnitOfWork
	events EventPublisher
}

func NewApprovePublicScopeUseCase(uow UnitOfWork, events EventPublisher) *ApprovePublicScopeUseCase {
	return &ApprovePublicScopeUseCase{uow: uow, events: events}
}

func (uc *ApprovePublicScopeUseCase) Execute(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)

	var cardID string
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		spec, err := tx.Specifications.GetByTokenHash(ctx, hash)
		if err != nil {
			return fmt.Errorf("escopo não localizado: %w", err)
		}
		cardID = spec.CardID

		if spec.ClientApproved {
			return nil // Já aprovado, idempotente
		}

		if err := tx.Specifications.ApproveByClient(ctx, hash); err != nil {
			return fmt.Errorf("falha ao registrar aprovação do escopo: %w", err)
		}

		_ = tx.CardEvents.Create(ctx, CardEventDTO{
			CardID:    spec.CardID,
			ActorType: "client",
			Kind:      "scope_approved_by_client",
			Payload: map[string]any{
				"approved_at": time.Now().Format(time.RFC3339),
			},
			At: time.Now(),
		})

		_ = tx.Audit.Record(ctx, AuditLogDTO{
			Action:     "client_approved_scope",
			ActorType:  "client",
			TargetType: "specification",
			TargetID:   &spec.ID,
			Meta:       map[string]any{"card_id": spec.CardID},
		})

		return nil
	})
	if err != nil {
		return err
	}

	_ = uc.events.Publish(ctx, TopicCardsChanged, map[string]string{"id": cardID, "event": "scope_approved"})
	return nil
}
