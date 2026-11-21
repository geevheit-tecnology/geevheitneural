package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

func TestGetSystemStateRequiresOwnerAndReportsUnavailable(t *testing.T) {
	repo := &fakeSystemStateRepo{state: &system.SystemState{Paused: true}}
	uc := app.NewGetSystemStateUseCase(&fakeUnitOfWork{repos: app.Repositories{SystemState: repo}})
	if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal1"}); !errors.Is(err, system.ErrOwnerRequired) {
		t.Fatal("aal1 allowed")
	}
	owner := system.OwnerPrincipal{ID: "owner", AAL: "aal2"}
	if paused, err := uc.Execute(context.Background(), owner); err != nil || !paused {
		t.Fatal("pause hidden")
	}
	repo.state = nil
	if _, err := uc.Execute(context.Background(), owner); err == nil {
		t.Fatal("missing state reported as resumed")
	}
}
