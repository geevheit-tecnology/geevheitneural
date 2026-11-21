package agents

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

func TestOpenAICompatRunner_MissingKey(t *testing.T) {
	r := NewOpenAIAgentRunner(OpenAICompatConfig{})
	_, err := r.RunStep(context.Background(), pipeline.RoleRequisitos, "teste", nil)
	if !errors.Is(err, ErrOpenAICompatKeyMissing) {
		t.Fatalf("expected ErrOpenAICompatKeyMissing, got %v", err)
	}
}

func TestOpenAICompatRunner_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": `{"output":"Plano pronto","stories":[{"title":"H1","given":"G","when":"W","then":"T","passed":true}]}`,
					},
				},
			},
		})
	}))
	defer srv.Close()

	r := NewOpenRouterAgentRunner(OpenAICompatConfig{
		APIKey:  "test-key",
		BaseURL: srv.URL,
	})
	res, err := r.RunStep(context.Background(), pipeline.RoleRequisitos, "Landing", map[string]any{"title": "Landing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Output != "Plano pronto" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestAnthropicRunner_MissingKeyAndSuccess(t *testing.T) {
	missing := NewAnthropicAgentRunner(AnthropicConfig{})
	if _, err := missing.RunStep(context.Background(), pipeline.RoleRequisitos, "teste", nil); !errors.Is(err, ErrAnthropicKeyMissing) {
		t.Fatalf("expected ErrAnthropicKeyMissing, got %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "ant-key" || r.Header.Get("anthropic-version") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": "```json\n{\"output\":\"Claude OK\",\"stories\":[{\"title\":\"H1\",\"given\":\"G\",\"when\":\"W\",\"then\":\"T\",\"passed\":true}]}\n```",
				},
			},
		})
	}))
	defer srv.Close()

	r := NewAnthropicAgentRunner(AnthropicConfig{
		APIKey:  "ant-key",
		BaseURL: srv.URL,
	})
	res, err := r.RunStep(context.Background(), pipeline.RoleRequisitos, "Landing", map[string]any{"title": "Landing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Output != "Claude OK" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
