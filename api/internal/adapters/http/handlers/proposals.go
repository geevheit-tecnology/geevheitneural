package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type ProposalHandler struct {
	uc *app.ProposalUseCases
}

func NewProposalHandler(uc *app.ProposalUseCases) *ProposalHandler {
	return &ProposalHandler{uc: uc}
}

// CreateProposal - Endpoint autenticado (Dono AAL2): POST /v1/proposals
func (h *ProposalHandler) CreateProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	var cmd app.CreateProposalCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "payload json inválido", http.StatusBadRequest)
		return
	}

	proposal, err := h.uc.CreateProposal(r.Context(), principal, cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(proposal)
}

// GetPublicProposal - Endpoint público: GET /v1/public/proposals/{token}
func (h *ProposalHandler) GetPublicProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Extrai token da URL (/v1/public/proposals/{token})
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "token não informado", http.StatusBadRequest)
		return
	}
	token := pathParts[3]

	view, err := h.uc.GetPublicProposal(r.Context(), token)
	if err != nil {
		http.Error(w, "proposta não encontrada ou expirada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(view)
}

// AcceptPublicProposal - Endpoint público: POST /v1/public/proposals/{token}/accept
func (h *ProposalHandler) AcceptPublicProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "token não informado", http.StatusBadRequest)
		return
	}
	token := pathParts[3]

	if err := h.uc.AcceptProposal(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Proposta aceita com sucesso! Nossa equipe iniciará a especificação do projeto.",
	})
}

// GetProposalDocument - Endpoint autenticado (Dono AAL2): GET/POST /v1/cards/{id}/proposal-document
func (h *ProposalHandler) GetProposalDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 3 && parts[0] == "v1" && parts[1] == "cards" {
			cardID = parts[2]
		}
	}

	if cardID == "" {
		http.Error(w, "id do cartão não informado", http.StatusBadRequest)
		return
	}

	doc, err := h.uc.GenerateProposalDocument(r.Context(), principal, cardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(doc)
}

