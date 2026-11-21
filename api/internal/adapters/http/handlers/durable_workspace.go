package handlers

import (
	"errors"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

// HandleGetDurableWorkspace exige o mesmo contexto autenticado usado pela RLS.
func HandleGetDurableWorkspace(uc *app.GetDurableWorkspaceUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}
		id := r.PathValue("id")
		if id == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_card_id", "id do cartão é obrigatório")
			return
		}
		files, err := uc.Execute(r.Context(), owner, id)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "workspace_unavailable", "não foi possível consultar os arquivos")
			return
		}
		if files == nil {
			files = []app.ProjectFile{}
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"card_id": id, "files": files, "count": len(files)})
	}
}

// HandleRunDurableJob executa uma tarefa durável específica na invocação serverless
// usando a identidade autenticada AAL2 do dono (sem armazenar JWT nem ignorar RLS).
func HandleRunDurableJob(queue *app.DurableFactoryQueue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := middleware.GetOwnerPrincipal(r.Context())
		if !ok {
			middleware.WriteError(w, http.StatusUnauthorized, "unauthorized", "identidade do dono ausente")
			return
		}
		jobID := r.PathValue("id")
		if jobID == "" {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_job_id", "id da tarefa é obrigatório")
			return
		}
		job, err := queue.RunJob(r.Context(), owner, jobID)
		if err != nil {
			if errors.Is(err, app.ErrDurableJobPending) {
				middleware.WriteError(w, http.StatusConflict, "job_pending", err.Error())
				return
			}
			if errors.Is(err, app.ErrDurableJobUnavailable) {
				middleware.WriteError(w, http.StatusNotFound, "job_not_found", err.Error())
				return
			}
			middleware.WriteError(w, http.StatusInternalServerError, "job_execution_failed", err.Error())
			return
		}
		if job == nil {
			middleware.WriteJSON(w, http.StatusOK, map[string]any{"id": jobID, "status": "completed"})
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{
			"id":      job.ID,
			"card_id": job.CardID,
			"status":  job.Status,
		})
	}
}
