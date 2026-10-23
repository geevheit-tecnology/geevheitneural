package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type fakeSubscriptionRepo struct {
	subs    map[string]*subscription.Subscription
	charges map[string]*subscription.SubscriptionCharge
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{
		subs:    make(map[string]*subscription.Subscription),
		charges: make(map[string]*subscription.SubscriptionCharge),
	}
}

func (r *fakeSubscriptionRepo) CreateSubscription(ctx context.Context, sub *subscription.Subscription) error {
	r.subs[sub.ID] = sub
	return nil
}

func (r *fakeSubscriptionRepo) GetSubscriptionByID(ctx context.Context, id string) (*subscription.Subscription, error) {
	s, ok := r.subs[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (r *fakeSubscriptionRepo) ListSubscriptions(ctx context.Context, clientID *string, status *string) ([]subscription.Subscription, error) {
	var list []subscription.Subscription
	for _, s := range r.subs {
		if clientID != nil && s.ClientID != *clientID {
			continue
		}
		if status != nil && string(s.Status) != *status {
			continue
		}
		list = append(list, *s)
	}
	return list, nil
}

func (r *fakeSubscriptionRepo) UpdateSubscriptionStatus(ctx context.Context, id string, status subscription.Status, nextDue *time.Time) error {
	if s, ok := r.subs[id]; ok {
		s.Status = status
		if nextDue != nil {
			s.NextDueDate = *nextDue
		}
	}
	return nil
}

func (r *fakeSubscriptionRepo) UpdateSubscriptionNextDueDate(ctx context.Context, id string, nextDue time.Time) error {
	if s, ok := r.subs[id]; ok {
		s.NextDueDate = nextDue
	}
	return nil
}

func (r *fakeSubscriptionRepo) CreateCharge(ctx context.Context, charge *subscription.SubscriptionCharge) error {
	// Simula a constraint UNIQUE (subscription_id, due_date)
	for _, c := range r.charges {
		if c.SubscriptionID == charge.SubscriptionID && c.DueDate.Equal(charge.DueDate) {
			return nil // Ignora idempotente
		}
	}
	r.charges[charge.ID] = charge
	return nil
}

func (r *fakeSubscriptionRepo) GetChargeByID(ctx context.Context, id string) (*subscription.SubscriptionCharge, error) {
	c, ok := r.charges[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (r *fakeSubscriptionRepo) ListCharges(ctx context.Context, status *string, dueBefore *time.Time) ([]subscription.SubscriptionCharge, error) {
	var list []subscription.SubscriptionCharge
	for _, c := range r.charges {
		if status != nil && string(c.Status) != *status {
			continue
		}
		if dueBefore != nil && c.DueDate.After(*dueBefore) {
			continue
		}
		list = append(list, *c)
	}
	return list, nil
}

func (r *fakeSubscriptionRepo) UpdateChargeStatus(ctx context.Context, id string, status subscription.ChargeStatus, paidAt *time.Time, method string) error {
	if c, ok := r.charges[id]; ok {
		c.Status = status
		c.PaidAt = paidAt
		c.PaymentMethod = method
	}
	return nil
}

type subFakeAuditRepo struct {
	records []app.AuditLogDTO
}

func (f *subFakeAuditRepo) Record(ctx context.Context, r app.AuditLogDTO) error {
	f.records = append(f.records, r)
	return nil
}

type subFakeFinanceRepo struct {
	entries []app.CreateLedgerEntryCommand
}

func (f *subFakeFinanceRepo) CreateEntry(ctx context.Context, entry app.CreateLedgerEntryCommand) (*app.LedgerEntryDTO, error) {
	f.entries = append(f.entries, entry)
	return &app.LedgerEntryDTO{ID: "entry-1"}, nil
}

func (f *subFakeFinanceRepo) GetOverview(ctx context.Context) (*app.FinancesOverviewDTO, error) {
	return &app.FinancesOverviewDTO{}, nil
}

type subFakeUoW struct {
	repos app.Repositories
}

func (u *subFakeUoW) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	return fn(u.repos)
}

func setupSubscriptionsTest() (*app.SubscriptionsUseCases, *fakeSubscriptionRepo, *subFakeFinanceRepo, *subFakeAuditRepo) {
	subRepo := newFakeSubscriptionRepo()
	finRepo := &subFakeFinanceRepo{}
	auditRepo := &subFakeAuditRepo{}

	uow := &subFakeUoW{
		repos: app.Repositories{
			Subscriptions: subRepo,
			Finances:      finRepo,
			Audit:         auditRepo,
		},
	}

	uc := app.NewSubscriptionsUseCases(uow, nil)
	return uc, subRepo, finRepo, auditRepo
}

func TestSubscriptionsUseCases_CreateAndPay(t *testing.T) {
	uc, subRepo, finRepo, auditRepo := setupSubscriptionsTest()
	principal := system.OwnerPrincipal{ID: "owner-1"}

	ctx := context.Background()
	in := app.CreateSubscriptionInput{
		ClientID:    "client-123",
		Plan:        "Suporte Pro",
		Description: "Manutenção mensal e monitoramento",
		AmountCents: 50000, // R$ 500,00
		Currency:    "BRL",
		Interval:    "mensal",
		DueDay:      15,
	}

	sub, err := uc.CreateSubscription(ctx, principal, in)
	if err != nil {
		t.Fatalf("CreateSubscription falhou: %v", err)
	}

	if sub.ID == "" || sub.Status != subscription.StatusAtiva {
		t.Errorf("Assinatura criada inválida: %+v", sub)
	}

	if len(subRepo.charges) != 1 {
		t.Fatalf("Esperava 1 cobrança gerada, obteve %d", len(subRepo.charges))
	}

	// Pega a cobrança criada
	var chargeID string
	for id := range subRepo.charges {
		chargeID = id
		break
	}

	// Paga a cobrança
	paidCharge, err := uc.PayCharge(ctx, principal, chargeID, "pix")
	if err != nil {
		t.Fatalf("PayCharge falhou: %v", err)
	}

	if paidCharge.Status != subscription.ChargeStatusPaga || paidCharge.PaidAt == nil {
		t.Errorf("Cobrança não foi marcada como paga: %+v", paidCharge)
	}

	// Verifica se registrou receita no financeiro
	if len(finRepo.entries) != 1 {
		t.Errorf("Esperava 1 entrada no financeiro, obteve %d", len(finRepo.entries))
	} else if finRepo.entries[0].AmountCents != 50000 || finRepo.entries[0].Kind != "receita" {
		t.Errorf("Entrada financeira incorreta: %+v", finRepo.entries[0])
	}

	// Verifica audit log
	if len(auditRepo.records) < 2 {
		t.Errorf("Esperava pelo menos 2 registros no audit log, obteve %d", len(auditRepo.records))
	}
}

func TestSubscriptionsUseCases_DailyProcess_Idempotence(t *testing.T) {
	uc, subRepo, _, _ := setupSubscriptionsTest()
	principal := system.OwnerPrincipal{ID: "owner-1"}
	ctx := context.Background()

	refDate := time.Date(2026, time.October, 23, 10, 0, 0, 0, time.UTC)
	dueSoon := time.Date(2026, time.October, 25, 0, 0, 0, 0, time.UTC)

	// Cria assinatura ativa com vencimento em 2 dias (dentro da janela de 7 dias)
	sub := &subscription.Subscription{
		ID:          "sub-active",
		OwnerID:     "owner-1",
		ClientID:    "client-1",
		Plan:        "Start",
		AmountCents: 20000,
		Interval:    subscription.IntervalMensal,
		DueDay:      25,
		Status:      subscription.StatusAtiva,
		NextDueDate: dueSoon,
	}
	_ = subRepo.CreateSubscription(ctx, sub)

	// Cria também uma cobrança antiga pendente (vencida ontem)
	overdueDue := time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC)
	overdueCharge := &subscription.SubscriptionCharge{
		ID:             "charge-overdue",
		SubscriptionID: "sub-active",
		AmountCents:    20000,
		DueDate:        overdueDue,
		Status:         subscription.ChargeStatusPendente,
	}
	_ = subRepo.CreateCharge(ctx, overdueCharge)

	// Executa rotina diária primeira vez
	rep1, err := uc.ProcessDailySubscriptionCharges(ctx, principal, refDate)
	if err != nil {
		t.Fatalf("ProcessDaily 1 falhou: %v", err)
	}

	if rep1.OverdueCount != 1 {
		t.Errorf("Esperava 1 cobrança marcada como atrasada, obteve %d", rep1.OverdueCount)
	}
	if overdueCharge.Status != subscription.ChargeStatusAtrasada {
		t.Errorf("Cobrança antiga não virou atrasada: %s", overdueCharge.Status)
	}

	totalChargesAfterRun1 := len(subRepo.charges)

	// Executa rotina diária segunda vez (idempotência)
	rep2, err := uc.ProcessDailySubscriptionCharges(ctx, principal, refDate)
	if err != nil {
		t.Fatalf("ProcessDaily 2 falhou: %v", err)
	}

	totalChargesAfterRun2 := len(subRepo.charges)
	if totalChargesAfterRun2 != totalChargesAfterRun1 {
		t.Errorf("ProcessDaily não foi idempotente! Criou cobranças duplicadas: %d != %d", totalChargesAfterRun2, totalChargesAfterRun1)
	}
	if rep2.GeneratedCount != 0 {
		t.Errorf("Segunda execução não deveria gerar novas cobranças na mesma data: %d", rep2.GeneratedCount)
	}
}
