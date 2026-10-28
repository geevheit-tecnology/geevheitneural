package app

import (
	"context"
	"fmt"
)

type DeleteCardUseCase struct {
	cardRepo  CardRepository
	publisher EventPublisher
}

func NewDeleteCardUseCase(cardRepo CardRepository, publisher EventPublisher) *DeleteCardUseCase {
	return &DeleteCardUseCase{cardRepo: cardRepo, publisher: publisher}
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id do cartão é obrigatório")
	}
	if err := uc.cardRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("falha ao excluir cartão: %w", err)
	}
	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicCardsChanged, map[string]string{"card_id": id, "action": "deleted"})
		_ = uc.publisher.Publish(ctx, TopicMoneyWaitingChanged, nil)
	}
	return nil
}
