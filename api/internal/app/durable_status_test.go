package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
)

type statusRepoFake struct {
	app.DurableFactoryRepository
	calls  int
	cardID string
	err    error
}

func (f *statusRepoFake) MarkExpiredForCard(_ context.Context, cardID string) (int64, error) {
	f.calls++
	f.cardID = cardID
	return 0, f.err
}

func TestGetPipelineChecksExpiredJobsOnlyWhenEnabled(t *testing.T) {
	id := uuid.NewString()
	repo := &statusRepoFake{}
	pipeline := &queuePipelineFake{state: &app.PipelineDTO{CardID: id, Status: "running"}}
	uow := &fakeUnitOfWork{repos: app.Repositories{DurableFactory: repo, Pipelines: pipeline}}
	if _, err := app.NewGetCardPipelineUseCase(uow).Execute(context.Background(), id); err != nil || repo.calls != 0 {
		t.Fatal("default enabled expiration check")
	}
	uc := app.NewGetCardPipelineUseCase(uow, true)
	if _, err := uc.Execute(context.Background(), id); err != nil || repo.calls != 1 || repo.cardID != id {
		t.Fatalf("check not scoped to card: %v", err)
	}
	if _, err := uc.Execute(context.Background(), "invalid"); err == nil || repo.calls != 1 {
		t.Fatal("invalid id reached repository")
	}
	repo.err = errors.New("database unavailable")
	if _, err := uc.Execute(context.Background(), id); err == nil {
		t.Fatal("database failure hidden")
	}
}
