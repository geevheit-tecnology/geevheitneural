package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type PauseRequest struct {
	Reason string `json:"reason"`
}

func HandlePauseSystem(uc *app.PauseSystemUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}

		var req PauseRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Reason == "" {
			req.Reason = "Pausa acionada pelo dono"
		}

		if err := uc.Execute(r.Context(), owner, req.Reason); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "pause_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"paused": true,
			"reason": req.Reason,
		})
	}
}

func HandleResumeSystem(uc *app.ResumeSystemUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}

		if err := uc.Execute(r.Context(), owner); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "resume_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"paused": false,
		})
	}
}
