package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/adapters/auth"
	"github.com/geevheit/neuralcenter/api/internal/adapters/http/handlers"
	"github.com/geevheit/neuralcenter/api/internal/adapters/http/middleware"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres"
	"github.com/geevheit/neuralcenter/api/internal/adapters/postgres/sqlcgen"
	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rollbackDeleteUOW struct{ tx pgx.Tx }

func (u rollbackDeleteUOW) RunInTx(ctx context.Context, fn func(app.Repositories) error) error {
	tx, err := u.tx.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	claims, ok := ctx.Value(postgres.ClaimsContextKey).(string)
	if !ok || claims == "" {
		return fmt.Errorf("claims ausentes")
	}
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE authenticated"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", claims); err != nil {
		return err
	}
	q := sqlcgen.New(tx)
	if err = fn(app.Repositories{Cards: postgres.NewCardRepository(q), Apps: postgres.NewAppRepository(q), Agents: postgres.NewAgentRepository(q), Audit: postgres.NewAuditRepository(q)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func TestDeleteCardHTTPPostgresRollback(t *testing.T) {
	url := os.Getenv("GEEV_FACTORY_ROLLBACK_DATABASE_URL")
	if os.Getenv("GEEV_FACTORY_ROLLBACK_TEST") != "1" || url == "" {
		t.Skip("isolated rollback database not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal("invalid test database")
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	migration, err := os.ReadFile("../../../../../db/migrations/00011_allow_card_deletion.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, strings.SplitN(string(migration), "-- +goose Down", 2)[0]); err != nil {
		t.Fatal(err)
	}
	owner, cardID, client := uuid.NewString(), uuid.NewString(), uuid.NewString()
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE authenticated"); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, owner)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.clients (id, name) VALUES ($1, 'synthetic')", client); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title) VALUES ($1, 'orcamento', $2, 'synthetic')", cardID, client); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.card_events (card_id, kind) VALUES ($1, 'synthetic')", cardID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.approvals (card_id, gate, decision) VALUES ($1, 'prompt_aprovar', 'aprovado')", cardID); err != nil {
		t.Fatal(err)
	}
	const secret = "synthetic-test-secret-at-least-32-bytes"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": owner, "aal": "aal2", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	h := middleware.RequireOwnerAuth(auth.NewTokenValidator(owner, secret))(handlers.HandleDeleteCard(app.NewDeleteCardUseCase(nil, nil, rollbackDeleteUOW{tx})))
	call := func(token, id string) int {
		req := httptest.NewRequest(http.MethodDelete, "/v1/cards/"+id, nil)
		req.SetPathValue("id", id)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w.Code
	}
	if got := call("", cardID); got != http.StatusUnauthorized {
		t.Fatalf("missing auth: %d", got)
	}
	if got := call(token, cardID); got != http.StatusOK {
		t.Fatalf("delete: %d", got)
	}
	var count int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM geev.cards WHERE id=$1", cardID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("card remains: %v", err)
	}
	if got := call(token, cardID); got == http.StatusOK {
		t.Fatal("missing card reported deleted")
	}
	other, otherClient, otherCard := uuid.NewString(), uuid.NewString(), uuid.NewString()
	if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, other)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.clients (id, name) VALUES ($1, 'other synthetic')", otherClient); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO geev.cards (id, type, client_id, title) VALUES ($1, 'orcamento', $2, 'other synthetic')", otherCard, otherClient); err != nil {
		t.Fatal(err)
	}
	if got := call(token, otherCard); got == http.StatusOK {
		t.Fatal("other owner's card reported deleted")
	}
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM geev.cards WHERE id=$1", otherCard).Scan(&count); err != nil || count != 1 {
		t.Fatalf("other owner's card removed: %v", err)
	}
	otherApp := uuid.NewString()
	if _, err := tx.Exec(ctx, "INSERT INTO geev.apps (id, name) VALUES ($1, 'other app')", otherApp); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", fmt.Sprintf(`{"sub":%q,"aal":"aal2"}`, owner)); err != nil {
		t.Fatal(err)
	}
	appRepo := postgres.NewAppRepository(sqlcgen.New(tx))
	appUOW := rollbackDeleteUOW{tx}
	appHandlers := handlers.NewAppHandlers(app.NewListAppsUseCase(appRepo, appUOW), app.NewCreateAppUseCase(appRepo, nil, appUOW), app.NewUpdateAppUseCase(appRepo, nil, appUOW), app.NewDeleteAppUseCase(appRepo, nil, appUOW))
	ownerAuth := middleware.RequireOwnerAuth(auth.NewTokenValidator(owner, secret))
	createApp := ownerAuth(appHandlers.Create())
	createReq := httptest.NewRequest(http.MethodPost, "/v1/apps", strings.NewReader(`{"name":"Aplicativo sintético"}`))
	createReq.Header.Set("Authorization", "Bearer "+token)
	created := httptest.NewRecorder()
	createApp.ServeHTTP(created, createReq)
	if created.Code != http.StatusCreated {
		t.Fatalf("create app: %d", created.Code)
	}
	var createdApp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdApp); err != nil || createdApp.ID == "" {
		t.Fatalf("created app missing ID: %v", err)
	}
	appID := createdApp.ID
	var storedOwner string
	if err := tx.QueryRow(ctx, "SELECT owner_id::text FROM geev.apps WHERE id=$1", appID).Scan(&storedOwner); err != nil || storedOwner != owner {
		t.Fatalf("incorrect app owner: %v", err)
	}
	deleteApp := ownerAuth(appHandlers.Delete())
	listApps := ownerAuth(appHandlers.List())
	otherReq := httptest.NewRequest(http.MethodDelete, "/v1/apps/"+otherApp, nil)
	otherReq.SetPathValue("id", otherApp)
	otherReq.Header.Set("Authorization", "Bearer "+token)
	otherResult := httptest.NewRecorder()
	deleteApp.ServeHTTP(otherResult, otherReq)
	if otherResult.Code == http.StatusOK {
		t.Fatal("other owner's app reported deleted")
	}
	appCall := func(handler http.Handler, method string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/v1/apps/"+appID, nil)
		req.SetPathValue("id", appID)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	if got := appCall(deleteApp, http.MethodDelete); got.Code != http.StatusOK {
		t.Fatalf("delete app: %d", got.Code)
	}
	if got := appCall(deleteApp, http.MethodDelete); got.Code == http.StatusOK {
		t.Fatal("missing app reported deleted")
	}
	listed := appCall(listApps, http.MethodGet)
	if listed.Code != http.StatusOK {
		t.Fatalf("list apps: %d", listed.Code)
	}
	var response struct {
		Apps []struct {
			ID string `json:"id"`
		} `json:"apps"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	for _, item := range response.Apps {
		if item.ID == appID || item.ID == otherApp {
			t.Fatal("deleted app returned in list")
		}
	}
	agentHandlers := handlers.NewAgentHandler(app.NewAgentUseCases(appUOW, nil))
	agentList := ownerAuth(http.HandlerFunc(agentHandlers.ListOwnerAgents))
	agentReq := httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	agentReq.Header.Set("Authorization", "Bearer "+token)
	agentResponse := httptest.NewRecorder()
	agentList.ServeHTTP(agentResponse, agentReq)
	if agentResponse.Code != http.StatusOK {
		t.Fatalf("list agents: %d %s", agentResponse.Code, agentResponse.Body.String())
	}
	var agents struct {
		Agents []struct {
			ID     string `json:"id"`
			Role   string `json:"role"`
			Active bool   `json:"active"`
		} `json:"agents"`
	}
	if err := json.Unmarshal(agentResponse.Body.Bytes(), &agents); err != nil || len(agents.Agents) != 7 {
		t.Fatalf("expected seven agents: %v", err)
	}
	var devID string
	for _, a := range agents.Agents {
		if a.Active {
			t.Fatal("agent started active")
		}
		if a.Role == "dev" {
			devID = a.ID
		}
	}
	if devID == "" {
		t.Fatal("developer agent missing")
	}
	agentToggle := ownerAuth(http.HandlerFunc(agentHandlers.ToggleOwnerAgent))
	toggleReq := httptest.NewRequest(http.MethodPost, "/v1/agents/"+devID+"/toggle", strings.NewReader(`{"active":true}`))
	toggleReq.SetPathValue("id", devID)
	toggleReq.Header.Set("Authorization", "Bearer "+token)
	toggleResponse := httptest.NewRecorder()
	agentToggle.ServeHTTP(toggleResponse, toggleReq)
	if toggleResponse.Code != http.StatusOK {
		t.Fatalf("toggle agent: %d %s", toggleResponse.Code, toggleResponse.Body.String())
	}
	var active bool
	if err := tx.QueryRow(ctx, "SELECT active FROM geev.agents WHERE id=$1", devID).Scan(&active); err != nil || !active {
		t.Fatalf("agent state not saved: %v", err)
	}
}
