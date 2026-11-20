package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type factorySettingsFake struct {
	selection *app.FactoryToolSelection
	err       error
}

type activeFactoryAgents struct {
	app.AgentRepository
	inactive string
}

func (a activeFactoryAgents) List(context.Context) ([]app.AgentDTO, error) {
	roles := []string{"requisitos", "arquiteto", "dev", "qa", "seguranca", "publicador", "suporte"}
	result := make([]app.AgentDTO, 0, len(roles))
	for _, role := range roles {
		result = append(result, app.AgentDTO{Role: role, Active: role != a.inactive})
	}
	return result, nil
}

func (s *factorySettingsFake) Get(context.Context) (*app.FactoryToolSelection, error) {
	return s.selection, s.err
}
func (s *factorySettingsFake) Save(_ context.Context, _ string, value app.FactoryToolSelection) error {
	if s.err != nil {
		return s.err
	}
	s.selection = &value
	return nil
}

func TestFactoryToolsSelectionControlsRealRunner(t *testing.T) {
	ctx := context.Background()
	m := newMockSprint02Repos()
	settings := &factorySettingsFake{}
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error {
			tx.FactoryTools = settings
			tx.Agents = activeFactoryAgents{}
			return fn(tx)
		})
	})
	var calls []string
	makeRunner := func(name string) pipeline.AgentRunner {
		return correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
			calls = append(calls, name)
			return &pipeline.RunnerResult{Output: "artifact"}, nil
		})
	}
	uc := app.NewFactoryTools(uow, system.NewExecutionGuard(), map[string]pipeline.AgentRunner{"gemini": makeRunner("gemini"), "groq": makeRunner("groq")}, "gemini")
	owner := system.OwnerPrincipal{ID: "owner-1", AAL: "aal2"}
	items, err := uc.List(ctx)
	if err != nil || !items[0].Active || items[1].Active {
		t.Fatalf("legacy selection: %v %v", items, err)
	}
	if err := uc.Select(ctx, owner, "groq", true, false); !errors.Is(err, app.ErrFactoryToolSelection) {
		t.Fatalf("missing free tier confirmation: %v", err)
	}
	if err := uc.Select(ctx, system.OwnerPrincipal{ID: "owner-1", AAL: "aal1"}, "groq", true, true); err == nil {
		t.Fatal("AAL1 selected a provider")
	}
	if err := uc.Select(ctx, owner, "unknown", true, true); !errors.Is(err, app.ErrFactoryToolSelection) {
		t.Fatal(err)
	}
	if err := uc.Select(ctx, owner, "groq", true, true); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.RunStep(ctx, pipeline.RoleDev, "build", nil); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "groq" {
		t.Fatalf("wrong runner: %v", calls)
	}
	// Uma nova instância deve ler a escolha persistida, sem depender do estado do processo.
	restarted := app.NewFactoryTools(uow, system.NewExecutionGuard(), map[string]pipeline.AgentRunner{"groq": makeRunner("groq")}, "")
	items, err = restarted.List(ctx)
	if err != nil || !items[1].Active || items[0].Active {
		t.Fatalf("stored selection lost: %v %v", items, err)
	}
	if err := uc.Select(ctx, owner, "groq", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.RunStep(ctx, pipeline.RoleDev, "build", nil); !errors.Is(err, system.ErrActorInactive) {
		t.Fatalf("disabled runner executed: %v", err)
	}
	build := app.NewStartBuildUseCase(nil, nil, system.NewExecutionGuard(), uc)
	if _, err := build.Execute(ctx, owner, app.StartBuildCommand{CardID: "not-created"}); !errors.Is(err, system.ErrActorInactive) {
		t.Fatalf("build changed card before checking the selected tool: %v", err)
	}
	m.systemState.Paused = true
	if err := uc.Select(ctx, owner, "groq", true, true); !errors.Is(err, system.ErrSystemPaused) {
		t.Fatalf("activation during pause: %v", err)
	}
	settings.selection.Active = true
	if _, err := uc.RunStep(ctx, pipeline.RoleDev, "build", nil); !errors.Is(err, system.ErrSystemPaused) {
		t.Fatalf("runner ignored pause: %v", err)
	}
	settings.err = errors.New("database unavailable")
	if _, err := uc.RunStep(ctx, pipeline.RoleDev, "build", nil); err == nil {
		t.Fatal("runner ignored settings read error")
	}
	if len(calls) != 1 {
		t.Fatalf("unexpected external calls: %v", calls)
	}
	if len(m.audits) != 2 {
		t.Fatalf("expected activation and deactivation audit: %v", m.audits)
	}
}

func TestFactoryToolsMissingKeyAndProviderErrors(t *testing.T) {
	m := newMockSprint02Repos()
	settings := &factorySettingsFake{}
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error {
			tx.FactoryTools = settings
			tx.Agents = activeFactoryAgents{}
			return fn(tx)
		})
	})
	providerErr := errors.New("quota exhausted")
	calls := 0
	uc := app.NewFactoryTools(uow, system.NewExecutionGuard(), map[string]pipeline.AgentRunner{
		"groq": correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
			calls++
			return nil, providerErr
		}),
	}, "")
	ctx := context.Background()
	owner := system.OwnerPrincipal{ID: "owner-1", AAL: "aal2"}
	if err := uc.Select(ctx, owner, "gemini", true, true); !errors.Is(err, app.ErrFactoryToolUnavailable) {
		t.Fatalf("missing key accepted: %v", err)
	}
	if err := uc.Select(ctx, owner, "groq", true, true); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.RunStep(ctx, pipeline.RoleDev, "build", nil); !errors.Is(err, providerErr) {
		t.Fatalf("provider error hidden: %v", err)
	}
	if calls != 1 {
		t.Fatalf("unexpected retries/fallback: %d", calls)
	}
}

func TestFactoryToolsRejectsInactiveAgent(t *testing.T) {
	m := newMockSprint02Repos()
	settings := &factorySettingsFake{selection: &app.FactoryToolSelection{Provider: "gemini", Active: true}}
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error {
			tx.FactoryTools = settings
			tx.Agents = activeFactoryAgents{inactive: "dev"}
			return fn(tx)
		})
	})
	calls := 0
	runner := correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
		calls++
		return &pipeline.RunnerResult{}, nil
	})
	uc := app.NewFactoryTools(uow, system.NewExecutionGuard(), map[string]pipeline.AgentRunner{"gemini": runner}, "")
	if err := uc.Ready(context.Background()); !errors.Is(err, system.ErrActorInactive) {
		t.Fatalf("build accepted disabled agent: %v", err)
	}
	if _, err := uc.RunStep(context.Background(), pipeline.RoleDev, "", nil); !errors.Is(err, system.ErrActorInactive) {
		t.Fatalf("disabled agent executed: %v", err)
	}
	if calls != 0 {
		t.Fatalf("runner called %d times", calls)
	}
}

func TestFactoryToolsStateReadFailsClosed(t *testing.T) {
	for _, mode := range []string{"no_settings", "no_state_repo", "no_state", "state_error", "transaction_error"} {
		t.Run(mode, func(t *testing.T) {
			m := newMockSprint02Repos()
			uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
				if mode == "transaction_error" {
					return errors.New("database unavailable")
				}
				return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error {
					tx.FactoryTools = &factorySettingsFake{selection: &app.FactoryToolSelection{Provider: "gemini", Active: true}}
					switch mode {
					case "no_settings":
						tx.FactoryTools = nil
					case "no_state_repo":
						tx.SystemState = nil
					case "no_state":
						tx.SystemState = correctionSystemState{}
					case "state_error":
						tx.SystemState = correctionSystemState{err: errors.New("state unavailable")}
					}
					return fn(tx)
				})
			})
			calls := 0
			runner := correctionRunner(func(context.Context, pipeline.Role, string, map[string]any) (*pipeline.RunnerResult, error) {
				calls++
				return nil, nil
			})
			uc := app.NewFactoryTools(uow, system.NewExecutionGuard(), map[string]pipeline.AgentRunner{"gemini": runner}, "gemini")
			if _, err := uc.RunStep(context.Background(), pipeline.RoleDev, "build", nil); err == nil {
				t.Fatal("missing state accepted")
			}
			if calls != 0 {
				t.Fatal("agent executed without verified state")
			}
		})
	}
}
