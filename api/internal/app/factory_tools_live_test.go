package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	localpub "github.com/geevheit/neuralcenter/api/internal/adapters/publisher"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

// Opt-in: IA e arquivos/Git reais; repositórios em memória, sem tocar dados de clientes.
func TestFactoryToolsLiveLandingPage(t *testing.T) {
	if os.Getenv("GEEV_FACTORY_LIVE_LANDING") != "1" {
		t.Skip("live provider opt-in required")
	}
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Fatal("Gemini key required")
	}
	m := newMockSprint02Repos()
	id := uuid.NewString()
	m.setCard(&card.Card{ID: id, ClientID: "synthetic-client", Title: "Landing page de demonstração para uma agência de veículos fictícia, com catálogo de três veículos claramente ilustrativos, filtro JavaScript e botão de contato sem número inventado. Não usar depoimentos ou preços reais. Entregar index.html funcional."})
	settings := &factorySettingsFake{selection: &app.FactoryToolSelection{Provider: "gemini", Active: true}}
	uow := correctionUOW(func(ctx context.Context, fn func(app.Repositories) error) error {
		return (&mockSprint02UOW{m: m}).RunInTx(ctx, func(tx app.Repositories) error { tx.FactoryTools = settings; return fn(tx) })
	})
	dir, err := os.MkdirTemp("", "geevheit-release-landing-")
	if err != nil {
		t.Fatal(err)
	}
	sm, err := sandbox.NewLocalSandboxManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	guard := system.NewExecutionGuard()
	tools := app.NewFactoryTools(uow, guard, map[string]pipeline.AgentRunner{"gemini": agents.NewGeminiAgentRunner(agents.GeminiConfig{APIKey: os.Getenv("GEMINI_API_KEY"), Model: os.Getenv("GEMINI_MODEL")})}, "")
	uc := app.NewStartBuildUseCase(uow, &mockEventPublisherSprint02{}, guard, tools, sm, localpub.NewLocalPublisher("http://localhost:8085", sm))
	if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "synthetic-owner", AAL: "aal2"}, app.StartBuildCommand{CardID: id}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(8 * time.Minute)
	previous := ""
	for time.Now().Before(deadline) {
		m.mu.RLock()
		p := *m.pipelines[id]
		m.mu.RUnlock()
		if p.ActiveRole != previous {
			t.Logf("real provider role=%s status=%s", p.ActiveRole, p.Status)
			previous = p.ActiveRole
		}
		if p.Status == "failed" {
			t.Fatalf("generation failed: %s", p.CurrentTask)
		}
		if p.Status == "completed" {
			path := filepath.Join(sm.GetWorkspacePath(id), "index.html")
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(content) < 1000 || !strings.Contains(strings.ToLower(string(content)), "<html") {
				t.Fatal("missing generated HTML")
			}
			if m.getCard(id).Stage != card.StageEntregaAprovar {
				t.Fatal("delivery approval gate was not preserved")
			}
			t.Logf("real Gemini + local files/Git completed; bytes=%d artifact=%s; repositories in memory", len(content), path)
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal("real generation exceeded the validation timeout")
}
