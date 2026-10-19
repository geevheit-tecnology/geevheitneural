package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/app"
)

type IntakeHandler struct {
	uc *app.IntakeUseCase
}

func NewIntakeHandler(uc *app.IntakeUseCase) *IntakeHandler {
	return &IntakeHandler{uc: uc}
}

func (h *IntakeHandler) SubmitIntake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var cmd app.IntakeCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "payload json inválido", http.StatusBadRequest)
		return
	}

	// Extração do IP para hashing seguro de auditoria/antifraude
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	} else {
		parts := strings.Split(clientIP, ",")
		clientIP = strings.TrimSpace(parts[0])
	}
	cmd.ClientIP = clientIP

	res, err := h.uc.Execute(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
