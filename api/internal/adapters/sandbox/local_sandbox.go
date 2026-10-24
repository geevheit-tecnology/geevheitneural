package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/app"
)

var (
	ErrInvalidCardID        = errors.New("identificador de cartão inválido para workspace")
	ErrPathTraversalAttempt = errors.New("tentativa de path traversal detectada no salvamento do arquivo")
	ErrWorkspaceNotFound    = errors.New("workspace do cartão não foi inicializado")
)

var validCardIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// LocalSandboxManager gerencia ambientes de compilação e teste efêmeros no disco local
type LocalSandboxManager struct {
	baseDir string
}

// NewLocalSandboxManager cria uma instância do gerenciador de sandbox local
func NewLocalSandboxManager(baseDir string) (*LocalSandboxManager, error) {
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "geevheit-sandboxes")
	}

	cleanBaseDir := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanBaseDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório base de sandbox: %w", err)
	}

	return &LocalSandboxManager{
		baseDir: cleanBaseDir,
	}, nil
}

func (s *LocalSandboxManager) validateCardID(cardID string) error {
	if !validCardIDPattern.MatchString(cardID) {
		return fmt.Errorf("%w: %s", ErrInvalidCardID, cardID)
	}
	return nil
}

func (s *LocalSandboxManager) GetWorkspacePath(cardID string) string {
	if err := s.validateCardID(cardID); err != nil {
		return ""
	}
	return filepath.Join(s.baseDir, cardID)
}

// InitWorkspace cria e prepara o diretório efêmero exclusivo do cartão
func (s *LocalSandboxManager) InitWorkspace(cardID string) (string, error) {
	if err := s.validateCardID(cardID); err != nil {
		return "", err
	}

	workspaceDir := s.GetWorkspacePath(cardID)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return "", fmt.Errorf("falha ao inicializar workspace: %w", err)
	}

	return workspaceDir, nil
}

// WriteFiles salva a coleção de arquivos com validação rígida anti-traversal
func (s *LocalSandboxManager) WriteFiles(cardID string, files []app.ProjectFile) error {
	if err := s.validateCardID(cardID); err != nil {
		return err
	}

	workspaceDir := s.GetWorkspacePath(cardID)
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		return ErrWorkspaceNotFound
	}

	for _, f := range files {
		cleanRelative := filepath.Clean(f.Path)

		// Proteção anti-traversal (rejeita .., /, \ ou tentativas de fuga)
		if strings.HasPrefix(cleanRelative, "..") || strings.HasPrefix(cleanRelative, "/") || strings.HasPrefix(cleanRelative, "\\") {
			return fmt.Errorf("%w: %s", ErrPathTraversalAttempt, f.Path)
		}

		fullPath := filepath.Join(workspaceDir, cleanRelative)
		rel, err := filepath.Rel(workspaceDir, fullPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("%w: %s", ErrPathTraversalAttempt, f.Path)
		}

		// Cria pastas pai se necessário
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("falha ao criar diretório para %s: %w", cleanRelative, err)
		}

		if err := os.WriteFile(fullPath, []byte(f.Content), 0644); err != nil {
			return fmt.Errorf("falha ao escrever arquivo %s: %w", cleanRelative, err)
		}
	}

	return nil
}

// ReadFiles recupera todos os arquivos do workspace excluindo pastas auxiliares (.git, node_modules)
func (s *LocalSandboxManager) ReadFiles(cardID string) ([]app.ProjectFile, error) {
	if err := s.validateCardID(cardID); err != nil {
		return nil, err
	}

	workspaceDir := s.GetWorkspacePath(cardID)
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		return nil, ErrWorkspaceNotFound
	}

	var result []app.ProjectFile

	err := filepath.WalkDir(workspaceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Pular diretórios pesados e de controle
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".next" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(workspaceDir, path)
		if err != nil {
			return err
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		result = append(result, app.ProjectFile{
			Path:    relPath,
			Content: string(contentBytes),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivos do workspace: %w", err)
	}

	return result, nil
}

// RunValidation executa comandos de compilação ou testes dentro do workspace do cartão
func (s *LocalSandboxManager) RunValidation(ctx context.Context, cardID string, cmd string, args ...string) (*app.ExecResult, error) {
	if err := s.validateCardID(cardID); err != nil {
		return nil, err
	}

	workspaceDir := s.GetWorkspacePath(cardID)
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		return nil, ErrWorkspaceNotFound
	}

	execCmd := exec.CommandContext(ctx, cmd, args...)
	execCmd.Dir = workspaceDir

	var outputBuf bytes.Buffer
	execCmd.Stdout = &outputBuf
	execCmd.Stderr = &outputBuf

	err := execCmd.Run()
	outputStr := outputBuf.String()

	exitCode := 0
	errMsg := ""
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
			errMsg = err.Error()
		} else {
			exitCode = 1
			errMsg = err.Error()
		}
	}

	return &app.ExecResult{
		ExitCode: exitCode,
		Output:   outputStr,
		Error:    errMsg,
	}, nil
}

// CleanupWorkspace exclui permanentemente o diretório do projeto da sandbox
func (s *LocalSandboxManager) CleanupWorkspace(cardID string) error {
	if err := s.validateCardID(cardID); err != nil {
		return err
	}

	workspaceDir := s.GetWorkspacePath(cardID)
	return os.RemoveAll(workspaceDir)
}
