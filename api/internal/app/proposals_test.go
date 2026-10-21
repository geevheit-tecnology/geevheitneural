package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/money"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
)

type mockProposalCardRepo struct {
	enriched map[string]*app.CardListItemDTO
	cards    map[string]*card.Card
}

func (m *mockProposalCardRepo) GetByID(ctx context.Context, id string) (*card.Card, error) {
	c, ok := m.cards[id]
	if !ok {
		return nil, fmt.Errorf("cartão não encontrado")
	}
	return c, nil
}

func (m *mockProposalCardRepo) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	c, ok := m.enriched[id]
	if !ok {
		return nil, fmt.Errorf("cartão não encontrado")
	}
	return c, nil
}

func (m *mockProposalCardRepo) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	return nil, nil
}
func (m *mockProposalCardRepo) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	return nil, nil
}
func (m *mockProposalCardRepo) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	return nil, nil
}
func (m *mockProposalCardRepo) Update(ctx context.Context, c *card.Card) error {
	m.cards[c.ID] = c
	return nil
}
func (m *mockProposalCardRepo) Create(ctx context.Context, c *app.CardListItemDTO) error {
	m.enriched[c.ID] = c
	return nil
}

type mockProposalSpecRepo struct{}

func (m *mockProposalSpecRepo) GetByCardID(ctx context.Context, cardID string) (*app.SpecificationDTO, error) {
	return &app.SpecificationDTO{
		CardID: cardID,
		UserStories: []map[string]any{
			{"title": "Módulo de Gestão de Motoristas e Rotas"},
			{"title": "Telemetria de Consumo e Manutenção"},
		},
		ScopeURL: "https://central.geevheit.com/escopo?token=test12345",
	}, nil
}
func (m *mockProposalSpecRepo) Upsert(ctx context.Context, spec *app.SpecificationDTO, tokenHash string) error {
	return nil
}
func (m *mockProposalSpecRepo) GetByTokenHash(ctx context.Context, hash string) (*app.SpecificationDTO, error) {
	return nil, nil
}
func (m *mockProposalSpecRepo) ApproveByClient(ctx context.Context, hash string) error {
	return nil
}

type mockProposalEventRepo struct {
	events []app.CardEventDTO
}

func (m *mockProposalEventRepo) Create(ctx context.Context, e app.CardEventDTO) error {
	m.events = append(m.events, e)
	return nil
}

type mockProposalAuditRepo struct {
	logs []app.AuditLogDTO
}

func (m *mockProposalAuditRepo) Record(ctx context.Context, l app.AuditLogDTO) error {
	m.logs = append(m.logs, l)
	return nil
}

type mockProposalUOW struct {
	repos app.Repositories
}

func (u *mockProposalUOW) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	return fn(u.repos)
}

func TestGenerateProposalDocument(t *testing.T) {
	ctx := context.Background()
	owner := system.OwnerPrincipal{ID: "00000000-0000-0000-0000-000000000001"}

	t.Run("gera documento de proposta com sucesso para orcamento", func(t *testing.T) {
		cardID := "card-orcamento-123"

		cRepo := &mockProposalCardRepo{
			enriched: map[string]*app.CardListItemDTO{
				cardID: {
					ID:         cardID,
					Title:      "Plataforma de Gestão de Frota",
					ClientName: "Transportadora Ágil",
					AppName:    "FrotaAgil",
					Type:       "orcamento",
					Stage:      "proposta_aprovar",
					PriceCents: 1500000,
					Currency:   "BRL",
				},
			},
			cards: map[string]*card.Card{
				cardID: {
					ID:       cardID,
					Title:    "Plataforma de Gestão de Frota",
					Type:     card.TypeOrcamento,
					Stage:    card.StagePropostaAprovar,
					Price:    money.NewBRL(1500000),
				},
			},
		}

		pRepo := newMockPaymentRepo()
		pRepo.orders["po-1"] = &app.PaymentOrderDTO{
			ID:          "po-1",
			CardID:      cardID,
			AmountCents: 1500000,
			PixEMV:      "00020126580014br.gov.bcb.pix0136testemv6304ABCD",
			PixTxID:     "TXID12345",
			Status:      "pending",
		}

		evRepo := &mockProposalEventRepo{}
		auditRepo := &mockProposalAuditRepo{}
		specRepo := &mockProposalSpecRepo{}

		repos := app.Repositories{
			Cards:          cRepo,
			PaymentOrders:  pRepo,
			Specifications: specRepo,
			CardEvents:     evRepo,
			Audit:          auditRepo,
		}
		uow := &mockProposalUOW{repos: repos}
		pub := &testPublisher{}

		uc := app.NewProposalUseCases(uow, pub, "https://central.geevheit.com")

		doc, err := uc.GenerateProposalDocument(ctx, owner, cardID)
		if err != nil {
			t.Fatalf("erro inesperado ao gerar proposta: %v", err)
		}

		if doc.CardID != cardID {
			t.Errorf("esperava card_id %s, obteve %s", cardID, doc.CardID)
		}
		if doc.ClientName != "Transportadora Ágil" {
			t.Errorf("esperava cliente Transportadora Ágil, obteve %s", doc.ClientName)
		}
		if doc.WarrantyDays != 90 {
			t.Errorf("esperava garantia de 90 dias, obteve %d", doc.WarrantyDays)
		}
		if doc.TimelineDays <= 0 {
			t.Errorf("esperava cronograma positivo, obteve %d", doc.TimelineDays)
		}
		if len(doc.Deliverables) == 0 {
			t.Errorf("esperava entregáveis definidos, lista vazia")
		}
		if doc.PixEMV != "00020126580014br.gov.bcb.pix0136testemv6304ABCD" {
			t.Errorf("esperava código PIX EMV preenchido")
		}
		if doc.FormattedWhatsApp == "" {
			t.Errorf("esperava mensagem de WhatsApp formatada")
		}

		// Verifica se gravou evento no histórico do cartão
		if len(evRepo.events) == 0 {
			t.Errorf("esperava evento no histórico do cartão")
		}
		if evRepo.events[len(evRepo.events)-1].Kind != "proposal_document_generated" {
			t.Errorf("esperava kind proposal_document_generated, obteve %s", evRepo.events[len(evRepo.events)-1].Kind)
		}

		// Verifica se registrou auditoria
		if len(auditRepo.logs) == 0 {
			t.Errorf("esperava log de auditoria registrado")
		}
		if auditRepo.logs[len(auditRepo.logs)-1].Action != "proposal.document_generated" {
			t.Errorf("esperava action proposal.document_generated, obteve %s", auditRepo.logs[len(auditRepo.logs)-1].Action)
		}
	})

	t.Run("falha quando cartão não existe", func(t *testing.T) {
		repos := app.Repositories{
			Cards: &mockProposalCardRepo{
				enriched: make(map[string]*app.CardListItemDTO),
				cards:    make(map[string]*card.Card),
			},
		}
		uow := &mockProposalUOW{repos: repos}
		pub := &testPublisher{}
		uc := app.NewProposalUseCases(uow, pub, "https://central.geevheit.com")

		_, err := uc.GenerateProposalDocument(ctx, owner, "inexistente")
		if err == nil {
			t.Errorf("esperava erro para cartão inexistente")
		}
	})
}
