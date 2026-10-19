package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/approval"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

// Fakes para teste de casos de uso sem banco
type fakeCardRepo struct {
	cards map[string]*card.Card
}

func (f *fakeCardRepo) GetByID(_ context.Context, id string) (*card.Card, error) {
	c, ok := f.cards[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (f *fakeCardRepo) Update(_ context.Context, c *card.Card) error {
	f.cards[c.ID] = c
	return nil
}

func (f *fakeCardRepo) GetEnrichedByID(_ context.Context, id string) (*app.CardListItemDTO, error) {
	c, ok := f.cards[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &app.CardListItemDTO{
		ID:           c.ID,
		Type:         string(c.Type),
		ClientID:     c.ClientID,
		ClientName:   "Cliente Fake",
		AppName:      "Novo",
		Title:        c.Title,
		PriceCents:   c.Price.Cents,
		EstCostCents: c.EstCost.Cents,
		Currency:     c.Price.Currency,
		Stage:        string(c.Stage),
	}, nil
}

func (f *fakeCardRepo) ListEnriched(_ context.Context) ([]app.CardListItemDTO, error) {
	out := make([]app.CardListItemDTO, 0, len(f.cards))
	for _, c := range f.cards {
		out = append(out, app.CardListItemDTO{
			ID:           c.ID,
			Type:         string(c.Type),
			ClientID:     c.ClientID,
			ClientName:   "Cliente Fake",
			AppName:      "Novo",
			Title:        c.Title,
			PriceCents:   c.Price.Cents,
			EstCostCents: c.EstCost.Cents,
			Currency:     c.Price.Currency,
			Stage:        string(c.Stage),
		})
	}
	return out, nil
}

func (f *fakeCardRepo) ListEvents(_ context.Context, _ string) ([]app.CardEventDTO, error) {
	return []app.CardEventDTO{
		{Kind: "criado", At: time.Now()},
	}, nil
}

func (f *fakeCardRepo) GetRiskReport(_ context.Context, _ string) (*app.RiskReportDTO, error) {
	return nil, nil
}

func (f *fakeCardRepo) Create(_ context.Context, _ *app.CardListItemDTO) error {
	return nil
}

type fakeApprovalRepo struct {
	approvals []*approval.Approval
}

func (f *fakeApprovalRepo) Create(_ context.Context, a *approval.Approval) error {
	f.approvals = append(f.approvals, a)
	return nil
}

type fakeCardEventRepo struct {
	events []app.CardEventDTO
}

func (f *fakeCardEventRepo) Create(_ context.Context, e app.CardEventDTO) error {
	f.events = append(f.events, e)
	return nil
}

type fakeAuditRepo struct {
	records []app.AuditLogDTO
}

func (f *fakeAuditRepo) Record(_ context.Context, log app.AuditLogDTO) error {
	f.records = append(f.records, log)
	return nil
}

type fakeSystemStateRepo struct {
	state *system.SystemState
}

func (f *fakeSystemStateRepo) Get(_ context.Context) (*system.SystemState, error) {
	return f.state, nil
}

func (f *fakeSystemStateRepo) Save(_ context.Context, s *system.SystemState) error {
	f.state = s
	return nil
}

type fakeUnitOfWork struct {
	repos app.Repositories
}

func (f *fakeUnitOfWork) RunInTx(_ context.Context, fn func(tx app.Repositories) error) error {
	return fn(f.repos)
}

type fakePublisher struct {
	published []string
}

func (f *fakePublisher) Publish(_ context.Context, topic string, _ any) error {
	f.published = append(f.published, topic)
	return nil
}

type fakeMoneyWaitingReader struct {
	cents int64
	count int
}

func (f *fakeMoneyWaitingReader) GetMoneyWaiting(_ context.Context) (int64, int, error) {
	return f.cents, f.count, nil
}

func TestDecideGateUseCase(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 14, 0, 0, 0, time.UTC)
	clk := clock.FrozenClock{Current: now}
	owner, _ := system.NewOwnerPrincipal("owner-1", "aal2")

	t.Run("Approve proposal gate successfully", func(t *testing.T) {
		c, _ := card.New("card-1", owner.ID, "cli-1", card.TypeOrcamento, "Projeto Teste", money.NewBRL(500000), money.NewBRL(50000), now)
		c.Stage = card.StagePropostaAprovar

		cardRepo := &fakeCardRepo{cards: map[string]*card.Card{"card-1": c}}
		apprRepo := &fakeApprovalRepo{}
		eventRepo := &fakeCardEventRepo{}
		auditRepo := &fakeAuditRepo{}
		sysRepo := &fakeSystemStateRepo{state: &system.SystemState{Paused: false}}
		pub := &fakePublisher{}

		uow := &fakeUnitOfWork{
			repos: app.Repositories{
				Cards:       cardRepo,
				Approvals:   apprRepo,
				CardEvents:  eventRepo,
				Audit:       auditRepo,
				SystemState: sysRepo,
			},
		}

		uc := app.NewDecideGateUseCase(uow, pub, clk)
		err := uc.Execute(ctx, app.DecideGateInput{
			Owner:    owner,
			CardID:   "card-1",
			Decision: card.DecisionAprovado,
			Note:     "Aprovado escopo e margem",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Valida estado atualizado do cartão
		if c.Stage != card.StageAguardandoCliente {
			t.Errorf("expected StageAguardandoCliente, got %s", c.Stage)
		}

		// Valida aprovação gerada
		if len(apprRepo.approvals) != 1 {
			t.Fatalf("expected 1 approval, got %d", len(apprRepo.approvals))
		}
		if apprRepo.approvals[0].Decision != card.DecisionAprovado {
			t.Errorf("expected DecisionAprovado, got %s", apprRepo.approvals[0].Decision)
		}

		// Valida evento e auditoria
		if len(eventRepo.events) != 1 {
			t.Errorf("expected 1 card event, got %d", len(eventRepo.events))
		}
		if len(auditRepo.records) != 1 {
			t.Errorf("expected 1 audit record, got %d", len(auditRepo.records))
		}

		// Valida publicação de evento de dinheiro parado
		if len(pub.published) != 1 || pub.published[0] != app.TopicMoneyWaitingChanged {
			t.Errorf("expected TopicMoneyWaitingChanged published, got %v", pub.published)
		}
	})

	t.Run("Blocked when system is paused on prompt gate", func(t *testing.T) {
		c, _ := card.New("card-2", owner.ID, "cli-1", card.TypeOrcamento, "App", money.NewBRL(200000), money.NewBRL(20000), now)
		c.Stage = card.StagePromptAprovar

		cardRepo := &fakeCardRepo{cards: map[string]*card.Card{"card-2": c}}
		apprRepo := &fakeApprovalRepo{}
		eventRepo := &fakeCardEventRepo{}
		auditRepo := &fakeAuditRepo{}
		sysRepo := &fakeSystemStateRepo{state: &system.SystemState{Paused: true}}
		pub := &fakePublisher{}

		uow := &fakeUnitOfWork{
			repos: app.Repositories{
				Cards:       cardRepo,
				Approvals:   apprRepo,
				CardEvents:  eventRepo,
				Audit:       auditRepo,
				SystemState: sysRepo,
			},
		}

		uc := app.NewDecideGateUseCase(uow, pub, clk)
		err := uc.Execute(ctx, app.DecideGateInput{
			Owner:    owner,
			CardID:   "card-2",
			Decision: card.DecisionAprovado,
			Note:     "Tentativa de liberar",
		})

		if !errors.Is(err, card.ErrSystemPaused) {
			t.Errorf("expected ErrSystemPaused, got %v", err)
		}
	})
}

func TestPauseAndResumeSystemUseCase(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 14, 0, 0, 0, time.UTC)
	clk := clock.FrozenClock{Current: now}
	owner, _ := system.NewOwnerPrincipal("owner-1", "aal2")

	sysRepo := &fakeSystemStateRepo{state: &system.SystemState{Paused: false}}
	auditRepo := &fakeAuditRepo{}
	pub := &fakePublisher{}

	uow := &fakeUnitOfWork{
		repos: app.Repositories{
			Audit:       auditRepo,
			SystemState: sysRepo,
		},
	}

	pauseUC := app.NewPauseSystemUseCase(uow, pub, clk)
	err := pauseUC.Execute(ctx, owner, "Manutenção emergencial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !sysRepo.state.Paused {
		t.Errorf("expected system to be paused")
	}

	resumeUC := app.NewResumeSystemUseCase(uow, pub, clk)
	err = resumeUC.Execute(ctx, owner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sysRepo.state.Paused {
		t.Errorf("expected system to be resumed")
	}
}

func TestGetMoneyWaitingUseCase(t *testing.T) {
	ctx := context.Background()
	reader := &fakeMoneyWaitingReader{
		cents: 1450000, // R$ 14.500,00
		count: 3,
	}

	uc := app.NewGetMoneyWaitingUseCase(reader)
	res, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalCents != 1450000 {
		t.Errorf("expected 1450000, got %d", res.TotalCents)
	}
	if res.Count != 3 {
		t.Errorf("expected 3, got %d", res.Count)
	}
	if res.Formatted != "R$ 14.500,00" {
		t.Errorf("expected 'R$ 14.500,00', got %q", res.Formatted)
	}
}

type fakeAppRepo struct {
	apps []app.AppDTO
}

func (f *fakeAppRepo) ListWithFinancials(_ context.Context) ([]app.AppDTO, error) {
	return f.apps, nil
}
func (f *fakeAppRepo) GetByID(_ context.Context, id string) (*app.AppDTO, error) {
	for _, a := range f.apps {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, errors.New("app not found")
}
func (f *fakeAppRepo) Create(_ context.Context, cmd app.CreateAppCommand) (*app.AppDTO, error) {
	a := app.AppDTO{ID: "app-new", Name: cmd.Name, Hosting: cmd.Hosting, Database: cmd.Database}
	f.apps = append(f.apps, a)
	return &a, nil
}
func (f *fakeAppRepo) Update(_ context.Context, cmd app.UpdateAppCommand) error {
	for i, a := range f.apps {
		if a.ID == cmd.ID {
			f.apps[i].Name = cmd.Name
			return nil
		}
	}
	return errors.New("app not found")
}
func (f *fakeAppRepo) Delete(_ context.Context, id string) error {
	for i, a := range f.apps {
		if a.ID == id {
			f.apps = append(f.apps[:i], f.apps[i+1:]...)
			return nil
		}
	}
	return errors.New("app not found")
}

type fakeFinanceRepo struct {
	overview *app.FinancesOverviewDTO
}

func (f *fakeFinanceRepo) GetOverview(_ context.Context) (*app.FinancesOverviewDTO, error) {
	return f.overview, nil
}
func (f *fakeFinanceRepo) CreateEntry(_ context.Context, cmd app.CreateLedgerEntryCommand) (*app.LedgerEntryDTO, error) {
	return &app.LedgerEntryDTO{
		ID:          "entry-1",
		Kind:        cmd.Kind,
		AmountCents: cmd.AmountCents,
		Service:     cmd.Service,
	}, nil
}

func TestPhase2UseCases(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("ListCards and GetCardDetails with margin", func(t *testing.T) {
		c, _ := card.New("card-10", "owner-1", "cli-1", card.TypeOrcamento, "Site Clínico", money.NewBRL(320000), money.NewBRL(26000), now)
		c.Stage = card.StagePropostaAprovar

		cardRepo := &fakeCardRepo{cards: map[string]*card.Card{"card-10": c}}
		listUC := app.NewListCardsUseCase(cardRepo)
		list, err := listUC.Execute(ctx)
		if err != nil || len(list) != 1 {
			t.Fatalf("falha ao listar cartões: %v", err)
		}

		detailUC := app.NewGetCardDetailsUseCase(cardRepo)
		details, err := detailUC.Execute(ctx, "card-10")
		if err != nil {
			t.Fatalf("falha ao obter detalhes: %v", err)
		}
		if !details.IsGate {
			t.Errorf("proposta_aprovar deveria ser portão")
		}
		if details.MarginCents != (320000 - 26000) {
			t.Errorf("margem calculada incorreta: %d", details.MarginCents)
		}
		if len(details.Actions) == 0 || details.Actions[0] != "Aprovar e enviar ao cliente" {
			t.Errorf("ação esperada incorreta: %v", details.Actions)
		}
	})

	t.Run("Apps CRUD and Financial check", func(t *testing.T) {
		appRepo := &fakeAppRepo{
			apps: []app.AppDTO{
				{ID: "app-1", Name: "pedidos-app", MonthlyRevenueCents: 18000, MonthlyCostCents: 3800, NetMarginCents: 14200, IsCostly: false},
				{ID: "app-2", Name: "api-antiga", MonthlyRevenueCents: 0, MonthlyCostCents: 9600, NetMarginCents: -9600, IsCostly: true},
			},
		}
		pub := &fakePublisher{}

		listUC := app.NewListAppsUseCase(appRepo)
		apps, err := listUC.Execute(ctx)
		if err != nil || len(apps) != 2 {
			t.Fatalf("falha ao listar apps: %v", err)
		}
		if !apps[1].IsCostly {
			t.Errorf("api-antiga deveria estar marcada como custosa")
		}

		createUC := app.NewCreateAppUseCase(appRepo, pub)
		created, err := createUC.Execute(ctx, app.CreateAppCommand{Name: "novo-app", Hosting: "Render"})
		if err != nil || created.Name != "novo-app" {
			t.Fatalf("falha ao criar app: %v", err)
		}

		deleteUC := app.NewDeleteAppUseCase(appRepo, pub)
		if err := deleteUC.Execute(ctx, "app-1"); err != nil {
			t.Fatalf("falha ao deletar app: %v", err)
		}
	})

	t.Run("Finances Overview and consolidation suggestion", func(t *testing.T) {
		financeRepo := &fakeFinanceRepo{
			overview: &app.FinancesOverviewDTO{
				TotalRevenueCents: 102000,
				TotalCostCents:    30100,
				NetProfitCents:    71900,
				CostsByService: []app.CostByServiceDTO{
					{ServiceName: "Render", CostCents: 10800},
					{ServiceName: "Neon", CostCents: 8300},
					{ServiceName: "Vercel", CostCents: 4500},
					{ServiceName: "Firebase", CostCents: 5200},
					{ServiceName: "VPS", CostCents: 1300},
				},
			},
		}

		uc := app.NewGetFinancesOverviewUseCase(financeRepo)
		res, err := uc.Execute(ctx)
		if err != nil {
			t.Fatalf("falha ao buscar finanças: %v", err)
		}
		if res.NetProfitCents != 71900 {
			t.Errorf("lucro líquido incorreto: %d", res.NetProfitCents)
		}
		if res.Suggestion != "São 5 serviços diferentes. Consolidar em 2 ou 3 reduz custo e contas para vigiar." {
			t.Errorf("sugestão inesperada: %s", res.Suggestion)
		}
	})
}

