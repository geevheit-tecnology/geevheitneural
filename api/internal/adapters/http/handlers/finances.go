package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/app"
)

type FinanceHandlers struct {
	overviewUC    *app.GetFinancesOverviewUseCase
	createEntryUC *app.CreateLedgerEntryUseCase
}

func NewFinanceHandlers(
	overviewUC *app.GetFinancesOverviewUseCase,
	createEntryUC *app.CreateLedgerEntryUseCase,
) *FinanceHandlers {
	return &FinanceHandlers{
		overviewUC:    overviewUC,
		createEntryUC: createEntryUC,
	}
}

func (h *FinanceHandlers) Overview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview, err := h.overviewUC.Execute(r.Context())
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "finances_overview_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusOK, overview)
	}
}

func (h *FinanceHandlers) CreateEntry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cmd app.CreateLedgerEntryCommand
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid_payload", "corpo da requisição inválido")
			return
		}

		if cmd.Currency == "" {
			cmd.Currency = "BRL"
		}

		entry, err := h.createEntryUC.Execute(r.Context(), cmd)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "create_entry_failed", err.Error())
			return
		}

		middleware.WriteJSON(w, http.StatusCreated, entry)
	}
}
