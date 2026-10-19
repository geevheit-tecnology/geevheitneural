package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
)

type DecideGateRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

func HandleDecideGate(uc *app.DecideGateUseCase) http.HandlerFunc {
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

		var req DecideGateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		dec := card.Decision(req.Decision)
		if !dec.IsValid() {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_decision", "decisão inválida (use: aprovado, mais_info, recusado)")
			return
		}

		err := uc.Execute(r.Context(), app.DecideGateInput{
			Owner:    owner,
			CardID:   cardID,
			Decision: dec,
			Note:     req.Note,
		})

		if err != nil {
			if errors.Is(err, card.ErrSystemPaused) {
				middleware.WriteError(w, http.StatusLocked, "system_paused", "sistema pausado: nenhuma execução pode ser iniciada")
				return
			}
			if errors.Is(err, card.ErrInvalidTransition) || errors.Is(err, card.ErrGateRequiresOwner) {
				middleware.WriteError(w, http.StatusBadRequest, "invalid_transition", err.Error())
				return
			}
			middleware.WriteError(w, http.StatusInternalServerError, "gate_decision_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"success":  true,
			"card_id":  cardID,
			"decision": dec,
		})
	}
}
