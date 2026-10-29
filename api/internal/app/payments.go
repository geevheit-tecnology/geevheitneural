package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/payment"
	"github.com/google/uuid"
)

var (
	ErrCardHasNoPrice = errors.New("cartão não possui valor financeiro para cobrança")
)

type GeneratePixPaymentUseCase struct {
	uow    UnitOfWork
	events EventPublisher
	clock  clock.Clock
}

func NewGeneratePixPaymentUseCase(uow UnitOfWork, events EventPublisher, clk clock.Clock) *GeneratePixPaymentUseCase {
	return &GeneratePixPaymentUseCase{
		uow:    uow,
		events: events,
		clock:  clk,
	}
}

func (uc *GeneratePixPaymentUseCase) Execute(ctx context.Context, cardID, ownerID, pixKey string) (*PaymentOrderDTO, error) {
	var result *PaymentOrderDTO
	now := uc.clock.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		c, err := tx.Cards.GetByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("cartão não encontrado: %w", err)
		}

		if c.Price.Cents <= 0 {
			return ErrCardHasNoPrice
		}

		// Idempotência: se já existe cobrança pendente para este cartão com o mesmo valor, reutiliza
		existing, err := tx.PaymentOrders.GetByCardID(ctx, cardID)
		if err == nil && existing != nil && existing.Status == string(payment.StatusPending) && existing.AmountCents == c.Price.Cents {
			result = existing
			return nil
		}

		val := money.NewBRL(c.Price.Cents)
		txid := fmt.Sprintf("GEEV%d", now.UnixNano())
		if len(txid) > 25 {
			txid = txid[:25]
		}

		order, err := payment.NewOrder(uuid.NewString(), ownerID, c.ID, c.ClientID, val, pixKey, txid, now)
		if err != nil {
			return fmt.Errorf("erro ao gerar ordem de pagamento: %w", err)
		}

		dto := &PaymentOrderDTO{
			ID:          order.ID,
			OwnerID:     order.OwnerID,
			CardID:      order.CardID,
			ClientID:    order.ClientID,
			AmountCents: order.Amount.Cents,
			Currency:    order.Amount.Currency,
			PixEMV:      order.PixEMV,
			PixTxID:     order.PixTxID,
			PixKey:      order.PixKey,
			Status:      string(order.Status),
			CreatedAt:   order.CreatedAt,
			UpdatedAt:   order.UpdatedAt,
		}

		if err := tx.PaymentOrders.Create(ctx, dto); err != nil {
			return fmt.Errorf("erro ao persistir ordem de pagamento: %w", err)
		}

		// Registra evento no histórico do cartão
		_ = tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.NewString(),
			OwnerID:   ownerID,
			CardID:    c.ID,
			Kind:      "pix_generated",
			ActorType: "owner",
			Payload: map[string]any{
				"amount_cents": c.Price.Cents,
				"txid":         order.PixTxID,
				"pix_key":      order.PixKey,
			},
			At: now,
		})

		result = dto
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

type GetPixPaymentUseCase struct {
	uow UnitOfWork
}

func NewGetPixPaymentUseCase(uow UnitOfWork) *GetPixPaymentUseCase {
	return &GetPixPaymentUseCase{uow: uow}
}

func (uc *GetPixPaymentUseCase) Execute(ctx context.Context, cardID string) (*PaymentOrderDTO, error) {
	var result *PaymentOrderDTO
	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		order, err := tx.PaymentOrders.GetByCardID(ctx, cardID)
		if err != nil {
			return payment.ErrOrderNotFound
		}
		result = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ProcessPixWebhookUseCase struct {
	uow    UnitOfWork
	events EventPublisher
	clock  clock.Clock
}

func NewProcessPixWebhookUseCase(uow UnitOfWork, events EventPublisher, clk clock.Clock) *ProcessPixWebhookUseCase {
	return &ProcessPixWebhookUseCase{
		uow:    uow,
		events: events,
		clock:  clk,
	}
}

func (uc *ProcessPixWebhookUseCase) Execute(ctx context.Context, txid string, paidAmountCents int64) (*PaymentOrderDTO, error) {
	var result *PaymentOrderDTO
	now := uc.clock.Now()

	err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
		order, err := tx.PaymentOrders.GetByTxID(ctx, txid)
		if err != nil {
			return payment.ErrOrderNotFound
		}

		// Idempotência: se já foi pago, retorna sem duplicar lançamentos
		if order.Status == string(payment.StatusPaid) {
			result = order
			return nil
		}

		// Marca pago
		if err := tx.PaymentOrders.UpdateStatus(ctx, order.ID, string(payment.StatusPaid), &now); err != nil {
			return fmt.Errorf("erro ao atualizar status da ordem: %w", err)
		}
		order.Status = string(payment.StatusPaid)
		order.PaidAt = &now
		order.UpdatedAt = now

		// 1. Registra lançamento financeiro automático (crédito em receita)
		_, _ = tx.Finances.CreateEntry(ctx, CreateLedgerEntryCommand{
			ClientID:    &order.ClientID,
			Kind:        "receita",
			AmountCents: order.AmountCents,
			Currency:    order.Currency,
			OccurredOn:  now.Format("2006-01-02"),
			Service:     "PIX",
			Description: fmt.Sprintf("Recebimento PIX ref. cartão %s (TXID %s)", order.CardID[:8], order.PixTxID),
		})

		// 2. Registra evento no cartão
		_ = tx.CardEvents.Create(ctx, CardEventDTO{
			ID:        uuid.NewString(),
			OwnerID:   order.OwnerID,
			CardID:    order.CardID,
			Kind:      "pix_confirmed",
			ActorType: "system",
			Payload: map[string]any{
				"amount_cents": order.AmountCents,
				"txid":         order.PixTxID,
				"paid_at":      now.Format(time.RFC3339),
			},
			At: now,
		})

		// 3. Auditoria imutável
		_ = tx.Audit.Record(ctx, AuditLogDTO{
			ID:         uuid.NewString(),
			OwnerID:    order.OwnerID,
			Action:     "payment.pix_confirmed",
			ActorType:  "system",
			TargetType: "payment_order",
			TargetID:   &order.ID,
			Meta: map[string]any{
				"card_id":      order.CardID,
				"amount_cents": order.AmountCents,
				"txid":         order.PixTxID,
			},
			At: now,
		})

		// 4. Se houver especificação de escopo com sinal pendente, liquida o sinal automaticamente
		if tx.Specifications != nil {
			spec, errSpec := tx.Specifications.GetByCardID(ctx, order.CardID)
			if errSpec == nil && spec != nil {
				spec.PaymentSignalPaid = true
				_ = tx.Specifications.Upsert(ctx, spec, "")
			}
		}

		result = order
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Publica eventos em tempo real para os clientes conectados via SSE
	_ = uc.events.Publish(ctx, TopicFinancesChanged, result)
	_ = uc.events.Publish(ctx, TopicCardsChanged, result)
	_ = uc.events.Publish(ctx, TopicMoneyWaitingChanged, result)

	return result, nil
}
