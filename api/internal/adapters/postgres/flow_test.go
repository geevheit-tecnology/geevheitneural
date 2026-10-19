package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestFlow_AppsAndFinances valida o fluxo completo da Fase 2 (F2-10):
// Cadastrar app, lançar custos e receitas, verificar margens e ver o lucro consolidado em milissegundos.
func TestFlow_AppsAndFinances(t *testing.T) {
	dbURL := "postgres://postgres:postgres@127.0.0.1:54332/postgres?sslmode=disable"
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
