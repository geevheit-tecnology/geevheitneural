package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type PortalHandler struct {
	uc *app.PortalUseCases
}

func NewPortalHandler(uc *app.PortalUseCases) *PortalHandler {
	return &PortalHandler{uc: uc}
}

// GeneratePortalLink - POST /v1/cards/{id}/portal-link (Owner AAL2)
func (h *PortalHandler) GeneratePortalLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "card_id inválido", http.StatusBadRequest)
		return
	}
	cardID := pathParts[2]

	var req struct {
		DaysValid int `json:"days_valid"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.DaysValid <= 0 {
		req.DaysValid = 30
	}

	rawToken, url, err := h.uc.GeneratePortalLink(r.Context(), principal, cardID, req.DaysValid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token":      rawToken,
		"portal_url": url,
		"card_id":    cardID,
		"days_valid": req.DaysValid,
	})
}

// RevokePortalLink - DELETE /v1/cards/{id}/portal-link (Owner AAL2)
func (h *PortalHandler) RevokePortalLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "card_id inválido", http.StatusBadRequest)
		return
	}
	cardID := pathParts[2]

	if err := h.uc.RevokePortalLink(r.Context(), principal, cardID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "link do portal revogado com sucesso",
	})
}

// GetPortalStatus - GET /v1/cards/{id}/portal-link (Owner AAL2)
func (h *PortalHandler) GetPortalStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	principal, ok := middleware.GetOwnerPrincipal(r.Context())
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "card_id inválido", http.StatusBadRequest)
		return
	}
	cardID := pathParts[2]

	status, err := h.uc.GetPortalTokenStatus(r.Context(), principal, cardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(status)
}

// GetPublicPortal - GET /v1/public/portal/{token} (Público)
func (h *PortalHandler) GetPublicPortal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "token de acesso não fornecido", http.StatusBadRequest)
		return
	}
	rawToken := pathParts[3]

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	portalView, err := h.uc.GetClientPortal(r.Context(), rawToken, clientIP, userAgent)
	if err != nil {
		// Retorno genérico sem vazar se o token existiu ou não
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "link de acesso inválido ou expirado",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(portalView)
}

// AcceptPublicDelivery - POST /v1/public/portal/{token}/accept (Público)
func (h *PortalHandler) AcceptPublicDelivery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		http.Error(w, "token de acesso não fornecido", http.StatusBadRequest)
		return
	}
	rawToken := pathParts[3]

	var req struct {
		AcceptedBy string `json:"accepted_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.AcceptedBy) == "" {
		http.Error(w, "o campo 'accepted_by' é obrigatório para registrar o aceite", http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	portalView, err := h.uc.AcceptClientDelivery(r.Context(), rawToken, strings.TrimSpace(req.AcceptedBy), clientIP, userAgent)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(portalView)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	return r.RemoteAddr
}
