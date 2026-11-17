package postgres

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Opt-in contra banco configurado, sempre em transação com ROLLBACK.
// Não altera a preferência do dono: usa UUID de teste e valida isolamento por RLS.
func TestFactoryToolsPostgresRollback(t *testing.T) {
	if os.Getenv("GEEV_FACTORY_ROLLBACK_TEST") != "1" {
		t.Skip("rollback integration opt-in required")
	}
	databaseURL := os.Getenv("GEEV_FACTORY_ROLLBACK_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("explicit rollback database URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatal("invalid database configuration")
	}
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal("database connection failed")
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal("database transaction unavailable")
	}
	defer tx.Rollback(context.Background())
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE authenticated"); err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	setClaims := func(id, aal string) {
		claims, _ := json.Marshal(map[string]string{"sub": id, "aal": aal})
		if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claims)); err != nil {
			t.Fatal(err)
		}
	}
	setClaims(owner, "aal2")
	repo := &FactoryToolSettingsRepository{q: sqlcgen.New(tx)}
	ctx = ContextWithClaims(ctx, `{"sub":"`+owner+`","aal":"aal2"}`)
	for _, selection := range []app.FactoryToolSelection{{Provider: "gemini", Active: true}, {Provider: "groq", Active: true}, {Provider: "groq", Active: false}} {
		if err := repo.Save(ctx, owner, selection); err != nil {
			t.Fatal(err)
		}
		stored, err := repo.Get(ctx)
		if err != nil || stored == nil || *stored != selection {
			t.Fatalf("selection roundtrip failed: %v", err)
		}
	}
	if err := NewAuditRepository(sqlcgen.New(tx)).Record(ctx, app.AuditLogDTO{OwnerID: owner, ActorType: "owner", ActorID: &owner, Action: "factory.tool_selected", TargetType: "connector", Meta: map[string]any{"provider": "groq", "active": false}}); err != nil {
		t.Fatal(err)
	}
	setClaims(uuid.NewString(), "aal2")
	if stored, err := repo.Get(ctx); err != nil || stored != nil {
		t.Fatalf("another owner could read the selection: %v", err)
	}
	setClaims(owner, "aal1")
	if stored, err := repo.Get(ctx); err != nil || stored != nil {
		t.Fatalf("AAL1 could read the selection: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	t.Log("real PostgreSQL: activate/switch/disable, audit JSONB and RLS passed; ROLLBACK confirmed")
}
