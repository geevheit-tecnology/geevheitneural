package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type durableReadFake struct {
	app.DurableFactoryRepository
	err   error
	calls int
}

func (f *durableReadFake) ListFiles(context.Context, string) ([]app.ProjectFile, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return []app.ProjectFile{{Path: "index.html", Content: "conteúdo"}}, nil
}

func TestGetDurableWorkspaceRequiresOwnerAndPropagatesFailure(t *testing.T) {
	repo := &durableReadFake{}
	uow := correctionUOW(func(_ context.Context, fn func(app.Repositories) error) error {
		return fn(app.Repositories{DurableFactory: repo})
	})
	uc := app.NewGetDurableWorkspaceUseCase(uow)
	if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal1"}, "card"); !errors.Is(err, system.ErrOwnerRequired) {
		t.Fatalf("aal1 accepted: %v", err)
	}
	if repo.calls != 0 {
		t.Fatal("repository accessed without AAL2")
	}
	files, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"}, "card")
	if err != nil || len(files) != 1 {
		t.Fatalf("read failed: %v", err)
	}
	repo.err = errors.New("database unavailable")
	if _, err := uc.Execute(context.Background(), system.OwnerPrincipal{ID: "owner", AAL: "aal2"}, "card"); !errors.Is(err, repo.err) {
		t.Fatalf("read error hidden: %v", err)
	}
}
