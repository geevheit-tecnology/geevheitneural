package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/auth"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	OwnerKey     contextKey = "owner_principal"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			reqID, _ := r.Context().Value(RequestIDKey).(string)

			// Não loga query strings nem corpo para garantir conformidade LGPD e evitar vazamento de PII
			logger.Info("http_request",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.statusCode),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID, _ := r.Context().Value(RequestIDKey).(string)
					logger.Error("panic_recovered",
						slog.String("request_id", reqID),
						slog.Any("error", err),
					)
					WriteError(w, http.StatusInternalServerError, "internal_error", "erro interno do servidor")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowed := false
				for _, o := range allowedOrigins {
					if o == "*" || strings.EqualFold(o, origin) {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireOwnerAuth(validator *auth.TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				if qToken := r.URL.Query().Get("token"); qToken != "" {
					authHeader = "Bearer " + qToken
				}
			}
			if authHeader == "" {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "token de autenticação ausente")
				return
			}

			res, err := validator.ValidateToken(authHeader)
			if err != nil {
				if errors.Is(err, auth.ErrNotOwner) {
					WriteError(w, http.StatusForbidden, "forbidden", "acesso exclusivo ao dono do sistema")
					return
				}
				if errors.Is(err, auth.ErrMFAEnforced) {
					WriteError(w, http.StatusForbidden, "mfa_required", "autenticação 2FA (AAL2) obrigatória")
					return
				}
				WriteError(w, http.StatusUnauthorized, "invalid_token", err.Error())
				return
			}

			// Injeta OwnerPrincipal e ClaimsJSON no contexto para os casos de uso e RLS
			ctx := context.WithValue(r.Context(), OwnerKey, res.Owner)
			ctx = postgres.ContextWithClaims(ctx, res.ClaimsJSON)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetOwnerPrincipal(ctx context.Context) (system.OwnerPrincipal, bool) {
	p, ok := ctx.Value(OwnerKey).(system.OwnerPrincipal)
	return p, ok
}
