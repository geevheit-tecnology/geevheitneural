package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

func HandleListFactoryTools(uc *app.FactoryTools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := uc.List(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "factory_tools_failed", "Não foi possível consultar as ferramentas de criação.")
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"tools": items})
	}
}

func HandleSelectFactoryTool(uc *app.FactoryTools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}
		var body struct {
			Active            *bool `json:"active"`
			FreeTierConfirmed bool  `json:"free_tier_confirmed"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || body.Active == nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "informe active e free_tier_confirmed")
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "envie somente um objeto JSON")
			return
		}
		err := uc.Select(r.Context(), owner, r.PathValue("id"), *body.Active, body.FreeTierConfirmed)
		if err != nil {
			status := http.StatusInternalServerError
			message := "Não foi possível salvar a ferramenta de criação."
			if errors.Is(err, app.ErrFactoryToolSelection) || errors.Is(err, app.ErrFactoryToolUnavailable) {
				status = http.StatusBadRequest
				message = err.Error()
			}
			if errors.Is(err, system.ErrSystemPaused) {
				status = http.StatusLocked
				message = err.Error()
			}
			if errors.Is(err, system.ErrOwnerRequired) {
				status = http.StatusForbidden
				message = err.Error()
			}
			middleware.WriteError(w, status, "factory_tool_selection_failed", message)
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
	}
}
