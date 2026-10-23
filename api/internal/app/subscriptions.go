package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("assinatura não encontrada")
	ErrChargeNotFound       = errors.New("cobrança de mensalidade não encontrada")
)

type SubscriptionsUseCases struct {
	uow       UnitOfWork
	publisher EventPublisher
}

func NewSubscriptionsUseCases(uow UnitOfWork, publisher EventPublisher) *SubscriptionsUseCases {
	return &SubscriptionsUseCases{
		uow:       uow,
		publisher: publisher,
	}
}

type CreateSubscriptionInput struct {
	ClientID    string `json:"client_id"`
	AppID       *string `json:"app_id,omitempty"`
	Plan        string `json:"plan"`
	Description string `json:"description"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	Interval    string `json:"interval"` // mensal, trimestral, semestral, anual
	DueDay      int    `json:"due_day"`
}

type DailyProcessReport struct {
	OverdueCount  int `json:"overdue_count"`
	GeneratedCount int `json:"generated_count"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// CreateSubscription registra uma nova assinatura e agenda a primeira cobrança
func (uc *SubscriptionsUseCases) CreateSubscription(ctx context.Context, principal system.OwnerPrincipal, in CreateSubscriptionInput) (*subscription.Subscription, error) {
	if in.ClientID == "" {
		return nil, errors.New("client_id é obrigatório")
	}
	if in.Plan == "" {
		return nil, errors.New("plano é obrigatório")
	}
	if in.AmountCents <= 0 {
		return nil, subscription.ErrInvalidAmount
	}
	if in.DueDay < 1 || in.DueDay > 31 {
		return nil, subscription.ErrInvalidDueDay
	}
	if in.Currency == "" {
		in.Currency = "BRL"
	}
	if in.Interval == "" {
		in.Interval = string(subscription.IntervalMensal)
	}

	interval := subscription.Interval(in.Interval)
	if interval != subscription.IntervalMensal &&
		interval != subscription.IntervalTrimestral &&
		interval != subscription.IntervalSemestral &&
		interval != subscription.IntervalAnual {
		return nil, subscription.ErrInvalidInterval
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Calcula o primeiro vencimento
	firstDue := time.Date(today.Year(), today.Month(), in.DueDay, 0, 0, 0, 0, time.UTC)
	if !firstDue.After(today) {
		// Se o dia já passou no mês corrente, primeiro vencimento no próximo mês
		firstDue = firstDue.AddDate(0, 1, 0)
	}

	sub := &subscription.Subscription{
		ID:          uuid.New().String(),
		OwnerID:     principal.ID,
		ClientID:    in.ClientID,
		AppID:       in.AppID,
		Plan:        in.Plan,
		Description: in.Description,
		AmountCents: in.AmountCents,
		Currency:    in.Currency,
		Interval:    interval,
		DueDay:      in.DueDay,
		Status:      subscription.StatusAtiva,
		StartDate:   today,
		NextDueDate: firstDue,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	charge := &subscription.SubscriptionCharge{
		ID:             uuid.New().String(),
		OwnerID:        principal.ID,
		SubscriptionID: sub.ID,
		AmountCents:    sub.AmountCents,
		Currency:       sub.Currency,
		DueDate:        firstDue,
		Status:         subscription.ChargeStatusPendente,
		PaymentMethod:  "pix",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		if err := tx.Subscriptions.CreateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("falha ao criar assinatura: %w", err)
		}
		if err := tx.Subscriptions.CreateCharge(ctx, charge); err != nil {
			return fmt.Errorf("falha ao gerar cobrança inicial: %w", err)
		}
		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "subscription_created",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "subscription",
				TargetID:   &sub.ID,
				Meta: map[string]any{
					"plan":         sub.Plan,
					"amount_cents": sub.AmountCents,
					"interval":     sub.Interval,
					"due_day":      sub.DueDay,
				},
				At: now,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicFinancesChanged, map[string]any{
			"event":           "subscription_created",
			"subscription_id": sub.ID,
		})
	}

	return sub, nil
}

// UpdateSubscriptionStatus altera o status de uma assinatura (ativa, pausada, cancelada)
func (uc *SubscriptionsUseCases) UpdateSubscriptionStatus(ctx context.Context, principal system.OwnerPrincipal, subscriptionID string, newStatus string) (*subscription.Subscription, error) {
	status := subscription.Status(newStatus)
	if status != subscription.StatusAtiva &&
		status != subscription.StatusPausada &&
		status != subscription.StatusCancelada &&
		status != subscription.StatusInadimplente {
		return nil, subscription.ErrInvalidStatus
	}

	var updatedSub *subscription.Subscription
	now := time.Now().UTC()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		sub, err := tx.Subscriptions.GetSubscriptionByID(ctx, subscriptionID)
		if err != nil || sub == nil {
			return ErrSubscriptionNotFound
		}

		sub.Status = status
		sub.UpdatedAt = now

		if err := tx.Subscriptions.UpdateSubscriptionStatus(ctx, subscriptionID, status, nil); err != nil {
			return fmt.Errorf("falha ao atualizar status da assinatura: %w", err)
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "subscription_status_updated",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "subscription",
				TargetID:   &subscriptionID,
				Meta: map[string]any{
					"old_status": sub.Status,
					"new_status": status,
				},
				At: now,
			})
		}

		updatedSub = sub
		return nil
	})

	if err != nil {
		return nil, err
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicFinancesChanged, map[string]any{
			"event":           "subscription_status_updated",
			"subscription_id": subscriptionID,
			"new_status":      newStatus,
		})
	}

	return updatedSub, nil
}

// GetMRRMetrics calcula o placar completo de MRR
func (uc *SubscriptionsUseCases) GetMRRMetrics(ctx context.Context, principal system.OwnerPrincipal) (*subscription.MRRMetrics, error) {
	var metrics subscription.MRRMetrics

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		subs, err := tx.Subscriptions.ListSubscriptions(ctx, nil, nil)
		if err != nil {
			return fmt.Errorf("falha ao listar assinaturas: %w", err)
		}
		metrics = subscription.CalculateMRR(subs)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

// ListSubscriptions retorna lista de assinaturas
func (uc *SubscriptionsUseCases) ListSubscriptions(ctx context.Context, principal system.OwnerPrincipal, clientID *string, status *string) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		subs, err = tx.Subscriptions.ListSubscriptions(ctx, clientID, status)
		return err
	})
	return subs, err
}

// ListCharges retorna cobranças a vencer ou atrasadas
func (uc *SubscriptionsUseCases) ListCharges(ctx context.Context, principal system.OwnerPrincipal, status *string, dueBefore *time.Time) ([]subscription.SubscriptionCharge, error) {
	var charges []subscription.SubscriptionCharge
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		var err error
		charges, err = tx.Subscriptions.ListCharges(ctx, status, dueBefore)
		return err
	})
	return charges, err
}

// PayCharge liquida uma mensalidade, atualiza status e lança no Ledger
func (uc *SubscriptionsUseCases) PayCharge(ctx context.Context, principal system.OwnerPrincipal, chargeID string, paymentMethod string) (*subscription.SubscriptionCharge, error) {
	if paymentMethod == "" {
		paymentMethod = "pix"
	}
	now := time.Now().UTC()
	var updatedCharge *subscription.SubscriptionCharge

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		charge, err := tx.Subscriptions.GetChargeByID(ctx, chargeID)
		if err != nil || charge == nil {
			return ErrChargeNotFound
		}

		if charge.Status == subscription.ChargeStatusPaga {
			return subscription.ErrChargeAlreadyPaid
		}

		charge.Status = subscription.ChargeStatusPaga
		charge.PaidAt = &now
		charge.PaymentMethod = paymentMethod
		charge.UpdatedAt = now

		if err := tx.Subscriptions.UpdateChargeStatus(ctx, chargeID, subscription.ChargeStatusPaga, &now, paymentMethod); err != nil {
			return fmt.Errorf("falha ao liquidar cobrança: %w", err)
		}

		// Buscar assinatura para registrar receita com nome do app / plano
		sub, _ := tx.Subscriptions.GetSubscriptionByID(ctx, charge.SubscriptionID)
		desc := "Mensalidade de assinatura"
		var appID *string
		if sub != nil {
			desc = fmt.Sprintf("Mensalidade Plano %s (%s)", sub.Plan, sub.Interval)
			appID = sub.AppID
		}

		// Cria lançamento financeiro no Ledger (receita)
		if tx.Finances != nil {
			_, _ = tx.Finances.CreateEntry(ctx, CreateLedgerEntryCommand{
				AppID:       appID,
				ClientID:    subClientID(sub),
				Kind:        "receita",
				AmountCents: charge.AmountCents,
				Currency:    charge.Currency,
				OccurredOn:  now.Format("2006-01-02"),
				Service:     "Mensalidade Recorrente",
				Description: desc,
			})
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "subscription_charge_paid",
				ActorType:  "owner",
				ActorID:    &actorID,
				TargetType: "subscription_charge",
				TargetID:   &chargeID,
				Meta: map[string]any{
					"subscription_id": charge.SubscriptionID,
					"amount_cents":    charge.AmountCents,
					"payment_method":  paymentMethod,
					"paid_at":         now,
				},
				At: now,
			})
		}

		updatedCharge = charge
		return nil
	})

	if err != nil {
		return nil, err
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicFinancesChanged, map[string]any{
			"event":     "subscription_charge_paid",
			"charge_id": chargeID,
		})
	}

	return updatedCharge, nil
}

// ProcessDailySubscriptionCharges é a rotina diária idempotente de geração e atualização de status
func (uc *SubscriptionsUseCases) ProcessDailySubscriptionCharges(ctx context.Context, principal system.OwnerPrincipal, refDate time.Time) (*DailyProcessReport, error) {
	if refDate.IsZero() {
		refDate = time.Now().UTC()
	}
	today := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)
	report := &DailyProcessReport{
		ProcessedAt: today,
	}

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		// 1. Marcar cobranças pendentes vencidas como "atrasada"
		charges, err := tx.Subscriptions.ListCharges(ctx, nil, nil)
		if err == nil {
			for _, ch := range charges {
				if ch.Status == subscription.ChargeStatusPendente && ch.DueDate.Before(today) {
					_ = tx.Subscriptions.UpdateChargeStatus(ctx, ch.ID, subscription.ChargeStatusAtrasada, nil, ch.PaymentMethod)
					report.OverdueCount++

					// Se atrasada, marcar assinatura como inadimplente
					_ = tx.Subscriptions.UpdateSubscriptionStatus(ctx, ch.SubscriptionID, subscription.StatusInadimplente, nil)
				}
			}
		}

		// 2. Para assinaturas ativas, gerar a próxima cobrança se o vencimento for nos próximos 7 dias
		horizon := today.AddDate(0, 0, 7)
		activeStatus := string(subscription.StatusAtiva)
		subs, err := tx.Subscriptions.ListSubscriptions(ctx, nil, &activeStatus)
		if err == nil {
			for _, sub := range subs {
				if !sub.NextDueDate.After(horizon) {
					newCharge := &subscription.SubscriptionCharge{
						ID:             uuid.New().String(),
						OwnerID:        principal.ID,
						SubscriptionID: sub.ID,
						AmountCents:    sub.AmountCents,
						Currency:       sub.Currency,
						DueDate:        sub.NextDueDate,
						Status:         subscription.ChargeStatusPendente,
						PaymentMethod:  "pix",
						CreatedAt:      refDate,
						UpdatedAt:      refDate,
					}
					// Criação é idempotente devido à constraint UNIQUE (subscription_id, due_date)
					if err := tx.Subscriptions.CreateCharge(ctx, newCharge); err == nil {
						report.GeneratedCount++
						// Avança o próximo vencimento
						nextDue := sub.CalculateNextDueDate(sub.NextDueDate)
						_ = tx.Subscriptions.UpdateSubscriptionNextDueDate(ctx, sub.ID, nextDue)
					}
				}
			}
		}

		if tx.Audit != nil {
			actorID := principal.ID
			_ = tx.Audit.Record(ctx, AuditLogDTO{
				Action:     "subscription_daily_process",
				ActorType:  "system",
				ActorID:    &actorID,
				TargetType: "subscription_batch",
				Meta: map[string]any{
					"overdue_count":   report.OverdueCount,
					"generated_count": report.GeneratedCount,
					"reference_date":  today,
				},
				At: refDate,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return report, nil
}

func subClientID(sub *subscription.Subscription) *string {
	if sub == nil || sub.ClientID == "" {
		return nil
	}
	return &sub.ClientID
}
