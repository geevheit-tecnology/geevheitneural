package app

import (
	"context"
	"fmt"
)

type GetFinancesOverviewUseCase struct {
	financeRepo FinanceRepository
}

func NewGetFinancesOverviewUseCase(financeRepo FinanceRepository) *GetFinancesOverviewUseCase {
	return &GetFinancesOverviewUseCase{financeRepo: financeRepo}
}

func (uc *GetFinancesOverviewUseCase) Execute(ctx context.Context) (*FinancesOverviewDTO, error) {
	overview, err := uc.financeRepo.GetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("buscar visão financeira: %w", err)
	}

	servicesCount := len(overview.CostsByService)
	if servicesCount > 3 {
		overview.Suggestion = fmt.Sprintf("São %d serviços diferentes. Consolidar em 2 ou 3 reduz custo e contas para vigiar.", servicesCount)
	} else if servicesCount > 0 {
		overview.Suggestion = "Serviços sob controle. Monitore a margem dos novos apps antes de contratar novas ferramentas."
	} else {
		overview.Suggestion = "Cadastre custos de hospedagem e banco dos seus apps para ver a sugestão de consolidação."
	}

	return overview, nil
}

type CreateLedgerEntryUseCase struct {
	financeRepo FinanceRepository
	publisher   EventPublisher
}

func NewCreateLedgerEntryUseCase(financeRepo FinanceRepository, publisher EventPublisher) *CreateLedgerEntryUseCase {
	return &CreateLedgerEntryUseCase{financeRepo: financeRepo, publisher: publisher}
}

func (uc *CreateLedgerEntryUseCase) Execute(ctx context.Context, cmd CreateLedgerEntryCommand) (*LedgerEntryDTO, error) {
	if cmd.Kind != "receita" && cmd.Kind != "custo" {
		return nil, fmt.Errorf("tipo deve ser 'receita' ou 'custo'")
	}
	if cmd.AmountCents <= 0 {
		return nil, fmt.Errorf("valor deve ser maior que zero")
	}

	entry, err := uc.financeRepo.CreateEntry(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicFinancesChanged, entry)
	}

	return entry, nil
}
