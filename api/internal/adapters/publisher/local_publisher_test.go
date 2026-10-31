package publisher_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/publisher"
	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func TestLocalPublisher_PublishWorkspace(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "test-geev-publisher-*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempBase)

	sm, err := sandbox.NewLocalSandboxManager(tempBase)
	if err != nil {
		t.Fatalf("erro ao instanciar LocalSandboxManager: %v", err)
	}

	cardID := "card-pub-123"
	_, err = sm.InitWorkspace(cardID)
	if err != nil {
		t.Fatalf("falha ao inicializar workspace: %v", err)
	}

	// Escrever arquivo de teste
	err = sm.WriteFiles(cardID, []app.ProjectFile{
		{
			Path:    "index.html",
			Content: "<html><body>App do Dono</body></html>",
		},
	})
	if err != nil {
		t.Fatalf("falha ao escrever arquivos: %v", err)
	}

	pub := publisher.NewLocalPublisher("https://api.geevheit.internal", sm)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := pub.PublishWorkspace(ctx, cardID, "feat: primeiro commit do robô")
	if err != nil {
		t.Fatalf("erro ao publicar workspace: %v", err)
	}

	if res.CommitHash == "" {
		t.Errorf("esperava commit hash não vazio")
	}
	if res.PreviewURL != "https://api.geevheit.internal/v1/cards/card-pub-123/preview" {
		t.Errorf("esperava PreviewURL correta, obteve: %s", res.PreviewURL)
	}
}

// Reproduz o ambiente sem Git que a imagem distroless entregava.
func TestLocalPublisher_MissingGitDoesNotPublish(t *testing.T) {
	sm, err := sandbox.NewLocalSandboxManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sm.InitWorkspace("missing-git"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	result, err := publisher.NewLocalPublisher("http://localhost:8085", sm).PublishWorkspace(context.Background(), "missing-git", "test")
	if !errors.Is(err, publisher.ErrGitInitFailed) {
		t.Fatalf("expected missing git initialization failure, got %v", err)
	}
	if result != nil {
		t.Fatal("publication reported a result without Git")
	}
}
