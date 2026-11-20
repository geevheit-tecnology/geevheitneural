package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The schema and synthetic data exist only until the enclosing transaction rolls back.
func TestDurableFactoryPostgresRollback(t *testing.T) {
	url := os.Getenv("GEEV_FACTORY_ROLLBACK_DATABASE_URL")
	if os.Getenv("GEEV_FACTORY_ROLLBACK_TEST") != "1" || url == "" {
		t.Skip("isolated rollback database not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal("invalid database configuration")
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal("database connection failed")
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	migration, err := os.ReadFile("../../../../db/migrations/00012_factory_durable_storage.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.SplitN(string(migration), "-- +goose Down", 2)[0]
	if _, err := tx.Exec(ctx, up); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE authenticated"); err != nil {
		t.Fatal(err)
	}
	owner, other, client, cardID, key := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	setOwner := func(id string) {
		t.Helper()
		_, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, id))
		if err != nil {
			t.Fatal(err)
		}
	}
	setOwner(owner)
	if _, err := tx.Exec(ctx, "INSERT INTO geev.clients (id, name) VALUES ($1, 'synthetic')", client); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title) VALUES ($1, 'orcamento', $2, 'synthetic')", cardID, client); err != nil {
		t.Fatal(err)
	}
	repo := NewDurableFactoryRepository(sqlcgen.New(tx))
	queued, err := repo.Queue(ctx, cardID, key)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate, err := repo.Queue(ctx, cardID, key); err == nil || duplicate != nil {
		t.Fatal("duplicate idempotency key accepted")
	}
	if err := repo.PutFile(ctx, cardID, app.ProjectFile{Path: "index.html", Content: "synthetic"}); err != nil {
		t.Fatal(err)
	}
	setOwner(other)
	files, err := repo.ListFiles(ctx, cardID)
	if err != nil || len(files) != 0 {
		t.Fatalf("other owner read files: %v", err)
	}
	setOwner(owner)
	files, err = repo.ListFiles(ctx, cardID)
	if err != nil || len(files) != 1 || files[0].Content != "synthetic" {
		t.Fatalf("files not persisted: %v", err)
	}
	token := uuid.NewString()
	claimed, err := repo.Claim(ctx, token)
	if err != nil || claimed.ID != queued.ID || claimed.Attempts != 1 {
		t.Fatalf("claim failed: %v", err)
	}
	if second, err := repo.Claim(ctx, uuid.NewString()); err == nil || second != nil {
		t.Fatal("running job claimed twice")
	}
	finished, err := repo.Finish(ctx, queued.ID, uuid.NewString(), "completed")
	if err != nil || finished {
		t.Fatalf("wrong token accepted: %v", err)
	}
	finished, err = repo.Finish(ctx, queued.ID, token, "completed")
	if err != nil || !finished {
		t.Fatalf("finish failed: %v", err)
	}
	next, err := repo.Queue(ctx, cardID, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	expiredToken := uuid.NewString()
	if _, err := repo.Claim(ctx, expiredToken); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "UPDATE geev.factory_jobs SET lease_expires_at = clock_timestamp() - interval '1 second' WHERE id = $1", next.ID); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.Finish(ctx, next.ID, expiredToken, "completed"); err != nil || ok {
		t.Fatalf("expired lease accepted: %v", err)
	}
	count, err := repo.MarkExpired(ctx)
	if err != nil || count != 1 {
		t.Fatalf("expired lease not interrupted: %v", err)
	}
	if job, err := repo.Claim(ctx, uuid.NewString()); err == nil || job != nil {
		t.Fatal("interrupted job automatically retried")
	}
}
