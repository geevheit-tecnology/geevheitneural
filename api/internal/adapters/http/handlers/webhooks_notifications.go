package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/notification"
	"github.com/geevheit/neuralcenter/api/internal/domain/webhook"
)

type NotificationWebhookHandlers struct {
	uc *app.NotificationWebhooksUseCases
}

func NewNotificationWebhookHandlers(uc *app.NotificationWebhooksUseCases) *NotificationWebhookHandlers {
	return &NotificationWebhookHandlers{uc: uc}
}

// ListEndpoints manipula GET /v1/webhooks
func (h *NotificationWebhookHandlers) ListEndpoints() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		eps, err := h.uc.ListWebhookEndpoints(r.Context(), principal)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_webhooks_failed", err.Error())
			return
		}

		if eps == nil {
			eps = make([]webhook.Endpoint, 0)
		}

		middleware.WriteJSON(w, http.StatusOK, eps)
	}
}

// CreateEndpoint manipula POST /v1/webhooks
func (h *NotificationWebhookHandlers) CreateEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var in app.CreateWebhookInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		ep, err := h.uc.CreateWebhookEndpoint(r.Context(), principal, in)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "create_webhook_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, ep)
	}
}

// DeleteEndpoint manipula DELETE /v1/webhooks/{id}
func (h *NotificationWebhookHandlers) DeleteEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "missing_id", "id do webhook é obrigatório")
			return
		}

		if err := h.uc.DeleteWebhookEndpoint(r.Context(), principal, id); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "delete_webhook_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "endpoint removido com sucesso"})
	}
}

// ListDeliveries manipula GET /v1/webhooks/deliveries
func (h *NotificationWebhookHandlers) ListDeliveries() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var webhookID *string
		if wID := r.URL.Query().Get("webhook_id"); wID != "" {
			webhookID = &wID
		}

		deliveries, err := h.uc.ListDeliveries(r.Context(), principal, webhookID)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_deliveries_failed", err.Error())
			return
		}

		if deliveries == nil {
			deliveries = make([]webhook.Delivery, 0)
		}

		middleware.WriteJSON(w, http.StatusOK, deliveries)
	}
}

// Redeliver manipula POST /v1/webhooks/{id}/redeliver
func (h *NotificationWebhookHandlers) Redeliver() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "missing_id", "id do webhook é obrigatório")
			return
		}

		var in struct {
			Event   string         `json:"event"`
			Payload map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo inválido")
			return
		}

		delivery, err := h.uc.Redeliver(r.Context(), principal, id, in.Event, in.Payload)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "redeliver_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, delivery)
	}
}

// GetPreferences manipula GET /v1/notifications/preferences
func (h *NotificationWebhookHandlers) GetPreferences() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		prefs, err := h.uc.GetNotificationPreferences(r.Context(), principal)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "get_preferences_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, prefs)
	}
}

// SavePreferences manipula PUT /v1/notifications/preferences
func (h *NotificationWebhookHandlers) SavePreferences() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var prefs notification.Preferences
		if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo inválido")
			return
		}

		if err := h.uc.SaveNotificationPreferences(r.Context(), principal, prefs); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "save_preferences_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, prefs)
	}
}
