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
	ErrOpenAICompatKeyMissing   = errors.New("chave da API do provedor não configurada")
	ErrOpenAICompatAPIError     = errors.New("erro retornado pela API do provedor")
	ErrOpenAICompatInvalidReply = errors.New("resposta da API em formato inesperado")
)

// OpenAICompatConfig configura provedores compatíveis com /chat/completions (OpenAI, OpenRouter).
type OpenAICompatConfig struct {
	ProviderName string
	APIKey       string
	Model        string
	BaseURL      string
	HTTPClient   *http.Client
	ExtraHeaders map[string]string
}

type OpenAICompatRunner struct {
	cfg OpenAICompatConfig
}

func NewOpenAIAgentRunner(cfg OpenAICompatConfig) *OpenAICompatRunner {
	if cfg.ProviderName == "" {
		cfg.ProviderName = "OpenAI"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4.1-mini"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	return &OpenAICompatRunner{cfg: cfg}
}

func NewOpenRouterAgentRunner(cfg OpenAICompatConfig) *OpenAICompatRunner {
	if cfg.ProviderName == "" {
		cfg.ProviderName = "OpenRouter"
	}
	if cfg.Model == "" {
		cfg.Model = "anthropic/claude-sonnet-4"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://openrouter.ai/api/v1"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	if cfg.ExtraHeaders == nil {
		cfg.ExtraHeaders = map[string]string{
			"X-Title": "GeevheitNeural",
		}
	}
	return &OpenAICompatRunner{cfg: cfg}
}

func (r *OpenAICompatRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	if r.cfg.APIKey == "" {
		return nil, fmt.Errorf("%w (%s)", ErrOpenAICompatKeyMissing, r.cfg.ProviderName)
	}

	systemInstruction := buildSystemInstruction(role, contextMeta)
	userContent := fmt.Sprintf("Instrução do Passo: %s\nContexto: %s", prompt, formatMeta(contextMeta))

	reqBody := groqChatRequest{
		Model: r.cfg.Model,
		Messages: []groqChatMessage{
			{Role: "system", Content: systemInstruction},
			{Role: "user", Content: userContent},
		},
		Temperature:    0.2,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload para %s: %w", r.cfg.ProviderName, err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", strings.TrimRight(r.cfg.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição HTTP para %s: %w", r.cfg.ProviderName, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
	for k, v := range r.cfg.ExtraHeaders {
		httpReq.Header.Set(k, v)
	}

	resp, err := r.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha de rede ao conectar com %s: %w", r.cfg.ProviderName, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler resposta da API %s: %w", r.cfg.ProviderName, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w (%s): status %d - %s", ErrOpenAICompatAPIError, r.cfg.ProviderName, resp.StatusCode, string(bodyBytes))
	}

	var chatResp groqChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON retornado por %s: %w", r.cfg.ProviderName, err)
	}
	if chatResp.Error != nil {
		return nil, fmt.Errorf("%w (%s): [%s] %s", ErrOpenAICompatAPIError, r.cfg.ProviderName, chatResp.Error.Type, chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 || strings.TrimSpace(chatResp.Choices[0].Message.Content) == "" {
		return nil, fmt.Errorf("%w (%s)", ErrOpenAICompatInvalidReply, r.cfg.ProviderName)
	}

	responseText := cleanJSONMarkdown(chatResp.Choices[0].Message.Content)
	return parseRoleOutput(role, responseText, contextMeta)
}
