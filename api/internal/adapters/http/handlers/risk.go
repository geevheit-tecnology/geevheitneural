package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type RiskHandlers struct {
	getUC  *app.GetRiskReportUseCase
	saveUC *app.SaveRiskReportUseCase
}

func NewRiskHandlers(getUC *app.GetRiskReportUseCase, saveUC *app.SaveRiskReportUseCase) *RiskHandlers {
	return &RiskHandlers{getUC: getUC, saveUC: saveUC}
}

func (h *RiskHandlers) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão obrigatório")
			return
		}

		report, err := h.getUC.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "get_risk_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, report)
	}
}

func (h *RiskHandlers) Save() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão obrigatório")
			return
		}

		var cmd app.SaveRiskReportCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}
		cmd.CardID = cardID

		saved, err := h.saveUC.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "save_risk_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, saved)
	}
}
