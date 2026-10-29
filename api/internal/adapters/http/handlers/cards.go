package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleCreateCard(uc *app.CreateCardUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cmd app.CreateCardCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "dados do cartão inválidos")
			return
		}
		created, err := uc.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "create_card_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusCreated, created)
	}
}

func HandleListCards(uc *app.ListCardsUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cards, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_cards_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"cards": cards,
			"count": len(cards),
		})
	}
}

func HandleGetCardDetails(uc *app.GetCardDetailsUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		details, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusNotFound, "card_not_found", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, details)
	}
}

func HandleDeleteCard(uc *app.DeleteCardUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		if cardID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}

		if err := uc.Execute(r.Context(), cardID); err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "delete_card_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"id":      cardID,
			"message": "Cartão excluído com sucesso",
		})
	}
}
