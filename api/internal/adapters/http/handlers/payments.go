package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type GeneratePixRequest struct {
	PixKey string `json:"pix_key,omitempty"`
}

func HandleGeneratePix(uc *app.GeneratePixPaymentUseCase, defaultOwnerID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "ID do cartão obrigatório")
			return
		}

		var req GeneratePixRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		ownerID := defaultOwnerID
		if p, ok := middleware.GetOwnerPrincipal(r.Context()); ok && p.ID != "" {
			ownerID = p.ID
		}

		order, err := uc.Execute(r.Context(), cardID, ownerID, req.PixKey)
		if err != nil {
			middleware.WriteError(w, http.StatusUnprocessableEntity, "pix_generation_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, order)
	}
}

func HandleGetPix(uc *app.GetPixPaymentUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "ID do cartão obrigatório")
			return
		}

		order, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusNotFound, "pix_not_found", "Nenhuma cobrança PIX gerada para este cartão")
			return
		}

		middleware.WriteJSON(w, http.StatusOK, order)
	}
}

type PixWebhookRequest struct {
	TxID        string `json:"txid"`
	AmountCents int64  `json:"amount_cents"`
}

func HandlePixWebhook(uc *app.ProcessPixWebhookUseCase, webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if webhookSecret != "" {
			sig := r.Header.Get("X-Webhook-Secret")
			if sig != webhookSecret {
				middleware.WriteError(w, http.StatusUnauthorized, "invalid_webhook_secret", "Segredo do webhook inválido")
				return
			}
		}

		var req PixWebhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TxID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "Payload do webhook deve conter txid")
			return
		}

		order, err := uc.Execute(r.Context(), req.TxID, req.AmountCents)
		if err != nil {
			middleware.WriteError(w, http.StatusUnprocessableEntity, "webhook_processing_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"status":  "paid",
			"txid":    order.PixTxID,
			"card_id": order.CardID,
			"amount":  order.AmountCents,
		})
	}
}
