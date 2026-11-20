package app_test

import (
	"context"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type enqueueFake struct {
	app.DurableFactoryRepository
	job *app.DurableFactoryJob
}

func (f *enqueueFake) Queue(_ context.Context, cardID, _ string) (*app.DurableFactoryJob, error) {
	if f.job == nil {
		f.job = &app.DurableFactoryJob{ID: uuid.NewString(), CardID: cardID, Status: "queued"}
		j := *f.job
		j.Created = true
		return &j, nil
	}
	j := *f.job
	return &j, nil
}

func TestStartBuildQueuesWithoutRunningAndRetryKeepsState(t *testing.T) {
	m := newMockSprint02Repos()
	id, key := uuid.NewString(), uuid.NewString()
	m.setCard(&card.Card{ID: id, Title: "Landing", Stage: card.StageEmExecucao})
	store := &enqueueFake{}
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error { tx.DurableFactory = store; return fn(tx) })
	})
	calls := 0
	runner := correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
		calls++
		return nil, nil
	})
	uc := app.NewStartBuildUseCase(uow, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner, app.EnqueueDurable{})
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	if _, err := uc.Execute(context.Background(), owner, app.StartBuildCommand{CardID: id}); err == nil {
		t.Fatal("missing key accepted")
	}
	result, err := uc.Execute(context.Background(), owner, app.StartBuildCommand{CardID: id, RequestKey: key})
	if err != nil || result.Status != "queued" || calls != 0 {
		t.Fatalf("not queued safely: %v", err)
	}
	m.mu.Lock()
	m.pipelines[id].Status = "running"
	m.pipelines[id].SpendTotalCents = 42
	m.mu.Unlock()
	result, err = uc.Execute(context.Background(), owner, app.StartBuildCommand{CardID: id, RequestKey: key})
	if err != nil || result.Status != "running" || result.SpendTotalCents != 42 || calls != 0 {
		t.Fatalf("retry reset state: %v", err)
	}
}
