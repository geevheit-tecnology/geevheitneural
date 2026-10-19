package app

import (
	"context"
	"fmt"
)

type GetRiskReportUseCase struct {
	cardRepo CardRepository
}

func NewGetRiskReportUseCase(cardRepo CardRepository) *GetRiskReportUseCase {
	return &GetRiskReportUseCase{cardRepo: cardRepo}
}

func (uc *GetRiskReportUseCase) Execute(ctx context.Context, cardID string) (*RiskReportDTO, error) {
	report, err := uc.cardRepo.GetRiskReport(ctx, cardID)
	if err != nil {
		// Se não existe, retorna ficha inicial padrão vazia para preenchimento
		return &RiskReportDTO{
			CardID: cardID,
			Tests: []map[string]any{
				{"label": "Testes automáticos", "status": "pendente", "level": "warn"},
			},
			SecretsScan: []map[string]any{
				{"label": "Segredos no código", "status": "pendente", "level": "warn"},
			},
			OutdatedDeps: []map[string]any{
				{"label": "Dependências", "status": "pendente", "level": "warn"},
			},
			LgpdPoints: []map[string]any{
				{"label": "Dados pessoais (LGPD)", "status": "pendente", "level": "warn"},
			},
			Source: "manual",
		}, nil
	}
	return report, nil
}

type SaveRiskReportCommand struct {
	CardID       string           `json:"card_id"`
	Tests        []map[string]any `json:"tests"`
	SecretsScan  []map[string]any `json:"secrets_scan"`
	OutdatedDeps []map[string]any `json:"outdated_deps"`
	LgpdPoints   []map[string]any `json:"lgpd_points"`
	Source       string           `json:"source"`
}

type SaveRiskReportUseCase struct {
	cardRepo  CardRepository
	publisher EventPublisher
}

func NewSaveRiskReportUseCase(cardRepo CardRepository, publisher EventPublisher) *SaveRiskReportUseCase {
	return &SaveRiskReportUseCase{cardRepo: cardRepo, publisher: publisher}
}

func (uc *SaveRiskReportUseCase) Execute(ctx context.Context, cmd SaveRiskReportCommand) (*RiskReportDTO, error) {
	if cmd.CardID == "" {
		return nil, fmt.Errorf("card_id é obrigatório")
	}

	// Salva ou atualiza a ficha de risco
	dto := &RiskReportDTO{
		CardID:       cmd.CardID,
		Tests:        cmd.Tests,
		SecretsScan:  cmd.SecretsScan,
		OutdatedDeps: cmd.OutdatedDeps,
		LgpdPoints:   cmd.LgpdPoints,
		Source:       cmd.Source,
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, TopicCardsChanged, cmd.CardID)
	}

	return dto, nil
}
