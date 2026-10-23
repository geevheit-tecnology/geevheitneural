package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/subscription"
)

type SubscriptionHandlers struct {
	uc *app.SubscriptionsUseCases
}

func NewSubscriptionHandlers(uc *app.SubscriptionsUseCases) *SubscriptionHandlers {
	return &SubscriptionHandlers{uc: uc}
}

// CreateSubscription manipula POST /v1/subscriptions
func (h *SubscriptionHandlers) CreateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var in app.CreateSubscriptionInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		sub, err := h.uc.CreateSubscription(r.Context(), principal, in)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "create_subscription_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, sub)
	}
}

// ListSubscriptions manipula GET /v1/subscriptions
func (h *SubscriptionHandlers) ListSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var clientID *string
		if c := r.URL.Query().Get("client_id"); c != "" {
			clientID = &c
		}

		var status *string
		if s := r.URL.Query().Get("status"); s != "" {
			status = &s
		}

		subs, err := h.uc.ListSubscriptions(r.Context(), principal, clientID, status)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_subscriptions_failed", err.Error())
			return
		}

		if subs == nil {
			subs = make([]subscription.Subscription, 0)
		}

		middleware.WriteJSON(w, http.StatusOK, subs)
	}
}

// UpdateStatus manipula PATCH /v1/subscriptions/{id}/status
func (h *SubscriptionHandlers) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "missing_id", "id da assinatura obrigatório")
			return
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		sub, err := h.uc.UpdateSubscriptionStatus(r.Context(), principal, id, payload.Status)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "update_status_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, sub)
	}
}

// GetMRRMetrics manipula GET /v1/finances/mrr
func (h *SubscriptionHandlers) GetMRRMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		metrics, err := h.uc.GetMRRMetrics(r.Context(), principal)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "get_mrr_metrics_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, metrics)
	}
}

// ListCharges manipula GET /v1/finances/charges
func (h *SubscriptionHandlers) ListCharges() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		var status *string
		if s := r.URL.Query().Get("status"); s != "" {
			status = &s
		}

		var dueBefore *time.Time
		if d := r.URL.Query().Get("due_before"); d != "" {
			if parsed, err := time.Parse("2006-01-02", d); err == nil {
				dueBefore = &parsed
			}
		}

		charges, err := h.uc.ListCharges(r.Context(), principal, status, dueBefore)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_charges_failed", err.Error())
			return
		}

		if charges == nil {
			charges = make([]subscription.SubscriptionCharge, 0)
		}

		middleware.WriteJSON(w, http.StatusOK, charges)
	}
}

// PayCharge manipula POST /v1/finances/charges/{id}/pay
func (h *SubscriptionHandlers) PayCharge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "missing_id", "id da cobrança obrigatório")
			return
		}

		var payload struct {
			PaymentMethod string `json:"payment_method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		charge, err := h.uc.PayCharge(r.Context(), principal, id, payload.PaymentMethod)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "pay_charge_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, charge)
	}
}

// ProcessDaily manipula POST /v1/subscriptions/process-daily
func (h *SubscriptionHandlers) ProcessDaily() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "autenticação do dono exigida")
			return
		}

		report, err := h.uc.ProcessDailySubscriptionCharges(r.Context(), principal, time.Now().UTC())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "process_daily_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, report)
	}
}
