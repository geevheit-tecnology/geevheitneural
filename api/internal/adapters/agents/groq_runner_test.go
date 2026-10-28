package agents_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/adapters/agents"
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
)

func TestGroqAgentRunner_KeyMissing(t *testing.T) {
	runner := agents.NewGroqAgentRunner(agents.GroqConfig{
		APIKey: "",
	})

	_, err := runner.RunStep(context.Background(), pipeline.RoleRequisitos, "Demanda teste", nil)
	if err != agents.ErrGroqKeyMissing {
		t.Fatalf("esperava ErrGroqKeyMissing, obteve: %v", err)
	}
}

func TestGroqAgentRunner_RunStep_Success(t *testing.T) {
	fakeGroqResponse := map[string]any{
		"choices": []map[string]any{
			{
				"message": map[string]any{
					"role": "assistant",
					"content": `{
						"output": "Arquitetura proposta com sucesso",
						"playbook_selected": "pwa_mobile",
						"files_plan": ["index.html", "app.js"],
						"tech_stack": "Next.js, Go"
					}`,
				},
				"finish_reason": "stop",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperava POST, obteve %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-groq-key" {
			t.Errorf("header Authorization inválido: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeGroqResponse)
	}))
	defer server.Close()

	runner := agents.NewGroqAgentRunner(agents.GroqConfig{
		APIKey:  "test-groq-key",
		Model:   "openai/gpt-oss-120b",
		BaseURL: server.URL,
	})

	res, err := runner.RunStep(context.Background(), pipeline.RoleArquiteto, "Planejar PWA", nil)
	if err != nil {
		t.Fatalf("esperava sucesso no Groq RunStep, obteve erro: %v", err)
	}

	if res.Role != pipeline.RoleArquiteto {
		t.Errorf("esperava role Arquiteto, obteve: %s", res.Role)
	}
	if res.Output != "Arquitetura proposta com sucesso" {
		t.Errorf("output inesperado: %s", res.Output)
	}
	if res.NextRole == nil || *res.NextRole != pipeline.RoleDev {
		t.Errorf("esperava NextRole RoleDev, obteve: %v", res.NextRole)
	}
}
