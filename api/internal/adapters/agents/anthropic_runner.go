package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

var (
	ErrAnthropicKeyMissing   = errors.New("chave da API Anthropic não configurada (ANTHROPIC_API_KEY)")
	ErrAnthropicAPIError     = errors.New("erro retornado pela API da Anthropic")
	ErrAnthropicInvalidReply = errors.New("resposta da API Anthropic em formato inesperado")
)

// AnthropicConfig configura o executor nativo da API Anthropic Messages (Claude).
type AnthropicConfig struct {
	APIKey     string
	Model      string
	BaseURL    string
	MaxTokens  int
	HTTPClient *http.Client
}

type AnthropicAgentRunner struct {
	cfg AnthropicConfig
}

func NewAnthropicAgentRunner(cfg AnthropicConfig) *AnthropicAgentRunner {
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 8192
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	return &AnthropicAgentRunner{cfg: cfg}
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	System      string             `json:"system"`
	Messages    []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (r *AnthropicAgentRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	if r.cfg.APIKey == "" {
		return nil, ErrAnthropicKeyMissing
	}

	systemInstruction := buildSystemInstruction(role, contextMeta)
	userContent := fmt.Sprintf("Instrução do Passo: %s\nContexto: %s", prompt, formatMeta(contextMeta))

	reqBody := anthropicRequest{
		Model:       r.cfg.Model,
		MaxTokens:   r.cfg.MaxTokens,
		Temperature: 0.2,
		System:      systemInstruction,
		Messages: []anthropicMessage{
			{Role: "user", Content: userContent},
		},
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload para Anthropic: %w", err)
	}

	endpoint := fmt.Sprintf("%s/messages", strings.TrimRight(r.cfg.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição HTTP para Anthropic: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", r.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := r.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha de rede ao conectar com Anthropic: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler resposta da API Anthropic: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d - %s", ErrAnthropicAPIError, resp.StatusCode, string(bodyBytes))
	}

	var antResp anthropicResponse
	if err := json.Unmarshal(bodyBytes, &antResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON retornado pela Anthropic: %w", err)
	}
	if antResp.Error != nil {
		return nil, fmt.Errorf("%w: [%s] %s", ErrAnthropicAPIError, antResp.Error.Type, antResp.Error.Message)
	}

	var textParts []string
	for _, c := range antResp.Content {
		if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
			textParts = append(textParts, c.Text)
		}
	}
	if len(textParts) == 0 {
		return nil, ErrAnthropicInvalidReply
	}

	responseText := cleanJSONMarkdown(strings.Join(textParts, "\n"))
	return parseRoleOutput(role, responseText, contextMeta)
}
