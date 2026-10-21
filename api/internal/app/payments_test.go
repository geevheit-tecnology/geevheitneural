package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/payment"
)

type mockPaymentRepo struct {
	orders map[string]*app.PaymentOrderDTO
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{orders: make(map[string]*app.PaymentOrderDTO)}
}

func (r *mockPaymentRepo) Create(ctx context.Context, order *app.PaymentOrderDTO) error {
	r.orders[order.ID] = order
	return nil
}

func (r *mockPaymentRepo) GetByID(ctx context.Context, id string) (*app.PaymentOrderDTO, error) {
	o, ok := r.orders[id]
	if !ok {
		return nil, payment.ErrOrderNotFound
	}
	return o, nil
}

func (r *mockPaymentRepo) GetByTxID(ctx context.Context, txid string) (*app.PaymentOrderDTO, error) {
	for _, o := range r.orders {
		if o.PixTxID == txid {
			return o, nil
		}
	}
	return nil, payment.ErrOrderNotFound
}

func (r *mockPaymentRepo) GetByCardID(ctx context.Context, cardID string) (*app.PaymentOrderDTO, error) {
	for _, o := range r.orders {
		if o.CardID == cardID {
			return o, nil
		}
	}
	return nil, payment.ErrOrderNotFound
}

func (r *mockPaymentRepo) UpdateStatus(ctx context.Context, id, status string, paidAt *time.Time) error {
	o, ok := r.orders[id]
	if !ok {
		return payment.ErrOrderNotFound
	}
	o.Status = status
	o.PaidAt = paidAt
	return nil
}

type mockPaymentFinanceRepo struct {
	entries []app.CreateLedgerEntryCommand
}

func (r *mockPaymentFinanceRepo) GetOverview(ctx context.Context) (*app.FinancesOverviewDTO, error) {
	return &app.FinancesOverviewDTO{}, nil
}

func (r *mockPaymentFinanceRepo) CreateEntry(ctx context.Context, cmd app.CreateLedgerEntryCommand) (*app.LedgerEntryDTO, error) {
	r.entries = append(r.entries, cmd)
	return &app.LedgerEntryDTO{
		ID:          "fin-entry-1",
		Kind:        cmd.Kind,
		AmountCents: cmd.AmountCents,
		Currency:    cmd.Currency,
		Service:     cmd.Service,
		Description: cmd.Description,
	}, nil
}

type paymentTestUOW struct {
	repos app.Repositories
}

func (u *paymentTestUOW) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	return fn(u.repos)
}

type testPublisher struct{}

func (p *testPublisher) Publish(ctx context.Context, topic string, payload any) error {
	return nil
}

func TestPaymentsWorkflow(t *testing.T) {
	clk := clock.FrozenClock{Current: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)}
	pub := &testPublisher{}

	pRepo := newMockPaymentRepo()
	fRepo := &mockPaymentFinanceRepo{}
	cRepo := &mockCardRepoPayment{
		card: &card.Card{
			ID:       "card-pix-1",
			OwnerID:  "owner-1",
			ClientID: "client-1",
			Title:    "Site Odonto",
			Price:    money.NewBRL(350000), // R$ 3.500,00
			Stage:    card.StagePropostaAprovar,
		},
	}
	evRepo := &mockCardEventRepoPayment{}
	auditRepo := &mockAuditRepoPayment{}

	repos := app.Repositories{
		Cards:         cRepo,
		PaymentOrders: pRepo,
		Finances:      fRepo,
		CardEvents:    evRepo,
		Audit:         auditRepo,
	}
	uow := &paymentTestUOW{repos: repos}

	generateUC := app.NewGeneratePixPaymentUseCase(uow, pub, clk)
	getUC := app.NewGetPixPaymentUseCase(uow)
	webhookUC := app.NewProcessPixWebhookUseCase(uow, pub, clk)

	// 1. Gerar PIX
	order, err := generateUC.Execute(context.Background(), "card-pix-1", "owner-1", "financeiro@geevheit.com")
	if err != nil {
		t.Fatalf("falha ao gerar pix: %v", err)
	}

	if order.AmountCents != 350000 {
		t.Errorf("esperado 350000 cents, obtido %d", order.AmountCents)
	}
	if order.Status != "pending" {
		t.Errorf("esperado status pending, obtido %s", order.Status)
	}
	if order.PixEMV == "" || order.PixTxID == "" {
		t.Error("esperado PIX EMV e TXID preenchidos")
	}

	// 2. Consulta PIX
	queried, err := getUC.Execute(context.Background(), "card-pix-1")
	if err != nil {
		t.Fatalf("falha ao consultar pix: %v", err)
	}
	if queried.ID != order.ID {
		t.Errorf("esperado mesmo id %s, obtido %s", order.ID, queried.ID)
	}

	// 3. Webhook de Confirmação de Pagamento
	webhookOrder, err := webhookUC.Execute(context.Background(), order.PixTxID, 350000)
	if err != nil {
		t.Fatalf("falha ao processar webhook: %v", err)
	}

	if webhookOrder.Status != "paid" {
		t.Errorf("esperado status paid, obtido %s", webhookOrder.Status)
	}
	if webhookOrder.PaidAt == nil {
		t.Error("esperado paid_at preenchido")
	}

	// 4. Verifica se gerou crédito no financeiro automaticamente (SOLID / Invariante)
	if len(fRepo.entries) != 1 {
		t.Fatalf("esperado 1 lançamento financeiro, obtido %d", len(fRepo.entries))
	}
	entry := fRepo.entries[0]
	if entry.Kind != "receita" || entry.AmountCents != 350000 {
		t.Errorf("lançamento inválido: kind=%s, cents=%d", entry.Kind, entry.AmountCents)
	}

	// 5. Idempotência do webhook: reprocessar o mesmo TXID não deve duplicar receita
	_, err = webhookUC.Execute(context.Background(), order.PixTxID, 350000)
	if err != nil {
		t.Fatalf("webhook idempotente falhou: %v", err)
	}
	if len(fRepo.entries) != 1 {
		t.Errorf("webhook duplicou lançamento financeiro!")
	}
}

type mockCardRepoPayment struct {
	card *card.Card
}

func (r *mockCardRepoPayment) GetByID(ctx context.Context, id string) (*card.Card, error) {
	if r.card != nil && r.card.ID == id {
		return r.card, nil
	}
	return nil, card.ErrEmptyTitle
}
func (r *mockCardRepoPayment) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPayment) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPayment) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPayment) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPayment) Update(ctx context.Context, c *card.Card) error {
	r.card = c
	return nil
}
func (r *mockCardRepoPayment) Create(ctx context.Context, c *app.CardListItemDTO) error {
	return nil
}

type mockCardEventRepoPayment struct{}

func (r *mockCardEventRepoPayment) Create(ctx context.Context, e app.CardEventDTO) error {
	return nil
}

type mockAuditRepoPayment struct{}

func (r *mockAuditRepoPayment) Record(ctx context.Context, l app.AuditLogDTO) error {
	return nil
}
