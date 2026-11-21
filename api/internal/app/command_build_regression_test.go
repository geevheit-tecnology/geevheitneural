package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/clock"
	"github.com/geevheit/neuralcenter/api/internal/domain/command"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type stubReadyRunner struct{}

func (stubReadyRunner) RunStep(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
	return &pipeline.RunnerResult{Output: "ok"}, nil
}

func TestCommandApprovesBeforeReportingFactoryUnavailable(t *testing.T) {
	now := time.Now()
	owner, _ := system.NewOwnerPrincipal("owner-1", "aal2")
	c, _ := card.New("card-1", owner.ID, "client-1", card.TypeOrcamento, "Landing Flores", money.NewBRL(10000), money.NewBRL(1000), now)
	c.Stage = card.StagePromptAprovar
	cards := &fakeCardRepo{cards: map[string]*card.Card{c.ID: c}}
	approvals := &fakeApprovalRepo{}
	pub := &fakePublisher{}
	uow := &fakeUnitOfWork{repos: app.Repositories{Cards: cards, Approvals: approvals, CardEvents: &fakeCardEventRepo{}, Audit: &fakeAuditRepo{}, SystemState: &fakeSystemStateRepo{state: &system.SystemState{}}}}
	decide := app.NewDecideGateUseCase(uow, pub, clock.FrozenClock{Current: now})
	uc := app.NewProcessCommandUseCase(cards, nil, nil, decide, nil, nil, nil)
	result, err := uc.Execute(context.Background(), app.ProcessCommandInput{Owner: owner, Text: "aprovar Landing Flores"})
	if err != nil {
		t.Fatal(err)
	}
	if len(approvals.approvals) != 1 || c.Stage != card.StageEmExecucao {
		t.Fatal("approval was not recorded")
	}
	if !strings.Contains(result.Message, "fábrica não está disponível") {
		t.Fatalf("unexpected message: %s", result.Message)
	}
}

func TestBuildCommandDoesNotCreateOnAmbiguousTarget(t *testing.T) {
	now := time.Now()
	c1, _ := card.New("card-1", "owner", "client", card.TypeOrcamento, "Landing Flores", money.NewBRL(10000), money.NewBRL(1000), now)
	c2, _ := card.New("card-2", "owner", "client", card.TypeOrcamento, "Landing Loja", money.NewBRL(10000), money.NewBRL(1000), now)
	cards := &fakeCardRepo{cards: map[string]*card.Card{c1.ID: c1, c2.ID: c2}}
	uc := app.NewProcessCommandUseCase(cards, nil, nil, nil, nil, nil, nil).WithCreateCard(app.NewCreateCardUseCase(nil, nil))
	_, err := uc.Execute(context.Background(), app.ProcessCommandInput{Text: "construir Landing"})
	if !errors.Is(err, command.ErrAmbiguousTarget) {
		t.Fatalf("expected ambiguous target, got %v", err)
	}
}

func TestBuildCommandDoesNotCreateOrphanCardWhenSystemPaused(t *testing.T) {
	owner, _ := system.NewOwnerPrincipal("owner-1", "aal2")
	cards := &fakeCardRepo{cards: map[string]*card.Card{}}
	uow := &fakeUnitOfWork{repos: app.Repositories{
		Cards:       cards,
		SystemState: &fakeSystemStateRepo{state: &system.SystemState{Paused: true}},
	}}
	startBuild := app.NewStartBuildUseCase(uow, &fakePublisher{}, system.NewExecutionGuard(), stubReadyRunner{})
	uc := app.NewProcessCommandUseCase(cards, nil, nil, nil, nil, nil, nil).
		WithCreateCard(app.NewCreateCardUseCase(uow, &fakePublisher{})).
		WithStartBuild(startBuild)

	_, err := uc.Execute(context.Background(), app.ProcessCommandInput{Owner: owner, Text: "construir Landing Page Nova"})
	if !errors.Is(err, card.ErrSystemPaused) {
		t.Fatalf("expected ErrSystemPaused, got %v", err)
	}
	if len(cards.cards) != 0 {
		t.Fatalf("expected 0 orphan cards created when paused, got %d", len(cards.cards))
	}
}

func TestListCardsUseCaseUsesUnitOfWorkWhenProvided(t *testing.T) {
	now := time.Now()
	c1, _ := card.New("card-1", "owner-1", "client-1", card.TypeOrcamento, "Landing", money.NewBRL(10000), money.NewBRL(1000), now)
	txCards := &fakeCardRepo{cards: map[string]*card.Card{c1.ID: c1}}
	uow := &fakeUnitOfWork{repos: app.Repositories{Cards: txCards}}
	uc := app.NewListCardsUseCase(&fakeCardRepo{cards: map[string]*card.Card{}}, uow)

	items, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(items) != 1 || items[0].ID != c1.ID {
		t.Fatalf("expected card from UoW transaction, got %+v", items)
	}
}
