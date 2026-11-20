package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type pendingJobsFake struct {
	app.DurableFactoryRepository
	ids   []string
	reads int
}

func (f *pendingJobsFake) ListQueuedIDs(_ context.Context, _ string, _ int32) ([]string, error) {
	f.reads++
	return f.ids, nil
}

func TestDurableReconcileRetriesPageAfterPublishFailure(t *testing.T) {
	ids := []string{uuid.NewString(), uuid.NewString()}
	repo := &pendingJobsFake{ids: ids}
	uow := &fakeUnitOfWork{repos: app.Repositories{DurableFactory: repo}}
	sends := []string{}
	fail := true
	signal := signalFunc(func(_ context.Context, _ string, id string) error {
		sends = append(sends, id)
		if id == ids[1] && fail {
			return errors.New("unavailable")
		}
		return nil
	})
	r := app.NewDurableQueueReconciler(uow, signal)
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	if _, err := r.ReconcilePage(context.Background(), owner, "", "", 2); err == nil || repo.reads != 0 {
		t.Fatal("missing identity allowed")
	}
	if _, err := r.ReconcilePage(context.Background(), owner, "oidc", "invalid", 2); err == nil || repo.reads != 0 {
		t.Fatal("invalid cursor allowed")
	}
	if _, err := r.ReconcilePage(context.Background(), owner, "oidc", "", 101); err == nil || repo.reads != 0 {
		t.Fatal("invalid limit allowed")
	}
	if _, err := r.ReconcilePage(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal1"}, "oidc", "", 2); !errors.Is(err, system.ErrOwnerRequired) {
		t.Fatal("aal1 allowed")
	}
	if next, err := r.ReconcilePage(context.Background(), owner, "oidc", "", 2); !errors.Is(err, app.ErrDurableDispatchPending) || next != "" {
		t.Fatalf("failure hidden: %v", err)
	}
	fail = false
	if next, err := r.ReconcilePage(context.Background(), owner, "oidc", "", 2); err != nil || next != ids[1] {
		t.Fatalf("retry failed: %v", err)
	}
	if len(sends) != 4 || sends[0] != sends[2] || sends[1] != sends[3] {
		t.Fatal("retry did not repeat same page")
	}
}
