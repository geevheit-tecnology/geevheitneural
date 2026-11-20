package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type queueRepoFake struct {
	app.DurableFactoryRepository
	job          *app.DurableFactoryJob
	known        *app.DurableFactoryJob
	finishes     int
	finishOK     bool
	pipeline     *queuePipelineFake
	expiredCalls int
}

func (f *queueRepoFake) MarkExpired(context.Context) (int64, error) {
	f.expiredCalls++
	return 1, nil
}

func TestMarkInterruptedRequiresOwnerAndDoesNotRunProcessor(t *testing.T) {
	repo := &queueRepoFake{}
	uow := &fakeUnitOfWork{repos: app.Repositories{DurableFactory: repo}}
	runs := 0
	q := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error { runs++; return nil }))
	if _, err := q.MarkInterrupted(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal1"}); !errors.Is(err, system.ErrOwnerRequired) || repo.expiredCalls != 0 {
		t.Fatal("unverified owner allowed")
	}
	count, err := q.MarkInterrupted(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"})
	if err != nil || count != 1 || runs != 0 {
		t.Fatalf("incorrect interruption handling: %v", err)
	}
}

func (f *queueRepoFake) MarkActivePipelineFailed(_ context.Context, _ string) error {
	if f.pipeline != nil && (f.pipeline.state.Status == "queued" || f.pipeline.state.Status == "running") {
		f.pipeline.state.Status = "failed"
	}
	return nil
}

func (f *queueRepoFake) Queue(_ context.Context, cardID, _ string) (*app.DurableFactoryJob, error) {
	return &app.DurableFactoryJob{ID: uuid.NewString(), CardID: cardID, Status: "queued"}, nil
}
func (f *queueRepoFake) Claim(_ context.Context, token string) (*app.DurableFactoryJob, error) {
	if f.job == nil {
		return nil, nil
	}
	j := *f.job
	j.LeaseToken = token
	j.Status = "running"
	f.known = &j
	f.job = nil
	return &j, nil
}
func (f *queueRepoFake) GetByID(_ context.Context, id string) (*app.DurableFactoryJob, error) {
	if f.known != nil && f.known.ID == id {
		copy := *f.known
		return &copy, nil
	}
	return nil, nil
}
func (f *queueRepoFake) MarkExpiredForCard(_ context.Context, cardID string) (int64, error) {
	if f.known != nil && f.known.CardID == cardID && !f.known.LeaseExpiresAt.IsZero() && f.known.LeaseExpiresAt.Before(time.Now()) {
		f.known.Status = "interrupted"
		return 1, nil
	}
	return 0, nil
}
func (f *queueRepoFake) ClaimByID(ctx context.Context, id, token string) (*app.DurableFactoryJob, error) {
	if f.job == nil || f.job.ID != id {
		return nil, nil
	}
	return f.Claim(ctx, token)
}
func (f *queueRepoFake) Finish(_ context.Context, _, _, status string) (bool, error) {
	f.finishes++
	if f.finishOK && f.known != nil {
		f.known.Status = status
	}
	return f.finishOK, nil
}

func TestDurableFactoryQueueClaimsOnlyRequestedJob(t *testing.T) {
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	id := uuid.NewString()
	repo := &queueRepoFake{job: &app.DurableFactoryJob{ID: id}, finishOK: true}
	uow := &fakeUnitOfWork{repos: app.Repositories{SystemState: &fakeSystemStateRepo{state: &system.SystemState{}}, DurableFactory: repo}}
	runs := 0
	q := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error { runs++; return nil }))
	if job, err := q.RunJob(context.Background(), owner, uuid.NewString()); !errors.Is(err, app.ErrDurableJobUnavailable) || job != nil || runs != 0 {
		t.Fatalf("wrong job executed: %v", err)
	}
	if job, err := q.RunJob(context.Background(), owner, id); err != nil || job == nil || runs != 1 {
		t.Fatalf("requested job not executed: %v", err)
	}
	if job, err := q.RunJob(context.Background(), owner, id); err != nil || job != nil || runs != 1 {
		t.Fatalf("duplicate executed: %v", err)
	}
}

func TestDurableDeliveryRunningIsNotAcknowledgedAndExpiredIsInterrupted(t *testing.T) {
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	id, cardID := uuid.NewString(), uuid.NewString()
	repo := &queueRepoFake{known: &app.DurableFactoryJob{ID: id, CardID: cardID, Status: "running", LeaseExpiresAt: time.Now().Add(time.Minute)}}
	uow := &fakeUnitOfWork{repos: app.Repositories{SystemState: &fakeSystemStateRepo{state: &system.SystemState{}}, DurableFactory: repo}}
	runs := 0
	q := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error { runs++; return nil }))
	if _, err := q.RunJob(context.Background(), owner, id); !errors.Is(err, app.ErrDurableJobPending) || runs != 0 {
		t.Fatal("running delivery acknowledged")
	}
	repo.known.LeaseExpiresAt = time.Now().Add(-time.Minute)
	if job, err := q.RunJob(context.Background(), owner, id); err != nil || job != nil || repo.known.Status != "interrupted" || runs != 0 {
		t.Fatalf("expired delivery not interrupted: %v", err)
	}
}

type processorFunc func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error

type queuePipelineFake struct {
	app.PipelineRepository
	state    *app.PipelineDTO
	writeErr error
}

func (f *queuePipelineFake) GetByCardID(context.Context, string) (*app.PipelineDTO, error) {
	return f.state, nil
}
func (f *queuePipelineFake) Upsert(_ context.Context, p *app.PipelineDTO) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.state = p
	return nil
}

func (f processorFunc) Process(ctx context.Context, owner system.OwnerPrincipal, job app.DurableFactoryJob) error {
	return f(ctx, owner, job)
}

func TestDurableQueueReportsWorkerFailureInPipeline(t *testing.T) {
	for _, initial := range []string{"queued", "running", "paused_spend_limit"} {
		t.Run(initial, func(t *testing.T) {
			id := uuid.NewString()
			repo := &queueRepoFake{job: &app.DurableFactoryJob{ID: uuid.NewString(), CardID: id}, finishOK: true}
			p := &queuePipelineFake{state: &app.PipelineDTO{CardID: id, Status: initial}}
			repo.pipeline = p
			uow := &fakeUnitOfWork{repos: app.Repositories{SystemState: &fakeSystemStateRepo{state: &system.SystemState{}}, DurableFactory: repo, Pipelines: p}}
			q := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error {
				return errors.New("provider unavailable")
			}))
			job, err := q.RunNext(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"})
			if err == nil || job == nil || job.Status != "failed" {
				t.Fatal("worker failure hidden")
			}
			want := "failed"
			if initial == "paused_spend_limit" {
				want = initial
			}
			if p.state.Status != want {
				t.Fatalf("status = %s; want %s", p.state.Status, want)
			}
		})
	}
}

func TestDurableFactoryQueue(t *testing.T) {
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	id := uuid.NewString()
	c, _ := card.New(id, owner.ID, "client", card.TypeOrcamento, "Landing", money.NewBRL(100), money.NewBRL(0), time.Now())
	c.Stage = card.StageEmExecucao
	repo := &queueRepoFake{finishOK: true}
	state := &fakeSystemStateRepo{state: &system.SystemState{}}
	uow := &fakeUnitOfWork{repos: app.Repositories{Cards: &fakeCardRepo{cards: map[string]*card.Card{id: c}}, SystemState: state, DurableFactory: repo}}
	processed := 0
	queue := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error { processed++; return nil }))
	if _, err := queue.Enqueue(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal1"}, id, uuid.NewString()); !errors.Is(err, system.ErrOwnerRequired) {
		t.Fatalf("aal1 accepted: %v", err)
	}
	state.state.Paused = true
	if _, err := queue.Enqueue(context.Background(), owner, id, uuid.NewString()); !errors.Is(err, system.ErrSystemPaused) {
		t.Fatalf("paused queue: %v", err)
	}
	state.state.Paused = false
	c.Stage = card.StagePromptAprovar
	if _, err := queue.Enqueue(context.Background(), owner, id, uuid.NewString()); err == nil {
		t.Fatal("unapproved gate enqueued")
	}
	c.Stage = card.StageEmExecucao
	job, err := queue.Enqueue(context.Background(), owner, id, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	repo.job = job
	got, err := queue.RunNext(context.Background(), owner)
	if err != nil || got.Status != "completed" || processed != 1 || repo.finishes != 1 {
		t.Fatalf("run failed: %v", err)
	}
	got, err = queue.RunNext(context.Background(), owner)
	if err != nil || got != nil || processed != 1 {
		t.Fatalf("empty queue executed: %v", err)
	}
}

func TestDurableFactoryQueueDoesNotConfirmCancelledOrExpiredWork(t *testing.T) {
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	for _, mode := range []string{"cancel", "expired", "failed"} {
		t.Run(mode, func(t *testing.T) {
			repo := &queueRepoFake{job: &app.DurableFactoryJob{ID: uuid.NewString()}, finishOK: mode != "expired"}
			uow := &fakeUnitOfWork{repos: app.Repositories{SystemState: &fakeSystemStateRepo{state: &system.SystemState{}}, DurableFactory: repo}}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			q := app.NewDurableFactoryQueue(uow, processorFunc(func(context.Context, system.OwnerPrincipal, app.DurableFactoryJob) error {
				if mode == "cancel" {
					cancel()
				}
				if mode == "failed" {
					return errors.New("provider failed")
				}
				return nil
			}))
			job, err := q.RunNext(ctx, owner)
			if err == nil {
				t.Fatal("failure hidden")
			}
			if mode == "cancel" && repo.finishes != 0 {
				t.Fatal("cancelled job confirmed")
			}
			if mode == "failed" && (job == nil || job.Status != "failed") {
				t.Fatal("failure not recorded")
			}
		})
	}
}
