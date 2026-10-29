package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type StartBuildRequest struct {
	SpendLimitCents int64  `json:"spend_limit_cents"`
	Title           string `json:"title,omitempty"`
	ClientName      string `json:"client_name,omitempty"`
	PriceCents      int64  `json:"price_cents,omitempty"`
}

func HandleStartBuild(uc *app.StartBuildUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}

		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		var req StartBuildRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		p, err := uc.Execute(r.Context(), owner, app.StartBuildCommand{
			CardID:          cardID,
			SpendLimitCents: req.SpendLimitCents,
			Title:           req.Title,
			ClientName:      req.ClientName,
			PriceCents:      req.PriceCents,
		})
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "start_build_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusAccepted, p)
	}
}

func HandleGetCardPipeline(uc *app.GetCardPipelineUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		p, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusNotFound, "pipeline_not_found", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, p)
	}
}
