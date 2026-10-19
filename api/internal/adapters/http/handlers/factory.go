package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleListPlaybooks(uc *app.ListPlaybooksUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_playbooks_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"playbooks": list})
	}
}

func HandleListModules(uc *app.ListModulesUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_modules_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"modules": list})
	}
}

func HandleGetFactoryScoreboard(uc *app.GetFactoryScoreboardUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		board, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "scoreboard_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, board)
	}
}

func HandleEstimatePrice(uc *app.EstimatePriceUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cmd app.EstimatePriceCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "parâmetros de estimativa inválidos")
			return
		}

		res, err := uc.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "estimate_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, res)
	}
}
