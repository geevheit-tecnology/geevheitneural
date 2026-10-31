package postgres_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/adapters/events"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	localpub "github.com/geevheit/neuralcenter/api/internal/adapters/publisher"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"path/filepath"
)

// TestFlow_AppsAndFinances valida o fluxo completo da Fase 2 (F2-10):
// Cadastrar app, lançar custos e receitas, verificar margens e ver o lucro consolidado em milissegundos.
func TestFlow_AppsAndFinances(t *testing.T) {
	dbURL := os.Getenv("GEEV_FACTORY_TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("isolated test database not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Supabase local offline no ambiente do teste, pulando teste de banco real: %v", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Supabase ping falhou, pulando teste de integração real: %v", err)
		return
	}

	q := sqlcgen.New(pool)
	appRepo := postgres.NewAppRepository(q)
	financeRepo := postgres.NewFinanceRepository(q)

	appName := "test-app-" + uuid.New().String()[:8]

	// 1. Cadastra app
	createdApp, err := appRepo.Create(ctx, app.CreateAppCommand{
		Name:     appName,
		Hosting:  "Render",
		Database: "Supabase",
	})
	if err != nil {
		t.Fatalf("falha ao criar app no banco: %v", err)
	}

	defer func() {
		_ = appRepo.Delete(ctx, createdApp.ID)
	}()

	// 2. Lança receita
	_, err = financeRepo.CreateEntry(ctx, app.CreateLedgerEntryCommand{
		AppID:       &createdApp.ID,
		Kind:        "receita",
		AmountCents: 25000, // R$ 250,00
		Currency:    "BRL",
		Service:     "Cliente Teste",
		Description: "Mensalidade",
	})
	if err != nil {
		t.Fatalf("falha ao lançar receita: %v", err)
	}

	// 3. Lança custo
	_, err = financeRepo.CreateEntry(ctx, app.CreateLedgerEntryCommand{
		AppID:       &createdApp.ID,
		Kind:        "custo",
		AmountCents: 4500, // R$ 45,00
		Currency:    "BRL",
		Service:     "Render",
		Description: "Instância web",
	})
	if err != nil {
		t.Fatalf("falha ao lançar custo: %v", err)
	}

	// 4. Verifica cálculo do app
	apps, err := appRepo.ListWithFinancials(ctx)
	if err != nil {
		t.Fatalf("falha ao listar apps com financeiro: %v", err)
	}

	var found *app.AppDTO
	for i := range apps {
		if apps[i].ID == createdApp.ID {
			found = &apps[i]
			break
		}
	}

	if found == nil {
		t.Fatalf("app %s não encontrado na listagem", createdApp.ID)
	}

	if found.MonthlyRevenueCents != 25000 {
		t.Errorf("esperava receita 25000, obteve %d", found.MonthlyRevenueCents)
	}
	if found.MonthlyCostCents != 4500 {
		t.Errorf("esperava custo 4500, obteve %d", found.MonthlyCostCents)
	}
	if found.NetMarginCents != 20500 {
		t.Errorf("esperava margem 20500, obteve %d", found.NetMarginCents)
	}
	if found.IsCostly {
		t.Errorf("app dá lucro (250 - 45 = 205), não deveria estar IsCostly=true")
	}

	// 5. Verifica visão consolidada de finanças
	overviewUC := app.NewGetFinancesOverviewUseCase(financeRepo)
	overview, err := overviewUC.Execute(ctx)
	if err != nil {
		t.Fatalf("falha ao obter visão geral de finanças: %v", err)
	}

	if overview.TotalRevenueCents < 25000 {
		t.Errorf("total receita consolidada inesperada: %d", overview.TotalRevenueCents)
	}
	if overview.TotalCostCents < 4500 {
		t.Errorf("total custo consolidado inesperado: %d", overview.TotalCostCents)
	}
}

func TestFactoryAuthenticatedOwner(t *testing.T) {
	url := os.Getenv("GEEV_FACTORY_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("isolated factory database not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	owner := uuid.NewString()
	ctx = postgres.ContextWithClaims(ctx, fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, owner))
	uow := postgres.NewUnitOfWork(pool)
	id := uuid.NewString()
	err = uow.RunInTx(ctx, func(tx app.Repositories) error {
		cl, err := tx.Clients.Create(ctx, "Factory isolated test", "")
		if err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Cards.Create(ctx, &app.CardListItemDTO{ID: id, ClientID: cl.ID, Title: "Factory isolated test", Type: "orcamento", Currency: "BRL", Stage: "em_execucao", CreatedAt: now, UpdatedAt: now}); err != nil {
			return err
		}
		if err := tx.Pipelines.Upsert(ctx, &app.PipelineDTO{CardID: id, Stage: "em_execucao", ActiveRole: "requisitos", Status: "running"}); err != nil {
			return err
		}
		if err := tx.Specifications.Upsert(ctx, &app.SpecificationDTO{CardID: id, UserStories: []map[string]any{{"title": "Test"}}}, uuid.NewString()); err != nil {
			return err
		}
		if err := tx.CardEvents.Create(ctx, app.CardEventDTO{CardID: id, ActorType: "owner", Kind: "test"}); err != nil {
			return err
		}
		if err := tx.Audit.Record(ctx, app.AuditLogDTO{Action: "factory_test", ActorType: "owner", TargetType: "card", TargetID: &id}); err != nil {
			return err
		}
		if err := tx.ProjectMemory.Upsert(ctx, &app.ProjectMemoryDTO{CardID: id, ClientID: cl.ID, Key: "test", Category: "nota", Content: "Test"}); err != nil {
			return err
		}
		if err := tx.Previews.Create(ctx, &app.PreviewDTO{CardID: id, PreviewURL: "http://localhost/preview", Status: "ready", ExpiresAt: now.Add(time.Hour)}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	other := postgres.ContextWithClaims(context.Background(), fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, uuid.NewString()))
	if err := uow.RunInTx(other, func(tx app.Repositories) error { _, err := tx.Pipelines.GetByCardID(other, id); return err }); err == nil {
		t.Fatal("other owner can read pipeline")
	}
}

// Opt-in: calls the configured real provider and writes only to an isolated test DB.
func TestFactoryLiveGeneration(t *testing.T) {
	if os.Getenv("GEEV_FACTORY_LIVE_TEST") != "1" {
		t.Skip("live generation opt-in required")
	}
	url := os.Getenv("GEEV_FACTORY_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("isolated database required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	owner := uuid.NewString()
	ctx = postgres.ContextWithClaims(ctx, fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, owner))
	uow := postgres.NewUnitOfWork(pool)
	id := uuid.NewString()
	err = uow.RunInTx(ctx, func(tx app.Repositories) error {
		cl, err := tx.Clients.Create(ctx, "Isolated live generation", "")
		if err != nil {
			return err
		}
		now := time.Now()
		return tx.Cards.Create(ctx, &app.CardListItemDTO{ID: id, ClientID: cl.ID, Title: "Uma página HTML estática com título Olá e botão que incrementa um contador, em um único index.html, sem dependências externas", Type: "orcamento", Currency: "BRL", Stage: "em_execucao", CreatedAt: now, UpdatedAt: now})
	})
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("", "geevheit-live-artifacts-")
	if err != nil {
		t.Fatal(err)
	}
	sm, err := sandbox.NewLocalSandboxManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{APIKey: os.Getenv("GEMINI_API_KEY"), Model: os.Getenv("GEMINI_MODEL"), HTTPClient: &http.Client{Timeout: 180 * time.Second}})
	uc := app.NewStartBuildUseCase(uow, events.NewInMemoryPublisher(), system.NewExecutionGuard(), runner, sm, localpub.NewLocalPublisher("http://localhost:8085", sm))
	if _, err := uc.Execute(ctx, system.OwnerPrincipal{ID: owner, AAL: "aal2"}, app.StartBuildCommand{CardID: id, SpendLimitCents: 100}); err != nil {
		t.Fatal(err)
	}
	previous := ""
	for {
		var p *app.PipelineDTO
		err = uow.RunInTx(ctx, func(tx app.Repositories) error { var err error; p, err = tx.Pipelines.GetByCardID(ctx, id); return err })
		if err != nil {
			t.Fatal(err)
		}
		if p.ActiveRole != previous {
			t.Logf("role=%s status=%s", p.ActiveRole, p.Status)
			previous = p.ActiveRole
		}
		if p.Status == "failed" {
			t.Fatalf("generation failed: %s; workspace %s", p.CurrentTask, sm.GetWorkspacePath(id))
		}
		if p.Status == "completed" {
			content, err := os.ReadFile(filepath.Join(sm.GetWorkspacePath(id), "index.html"))
			if err != nil {
				t.Fatal(err)
			}
			if len(content) < 100 {
				t.Fatal("generated HTML is unexpectedly empty")
			}
			var preview *app.PreviewDTO
			if err := uow.RunInTx(ctx, func(tx app.Repositories) error {
				var err error
				preview, err = tx.Previews.GetByCardID(ctx, id)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			if preview.Status != "ready" {
				t.Fatalf("preview not ready: %s", preview.Status)
			}
			t.Logf("real Gemini pipeline completed; bytes=%d workspace=%s", len(content), sm.GetWorkspacePath(id))
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
}
