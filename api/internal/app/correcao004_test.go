package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

type correctionRunner func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error)

func (f correctionRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, meta map[string]any) (*pipeline.RunnerResult, error) {
	return f(ctx, role, prompt, meta)
}

type correctionPublisher struct{ fail bool }

func (p correctionPublisher) PublishWorkspace(context.Context, string, string) (*app.PublishedPreviewDTO, error) {
	if p.fail {
		return nil, errors.New("publication failed")
	}
	return &app.PublishedPreviewDTO{PreviewURL: "http://localhost/preview"}, nil
}

type correctionSandbox struct {
	app.SandboxManager
	fail bool
}

func (s correctionSandbox) WriteFiles(id string, files []app.ProjectFile) error {
	if s.fail {
		return errors.New("disk write failed")
	}
	return s.SandboxManager.WriteFiles(id, files)
}

type correctionContextKey struct{}

func TestCorrection004_ArtifactsAndFailures(t *testing.T) {
	for _, mode := range []string{"success", "empty_files", "write_error", "publish_error", "nil_result", "provider_error", "qa_rejected"} {
		t.Run(mode, func(t *testing.T) {
			m := newMockSprint02Repos()
			id := uuid.NewString()
			m.setCard(&card.Card{ID: id, Title: "Landing", ClientID: "client-1", Stage: card.StageEmExecucao})
			sm, err := sandbox.NewLocalSandboxManager(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			base := &fixtureFileRunner{AgentRunner: agents.NewMockAgentRunner()}
			runner := correctionRunner(func(ctx context.Context, role pipeline.Role, prompt string, meta map[string]any) (*pipeline.RunnerResult, error) {
				if ctx.Value(correctionContextKey{}) != "claims" {
					return nil, errors.New("lost request context")
				}
				if role == pipeline.RoleDev {
					if meta["architecture"] == nil {
						return nil, errors.New("missing architecture")
					}
					if mode == "nil_result" {
						return nil, nil
					}
					if mode == "provider_error" {
						return nil, errors.New("provider rejected request")
					}
				}
				if role == pipeline.RoleQA || role == pipeline.RoleSeguranca {
					files, ok := meta["files"].([]app.ProjectFile)
					if !ok || len(files) != 1 {
						return nil, errors.New("missing code for review")
					}
				}
				result, err := base.RunStep(ctx, role, prompt, meta)
				if role == pipeline.RoleDev && mode == "empty_files" {
					delete(result.Artifacts, "files")
				}
				if role == pipeline.RoleQA && mode == "qa_rejected" {
					result.Artifacts["ready_for_review"] = false
				}
				return result, err
			})
			uc := app.NewStartBuildUseCase(&mockSprint02UOW{m: m}, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner, correctionSandbox{sm, mode == "write_error"}, correctionPublisher{mode == "publish_error"})
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), correctionContextKey{}, "claims"))
			_, err = uc.Execute(ctx, system.OwnerPrincipal{ID: "owner-1", AAL: "aal2"}, app.StartBuildCommand{CardID: id})
			cancel()
			if err != nil {
				t.Fatal(err)
			}
			want := string(pipeline.StatusFailed)
			if mode == "success" {
				want = string(pipeline.StatusCompleted)
			}
			deadline := time.Now().Add(3 * time.Second)
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
			if mode == "success" {
				content, err := os.ReadFile(filepath.Join(sm.GetWorkspacePath(id), "index.html"))
				if err != nil || string(content) != "<!doctype html><title>Fixture</title>" {
					t.Fatalf("artifact missing: %v", err)
				}
			} else {
				m.mu.RLock()
				preview := m.previews[id]
				m.mu.RUnlock()
				if preview != nil {
					t.Fatal("failure produced ready preview")
				}
				if m.getCard(id).Stage == card.StageEntregaAprovar {
					t.Fatal("failure advanced delivery gate")
				}
			}
		})
	}
}

func TestCorrection004_MissingProvider(t *testing.T) {
	uc := app.NewStartBuildUseCase(nil, nil, system.NewExecutionGuard(), nil)
	if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{}, app.StartBuildCommand{CardID: uuid.NewString()}); err == nil {
		t.Fatal("unconfigured provider accepted")
	}
}

func TestCorrection004_DuplicateBuild(t *testing.T) {
	m := newMockSprint02Repos()
	id := uuid.NewString()
	m.setCard(&card.Card{ID: id, Title: "Landing"})
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	runner := correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
		close(started)
		<-release
		close(done)
		return nil, errors.New("stop fixture")
	})
	uc := app.NewStartBuildUseCase(&mockSprint02UOW{m: m}, &mockEventPublisherSprint02{}, system.NewExecutionGuard(), runner)
	owner := system.OwnerPrincipal{ID: "owner-1", AAL: "aal2"}
	if _, err := uc.Execute(context.Background(), owner, app.StartBuildCommand{CardID: id}); err != nil {
		t.Fatal(err)
	}
	<-started
	_, err := uc.Execute(context.Background(), owner, app.StartBuildCommand{CardID: id})
	close(release)
	<-done
	if err == nil {
		t.Fatal("duplicate build accepted")
	}
}
