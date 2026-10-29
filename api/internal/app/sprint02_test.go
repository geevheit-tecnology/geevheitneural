package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/command"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type mockSprint02Repos struct {
	mu             sync.RWMutex
	systemState    *system.SystemState
	cards          map[string]*card.Card
	pipelines      map[string]*app.PipelineDTO
	specifications map[string]*app.SpecificationDTO
	specTokens     map[string]string // tokenHash -> cardID
	playbooks      map[string]*app.PlaybookDTO
	modules        map[string]*app.ModuleDTO
	previews       map[string]*app.PreviewDTO
	metrics        []map[string]any
	cardEvents     []app.CardEventDTO
	audits         []app.AuditLogDTO
}

func newMockSprint02Repos() *mockSprint02Repos {
	return &mockSprint02Repos{
		systemState:    &system.SystemState{Paused: false, UpdatedAt: time.Now()},
		cards:          make(map[string]*card.Card),
		pipelines:      make(map[string]*app.PipelineDTO),
		specifications: make(map[string]*app.SpecificationDTO),
		specTokens:     make(map[string]string),
		playbooks:      make(map[string]*app.PlaybookDTO),
		modules:        make(map[string]*app.ModuleDTO),
		previews:       make(map[string]*app.PreviewDTO),
	}
}

type mockSprint02UOW struct {
	m *mockSprint02Repos
}

func (u *mockSprint02UOW) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	repos := app.Repositories{
		SystemState:    &mockSystemStateRepoSprint02{m: u.m},
		Cards:          &mockCardRepoSprint02{m: u.m},
		Clients:        &mockClientRepoSprint02{},
		Finances:       &mockFinanceRepoSprint02{},
		Pipelines:      &mockPipelineRepoSprint02{m: u.m},
		Specifications: &mockSpecRepoSprint02{m: u.m},
		Playbooks:      &mockPlaybookRepoSprint02{m: u.m},
		Previews:       &mockPreviewRepoSprint02{m: u.m},
		FactoryMetrics: &mockMetricsRepoSprint02{m: u.m},
		CardEvents:     &mockCardEventsRepoSprint02{m: u.m},
		Audit:          &mockAuditRepoSprint02{m: u.m},
	}
	return fn(repos)
}

type mockClientRepoSprint02 struct{}

func (r *mockClientRepoSprint02) FindByName(ctx context.Context, name string) (*app.ClientDTO, error) {
	return nil, nil
}
func (r *mockClientRepoSprint02) GetByID(ctx context.Context, id string) (*app.ClientDTO, error) {
	return &app.ClientDTO{ID: id, Name: "Cliente"}, nil
}
func (r *mockClientRepoSprint02) Create(ctx context.Context, name, contactRef string) (*app.ClientDTO, error) {
	return &app.ClientDTO{ID: uuid.NewString(), Name: name, ContactRef: contactRef}, nil
}
func (r *mockClientRepoSprint02) Delete(ctx context.Context, id string) error {
	return nil
}
func (r *mockClientRepoSprint02) Export(ctx context.Context, id string) (*app.ClientExportDTO, error) {
	return nil, nil
}

type mockMoneyReaderSprint02 struct{}

func (r *mockMoneyReaderSprint02) GetMoneyWaiting(ctx context.Context) (int64, int, error) {
	return 0, 0, nil
}

type mockFinanceRepoSprint02 struct{}

func (r *mockFinanceRepoSprint02) GetOverview(ctx context.Context) (*app.FinancesOverviewDTO, error) {
	return &app.FinancesOverviewDTO{}, nil
}
func (r *mockFinanceRepoSprint02) CreateEntry(ctx context.Context, cmd app.CreateLedgerEntryCommand) (*app.LedgerEntryDTO, error) {
	return &app.LedgerEntryDTO{}, nil
}

type mockSystemStateRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockSystemStateRepoSprint02) Get(ctx context.Context) (*system.SystemState, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	return r.m.systemState, nil
}
func (r *mockSystemStateRepoSprint02) Save(ctx context.Context, s *system.SystemState) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.systemState = s
	return nil
}

type mockCardRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockCardRepoSprint02) GetByID(ctx context.Context, id string) (*card.Card, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	c, ok := r.m.cards[id]
	if !ok {
		return nil, errors.New("cartão não encontrado")
	}
	cp := *c
	return &cp, nil
}
func (r *mockCardRepoSprint02) Update(ctx context.Context, c *card.Card) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	cp := *c
	r.m.cards[c.ID] = &cp
	return nil
}
func (r *mockCardRepoSprint02) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoSprint02) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoSprint02) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	return nil, nil
}
func (r *mockCardRepoSprint02) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	return nil, nil
}
func (r *mockCardRepoSprint02) Create(ctx context.Context, c *app.CardListItemDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.cards[c.ID] = &card.Card{
		ID:        c.ID,
		ClientID:  c.ClientID,
		Type:      card.Type(c.Type),
		Title:     c.Title,
		Price:     money.Money{Cents: c.PriceCents, Currency: "BRL"},
		EstCost:   money.Money{Cents: c.EstCostCents, Currency: "BRL"},
		Stage:     card.Stage(c.Stage),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	return nil
}
func (r *mockCardRepoSprint02) Delete(ctx context.Context, id string) error {
	return nil
}

type mockPipelineRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockPipelineRepoSprint02) Upsert(ctx context.Context, p *app.PipelineDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	// Copia o valor para evitar corrida em campos internos
	cp := *p
	r.m.pipelines[p.CardID] = &cp
	return nil
}
func (r *mockPipelineRepoSprint02) GetByCardID(ctx context.Context, cardID string) (*app.PipelineDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	p, ok := r.m.pipelines[cardID]
	if !ok {
		return nil, errors.New("pipeline não encontrado")
	}
	cp := *p
	return &cp, nil
}
func (r *mockPipelineRepoSprint02) List(ctx context.Context) ([]app.PipelineDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []app.PipelineDTO
	for _, p := range r.m.pipelines {
		list = append(list, *p)
	}
	return list, nil
}
func (r *mockPipelineRepoSprint02) AddSpend(ctx context.Context, cardID string, cents int64) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if p, ok := r.m.pipelines[cardID]; ok {
		p.SpendTotalCents += cents
	}
	return nil
}

type mockSpecRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockSpecRepoSprint02) Upsert(ctx context.Context, spec *app.SpecificationDTO, tokenHash string) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.specifications[spec.CardID] = spec
	if tokenHash != "" {
		r.m.specTokens[tokenHash] = spec.CardID
	}
	return nil
}
func (r *mockSpecRepoSprint02) GetByCardID(ctx context.Context, cardID string) (*app.SpecificationDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	s, ok := r.m.specifications[cardID]
	if !ok {
		return nil, errors.New("especificação não encontrada")
	}
	return s, nil
}
func (r *mockSpecRepoSprint02) GetByTokenHash(ctx context.Context, hash string) (*app.SpecificationDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	cardID, ok := r.m.specTokens[hash]
	if !ok {
		return nil, errors.New("token não encontrado")
	}
	return r.m.specifications[cardID], nil
}
func (r *mockSpecRepoSprint02) ApproveByClient(ctx context.Context, hash string) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	cardID, ok := r.m.specTokens[hash]
	if !ok {
		return errors.New("token não encontrado")
	}
	r.m.specifications[cardID].ClientApproved = true
	return nil
}

func (m *mockSprint02Repos) setCard(c *card.Card) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *c
	m.cards[c.ID] = &cp
}

func (m *mockSprint02Repos) getCard(id string) *card.Card {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c := m.cards[id]
	if c == nil {
		return nil
	}
	cp := *c
	return &cp
}

type mockPlaybookRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockPlaybookRepoSprint02) ListPlaybooks(ctx context.Context) ([]app.PlaybookDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []app.PlaybookDTO
	for _, p := range r.m.playbooks {
		list = append(list, *p)
	}
	return list, nil
}
func (r *mockPlaybookRepoSprint02) GetPlaybook(ctx context.Context, productType string) (*app.PlaybookDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	p, ok := r.m.playbooks[productType]
	if !ok {
		return nil, errors.New("playbook não encontrado")
	}
	return p, nil
}
func (r *mockPlaybookRepoSprint02) UpsertPlaybook(ctx context.Context, p *app.PlaybookDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.playbooks[p.ProductType] = p
	return nil
}
func (r *mockPlaybookRepoSprint02) ListModules(ctx context.Context) ([]app.ModuleDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []app.ModuleDTO
	for _, m := range r.m.modules {
		list = append(list, *m)
	}
	return list, nil
}
func (r *mockPlaybookRepoSprint02) UpsertModule(ctx context.Context, m *app.ModuleDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.modules[m.Name] = m
	return nil
}

type mockPreviewRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockPreviewRepoSprint02) Create(ctx context.Context, prev *app.PreviewDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.previews[prev.CardID] = prev
	return nil
}
func (r *mockPreviewRepoSprint02) GetByCardID(ctx context.Context, cardID string) (*app.PreviewDTO, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	p, ok := r.m.previews[cardID]
	if !ok {
		return nil, errors.New("preview não encontrado")
	}
	return p, nil
}

type mockMetricsRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockMetricsRepoSprint02) Record(ctx context.Context, cardID string, leadTimeSec, aiCost, infraCost, margin, decisionTimeSec int64, firstTimeApproved bool, defects int) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.metrics = append(r.m.metrics, map[string]any{"card_id": cardID, "margin": margin})
	return nil
}
func (r *mockMetricsRepoSprint02) GetScoreboard(ctx context.Context) (*app.FactoryScoreboardDTO, error) {
	return &app.FactoryScoreboardDTO{
		AvgLeadTimeHours:     36.0,
		TotalAICostCents:     12500,
		TotalInfraCostCents:  4500,
		TotalMarginCents:     450000,
		AvgDecisionTimeHours: 0.8,
		TotalProjects:        3,
		FirstTimeApprovalPct: 100.0,
		TotalDefects:         0,
	}, nil
}

type mockCardEventsRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockCardEventsRepoSprint02) Create(ctx context.Context, event app.CardEventDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.cardEvents = append(r.m.cardEvents, event)
	return nil
}

type mockAuditRepoSprint02 struct{ m *mockSprint02Repos }

func (r *mockAuditRepoSprint02) Record(ctx context.Context, log app.AuditLogDTO) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.audits = append(r.m.audits, log)
	return nil
}

type mockEventPublisherSprint02 struct{}

func (p *mockEventPublisherSprint02) Publish(ctx context.Context, topic string, payload any) error {
	return nil
}

// Testes de validação dos requisitos do Sprint 02
func TestSprint02_FactoryAndPipelines(t *testing.T) {
	m := newMockSprint02Repos()
	uow := &mockSprint02UOW{m: m}
	pub := &mockEventPublisherSprint02{}
	guard := system.NewExecutionGuard()
	runner := agents.NewMockAgentRunner()

	owner := system.OwnerPrincipal{ID: "owner-1", AAL: "aal2"}

	cardID := uuid.NewString()
	m.setCard(&card.Card{
		ID:        cardID,
		Title:     "Portal SaaS B2B",
		ClientID:  "client-1",
		Stage:     card.StagePromptAprovar,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	startBuildUC := app.NewStartBuildUseCase(uow, pub, guard, runner)
	getPipelineUC := app.NewGetCardPipelineUseCase(uow)

	t.Run("Fase 6: Iniciar Build transita cartão para em_execucao e inicializa pipeline", func(t *testing.T) {
		p, err := startBuildUC.Execute(context.Background(), owner, app.StartBuildCommand{
			CardID:          cardID,
			SpendLimitCents: 5000,
		})
		if err != nil {
			t.Fatalf("erro inesperado ao iniciar build: %v", err)
		}
		if p.Status != string(pipeline.StatusRunning) {
			t.Errorf("status esperado %s, obteve %s", pipeline.StatusRunning, p.Status)
		}

		c := m.getCard(cardID)
		if c.Stage != card.StageEmExecucao && c.Stage != card.StageEntregaAprovar {
			t.Errorf("estágio esperado em_execucao ou entrega_aprovar, obteve %s", c.Stage)
		}

		// Aguarda o pipeline assíncrono processar as etapas
		time.Sleep(100 * time.Millisecond)

		pAfter, err := getPipelineUC.Execute(context.Background(), cardID)
		if err != nil {
			t.Fatalf("erro ao consultar pipeline: %v", err)
		}
		if pAfter.SpendTotalCents == 0 {
			t.Errorf("spend total deve ser maior que 0 após execução dos agentes")
		}
	})

	t.Run("Fase 6 & 9: Invariante dos Portões - Agente Publicador nunca aprova direto", func(t *testing.T) {
		time.Sleep(50 * time.Millisecond)
		c := m.getCard(cardID)
		// O agente Publicador DEVE transitar para entrega_aprovar (Portão 3), NUNCA para entregue!
		if c.Stage == card.StageEntregue {
			t.Fatalf("VIOLAÇÃO CRÍTICA: agente publicador aprovou e entregou o projeto direto sem decisão do dono!")
		}
		if c.Stage != card.StageEntregaAprovar {
			t.Errorf("estágio esperado entrega_aprovar (Portão 3), obteve %s", c.Stage)
		}
	})

	t.Run("Fase 6: Bloqueio imediato quando o sistema estiver pausado", func(t *testing.T) {
		m.mu.Lock()
		m.systemState.Paused = true
		m.mu.Unlock()

		otherCardID := uuid.NewString()
		m.setCard(&card.Card{ID: otherCardID, Title: "Outro App"})

		_, err := startBuildUC.Execute(context.Background(), owner, app.StartBuildCommand{CardID: otherCardID})
		if err == nil {
			t.Fatalf("esperava erro com sistema pausado, obteve sucesso")
		}
		if !errors.Is(err, card.ErrSystemPaused) {
			t.Errorf("esperava card.ErrSystemPaused, obteve %v", err)
		}

		m.mu.Lock()
		m.systemState.Paused = false
		m.mu.Unlock()
	})

	t.Run("Fase 8: Playbooks e Módulos Catalogados", func(t *testing.T) {
		_ = app.SeedPlaybooksAndModules(context.Background(), uow)

		listPlaybooksUC := app.NewListPlaybooksUseCase(uow)
		playbooks, err := listPlaybooksUC.Execute(context.Background())
		if err != nil {
			t.Fatalf("erro ao listar playbooks: %v", err)
		}
		if len(playbooks) == 0 {
			t.Errorf("esperava playbooks catalogados, lista veio vazia")
		}

		listModulesUC := app.NewListModulesUseCase(uow)
		modules, err := listModulesUC.Execute(context.Background())
		if err != nil {
			t.Fatalf("erro ao listar módulos: %v", err)
		}
		if len(modules) == 0 {
			t.Errorf("esperava módulos catalogados, lista veio vazia")
		}
	})

	t.Run("Fase 10: Placar da Fábrica e Estimativa Inteligente de Preço", func(t *testing.T) {
		sbUC := app.NewGetFactoryScoreboardUseCase(uow)
		sb, err := sbUC.Execute(context.Background())
		if err != nil {
			t.Fatalf("erro ao consultar placar: %v", err)
		}
		if sb.AvgLeadTimeHours <= 0 {
			t.Errorf("lead time deve ser positivo, obteve %f", sb.AvgLeadTimeHours)
		}

		estUC := app.NewEstimatePriceUseCase(uow)
		est, err := estUC.Execute(context.Background(), app.EstimatePriceCommand{
			Title:       "Novo CRM",
			ProductType: "saas_b2b",
			Complexity:  "high",
		})
		if err != nil {
			t.Fatalf("erro ao estimar preço: %v", err)
		}
		if est.SuggestedPriceCents <= 0 || est.EstimatedDays <= 0 {
			t.Errorf("estimativa inválida: preço=%d dias=%d", est.SuggestedPriceCents, est.EstimatedDays)
		}
	})

	t.Run("Comando Jarvis/Jake cria demanda e inicia produção autônoma", func(t *testing.T) {
		clk := clock.SystemClock{}
		cardRepo := &mockCardRepoSprint02{m: m}
		moneyReader := &mockMoneyReaderSprint02{}
		finRepo := &mockFinanceRepoSprint02{}
		createCardUC := app.NewCreateCardUseCase(uow, pub)
		confirmSvc := command.NewConfirmationService(5 * time.Minute)
		decideGateUC := app.NewDecideGateUseCase(uow, pub, clk)
		pauseUC := app.NewPauseSystemUseCase(uow, pub, clk)
		resumeUC := app.NewResumeSystemUseCase(uow, pub, clk)

		processCmdUC := app.NewProcessCommandUseCase(cardRepo, moneyReader, finRepo, decideGateUC, pauseUC, resumeUC, confirmSvc)
		processCmdUC.WithStartBuild(startBuildUC)
		processCmdUC.WithCreateCard(createCardUC)

		// 1. Comando de criação assistido por Jarvis
		resCreate, err := processCmdUC.Execute(context.Background(), app.ProcessCommandInput{
			Owner: owner,
			Text:  "Jarvis, criar projeto Aplicativo de Logística para Transportadora Sul",
		})
		if err != nil {
			t.Fatalf("erro ao executar comando de criação com Jarvis: %v", err)
		}
		if resCreate.Intent != "create" {
			t.Errorf("esperava intent create, obteve %s", resCreate.Intent)
		}

		// 2. Comando de construção direta na fábrica autônoma
		resBuild, err := processCmdUC.Execute(context.Background(), app.ProcessCommandInput{
			Owner: owner,
			Text:  "Jake, construir Sistema de Delivery Express para Pizzaria Napoli",
		})
		if err != nil {
			t.Fatalf("erro ao executar comando de construção com Jake: %v", err)
		}
		if resBuild.Intent != "build" {
			t.Errorf("esperava intent build, obteve %s", resBuild.Intent)
		}
	})

	t.Run("Segurança de Token de Escopo Criptográfico e Pagamento de Sinal", func(t *testing.T) {
		cardID := uuid.NewString()
		m.setCard(&card.Card{ID: cardID, Title: "Projeto Seguro LGPD", Price: money.Money{Cents: 300000, Currency: "BRL"}})

		getSpecUC := app.NewGetCardSpecificationUseCase(uow)
		view, err := getSpecUC.Execute(context.Background(), cardID)
		if err != nil {
			t.Fatalf("erro ao obter especificação: %v", err)
		}
		if view.PublicToken == "" || view.PublicToken == "scope-"+cardID {
			t.Errorf("esperava token criptográfico seguro não previsível, obteve %s", view.PublicToken)
		}

		// Cliente acessa com o token criptográfico seguro
		getPublicScopeUC := app.NewGetPublicScopeUseCase(uow)
		publicView, err := getPublicScopeUC.Execute(context.Background(), view.PublicToken)
		if err != nil {
			t.Fatalf("erro ao resolver escopo com token criptográfico: %v", err)
		}
		if publicView.CardID != cardID {
			t.Errorf("card id divergente no escopo público: %s", publicView.CardID)
		}

		// Cliente aprova escopo
		approveScopeUC := app.NewApprovePublicScopeUseCase(uow, pub)
		if err := approveScopeUC.Execute(context.Background(), view.PublicToken); err != nil {
			t.Fatalf("falha ao aprovar escopo público: %v", err)
		}

		// Token errado deve ser rejeitado com erro
		if err := approveScopeUC.Execute(context.Background(), "token_falso_e_invalido"); err == nil {
			t.Fatalf("esperava erro ao tentar aprovar com token falso")
		}
	})
}
