package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleGetCardSpecification(uc *app.GetCardSpecificationUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão ausente")
			return
		}

		view, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusNotFound, "spec_not_found", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, view)
	}
}

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

func HandleRequestScopeChange(uc *app.RequestScopeChangeUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		if token == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_token", "token de escopo ausente")
			return
		}

		var cmd app.ScopeChangeRequestCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "dados do aditivo inválidos")
			return
		}

		card, err := uc.Execute(r.Context(), token, cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "change_request_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, map[string]any{
			"success":       true,
			"additive_card": card,
			"message":       "Aditivo de Escopo criado automaticamente para avaliação do dono.",
		})
	}
}

func HandlePayScopeSignal(uc *app.PayScopeSignalUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.PathValue("token")
		if token == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_token", "token de escopo ausente")
			return
		}

		if err := uc.Execute(r.Context(), token); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "signal_payment_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "Pagamento do sinal confirmado e escopo liberado para construção!",
		})
	}
}
