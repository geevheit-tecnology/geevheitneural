package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type AppHandlers struct {
	listUC   *app.ListAppsUseCase
	createUC *app.CreateAppUseCase
	updateUC *app.UpdateAppUseCase
	deleteUC *app.DeleteAppUseCase
}

func NewAppHandlers(
	listUC *app.ListAppsUseCase,
	createUC *app.CreateAppUseCase,
	updateUC *app.UpdateAppUseCase,
	deleteUC *app.DeleteAppUseCase,
) *AppHandlers {
	return &AppHandlers{
		listUC:   listUC,
		createUC: createUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

func (h *AppHandlers) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apps, err := h.listUC.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_apps_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"apps":  apps,
			"count": len(apps),
		})
	}
}

func (h *AppHandlers) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cmd app.CreateAppCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		created, err := h.createUC.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "create_app_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, created)
	}
}

func (h *AppHandlers) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_id", "id do app é obrigatório")
			return
		}

		var cmd app.UpdateAppCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}
		cmd.ID = id

		if err := h.updateUC.Execute(r.Context(), cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "update_app_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "id": id})
	}
}

func (h *AppHandlers) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_id", "id do app é obrigatório")
			return
		}

		if err := h.deleteUC.Execute(r.Context(), id); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "delete_app_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "id": id})
	}
}
