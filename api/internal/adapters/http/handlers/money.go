package handlers

import (
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleGetMoneyWaiting(uc *app.GetMoneyWaitingUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "money_waiting_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"total_cents": res.TotalCents,
			"count":       res.Count,
			"formatted":   res.Formatted,
		})
	}
}
