package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type AgentHandler struct {
	uc *app.AgentUseCases
}

func NewAgentHandler(uc *app.AgentUseCases) *AgentHandler {
	return &AgentHandler{uc: uc}
}

func handleAgentLockError(w http.ResponseWriter, err error) bool {
	if errors.Is(err, app.ErrExecutionLocked) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusLocked) // 423 Locked
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   "execution_locked",
			"message": "Execução bloqueada: sistema pausado ou agente inativo pelo dono.",
		})
		return true
	}
	return false
}

// ListCards - GET /v1/agent/cards
func (h *AgentHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetAgentPrincipal(r.Context())
	if !ok {
		http.Error(w, "agente não autenticado", http.StatusUnauthorized)
		return
	}

	cards, err := h.uc.ListAvailableCards(r.Context(), principal)
	if err != nil {
		if handleAgentLockError(w, err) {
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cards)
}

// GetCard - GET /v1/agent/cards/{id}
func (h *AgentHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetAgentPrincipal(r.Context())
	if !ok {
		http.Error(w, "agente não autenticado", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "id do cartão não informado", http.StatusBadRequest)
		return
	}
	cardID := parts[3]

	details, err := h.uc.GetCardDetails(r.Context(), principal, cardID)
	if err != nil {
		if handleAgentLockError(w, err) {
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(details)
}

// RecordProgress - POST /v1/agent/cards/{id}/progress
func (h *AgentHandler) RecordProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetAgentPrincipal(r.Context())
	if !ok {
		http.Error(w, "agente não autenticado", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "url inválida", http.StatusBadRequest)
		return
	}
	cardID := parts[3]

	var cmd app.AgentProgressCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "payload json inválido", http.StatusBadRequest)
		return
	}

	if err := h.uc.RecordProgress(r.Context(), principal, cardID, cmd); err != nil {
		if handleAgentLockError(w, err) {
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Progresso registrado com sucesso.",
	})
}

// SubmitDelivery - POST /v1/agent/cards/{id}/submit-delivery
func (h *AgentHandler) SubmitDelivery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetAgentPrincipal(r.Context())
	if !ok {
		http.Error(w, "agente não autenticado", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "url inválida", http.StatusBadRequest)
		return
	}
	cardID := parts[3]

	var cmd app.AgentSubmitDeliveryCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "payload json inválido", http.StatusBadRequest)
		return
	}

	if err := h.uc.SubmitDelivery(r.Context(), principal, cardID, cmd); err != nil {
		if handleAgentLockError(w, err) {
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Entrega submetida com sucesso para aprovação do dono (Portão 3).",
	})
}
