package app

import (
	"context"
	"fmt"

	"github.com/geevheit/neuralcenter/api/internal/domain/card"
)

type ListCardsUseCase struct {
	cardRepo CardRepository
	uow      UnitOfWork
}

func NewListCardsUseCase(cardRepo CardRepository, uow ...UnitOfWork) *ListCardsUseCase {
	uc := &ListCardsUseCase{cardRepo: cardRepo}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *ListCardsUseCase) Execute(ctx context.Context) ([]CardListItemDTO, error) {
	if uc.uow != nil {
		var items []CardListItemDTO
		err := uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Cards == nil {
				return fmt.Errorf("repositório de cartões indisponível")
			}
			var err error
			items, err = tx.Cards.ListEnriched(ctx)
			return err
		})
		return items, err
	}
	return uc.cardRepo.ListEnriched(ctx)
}

type GetCardDetailsUseCase struct {
	cardRepo CardRepository
	uow      UnitOfWork
}

func NewGetCardDetailsUseCase(cardRepo CardRepository, uow ...UnitOfWork) *GetCardDetailsUseCase {
	uc := &GetCardDetailsUseCase{cardRepo: cardRepo}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *GetCardDetailsUseCase) Execute(ctx context.Context, id string) (*CardDetailsDTO, error) {
	var (
		c      *CardListItemDTO
		events []CardEventDTO
		risk   *RiskReportDTO
		err    error
	)
	load := func(repo CardRepository) error {
		c, err = repo.GetEnrichedByID(ctx, id)
		if err != nil {
			return fmt.Errorf("buscar cartão: %w", err)
		}
		events, err = repo.ListEvents(ctx, id)
		if err != nil {
			return fmt.Errorf("buscar eventos do cartão: %w", err)
		}
		risk, err = repo.GetRiskReport(ctx, id)
		if err != nil {
			risk = nil
		}
		return nil
	}
	if uc.uow != nil {
		if txErr := uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Cards == nil {
				return fmt.Errorf("repositório de cartões indisponível")
			}
			return load(tx.Cards)
		}); txErr != nil {
			return nil, txErr
		}
	} else {
		if err := load(uc.cardRepo); err != nil {
			return nil, err
		}
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
