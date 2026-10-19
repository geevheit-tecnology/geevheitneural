package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type ClientLGPDHandler struct {
	uc *app.ClientLGPDUseCases
}

func NewClientLGPDHandler(uc *app.ClientLGPDUseCases) *ClientLGPDHandler {
	return &ClientLGPDHandler{uc: uc}
}

// ExportClient - GET /v1/clients/{id}/export
func (h *ClientLGPDHandler) ExportClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	// Extrai ID do path (/v1/clients/{id}/export)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "id do cliente não informado", http.StatusBadRequest)
		return
	}
	clientID := parts[2]

	data, err := h.uc.Export(r.Context(), principal, clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

// PurgeClient - DELETE /v1/clients/{id}/purge
func (h *ClientLGPDHandler) PurgeClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "id do cliente não informado", http.StatusBadRequest)
		return
	}
	clientID := parts[2]

	if err := h.uc.Purge(r.Context(), principal, clientID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Dados do cliente expurgados com sucesso conforme LGPD.",
	})
}
