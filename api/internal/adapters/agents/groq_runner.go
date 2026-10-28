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
	ErrGroqKeyMissing   = errors.New("chave da API Groq não configurada (GROQ_API_KEY)")
	ErrGroqAPIError     = errors.New("erro retornado pela API da Groq")
	ErrGroqInvalidReply = errors.New("resposta da API Groq em formato inesperado")
)

// GroqConfig define as opções de conexão com a API da Groq
type GroqConfig struct {
	APIKey     string
	Model      string // ex: "openai/gpt-oss-120b", "qwen/qwen3.8-27b", etc.
	BaseURL    string // URL base da API compatível OpenAI
	HTTPClient *http.Client
}

// GroqAgentRunner implementa pipeline.AgentRunner conectando-se aos modelos de ultra velocidade da Groq
type GroqAgentRunner struct {
	cfg GroqConfig
}

// NewGroqAgentRunner instancia o executor da Groq
func NewGroqAgentRunner(cfg GroqConfig) *GroqAgentRunner {
	if cfg.Model == "" {
		cfg.Model = "openai/gpt-oss-120b"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.groq.com/openai/v1"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}

	return &GroqAgentRunner{
		cfg: cfg,
	}
}

type groqChatRequest struct {
	Model          string              `json:"model"`
	Messages       []groqChatMessage   `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// RunStep executa um passo do pipeline utilizando os modelos da Groq
func (r *GroqAgentRunner) RunStep(ctx context.Context, role pipeline.Role, prompt string, contextMeta map[string]any) (*pipeline.RunnerResult, error) {
	if r.cfg.APIKey == "" {
		return nil, ErrGroqKeyMissing
	}

	systemInstruction := buildSystemInstruction(role, contextMeta)
	userContent := fmt.Sprintf("Instrução do Passo: %s\nContexto: %s", prompt, formatMeta(contextMeta))

	reqBody := groqChatRequest{
		Model: r.cfg.Model,
		Messages: []groqChatMessage{
			{
				Role:    "system",
				Content: systemInstruction,
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
		Temperature: 0.2,
		ResponseFormat: &groqResponseFormat{
			Type: "json_object",
		},
	}

	rawJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar payload para Groq: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", strings.TrimRight(r.cfg.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(rawJSON))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição HTTP para Groq: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)

	resp, err := r.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha de rede ao conectar com Groq: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler resposta da API Groq: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d - %s", ErrGroqAPIError, resp.StatusCode, string(bodyBytes))
	}

	var groqResp groqChatResponse
	if err := json.Unmarshal(bodyBytes, &groqResp); err != nil {
		return nil, fmt.Errorf("falha ao decodificar JSON retornado pela Groq: %w", err)
	}

	if groqResp.Error != nil {
		return nil, fmt.Errorf("%w: [%s] %s", ErrGroqAPIError, groqResp.Error.Type, groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 || groqResp.Choices[0].Message.Content == "" {
		return nil, ErrGroqInvalidReply
	}

	responseText := cleanJSONMarkdown(groqResp.Choices[0].Message.Content)
	return parseRoleOutput(role, responseText, contextMeta)
}
