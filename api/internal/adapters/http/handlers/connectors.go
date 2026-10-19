package handlers

import (
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleListConnectors(uc *app.ListConnectorsUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectors, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_connectors_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"connectors": connectors,
			"count":      len(connectors),
		})
	}
}
