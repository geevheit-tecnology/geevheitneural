package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/command"
)

type ProcessCommandRequest struct {
	Input string `json:"input"`
}

type ConfirmActionRequest struct {
	Token string `json:"token"`
}

type CommandHandlers struct {
	processUC *app.ProcessCommandUseCase
	confirmUC *app.ConfirmActionUseCase
}

func NewCommandHandlers(processUC *app.ProcessCommandUseCase, confirmUC *app.ConfirmActionUseCase) *CommandHandlers {
	return &CommandHandlers{
		processUC: processUC,
		confirmUC: confirmUC,
	}
}

func (h *CommandHandlers) HandleProcessCommand() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}

		var req ProcessCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		if req.Input == "" {
			middleware.WriteError(w, http.StatusBadRequest, "empty_command", "digite ou fale uma ordem válida")
			return
		}

		res, err := h.processUC.Execute(r.Context(), app.ProcessCommandInput{
			Owner: owner,
			Text:  req.Input,
		})
		if err != nil {
			if errors.Is(err, command.ErrAmbiguousTarget) {
				middleware.WriteError(w, http.StatusConflict, "ambiguous_target", err.Error())
				return
			}
			if errors.Is(err, command.ErrTargetNotFound) {
				middleware.WriteError(w, http.StatusNotFound, "target_not_found", err.Error())
				return
			}
			middleware.WriteError(w, http.StatusBadRequest, "command_error", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, res)
	}
}

func (h *CommandHandlers) HandleConfirmAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}

		var req ConfirmActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		if req.Token == "" {
			middleware.WriteError(w, http.StatusBadRequest, "empty_token", "token de confirmação obrigatório")
			return
		}

		msg, err := h.confirmUC.Execute(r.Context(), owner, req.Token)
		if err != nil {
			if errors.Is(err, command.ErrTokenExpired) {
				middleware.WriteError(w, http.StatusGone, "token_expired", "token de confirmação expirado, repita a ordem")
				return
			}
			if errors.Is(err, command.ErrTokenNotFound) || errors.Is(err, command.ErrTokenAlreadyUsed) {
				middleware.WriteError(w, http.StatusBadRequest, "token_invalid", "token de confirmação inválido ou já utilizado")
				return
			}
			middleware.WriteError(w, http.StatusInternalServerError, "confirm_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": msg,
		})
	}
}
