package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

func HandleListPlaybooks(uc *app.ListPlaybooksUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_playbooks_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"playbooks": list})
	}
}

func HandleListModules(uc *app.ListModulesUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_modules_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"modules": list})
	}
}

func HandleGetFactoryScoreboard(uc *app.GetFactoryScoreboardUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		board, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "scoreboard_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, board)
	}
}

func HandleEstimatePrice(uc *app.EstimatePriceUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cmd app.EstimatePriceCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "parâmetros de estimativa inválidos")
			return
		}

		res, err := uc.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "estimate_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, res)
	}
}

func HandleCreatePlaybook(uc *app.CreatePlaybookUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p app.PlaybookDTO
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "dados do playbook inválidos")
			return
		}
		res, err := uc.Execute(r.Context(), p)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "create_playbook_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusCreated, res)
	}
}

func HandleCreateModule(uc *app.CreateModuleUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m app.ModuleDTO
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "dados do módulo inválidos")
			return
		}
		res, err := uc.Execute(r.Context(), m)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "create_module_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusCreated, res)
	}
}

func HandleGetCardPreviewInfo(uc *app.GetCardPreviewInfoUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		res, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "preview_info_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, res)
	}
}

func HandleExecuteRollback(uc *app.ExecuteRollbackUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		var body struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		res, err := uc.Execute(r.Context(), cardID, body.Reason)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "rollback_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, res)
	}
}

func HandleListProjectMemory(uc *app.ListProjectMemoryUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		items, err := uc.Execute(r.Context(), cardID)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "list_memory_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, map[string]any{"memories": items})
	}
}

func HandleAddProjectMemory(uc *app.AddProjectMemoryUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cardID := r.PathValue("id")
		var m app.ProjectMemoryDTO
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "dados de memória inválidos")
			return
		}
		res, err := uc.Execute(r.Context(), cardID, m)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "add_memory_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusCreated, res)
	}
}

func HandleRunFactoryEvals(uc *app.RunFactoryEvalsUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := uc.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "evals_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusOK, res)
	}
}

func HandleMonitorAppHealth(uc *app.MonitorAppHealthUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := r.PathValue("id")
		var cmd app.MonitorAppHealthCommand
		_ = json.NewDecoder(r.Body).Decode(&cmd)
		if cmd.AppID == "" {
			cmd.AppID = appID
		}
		res, err := uc.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "monitor_failed", err.Error())
			return
		}
		middleware.WriteJSON(w, http.StatusCreated, map[string]any{
			"success":  true,
			"bug_card": res,
			"message":  "Agente de Suporte abriu automaticamente um cartão de Bug com logs e prioridade sugerida.",
		})
	}
}
