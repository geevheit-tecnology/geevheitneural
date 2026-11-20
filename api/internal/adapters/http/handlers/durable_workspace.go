package handlers

import (
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
