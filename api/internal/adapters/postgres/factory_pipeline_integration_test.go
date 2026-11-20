package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Savepoints permitem executar o pipeline com repositórios reais e reverter tudo ao fim do teste externo.
type rollbackPipelineUOW struct{ tx pgx.Tx }

func (u rollbackPipelineUOW) RunInTx(ctx context.Context, fn func(app.Repositories) error) error {
	tx, err := u.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	q := sqlcgen.New(tx)
	repos := app.Repositories{
		DurableFactory: NewDurableFactoryRepository(q), Cards: NewCardRepository(q),
		Pipelines: NewPipelineRepository(q), Specifications: NewSpecificationRepository(q),
		Previews: NewPreviewRepository(q), ProjectMemory: NewProjectMemoryRepository(q),
		FactoryMetrics: NewFactoryMetricsRepository(q), CardEvents: NewCardEventRepository(q),
		Audit: NewAuditRepository(q), SystemState: NewSystemStateRepository(q),
	}
	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type integrationEvents struct{}

func (integrationEvents) Publish(context.Context, string, any) error { return nil }

type integrationRunner struct{ base *agents.MockAgentRunner }

func (r integrationRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, meta map[string]any) (*pipeline.RunnerResult, error) {
	res, err := r.base.RunStep(ctx, role, prompt, meta)
	if err == nil && role == pipeline.RoleDev {
		res.Artifacts["files"] = []any{map[string]any{"path": "index.html", "content": "<!doctype html><h1>Landing sintética</h1>"}}
	}
	return res, err
}

type integrationPublisher struct{}

type neverProcessor struct{}

func (neverProcessor) Process(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error {
	return fmt.Errorf("processor must not run for a repeated delivery")
}

func (integrationPublisher) PublishWorkspace(context.Context, string, string) (*app.PublishedPreviewDTO, error) {
	return &app.PublishedPreviewDTO{PreviewURL: "https://example.invalid/synthetic"}, nil
}

func verifyDurablePipelineWithRealDatabase(t *testing.T, ctx context.Context, tx pgx.Tx, owner, client string) {
	t.Helper()
	cardID := uuid.NewString()
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title, stage) VALUES ($1, 'orcamento', $2, 'Landing sintética', 'em_execucao')", cardID, client); err != nil {
		t.Fatal(err)
	}
	ctx = ContextWithClaims(ctx, fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, owner))
	uow := rollbackPipelineUOW{tx: tx}
	sm, err := sandbox.NewLocalSandboxManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	build := app.NewStartBuildUseCase(uow, integrationEvents{}, system.NewExecutionGuard(), integrationRunner{base: agents.NewMockAgentRunner()}, sm, integrationPublisher{}, app.PersistGeneratedFiles{}, app.EnqueueDurable{})
	principal := system.OwnerPrincipal{ID: owner, AAL: "aal2"}
	queued, err := build.Execute(ctx, principal, app.StartBuildCommand{CardID: cardID, RequestKey: uuid.NewString(), SpendLimitCents: 5000})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	queue := app.NewDurableFactoryQueue(uow, app.NewDurablePipelineProcessor(build))
	job, err := queue.RunJob(ctx, principal, queued.JobID)
	if err != nil || job == nil || job.Status != "completed" {
		t.Fatalf("pipeline run: job=%+v err=%v", job, err)
	}
	if duplicate, err := queue.RunJob(ctx, principal, queued.JobID); err != nil || duplicate != nil {
		t.Fatalf("duplicate delivery reran job: job=%+v err=%v", duplicate, err)
	}
	var stage, status, content string
	if err := tx.QueryRow(ctx, "SELECT c.stage::text, p.status, f.content FROM geev.cards c JOIN geev.card_pipelines p ON p.card_id=c.id JOIN geev.factory_workspace_files f ON f.card_id=c.id WHERE c.id=$1 AND f.path='index.html'", cardID).Scan(&stage, &status, &content); err != nil {
		t.Fatal(err)
	}
	if stage != "entrega_aprovar" || status != "completed" || content != "<!doctype html><h1>Landing sintética</h1>" {
		t.Fatalf("unexpected persisted result: stage=%s status=%s content=%q", stage, status, content)
	}
}
