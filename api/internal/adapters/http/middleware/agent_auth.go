package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type agentContextKey string

const AgentPrincipalKey agentContextKey = "agent_principal"

func GetAgentPrincipal(ctx context.Context) (system.AgentPrincipal, bool) {
	val := ctx.Value(AgentPrincipalKey)
	if p, ok := val.(system.AgentPrincipal); ok {
		return p, true
	}
	return system.AgentPrincipal{}, false
}

// RequireAgentAuth valida o Bearer token do agente via hash SHA-256 no banco
func RequireAgentAuth(agentRepo app.AgentRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing_authorization","message":"Header Authorization obrigatório"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid_authorization","message":"Formato Bearer token inválido"}`, http.StatusUnauthorized)
				return
			}

			rawToken := strings.TrimSpace(parts[1])
			if rawToken == "" {
				http.Error(w, `{"error":"empty_token","message":"Token vazio"}`, http.StatusUnauthorized)
				return
			}

			// Calcula SHA-256 do token
			h := sha256.Sum256([]byte(rawToken))
			tokenHash := hex.EncodeToString(h[:])

			agent, err := agentRepo.GetByTokenHash(r.Context(), tokenHash)
			if err != nil || agent == nil {
				http.Error(w, `{"error":"unauthorized_agent","message":"Token de agente não reconhecido"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AgentPrincipalKey, *agent)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
