package app

import (
	"context"

	"github.com/geevheit/neuralcenter/api/internal/domain/money"
)

type MoneyWaitingResult struct {
	Total      money.Money
	TotalCents int64
	Count      int
	Formatted  string
}

type GetMoneyWaitingUseCase struct {
	reader MoneyWaitingReader
}

func NewGetMoneyWaitingUseCase(reader MoneyWaitingReader) *GetMoneyWaitingUseCase {
	return &GetMoneyWaitingUseCase{
		reader: reader,
	}
}

// Execute retorna a soma de dinheiro parado em portões do dono e o número de decisões pendentes.
func (uc *GetMoneyWaitingUseCase) Execute(ctx context.Context) (MoneyWaitingResult, error) {
	totalCents, count, err := uc.reader.GetMoneyWaiting(ctx)
	if err != nil {
		return MoneyWaitingResult{}, err
	}

	m := money.NewBRL(totalCents)
	return MoneyWaitingResult{
		Total:      m,
		TotalCents: totalCents,
		Count:      count,
		Formatted:  m.FormatBRL(),
	}, nil
}
