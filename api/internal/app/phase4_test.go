package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/geevheit/neuralcenter/api/internal/app"
	"github.com/geevheit/neuralcenter/api/internal/domain/card"
	"github.com/geevheit/neuralcenter/api/internal/domain/system"
	"github.com/google/uuid"
)

// Mock repositórios para Fase 4
type mockPhase4Repos struct {
	clients   map[string]*app.ClientDTO
	proposals map[string]*app.ProposalDTO
	cards     map[string]*card.Card
	enriched  map[string]*app.CardListItemDTO
	events    []app.CardEventDTO
	audits    []app.AuditLogDTO
}

func newMockPhase4Repos() *mockPhase4Repos {
	return &mockPhase4Repos{
		clients:   make(map[string]*app.ClientDTO),
		proposals: make(map[string]*app.ProposalDTO),
		cards:     make(map[string]*card.Card),
		enriched:  make(map[string]*app.CardListItemDTO),
	}
}

type mockClientRepo struct{ m *mockPhase4Repos }

func (r *mockClientRepo) FindByName(ctx context.Context, name string) (*app.ClientDTO, error) {
	for _, c := range r.m.clients {
		if strings.EqualFold(c.Name, name) {
			return c, nil
		}
	}
	return nil, nil
}
func (r *mockClientRepo) GetByID(ctx context.Context, id string) (*app.ClientDTO, error) {
	c, ok := r.m.clients[id]
	if !ok {
		return nil, card.ErrEmptyTitle
	}
	return c, nil
}
func (r *mockClientRepo) Create(ctx context.Context, name, contactRef string) (*app.ClientDTO, error) {
	c := &app.ClientDTO{
		ID:         uuid.New().String(),
		Name:       name,
		ContactRef: contactRef,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	r.m.clients[c.ID] = c
	return c, nil
}
func (r *mockClientRepo) Delete(ctx context.Context, id string) error {
	delete(r.m.clients, id)
	return nil
}
func (r *mockClientRepo) Export(ctx context.Context, id string) (*app.ClientExportDTO, error) {
	c, ok := r.m.clients[id]
	if !ok {
		return nil, card.ErrEmptyTitle
	}
	return &app.ClientExportDTO{
		Client:   *c,
		Cards:    []app.CardListItemDTO{},
		Ledger:   []app.LedgerEntryDTO{},
		ExportAt: time.Now(),
	}, nil
}

type mockProposalRepo struct{ m *mockPhase4Repos }

func (r *mockProposalRepo) Create(ctx context.Context, p *app.ProposalDTO) error {
	r.m.proposals[p.AcceptTokenHash] = p
	return nil
}
func (r *mockProposalRepo) GetByTokenHash(ctx context.Context, hash string) (*app.PublicProposalViewDTO, string, error) {
	p, ok := r.m.proposals[hash]
	if !ok {
		return nil, "", card.ErrEmptyTitle
	}
	return &app.PublicProposalViewDTO{
		CardTitle:   "Serviço Teste",
		ClientName:  "Cliente Teste",
		AmountCents: p.AmountCents,
		Currency:    p.Currency,
		Body:        p.Body,
		Status:      p.Status,
	}, p.CardID, nil
}
func (r *mockProposalRepo) Accept(ctx context.Context, hash string) error {
	p, ok := r.m.proposals[hash]
	if !ok {
		return card.ErrEmptyTitle
	}
	p.Status = "aceita"
	now := time.Now()
	p.AcceptedAt = &now
	return nil
}

type mockCardRepoPhase4 struct{ m *mockPhase4Repos }

func (r *mockCardRepoPhase4) GetByID(ctx context.Context, id string) (*card.Card, error) {
	c, ok := r.m.cards[id]
	if !ok {
		return nil, card.ErrEmptyTitle
	}
	return c, nil
}
func (r *mockCardRepoPhase4) GetEnrichedByID(ctx context.Context, id string) (*app.CardListItemDTO, error) {
	return r.m.enriched[id], nil
}
func (r *mockCardRepoPhase4) ListEnriched(ctx context.Context) ([]app.CardListItemDTO, error) {
	var list []app.CardListItemDTO
	for _, v := range r.m.enriched {
		list = append(list, *v)
	}
	return list, nil
}
func (r *mockCardRepoPhase4) ListEvents(ctx context.Context, cardID string) ([]app.CardEventDTO, error) {
	return r.m.events, nil
}
func (r *mockCardRepoPhase4) GetRiskReport(ctx context.Context, cardID string) (*app.RiskReportDTO, error) {
	return nil, nil
}
func (r *mockCardRepoPhase4) Update(ctx context.Context, c *card.Card) error {
	r.m.cards[c.ID] = c
	return nil
}
func (r *mockCardRepoPhase4) Create(ctx context.Context, c *app.CardListItemDTO) error {
	r.m.enriched[c.ID] = c
	r.m.cards[c.ID] = &card.Card{
		ID:        c.ID,
		ClientID:  c.ClientID,
		Stage:     card.Stage(c.Stage),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	return nil
}
func (r *mockCardRepoPhase4) Delete(ctx context.Context, id string) error {
	delete(r.m.cards, id)
	delete(r.m.enriched, id)
	return nil
}

type mockUoWPhase4 struct{ m *mockPhase4Repos }

func (u *mockUoWPhase4) RunInTx(ctx context.Context, fn func(tx app.Repositories) error) error {
	repos := app.Repositories{
		Clients:   &mockClientRepo{m: u.m},
		Proposals: &mockProposalRepo{m: u.m},
		Cards:     &mockCardRepoPhase4{m: u.m},
		CardEvents: &mockCardEventRepoPhase4{m: u.m},
		Audit:      &mockAuditRepoPhase4{m: u.m},
	}
	return fn(repos)
}

type mockCardEventRepoPhase4 struct{ m *mockPhase4Repos }

func (r *mockCardEventRepoPhase4) Create(ctx context.Context, event app.CardEventDTO) error {
	r.m.events = append(r.m.events, event)
	return nil
}

type mockAuditRepoPhase4 struct{ m *mockPhase4Repos }

func (r *mockAuditRepoPhase4) Record(ctx context.Context, log app.AuditLogDTO) error {
	r.m.audits = append(r.m.audits, log)
	return nil
}

type mockPublisherPhase4 struct{}

func (p *mockPublisherPhase4) Publish(ctx context.Context, topic string, payload any) error {
	return nil
}

func TestPhase4_IntakeAndProposals(t *testing.T) {
	ctx := context.Background()
	mockState := newMockPhase4Repos()
	uow := &mockUoWPhase4{m: mockState}
	pub := &mockPublisherPhase4{}

	intakeUC := app.NewIntakeUseCase(uow, pub)
	proposalUC := app.NewProposalUseCases(uow, pub, "https://geevheit.test")
	clientLGPD_UC := app.NewClientLGPDUseCases(uow)

	owner, _ := system.NewOwnerPrincipal("dono-id-01", "aal2")

	t.Run("Honeypot blocks spam bot without saving card", func(t *testing.T) {
		res, err := intakeUC.Execute(ctx, app.IntakeCommand{
			ClientName:  "Spam Bot",
			Contact:     "bot@spam.com",
			Description: "Buy cheap meds",
			Website:     "https://spam.com", // Honeypot preenchido!
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Spam {
			t.Errorf("expected spam to be true")
		}
		if len(mockState.cards) != 0 {
			t.Errorf("spam intake must not create cards in database")
		}
	})

	t.Run("Legitimate intake creates client and card in proposta_aprovar gate", func(t *testing.T) {
		res, err := intakeUC.Execute(ctx, app.IntakeCommand{
			ClientName:  "Padaria São João",
			Contact:     "joao@padaria.com",
			Description: "Cardápio digital mobile e painel de pedidos",
			Website:     "", // Honeypot limpo
			ClientIP:    "192.168.1.100",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.CardID == "" {
			t.Fatalf("expected non-empty card_id")
		}

		c, exists := mockState.cards[res.CardID]
		if !exists {
			t.Fatalf("card not created in repository")
		}
		if c.Stage != card.StagePropostaAprovar {
			t.Errorf("expected stage %s, got %s", card.StagePropostaAprovar, c.Stage)
		}
	})

	t.Run("Owner creates proposal with secure token and client accepts", func(t *testing.T) {
		// Pega o card criado pelo intake
		var cardID string
		for id := range mockState.cards {
			cardID = id
			break
		}

		proposal, err := proposalUC.CreateProposal(ctx, owner, app.CreateProposalCommand{
			CardID:      cardID,
			AmountCents: 450000, // R$ 4.500,00
			Currency:    "BRL",
			Body:        "Proposta de MVP com entrega em 14 dias.",
		})
		if err != nil {
			t.Fatalf("failed to create proposal: %v", err)
		}

		if proposal.AcceptTokenRaw == "" {
			t.Fatalf("expected raw accept token")
		}
		if !strings.Contains(proposal.AcceptURL, proposal.AcceptTokenRaw) {
			t.Errorf("accept URL does not contain token")
		}

		// Cliente acessa a visão pública
		view, err := proposalUC.GetPublicProposal(ctx, proposal.AcceptTokenRaw)
		if err != nil {
			t.Fatalf("failed to get public proposal view: %v", err)
		}
		if view.AmountCents != 450000 {
			t.Errorf("expected 450000 cents, got %d", view.AmountCents)
		}
		if view.Status != "pendente" {
			t.Errorf("expected status pendente, got %s", view.Status)
		}

		// Cliente aceita a proposta -> Cartão DEVE avançar para prompt_aprovar
		err = proposalUC.AcceptProposal(ctx, proposal.AcceptTokenRaw)
		if err != nil {
			t.Fatalf("failed to accept proposal: %v", err)
		}

		c := mockState.cards[cardID]
		if c.Stage != card.StagePromptAprovar {
			t.Errorf("card expected stage %s after acceptance, got %s", card.StagePromptAprovar, c.Stage)
		}
	})

	t.Run("LGPD export and purge compliance", func(t *testing.T) {
		var clientID string
		for id := range mockState.clients {
			clientID = id
			break
		}

		// Export
		export, err := clientLGPD_UC.Export(ctx, owner, clientID)
		if err != nil {
			t.Fatalf("failed to export client data: %v", err)
		}
		if export.Client.Name != "Padaria São João" {
			t.Errorf("unexpected exported client name: %s", export.Client.Name)
		}

		// Purge
		err = clientLGPD_UC.Purge(ctx, owner, clientID)
		if err != nil {
			t.Fatalf("failed to purge client: %v", err)
		}

		_, exists := mockState.clients[clientID]
		if exists {
			t.Errorf("client was not purged")
		}
	})
}
