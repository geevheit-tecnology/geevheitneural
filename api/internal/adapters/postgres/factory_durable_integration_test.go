package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
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
	if _, err := tx.Exec(ctx, "INSERT INTO geev.card_pipelines (card_id, status) VALUES ($1, 'queued')", cardID); err != nil {
		t.Fatal(err)
	}
	setOwner(other)
	if err := repo.MarkActivePipelineFailed(ctx, cardID); err != nil {
		t.Fatal(err)
	}
	setOwner(owner)
	var pipelineStatus string
	if err := tx.QueryRow(ctx, "SELECT status FROM geev.card_pipelines WHERE card_id = $1", cardID).Scan(&pipelineStatus); err != nil || pipelineStatus != "queued" {
		t.Fatalf("other owner changed pipeline: %v", err)
	}
	if err := repo.MarkActivePipelineFailed(ctx, cardID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, "SELECT status FROM geev.card_pipelines WHERE card_id = $1", cardID).Scan(&pipelineStatus); err != nil || pipelineStatus != "failed" {
		t.Fatalf("failed pipeline not visible: %v", err)
	}
	if _, err := tx.Exec(ctx, "UPDATE geev.card_pipelines SET status = 'paused_spend_limit' WHERE card_id = $1", cardID); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkActivePipelineFailed(ctx, cardID); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, "SELECT status FROM geev.card_pipelines WHERE card_id = $1", cardID).Scan(&pipelineStatus); err != nil || pipelineStatus != "paused_spend_limit" {
		t.Fatalf("paused pipeline overwritten: %v", err)
	}
	queued, err := repo.Queue(ctx, cardID, key)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate, err := repo.Queue(ctx, cardID, key); err != nil || duplicate == nil || duplicate.ID != queued.ID {
		t.Fatal("duplicate request did not return the original job")
	}
	otherCard := uuid.NewString()
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title) VALUES ($1, 'orcamento', $2, 'synthetic other')", otherCard, client); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Queue(ctx, otherCard, key); err == nil {
		t.Fatal("idempotency key reused for another card")
	}
	if err := repo.PutFile(ctx, cardID, app.ProjectFile{Path: "index.html", Content: "synthetic"}); err != nil {
		t.Fatal(err)
	}
	setOwner(other)
	if ids, err := repo.ListQueuedIDs(ctx, uuid.Nil.String(), 100); err != nil || len(ids) != 0 {
		t.Fatalf("other owner listed jobs: %v", err)
	}
	files, err := repo.ListFiles(ctx, cardID)
	if err != nil || len(files) != 0 {
		t.Fatalf("other owner read files: %v", err)
	}
	setOwner(owner)
	if ids, err := repo.ListQueuedIDs(ctx, uuid.Nil.String(), 100); err != nil || len(ids) != 1 || ids[0] != queued.ID {
		t.Fatalf("queued job not listed: %v", err)
	}
	files, err = repo.ListFiles(ctx, cardID)
	if err != nil || len(files) != 1 || files[0].Content != "synthetic" {
		t.Fatalf("files not persisted: %v", err)
	}
	if err := repo.ReplaceFiles(ctx, cardID, []app.ProjectFile{{Path: "index.html", Content: "v2"}, {Path: "style.css", Content: "body{}"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SAVEPOINT before_invalid_replacement"); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceFiles(ctx, cardID, []app.ProjectFile{{Path: "../bad", Content: "invalid"}}); err == nil {
		t.Fatal("invalid path accepted")
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT before_invalid_replacement"); err != nil {
		t.Fatal(err)
	}
	files, err = repo.ListFiles(ctx, cardID)
	if err != nil || len(files) != 2 || files[0].Content != "v2" {
		t.Fatalf("replacement was not rolled back: %v", err)
	}
	token := uuid.NewString()
	setOwner(other)
	if job, err := repo.ClaimByID(ctx, queued.ID, uuid.NewString()); err != nil || job != nil {
		t.Fatalf("another owner claimed job: %v", err)
	}
	setOwner(owner)
	if job, err := repo.ClaimByID(ctx, uuid.NewString(), uuid.NewString()); err != nil || job != nil {
		t.Fatalf("wrong ID claimed a job: %v", err)
	}
	claimed, err := repo.ClaimByID(ctx, queued.ID, token)
	if err != nil || claimed.ID != queued.ID || claimed.Attempts != 1 {
		t.Fatalf("claim failed: %v", err)
	}
	if ids, err := repo.ListQueuedIDs(ctx, uuid.Nil.String(), 100); err != nil || len(ids) != 0 {
		t.Fatalf("running job listed for resignal: %v", err)
	}
	if valid, err := repo.LockValidLease(ctx, queued.ID, token); err != nil || !valid {
		t.Fatalf("valid lease rejected: %v", err)
	}
	if valid, err := repo.LockValidLease(ctx, queued.ID, uuid.NewString()); err != nil || valid {
		t.Fatalf("invalid lease accepted: %v", err)
	}
	if second, err := repo.Claim(ctx, uuid.NewString()); err != nil || second != nil {
		t.Fatal("running job claimed twice")
	}
	if second, err := repo.ClaimByID(ctx, queued.ID, uuid.NewString()); err != nil || second != nil {
		t.Fatal("duplicate delivery claimed running job")
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
	if _, err := tx.Exec(ctx, "UPDATE geev.card_pipelines SET status = 'running' WHERE card_id = $1", cardID); err != nil {
		t.Fatal(err)
	}
	setOwner(other)
	if count, err := repo.MarkExpiredForCard(ctx, cardID); err != nil || count != 0 {
		t.Fatalf("other owner marked interrupted: %v", err)
	}
	setOwner(owner)
	if count, err := repo.MarkExpiredForCard(ctx, otherCard); err != nil || count != 0 {
		t.Fatalf("other card marked interrupted: %v", err)
	}
	if valid, err := repo.LockValidLease(ctx, next.ID, expiredToken); err != nil || valid {
		t.Fatalf("expired lease accepted: %v", err)
	}
	if ok, err := repo.Finish(ctx, next.ID, expiredToken, "completed"); err != nil || ok {
		t.Fatalf("expired lease accepted: %v", err)
	}
	count, err := repo.MarkExpiredForCard(ctx, cardID)
	if err != nil || count != 1 {
		t.Fatalf("expired lease not interrupted: %v", err)
	}
	if err := tx.QueryRow(ctx, "SELECT status FROM geev.card_pipelines WHERE card_id = $1", cardID).Scan(&pipelineStatus); err != nil || pipelineStatus != "interrupted" {
		t.Fatalf("interrupted pipeline not visible: %v", err)
	}
	if count, err := repo.MarkExpired(ctx); err != nil || count != 0 {
		t.Fatalf("interruption not idempotent: %v", err)
	}
	if job, err := repo.Claim(ctx, uuid.NewString()); err != nil || job != nil {
		t.Fatal("interrupted job automatically retried")
	}
	if _, err := repo.Queue(ctx, otherCard, uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	thirdCard := uuid.NewString()
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title) VALUES ($1, 'orcamento', $2, 'synthetic third')", thirdCard, client); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Queue(ctx, thirdCard, uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	first, err := repo.ListQueuedIDs(ctx, uuid.Nil.String(), 1)
	if err != nil || len(first) != 1 {
		t.Fatalf("first page: %v", err)
	}
	second, err := repo.ListQueuedIDs(ctx, first[0], 1)
	if err != nil || len(second) != 1 || first[0] == second[0] {
		t.Fatalf("second page: %v", err)
	}
	last, err := repo.ListQueuedIDs(ctx, second[0], 1)
	if err != nil || len(last) != 0 {
		t.Fatalf("pagination did not finish: %v", err)
	}
	if _, err := repo.ClaimByID(ctx, first[0], uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	queue := app.NewDurableFactoryQueue(rollbackPipelineUOW{tx}, neverProcessor{})
	principal := system.OwnerPrincipal{ID: owner, AAL: "aal2"}
	if _, err := queue.RunJob(ctx, principal, first[0]); !errors.Is(err, app.ErrDurableJobPending) {
		t.Fatalf("running delivery acknowledged: %v", err)
	}
	if _, err := tx.Exec(ctx, "UPDATE geev.factory_jobs SET lease_expires_at = clock_timestamp() - interval '1 second' WHERE id = $1", first[0]); err != nil {
		t.Fatal(err)
	}
	if job, err := queue.RunJob(ctx, principal, first[0]); err != nil || job != nil {
		t.Fatalf("expired delivery not interrupted: %v", err)
	}
	if existing, err := repo.GetByID(ctx, first[0]); err != nil || existing == nil || existing.Status != "interrupted" {
		t.Fatalf("interruption not persisted: %v", err)
	}
	verifyDurablePipelineWithRealDatabase(t, ctx, tx, owner, client)
}
