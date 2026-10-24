package sandbox_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/sandbox"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func TestLocalSandboxManager_LifecycleAndSecurity(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "test-geev-sandbox-*")
	if err != nil {
		t.Fatalf("falha ao criar temp dir: %v", err)
	}
	defer os.RemoveAll(tempBase)

	sm, err := sandbox.NewLocalSandboxManager(tempBase)
	if err != nil {
		t.Fatalf("erro ao instanciar LocalSandboxManager: %v", err)
	}

	cardID := "card-test-123"

	// 1. Inicializar workspace
	wsPath, err := sm.InitWorkspace(cardID)
	if err != nil {
		t.Fatalf("falha ao inicializar workspace: %v", err)
	}
	if !strings.Contains(wsPath, cardID) {
		t.Errorf("caminho esperado continha cardID, obteve %s", wsPath)
	}

	// 2. Escrever arquivos válidos
	files := []app.ProjectFile{
		{
			Path:    "index.html",
			Content: "<!DOCTYPE html><html><body><h1>Teste Sandbox</h1></body></html>",
		},
		{
			Path:    "css/style.css",
			Content: "body { background: #000; }",
		},
	}

	if err := sm.WriteFiles(cardID, files); err != nil {
		t.Fatalf("erro ao escrever arquivos válidos: %v", err)
	}

	// 3. Ler arquivos
	readFiles, err := sm.ReadFiles(cardID)
	if err != nil {
		t.Fatalf("erro ao ler arquivos: %v", err)
	}
	if len(readFiles) != 2 {
		t.Fatalf("esperava 2 arquivos lidos, obteve %d", len(readFiles))
	}

	// 4. Teste de tentativa de Path Traversal
	maliciousFiles := []app.ProjectFile{
		{
			Path:    "../../etc/passwd",
			Content: "malicious payload",
		},
	}
	err = sm.WriteFiles(cardID, maliciousFiles)
	if err == nil {
		t.Fatalf("esperava erro de tentativa de path traversal, mas passou com sucesso!")
	}

	// 5. Teste de validação/execução de comando
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	execRes, err := sm.RunValidation(ctx, cardID, "ls", "-la")
	if err != nil {
		t.Fatalf("erro ao rodar comando de validação: %v", err)
	}
	if execRes.ExitCode != 0 {
		t.Errorf("esperava exitCode 0, obteve %d", execRes.ExitCode)
	}
	if !strings.Contains(execRes.Output, "index.html") {
		t.Errorf("esperava encontrar index.html na saída, obteve:\n%s", execRes.Output)
	}

	// 6. Cleanup
	if err := sm.CleanupWorkspace(cardID); err != nil {
		t.Fatalf("erro ao limpar workspace: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempBase, cardID)); !os.IsNotExist(err) {
		t.Errorf("workspace ainda existe após cleanup")
	}
}

func TestLocalSandboxManager_InvalidCardID(t *testing.T) {
	sm, _ := sandbox.NewLocalSandboxManager("")
	invalidCardID := "../bad-id/etc"

	_, err := sm.InitWorkspace(invalidCardID)
	if err == nil {
		t.Fatalf("esperava erro de ID de cartão inválido")
	}
}
