package app_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type pipelineLeaseFake struct {
	app.DurableFactoryRepository
	valid   atomic.Bool
	claimed atomic.Bool
	saved   atomic.Int32
	job     app.DurableFactoryJob
}

func TestDurablePipelineDoesNotCallAgentWhenStagePreparationFails(t *testing.T) {
	m := newMockSprint02Repos()
	cardID := uuid.NewString()
	m.setCard(&card.Card{ID: cardID, Title: "Landing", ClientID: "client", Stage: card.StageEmExecucao})
	m.pipelines[cardID] = &app.PipelineDTO{CardID: cardID, SpendLimitCents: 5000, Status: "idle"}
	store := &pipelineLeaseFake{job: app.DurableFactoryJob{ID: uuid.NewString(), CardID: cardID, Status: "running"}}
	store.valid.Store(true)
	transactions := 0
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		transactions++
		if transactions == 4 {
			return errors.New("stage transaction failed")
		}
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error { tx.DurableFactory = store; return fn(tx) })
	})
	calls := 0
	runner := correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
		calls++
		return nil, errors.New("unexpected agent call")
	})
	build := app.NewStartBuildUseCase(uow, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner, app.PersistGeneratedFiles{})
	queue := app.NewDurableFactoryQueue(uow, app.NewDurablePipelineProcessor(build))
	if _, err := queue.RunNext(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"}); err == nil {
		t.Fatal("transaction failure hidden")
	}
	if calls != 0 {
		t.Fatal("agent called after stage preparation failed")
	}
}

func (f *pipelineLeaseFake) Claim(_ context.Context, token string) (*app.DurableFactoryJob, error) {
	if f.claimed.Swap(true) {
		return nil, nil
	}
	j := f.job
	j.LeaseToken = token
	j.LeaseExpiresAt = time.Now().Add(time.Minute)
	return &j, nil
}
func (f *pipelineLeaseFake) LockValidLease(context.Context, string, string) (bool, error) {
	return f.valid.Load(), nil
}
func (f *pipelineLeaseFake) ReplaceFiles(_ context.Context, _ string, files []app.ProjectFile) error {
	if len(files) > 0 {
		f.saved.Add(1)
	}
	return nil
}
func (f *pipelineLeaseFake) Finish(context.Context, string, string, string) (bool, error) {
	return f.valid.Load(), nil
}
func (f *pipelineLeaseFake) MarkActivePipelineFailed(context.Context, string) error { return nil }

func TestDurablePipelineProcessesAndRejectsLostLease(t *testing.T) {
	for _, loseLease := range []bool{false, true} {
		name := "success"
		if loseLease {
			name = "lease_lost_during_dev"
		}
		t.Run(name, func(t *testing.T) {
			m := newMockSprint02Repos()
			cardID := uuid.NewString()
			m.setCard(&card.Card{ID: cardID, Title: "Landing", ClientID: "client", Stage: card.StageEmExecucao})
			m.pipelines[cardID] = &app.PipelineDTO{CardID: cardID, SpendLimitCents: 5000, Status: "idle"}
			store := &pipelineLeaseFake{job: app.DurableFactoryJob{ID: uuid.NewString(), CardID: cardID, Status: "running"}}
			store.valid.Store(true)
			uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
				return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error { tx.DurableFactory = store; return fn(tx) })
			})
			sm, err := sandbox.NewLocalSandboxManager(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			base := &fixtureFileRunner{AgentRunner: agents.NewMockAgentRunner()}
			runner := correctionRunner(func(ctx context.Context, role pipeline.Role, prompt string, meta map[string]any) (*pipeline.RunnerResult, error) {
				res, err := base.RunStep(ctx, role, prompt, meta)
				if loseLease && role == pipeline.RoleDev {
					store.valid.Store(false)
				}
				return res, err
			})
			build := app.NewStartBuildUseCase(uow, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner, sm, correctionPublisher{}, app.PersistGeneratedFiles{})
			queue := app.NewDurableFactoryQueue(uow, app.NewDurablePipelineProcessor(build))
			job, err := queue.RunNext(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"})
			if loseLease {
				if err == nil {
					t.Fatal("lost lease was reported as success")
				}
				if store.saved.Load() != 0 || m.previews[cardID] != nil {
					t.Fatal("lost lease saved files or published preview")
				}
				return
			}
			if err != nil || job == nil || job.Status != "completed" {
				t.Fatalf("pipeline failed: %v", err)
			}
			if store.saved.Load() != 1 || m.getCard(cardID).Stage != card.StageEntregaAprovar {
				t.Fatal("missing files or owner approval gate")
			}
		})
	}
}
