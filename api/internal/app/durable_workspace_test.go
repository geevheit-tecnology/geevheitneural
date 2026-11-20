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

type durableFilesFake struct {
	app.DurableFactoryRepository
	calls atomic.Int32
	fail  bool
}

func (f *durableFilesFake) ReplaceFiles(_ context.Context, _ string, files []app.ProjectFile) error {
	f.calls.Add(1)
	if f.fail {
		return errors.New("durable storage unavailable")
	}
	if len(files) == 0 {
		return errors.New("no files")
	}
	return nil
}

func TestPipelineDurableFilePersistence(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "storage_error"}[fail], func(t *testing.T) {
			m := newMockSprint02Repos()
			id := uuid.NewString()
			m.setCard(&card.Card{ID: id, Title: "Landing", ClientID: "client", Stage: card.StageEmExecucao})
			store := &durableFilesFake{fail: fail}
			uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
				return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error { tx.DurableFactory = store; return fn(tx) })
			})
			sm, err := sandbox.NewLocalSandboxManager(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			runner := &fixtureFileRunner{AgentRunner: agents.NewMockAgentRunner()}
			uc := app.NewStartBuildUseCase(uow, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner, sm, correctionPublisher{}, app.PersistGeneratedFiles{})
			if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"}, app.StartBuildCommand{CardID: id}); err != nil {
				t.Fatal(err)
			}
			want := string(pipeline.StatusCompleted)
			if fail {
				want = string(pipeline.StatusFailed)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				m.mu.RLock()
				p := m.pipelines[id]
				status := ""
				if p != nil {
					status = p.Status
				}
				m.mu.RUnlock()
				if status == want {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("expected %s, got %s", want, status)
				}
				time.Sleep(time.Millisecond)
			}
			if store.calls.Load() != 1 {
				t.Fatalf("writes: %d", store.calls.Load())
			}
			if fail {
				m.mu.RLock()
				preview := m.previews[id]
				m.mu.RUnlock()
				if preview != nil {
					t.Fatal("preview created after storage failure")
				}
			}
		})
	}
}
