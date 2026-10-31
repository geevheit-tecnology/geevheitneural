package agents_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

func TestGeminiAgentRunner_KeyMissing(t *testing.T) {
	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{
		APIKey: "",
	})

	_, err := runner.RunStep(context.Background(), pipeline.RoleRequisitos, "Criar sistema", nil)
	if err != agents.ErrGeminiKeyMissing {
		t.Fatalf("esperava ErrGeminiKeyMissing, obteve: %v", err)
	}
}

func TestGeminiAgentRunner_RunStep_Success(t *testing.T) {
	fakeGeminiResponse := map[string]any{
		"candidates": []map[string]any{
			{
				"content": map[string]any{
					"parts": []map[string]any{
						{
							"text": `{
								"output": "Requisitos estruturados com sucesso",
								"stories": [
									{
										"title": "Login com TOTP",
										"given": "Dado que o dono acessa",
										"when": "Quando insere 2FA",
										"then": "Então acessa o sistema",
										"passed": true
									}
								]
							}`,
						},
					},
					"role": "model",
				},
				"finishReason": "STOP",
			},
		},
		"usageMetadata": map[string]any{
			"promptTokenCount":     150,
			"candidatesTokenCount": 200,
			"totalTokenCount":      350,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperava método POST, recebeu %s", r.Method)
		}
		if r.Header.Get("x-goog-api-key") != "test-api-key" || r.URL.Query().Has("key") {
			t.Error("a chave deve estar somente no cabeçalho")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeGeminiResponse)
	}))
	defer server.Close()

	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{
		APIKey:      "test-api-key",
		Model:       "gemini-1.5-flash",
		BaseURL:     server.URL,
		HTTPClient:  server.Client(),
		MinInterval: 10 * time.Millisecond, // Reduzido para o teste rodar rápido
	})

	res, err := runner.RunStep(context.Background(), pipeline.RoleRequisitos, "Projeto de Padaria", map[string]any{"client": "Padaria Silva"})
	if err != nil {
		t.Fatalf("erro inesperado ao executar RunStep: %v", err)
	}

	if res.Role != pipeline.RoleRequisitos {
		t.Errorf("esperava Role %s, obteve %s", pipeline.RoleRequisitos, res.Role)
	}
	if res.CostCents != 0 {
		t.Errorf("esperava Custo 0 (Free Tier), obteve %d", res.CostCents)
	}
	if res.NextRole == nil || *res.NextRole != pipeline.RoleArquiteto {
		t.Errorf("esperava NextRole %s, obteve %v", pipeline.RoleArquiteto, res.NextRole)
	}
	if res.RequiresGate != false {
		t.Errorf("esperava RequiresGate false para Requisitos, obteve %v", res.RequiresGate)
	}
}

func TestGeminiAgentRunner_Publicador_RequiresGate(t *testing.T) {
	fakeGeminiResponse := map[string]any{
		"candidates": []map[string]any{
			{
				"content": map[string]any{
					"parts": []map[string]any{
						{
							"text": `{"output": "Prévia publicada", "preview_url": "https://preview.vercel.app"}`,
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeGeminiResponse)
	}))
	defer server.Close()

	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{
		APIKey:      "test-key",
		BaseURL:     server.URL,
		HTTPClient:  server.Client(),
		MinInterval: 5 * time.Millisecond,
	})

	res, err := runner.RunStep(context.Background(), pipeline.RolePublicador, "Publicar prévia", nil)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// Portão 3 deve ser obrigatoriamente exigido
	if !res.RequiresGate {
		t.Errorf("esperava RequiresGate=true para RolePublicador (Portão de Entrega do Dono)")
	}
}

func TestGeminiAgentRunner_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error": {"code": 429, "message": "Resource has been exhausted", "status": "RESOURCE_EXHAUSTED"}}`))
	}))
	defer server.Close()

	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{
		APIKey:      "test-key",
		BaseURL:     server.URL,
		HTTPClient:  server.Client(),
		MinInterval: 5 * time.Millisecond,
	})

	_, err := runner.RunStep(context.Background(), pipeline.RoleDev, "Construir código", nil)
	if err == nil {
		t.Fatalf("esperava erro HTTP 429, mas obteve sucesso")
	}
}

func TestGeminiAgentRunner_TransientFailureRetry(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"output\":\"ok\"}"}]}}]}`))
	}))
	defer server.Close()
	runner := agents.NewGeminiAgentRunner(agents.GeminiConfig{APIKey: "test", BaseURL: server.URL, HTTPClient: server.Client(), MinInterval: time.Millisecond})
	result, err := runner.RunStep(context.Background(), pipeline.RoleDev, "test", nil)
	if err != nil || result.Output != "ok" || calls != 2 {
		t.Fatalf("calls=%d result=%v err=%v", calls, result, err)
	}
}
