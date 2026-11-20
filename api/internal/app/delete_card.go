package app

import (
	"context"
	"fmt"
)

type DeleteCardUseCase struct {
	cardRepo  CardRepository
	publisher EventPublisher
	uow       UnitOfWork
}

func NewDeleteCardUseCase(cardRepo CardRepository, publisher EventPublisher, uow ...UnitOfWork) *DeleteCardUseCase {
	uc := &DeleteCardUseCase{cardRepo: cardRepo, publisher: publisher}
	if len(uow) > 0 {
		uc.uow = uow[0]
	}
	return uc
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id do cartão é obrigatório")
	}
	var err error
	if uc.uow != nil {
		err = uc.uow.RunInTx(ctx, func(tx Repositories) error {
			if tx.Cards == nil {
				return fmt.Errorf("repositório de cartões indisponível")
			}
			return tx.Cards.Delete(ctx, id)
		})
	} else if uc.cardRepo != nil {
		err = uc.cardRepo.Delete(ctx, id)
	} else {
		err = fmt.Errorf("repositório de cartões indisponível")
	}
	if err != nil {
		return fmt.Errorf("falha ao excluir cartão: %w", err)
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": id, "action": "deleted"})
		_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)
	}
	return nil
}
