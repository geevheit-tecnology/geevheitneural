package handlers

import (
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleGetPublicScope(uc *app.GetPublicScopeUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		if token == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_token", "token de escopo ausente")
			return
		}

		view, err := uc.Execute(r.Context(), token)
		if err != nil {
			middleware.WriteError(w, http.StatusNotFound, "scope_not_found", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, view)
	}
}

func HandleApprovePublicScope(uc *app.ApprovePublicScopeUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		if token == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_token", "token de escopo ausente")
			return
		}

		if err := uc.Execute(r.Context(), token); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "scope_approval_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "Escopo aprovado com sucesso! A fábrica já pode iniciar a construção.",
		})
	}
}
