package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
)

type ListCardsUseCase struct {
	cardRepo CardRepository
}

func NewListCardsUseCase(cardRepo CardRepository) *ListCardsUseCase {
	return &ListCardsUseCase{cardRepo: cardRepo}
}

func (uc *ListCardsUseCase) Execute(ctx context.Context) ([]CardListItemDTO, error) {
	return uc.cardRepo.ListEnriched(ctx)
}

type GetCardDetailsUseCase struct {
	cardRepo CardRepository
}

func NewGetCardDetailsUseCase(cardRepo CardRepository) *GetCardDetailsUseCase {
	return &GetCardDetailsUseCase{cardRepo: cardRepo}
}

func (uc *GetCardDetailsUseCase) Execute(ctx context.Context, id string) (*CardDetailsDTO, error) {
	c, err := uc.cardRepo.GetEnrichedByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("buscar cartão: %w", err)
	}

	events, err := uc.cardRepo.ListEvents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("buscar eventos do cartão: %w", err)
	}

	risk, err := uc.cardRepo.GetRiskReport(ctx, id)
	if err != nil {
		// Não é erro fatal se não houver ficha de risco
		risk = nil
	}

	st := card.Stage(c.Stage)
	isGate := st.IsGate()

	actions := make([]string, 0)
	switch st {
	case card.StageEntrada:
		actions = append(actions, "Mover para proposta")
	case card.StagePropostaAprovar:
		actions = append(actions, "Aprovar e enviar ao cliente")
	case card.StageAguardandoCliente:
		actions = append(actions, "Cliente aceitou")
	case card.StagePromptAprovar:
		actions = append(actions, "Aprovar prompt e liberar execução")
	case card.StageEmExecucao:
		actions = append(actions, "Marcar prévia pronta")
	case card.StageEntregaAprovar:
		actions = append(actions, "Aprovar e entregar ao cliente")
	}

	marginCents := c.PriceCents - c.EstCostCents

	return &CardDetailsDTO{
		Card:        *c,
		Events:      events,
		RiskReport:  risk,
		IsGate:      isGate,
		Actions:     actions,
		MarginCents: marginCents,
	}, nil
}
