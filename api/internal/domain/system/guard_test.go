package system_test

import (
	"errors"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

func TestExecutionGuard(t *testing.T) {
	guard := system.NewExecutionGuard()

	t.Run("Blocked when system is paused even if actor is active", func(t *testing.T) {
		err := guard.Allow(true, true)
		if !errors.Is(err, system.ErrSystemPaused) {
			t.Errorf("expected ErrSystemPaused, got %v", err)
		}
	})

	t.Run("Blocked when actor is inactive even if system is running", func(t *testing.T) {
		err := guard.Allow(false, false)
		if !errors.Is(err, system.ErrActorInactive) {
			t.Errorf("expected ErrActorInactive, got %v", err)
		}
	})

	t.Run("Allowed only when system is running AND actor is active", func(t *testing.T) {
		err := guard.Allow(false, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestOwnerPrincipal(t *testing.T) {
	t.Run("Requires AAL2", func(t *testing.T) {
		_, err := system.NewOwnerPrincipal("owner-123", "aal1")
		if !errors.Is(err, system.ErrOwnerRequired) {
			t.Errorf("expected ErrOwnerRequired for aal1, got %v", err)
		}

		p, err := system.NewOwnerPrincipal("owner-123", "aal2")
		if err != nil {
			t.Fatalf("unexpected error for aal2: %v", err)
		}
		if !p.IsOwner() {
			t.Errorf("expected IsOwner to be true")
		}
	})
}

func TestSystemState_PauseResume(t *testing.T) {
	now := time.Date(2026, 10, 18, 15, 0, 0, 0, time.UTC)
	owner, _ := system.NewOwnerPrincipal("owner-1", "aal2")

	st := &system.SystemState{Paused: false}
	st.Pause(owner, "Manutenção emergencial", now)

	if !st.Paused {
		t.Errorf("expected paused to be true")
	}
	if st.Reason != "Manutenção emergencial" {
		t.Errorf("unexpected reason: %s", st.Reason)
	}

	st.Resume(owner, now.Add(time.Minute))
	if st.Paused {
		t.Errorf("expected paused to be false after resume")
	}
}
