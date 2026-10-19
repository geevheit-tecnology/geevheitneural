package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type mockPhase5Repos struct {
	systemState *system.SystemState
	agents      map[string]*app.AgentDTO
	cards       map[string]*card.Card
	enriched    map[string]*app.CardListItemDTO
	events      []app.CardEventDTO
	audits      []app.AuditLogDTO
}

func newMockPhase5Repos() *mockPhase5Repos {
	return &mockPhase5Repos{
		systemState: &system.SystemState{Paused: false, UpdatedAt: time.Now()},
		agents:      make(map[string]*app.AgentDTO),
		cards:       make(map[string]*card.Card),
		enriched:    make(map[string]*app.CardListItemDTO),
	}
}

type mockSystemStateRepoPhase5 struct{ m *mockPhase5Repos }

func (r *mockSystemStateRepoPhase5) Get(ctx context.Context) (*system.SystemState, error) {
	return r.m.systemState, nil
}
func (r *mockSystemStateRepoPhase5) Save(ctx context.Context, s *system.SystemState) error {
	r.m.systemState = s
	return nil
}

type mockAgentRepoPhase5 struct{ m *mockPhase5Repos }

func (r *mockAgentRepoPhase5) GetByTokenHash(ctx context.Context, hash string) (*system.AgentPrincipal, error) {
	for _, a := range r.m.agents {
		if a.ID == hash {
			return &system.AgentPrincipal{
				ID:              a.ID,
				Role:            a.Role,
				Active:          a.Active,
				SpendLimitCents: a.SpendLimitCents,
			}, nil
		}
	}
	return nil, errors.New("agente não encontrado")
}
func (r *mockAgentRepoPhase5) List(ctx context.Context) ([]app.AgentDTO, error) {
	var list []app.AgentDTO
	for _, a := range r.m.agents {
		list = append(list, *a)
	}
	return list, nil
}
func (r *mockAgentRepoPhase5) Create(ctx context.Context, role, scope string, spendLimitCents int64, tokenHash string) (*app.AgentDTO, error) {
	a := &app.AgentDTO{
		ID:              uuid.New().String(),
		Role:            role,
		Active:          false, // Desligado por padrão
		SpendLimitCents: spendLimitCents,
		Scope:           scope,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	r.m.agents[a.ID] = a
	return a, nil
}
func (r *mockAgentRepoPhase5) AddSpend(ctx context.Context, id string, cents int64) error {
	a, ok := r.m.agents[id]
	if !ok {
		return errors.New("agente não encontrado")
	}
	a.SpendCurrentCents += cents
	return nil
}
func (r *mockAgentRepoPhase5) Toggle(ctx context.Context, id string, active bool) error {
	a, ok := r.m.agents[id]
	if !ok {
		return errors.New("agente não encontrado")
	}
	a.Active = active
	return nil
}

type mockUoWPhase5 struct{ m *mockPhase5Repos }

func (u *mockUoWPhase5) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	repos := app.Repositories{
		SystemState: &mockSystemStateRepoPhase5{m: u.m},
		Agents:      &mockAgentRepoPhase5{m: u.m},
		Cards:       &mockCardRepoPhase4{m: &mockPhase4Repos{cards: u.m.cards, enriched: u.m.enriched, events: u.m.events, audits: u.m.audits}},
		CardEvents:  &mockCardEventRepoPhase4{m: &mockPhase4Repos{cards: u.m.cards, enriched: u.m.enriched, events: u.m.events, audits: u.m.audits}},
		Audit:       &mockAuditRepoPhase4{m: &mockPhase4Repos{cards: u.m.cards, enriched: u.m.enriched, events: u.m.events, audits: u.m.audits}},
	}
	return fn(repos)
}

func TestPhase5_AgentContractsAndExecutionGuard(t *testing.T) {
	ctx := context.Background()
	mockState := newMockPhase5Repos()
	uow := &mockUoWPhase5{m: mockState}
	pub := &mockPublisherPhase4{}

	agentUC := app.NewAgentUseCases(uow, pub)

	// Cadastra um cartão em execução
	cardID := "card-exec-01"
	mockState.cards[cardID] = &card.Card{
		ID:        cardID,
		Stage:     card.StageEmExecucao,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockState.enriched[cardID] = &app.CardListItemDTO{
		ID:    cardID,
		Title: "Implementar autenticação JWT",
		Stage: string(card.StageEmExecucao),
	}

	inactiveAgent := system.AgentPrincipal{
		ID:              "agent-01",
		Role:            "backend_builder",
		Active:          false, // Desligado por padrão!
		SpendLimitCents: 10000,
	}

	activeAgent := system.AgentPrincipal{
		ID:              "agent-02",
		Role:            "backend_builder",
		Active:          true,
		SpendLimitCents: 10000,
	}

	mockState.agents["agent-01"] = &app.AgentDTO{
		ID:              "agent-01",
		Role:            "backend_builder",
		Active:          false,
		SpendLimitCents: 10000,
	}
	mockState.agents["agent-02"] = &app.AgentDTO{
		ID:              "agent-02",
		Role:            "backend_builder",
		Active:          true,
		SpendLimitCents: 10000,
	}

	owner, _ := system.NewOwnerPrincipal("dono-id-01", "aal2")

	t.Run("ExecutionGuard blocks inactive agent with ErrExecutionLocked (423)", func(t *testing.T) {
		_, err := agentUC.ListAvailableCards(ctx, inactiveAgent)
		if !errors.Is(err, app.ErrExecutionLocked) {
			t.Fatalf("expected ErrExecutionLocked, got: %v", err)
		}

		err = agentUC.RecordProgress(ctx, inactiveAgent, cardID, app.AgentProgressCommand{
			Message: "Tentativa de escrita com agente desligado",
		})
		if !errors.Is(err, app.ErrExecutionLocked) {
			t.Fatalf("expected ErrExecutionLocked on RecordProgress, got: %v", err)
		}

		err = agentUC.SubmitDelivery(ctx, inactiveAgent, cardID, app.AgentSubmitDeliveryCommand{
			Summary: "Entrega não autorizada",
		})
		if !errors.Is(err, app.ErrExecutionLocked) {
			t.Fatalf("expected ErrExecutionLocked on SubmitDelivery, got: %v", err)
		}
	})

	t.Run("ExecutionGuard blocks active agent when system is paused with ErrExecutionLocked (423)", func(t *testing.T) {
		// Dono pausa o sistema
		mockState.systemState.Pause(owner, "Pausa de emergência", time.Now())

		_, err := agentUC.ListAvailableCards(ctx, activeAgent)
		if !errors.Is(err, app.ErrExecutionLocked) {
			t.Fatalf("expected ErrExecutionLocked when system is paused, got: %v", err)
		}

		err = agentUC.RecordProgress(ctx, activeAgent, cardID, app.AgentProgressCommand{
			Message: "Escrita com sistema pausado",
		})
		if !errors.Is(err, app.ErrExecutionLocked) {
			t.Fatalf("expected ErrExecutionLocked when system is paused, got: %v", err)
		}

		// Dono retoma o sistema
		mockState.systemState.Resume(owner, time.Now())
	})

	t.Run("Active agent executes progress and submits delivery to Gate 3 (entrega_aprovar)", func(t *testing.T) {
		// 1. Listagem de cartões
		cards, err := agentUC.ListAvailableCards(ctx, activeAgent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected 1 card in execution, got %d", len(cards))
		}

		// 2. Registro de progresso técnico
		err = agentUC.RecordProgress(ctx, activeAgent, cardID, app.AgentProgressCommand{
			Message: "Testes unitários concluídos com 98% de cobertura.",
			Cost:    150, // R$ 1,50 de custo de API de modelo
		})
		if err != nil {
			t.Fatalf("failed to record progress: %v", err)
		}

		// 3. Submissão da entrega pelo agente
		err = agentUC.SubmitDelivery(ctx, activeAgent, cardID, app.AgentSubmitDeliveryCommand{
			Summary: "Módulo de auth pronto com testes passando.",
			PRURL:   "https://github.com/geevheit/neuralcenter/pull/1",
		})
		if err != nil {
			t.Fatalf("failed to submit delivery: %v", err)
		}

		// INVARIANTE: O cartão DEVE estar em entrega_aprovar, NUNCA em entregue!
		c := mockState.cards[cardID]
		if c.Stage != card.StageEntregaAprovar {
			t.Fatalf("CRITICAL INVARIANT VIOLATION: agent submission must transition to %s, got: %s",
				card.StageEntregaAprovar, c.Stage)
		}
	})
}
