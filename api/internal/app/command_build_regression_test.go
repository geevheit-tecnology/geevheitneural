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
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

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
