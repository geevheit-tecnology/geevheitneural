package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type starterFunc func(context.Context, system.OwnerPrincipal, app.StartBuildCommand) (*app.PipelineDTO, error)

func (f starterFunc) DurableQueueEnabled() bool { return true }

func (f starterFunc) Execute(ctx context.Context, owner system.OwnerPrincipal, cmd app.StartBuildCommand) (*app.PipelineDTO, error) {
	return f(ctx, owner, cmd)
}

type signalFunc func(context.Context, string, string) error

func (f signalFunc) Publish(ctx context.Context, token, id string) error { return f(ctx, token, id) }

func TestDurableDispatchReportsFailureAfterSavingAndCanRetry(t *testing.T) {
	starts, sends := 0, 0
	starter := starterFunc(func(_ context.Context, _ system.OwnerPrincipal, cmd app.StartBuildCommand) (*app.PipelineDTO, error) {
		starts++
		if cmd.RequestKey != "same-key" {
			t.Fatal("idempotency key changed")
		}
		return &app.PipelineDTO{JobID: "job-id", Status: "queued"}, nil
	})
	signal := signalFunc(func(_ context.Context, token, id string) error {
		sends++
		if token != "oidc" || id != "job-id" {
			t.Fatal("incorrect signal")
		}
		if sends == 1 {
			return errors.New("unavailable")
		}
		return nil
	})
	d := app.NewDurableBuildDispatcher(starter, signal)
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	cmd := app.StartBuildCommand{RequestKey: "same-key"}
	if _, err := d.Dispatch(context.Background(), owner, cmd, ""); err == nil || starts != 0 {
		t.Fatal("missing OIDC started job")
	}
	result, err := d.Dispatch(context.Background(), owner, cmd, "oidc")
	if !errors.Is(err, app.ErrDurableDispatchPending) || result == nil || result.Status != "queued" {
		t.Fatalf("dispatch failure hidden: %v", err)
	}
	if _, err := d.Dispatch(context.Background(), owner, cmd, "oidc"); err != nil || starts != 2 || sends != 2 {
		t.Fatalf("retry failed: %v", err)
	}
}
