package app_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/portal"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type mockPortalTokenRepo struct {
	tokensByHash   map[string]*portal.PortalToken
	tokensByCardID map[string]*portal.PortalToken
}

func newMockPortalTokenRepo() *mockPortalTokenRepo {
	return &mockPortalTokenRepo{
		tokensByHash:   make(map[string]*portal.PortalToken),
		tokensByCardID: make(map[string]*portal.PortalToken),
	}
}

func (r *mockPortalTokenRepo) Create(ctx context.Context, token *portal.PortalToken) error {
	r.tokensByHash[token.TokenHash] = token
	r.tokensByCardID[token.CardID] = token
	return nil
}

func (r *mockPortalTokenRepo) GetByHash(ctx context.Context, tokenHash string) (*portal.PortalToken, error) {
	t, ok := r.tokensByHash[tokenHash]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *mockPortalTokenRepo) GetByCardID(ctx context.Context, cardID string) (*portal.PortalToken, error) {
	t, ok := r.tokensByCardID[cardID]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (r *mockPortalTokenRepo) Update(ctx context.Context, token *portal.PortalToken) error {
	r.tokensByHash[token.TokenHash] = token
	r.tokensByCardID[token.CardID] = token
	return nil
}

func (r *mockPortalTokenRepo) RevokeByCardID(ctx context.Context, cardID string, revokedAt time.Time) error {
	if t, ok := r.tokensByCardID[cardID]; ok {
		t.Revoke(revokedAt)
	}
	return nil
}

type mockAuditRepo struct {
	records []app.AuditLogDTO
}

func (r *mockAuditRepo) Record(ctx context.Context, log app.AuditLogDTO) error {
	r.records = append(r.records, log)
	return nil
}

type mockCardRepoPortal struct {
	card *card.Card
}

func (r *mockCardRepoPortal) GetByID(ctx context.Context, id string) (*card.Card, error) {
	if r.card != nil && r.card.ID == id {
		return r.card, nil
	}
	return nil, nil
}
func (r *mockCardRepoPortal) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPortal) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPortal) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPortal) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPortal) Update(ctx context.Context, c *card.Card) error {
	r.card = c
	return nil
}
func (r *mockCardRepoPortal) Create(ctx context.Context, c *app.CardListItemDTO) error { return nil }
func (r *mockCardRepoPortal) Delete(ctx context.Context, id string) error               { return nil }

type mockClientRepoPortal struct{}

func (r *mockClientRepoPortal) FindByName(ctx context.Context, name string) (*app.ClientDTO, error) {
	return nil, nil
}
func (r *mockClientRepoPortal) GetByID(ctx context.Context, id string) (*app.ClientDTO, error) {
	return &app.ClientDTO{ID: id, Name: "Acme Corp", ContactRef: "contato@acme.com"}, nil
}
func (r *mockClientRepoPortal) Create(ctx context.Context, name, contactRef string) (*app.ClientDTO, error) {
	return nil, nil
}
func (r *mockClientRepoPortal) Delete(ctx context.Context, id string) error { return nil }
func (r *mockClientRepoPortal) Export(ctx context.Context, id string) (*app.ClientExportDTO, error) {
	return nil, nil
}

type mockAppRepoPortal struct{}

func (r *mockAppRepoPortal) ListWithFinancials(ctx context.Context) ([]app.AppDTO, error) {
	return nil, nil
}
func (r *mockAppRepoPortal) GetByID(ctx context.Context, id string) (*app.AppDTO, error) {
	return &app.AppDTO{ID: id, Name: "acme-store", Hosting: "Vercel", Database: "Supabase"}, nil
}
func (r *mockAppRepoPortal) Create(ctx context.Context, cmd app.CreateAppCommand) (*app.AppDTO, error) {
	return nil, nil
}
func (r *mockAppRepoPortal) Update(ctx context.Context, cmd app.UpdateAppCommand) error {
	return nil
}
func (r *mockAppRepoPortal) Delete(ctx context.Context, id string) error { return nil }

type portalTestUOW struct {
	repos app.Repositories
}

func (u *portalTestUOW) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	return fn(u.repos)
}

func TestClientPortal_GenerationAndStatus(t *testing.T) {
	tokenRepo := newMockPortalTokenRepo()
	cardRepo := &mockCardRepoPortal{
		card: &card.Card{
			ID:       "card-portal-1",
			OwnerID:  "owner-1",
			ClientID: "client-1",
			Title:    "Portal E-Commerce",
			Price:    money.NewBRL(500000), // R$ 5.000,00
			EstCost:  money.NewBRL(50000),
			Stage:    card.StageEntregaAprovar,
		},
	}
	auditRepo := &mockAuditRepo{}

	uow := &portalTestUOW{
		repos: app.Repositories{
			Cards:        cardRepo,
			PortalTokens: tokenRepo,
			Clients:      &mockClientRepoPortal{},
			Apps:         &mockAppRepoPortal{},
			Audit:        auditRepo,
		},
	}

	uc := app.NewPortalUseCases(uow, &testPublisher{}, "https://geevheit.com")
	owner := system.OwnerPrincipal{ID: "owner-1"}

	// 1. Gera link do portal
	rawToken, portalURL, err := uc.GeneratePortalLink(context.Background(), owner, "card-portal-1", 15)
	if err != nil {
		t.Fatalf("erro inesperado ao gerar link do portal: %v", err)
	}

	if len(rawToken) != 64 { // 32 bytes em hexadecimal = 64 caracteres
		t.Errorf("token deve ter 64 caracteres hexadecimais, obteve %d", len(rawToken))
	}

	if !strings.Contains(portalURL, "/portal?token="+rawToken) {
		t.Errorf("url do portal inválida: %s", portalURL)
	}

	// 2. Verifica status pelo dono
	status, err := uc.GetPortalTokenStatus(context.Background(), owner, "card-portal-1")
	if err != nil {
		t.Fatalf("erro ao obter status do token: %v", err)
	}
	if !status.IsActive {
		t.Errorf("status deveria ser ativo")
	}

	// 3. Revoga o link
	err = uc.RevokePortalLink(context.Background(), owner, "card-portal-1")
	if err != nil {
		t.Fatalf("erro ao revogar link: %v", err)
	}

	// 4. Token revogado deve falhar com erro genérico
	_, err = uc.GetClientPortal(context.Background(), rawToken, "192.168.1.1", "curl/8.0")
	if err != app.ErrInvalidPortalToken {
		t.Errorf("esperava ErrInvalidPortalToken após revogação, obteve %v", err)
	}
}

func TestClientPortal_StrictDataIsolation(t *testing.T) {
	tokenRepo := newMockPortalTokenRepo()
	cardRepo := &mockCardRepoPortal{
		card: &card.Card{
			ID:       "card-secret-1",
			OwnerID:  "owner-1",
			ClientID: "client-1",
			Title:    "Sistema de Gestão Financeira",
			Price:    money.NewBRL(1200000), // R$ 12.000,00
			EstCost:  money.NewBRL(150000),  // R$ 1.500,00 (Margem 10.500)
			Stage:    card.StageEntregaAprovar,
		},
	}
	auditRepo := &mockAuditRepo{}

	uow := &portalTestUOW{
		repos: app.Repositories{
			Cards:        cardRepo,
			PortalTokens: tokenRepo,
			Clients:      &mockClientRepoPortal{},
			Apps:         &mockAppRepoPortal{},
			Audit:        auditRepo,
		},
	}

	uc := app.NewPortalUseCases(uow, &testPublisher{}, "https://geevheit.com")
	owner := system.OwnerPrincipal{ID: "owner-1"}

	rawToken, _, err := uc.GeneratePortalLink(context.Background(), owner, "card-secret-1", 30)
	if err != nil {
		t.Fatalf("falha ao gerar link: %v", err)
	}

	// Obtém a visão pública do cliente
	clientView, err := uc.GetClientPortal(context.Background(), rawToken, "200.100.50.25", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("falha ao carregar visão do portal: %v", err)
	}

	if clientView.ProjectTitle != "Sistema de Gestão Financeira" {
		t.Errorf("título do projeto incorreto: %s", clientView.ProjectTitle)
	}
	if clientView.ClientName != "Acme Corp" {
		t.Errorf("nome do cliente incorreto: %s", clientView.ClientName)
	}
	if len(clientView.Stages) != 4 {
		t.Errorf("esperava 4 etapas de linha do tempo, obteve %d", len(clientView.Stages))
	}
	if !clientView.CanAccept {
		t.Errorf("em entrega_aprovar, CanAccept deveria ser true")
	}

	// TESTE DE ISOLAMENTO RIGOROSO: Serializa para JSON e garante ausência de campos confidenciais
	jsonBytes, err := json.Marshal(clientView)
	if err != nil {
		t.Fatalf("falha ao serializar DTO para json: %v", err)
	}
	jsonStr := string(jsonBytes)

	forbiddenTerms := []string{
		"price_cents",
		"est_cost_cents",
		"margin_cents",
		"1200000",
		"150000",
		"spend_limit",
		"agent",
		"token_hash",
		"owner_id",
	}

	for _, term := range forbiddenTerms {
		if strings.Contains(jsonStr, term) {
			t.Errorf("VIOLAÇÃO DE ISOLAMENTO: termo confidencial '%s' vazou no DTO do cliente: %s", term, jsonStr)
		}
	}
}

func TestClientPortal_AcceptDeliveryIdempotenceAndAudit(t *testing.T) {
	tokenRepo := newMockPortalTokenRepo()
	appID := "app-123"
	cardRepo := &mockCardRepoPortal{
		card: &card.Card{
			ID:       "card-accept-1",
			OwnerID:  "owner-1",
			ClientID: "client-1",
			AppID:    &appID,
			Title:    "App Mobile de Entregas",
			Price:    money.NewBRL(800000),
			EstCost:  money.NewBRL(100000),
			Stage:    card.StageEntregaAprovar,
		},
	}
	auditRepo := &mockAuditRepo{}

	uow := &portalTestUOW{
		repos: app.Repositories{
			Cards:        cardRepo,
			PortalTokens: tokenRepo,
			Clients:      &mockClientRepoPortal{},
			Apps:         &mockAppRepoPortal{},
			Audit:        auditRepo,
		},
	}

	uc := app.NewPortalUseCases(uow, &testPublisher{}, "https://geevheit.com")
	owner := system.OwnerPrincipal{ID: "owner-1"}

	rawToken, _, err := uc.GeneratePortalLink(context.Background(), owner, "card-accept-1", 30)
	if err != nil {
		t.Fatalf("falha ao gerar link: %v", err)
	}

	// 1. Tentar aceitar sem identificar quem aceita deve falhar
	_, err = uc.AcceptClientDelivery(context.Background(), rawToken, "", "10.0.0.1", "Chrome")
	if err != portal.ErrMissingAcceptedBy {
		t.Errorf("esperava ErrMissingAcceptedBy, obteve %v", err)
	}

	// 2. Aceitar formalmente com sucesso
	view, err := uc.AcceptClientDelivery(context.Background(), rawToken, "Carlos Diretor de TI", "187.50.10.5", "Chrome/120.0")
	if err != nil {
		t.Fatalf("falha ao registrar aceite formal: %v", err)
	}

	if !view.IsAccepted {
		t.Errorf("IsAccepted deveria ser true após aceite")
	}
	if view.AcceptedBy != "Carlos Diretor de TI" {
		t.Errorf("AcceptedBy incorreto: %s", view.AcceptedBy)
	}
	if view.CanAccept {
		t.Errorf("CanAccept deveria ser false após já aceito")
	}

	// 3. Auditoria deve conter o registro com hash de estado
	if len(auditRepo.records) != 1 {
		t.Fatalf("esperava 1 registro em audit_log, obteve %d", len(auditRepo.records))
	}
	logEntry := auditRepo.records[0]
	if logEntry.Action != "client_portal_delivery_accepted" {
		t.Errorf("ação de auditoria incorreta: %s", logEntry.Action)
	}
	if logEntry.ActorType != "client" {
		t.Errorf("actor_type deveria ser client, obteve %s", logEntry.ActorType)
	}
	if stateHash, ok := logEntry.Meta["state_hash"].(string); !ok || len(stateHash) == 0 {
		t.Errorf("esperava state_hash gravado na auditoria")
	}

	// 4. Teste de Idempotência: Tentar aceitar novamente deve falhar
	_, err = uc.AcceptClientDelivery(context.Background(), rawToken, "Outra Pessoa", "187.50.10.5", "Chrome/120.0")
	if err != portal.ErrAlreadyAccepted {
		t.Errorf("esperava ErrAlreadyAccepted na segunda tentativa, obteve %v", err)
	}
}
